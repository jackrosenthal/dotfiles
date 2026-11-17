package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
)

const (
	hardSep     = "\ue0b0" // nf-powerline-right_hard_divider
	iconDir     = "\uf07b" // nf-fa-folder
	iconBranch  = "\ue725" // nf-dev-git_branch
	iconModel   = "\uf2db" // nf-fa-microchip
	iconSession = "\uf02b" // nf-fa-tag
	iconCtx     = "\uf080" // nf-fa-bar_chart
	icon5h      = "\uf017" // nf-fa-clock_o
	icon7d      = "\uf073" // nf-fa-calendar
)

const (
	usageURL      = "https://api.anthropic.com/api/oauth/usage?skip_spend=1"
	usageCacheTTL = 2 * time.Minute
	usageTimeout  = 3 * time.Second
)

type statusInput struct {
	Model struct {
		DisplayName string `json:"display_name"`
	} `json:"model"`
	Workspace struct {
		CurrentDir string `json:"current_dir"`
	} `json:"workspace"`
	CWD           string `json:"cwd"`
	SessionName   string `json:"session_name"`
	ContextWindow struct {
		UsedPercentage *float64 `json:"used_percentage"`
		CurrentUsage   *struct {
			InputTokens              int `json:"input_tokens"`
			CacheCreationInputTokens int `json:"cache_creation_input_tokens"`
			CacheReadInputTokens     int `json:"cache_read_input_tokens"`
		} `json:"current_usage"`
		ContextWindowSize *int `json:"context_window_size"`
	} `json:"context_window"`
	RateLimits struct {
		FiveHour *rateLimitInfo `json:"five_hour"`
		SevenDay *rateLimitInfo `json:"seven_day"`
	} `json:"rate_limits"`
}

type rateLimitInfo struct {
	UsedPercentage float64         `json:"used_percentage"`
	ResetsAt       json.RawMessage `json:"resets_at"`
}

type oauthCredentials struct {
	ClaudeAiOauth *struct {
		AccessToken string `json:"accessToken"`
		ExpiresAt   int64  `json:"expiresAt"`
	} `json:"claudeAiOauth"`
}

// usageResponse is the subset of the plan usage endpoint response needed to
// find per-model weekly limits.
type usageResponse struct {
	Limits []struct {
		Percent  float64         `json:"percent"`
		ResetsAt json.RawMessage `json:"resets_at"`
		Scope    *struct {
			Model *struct {
				DisplayName string `json:"display_name"`
			} `json:"model"`
		} `json:"scope"`
	} `json:"limits"`
}

type usageCache struct {
	FetchedAt time.Time      `json:"fetched_at"`
	Fable     *rateLimitInfo `json:"fable,omitempty"`
}

func ansiSetBg(n int) string { return fmt.Sprintf("\033[48;5;%dm", n) }
func ansiSetFg(n int) string { return fmt.Sprintf("\033[38;5;%dm", n) }

const ansiReset = "\033[0m"

type statusLine struct {
	sb   strings.Builder
	prev int // previous bg color; 0 means no previous segment
}

// firstSeg starts the bar with no leading separator.
func (s *statusLine) firstSeg(bgColor int, text string) {
	fmt.Fprintf(&s.sb, "%s%s %s ", ansiSetBg(bgColor), ansiSetFg(231), text)
	s.prev = bgColor
}

// seg appends a powerline segment with a hard separator from the previous bg.
func (s *statusLine) seg(nextBg int, text string) {
	fmt.Fprintf(&s.sb, "%s%s%s%s %s ", ansiSetBg(nextBg), ansiSetFg(s.prev), hardSep, ansiSetFg(231), text)
	s.prev = nextBg
}

// endCap prints the closing powerline triangle back to the terminal background.
func (s *statusLine) endCap() {
	fmt.Fprintf(&s.sb, "%s%s%s%s", ansiReset, ansiSetFg(s.prev), hardSep, ansiReset)
}

// parseResetTime parses a reset time given as either a Unix timestamp or an
// RFC 3339 string.  The result is rounded to the minute because the usage
// endpoint reports fractional seconds just before the hour, which would
// otherwise show as e.g. 19:59 rather than 20:00 and fail to match the
// weekly limit's reset time.
func parseResetTime(raw json.RawMessage) (time.Time, bool) {
	if len(raw) == 0 || string(raw) == "null" {
		return time.Time{}, false
	}
	var ts float64
	if err := json.Unmarshal(raw, &ts); err == nil {
		return time.Unix(int64(ts), 0).Round(time.Minute), true
	}
	var str string
	if err := json.Unmarshal(raw, &str); err == nil {
		if t, err := time.Parse(time.RFC3339, str); err == nil {
			return t.Round(time.Minute).Local(), true
		}
	}
	return time.Time{}, false
}

func resetSuffix(raw json.RawMessage, layout string) string {
	t, ok := parseResetTime(raw)
	if !ok {
		return ""
	}
	return fmt.Sprintf(" @ %s", t.Format(layout))
}

func sameReset(a, b *rateLimitInfo) bool {
	ta, okA := parseResetTime(a.ResetsAt)
	tb, okB := parseResetTime(b.ResetsAt)
	return okA && okB && ta.Equal(tb)
}

func usedPct(rl *rateLimitInfo) int {
	return int(math.Round(rl.UsedPercentage))
}

// fableLimit returns the Fable-only weekly limit.  The statusline input does
// not carry per-model limits, so they come from the plan usage endpoint,
// served from a short-lived on-disk cache so that the endpoint is not
// queried on every refresh.  Failed fetches are cached too, which backs off
// retries for the cache TTL.
func fableLimit() *rateLimitInfo {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return nil
	}
	cachePath := filepath.Join(cacheDir, "claude-statusline", "usage.json")

	if data, err := os.ReadFile(cachePath); err == nil {
		var c usageCache
		if err := json.Unmarshal(data, &c); err == nil {
			if age := time.Since(c.FetchedAt); age >= 0 && age < usageCacheTTL {
				return c.Fable
			}
		}
	}

	c := usageCache{FetchedAt: time.Now()}
	if fable, err := fetchFableLimit(); err == nil {
		c.Fable = fable
	}
	writeUsageCache(cachePath, c)
	return c.Fable
}

// fetchFableLimit queries the plan usage endpoint using Claude Code's OAuth
// access token.  It never refreshes the token, leaving that to Claude Code.
func fetchFableLimit() (*rateLimitInfo, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(filepath.Join(home, ".claude", ".credentials.json"))
	if err != nil {
		return nil, err
	}
	var creds oauthCredentials
	if err := json.Unmarshal(data, &creds); err != nil {
		return nil, err
	}
	oauth := creds.ClaudeAiOauth
	if oauth == nil || oauth.AccessToken == "" {
		return nil, errors.New("no OAuth credentials")
	}
	if !time.Now().Before(time.UnixMilli(oauth.ExpiresAt)) {
		return nil, errors.New("OAuth access token expired")
	}

	ctx, cancel := context.WithTimeout(context.Background(), usageTimeout)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, usageURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", oauth.AccessToken))
	req.Header.Set("anthropic-beta", "oauth-2025-04-20")
	req.Header.Set("User-Agent", "claude-statusline")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("usage endpoint returned %s", resp.Status)
	}

	var usage usageResponse
	if err := json.NewDecoder(resp.Body).Decode(&usage); err != nil {
		return nil, err
	}
	for _, l := range usage.Limits {
		if l.Scope != nil && l.Scope.Model != nil && strings.EqualFold(l.Scope.Model.DisplayName, "Fable") {
			return &rateLimitInfo{UsedPercentage: l.Percent, ResetsAt: l.ResetsAt}, nil
		}
	}
	return nil, nil
}

// writeUsageCache atomically replaces the cache file so that concurrent
// statusline invocations never read a partial write.
func writeUsageCache(path string, c usageCache) {
	data, err := json.Marshal(c)
	if err != nil {
		return
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return
	}
	f, err := os.CreateTemp(dir, "usage-*.json")
	if err != nil {
		return
	}
	_, writeErr := f.Write(data)
	closeErr := f.Close()
	if writeErr != nil || closeErr != nil {
		_ = os.Remove(f.Name())
		return
	}
	if err := os.Rename(f.Name(), path); err != nil {
		_ = os.Remove(f.Name())
	}
}

func gitBranch(dir string) string {
	if dir == "" {
		return ""
	}
	repo, err := git.PlainOpenWithOptions(dir, &git.PlainOpenOptions{DetectDotGit: true})
	if err != nil {
		return ""
	}
	head, err := repo.Head()
	if err != nil {
		return ""
	}
	if head.Name().IsBranch() {
		return head.Name().Short()
	}
	return head.Hash().String()[:7]
}

func main() {
	var in statusInput
	if err := json.NewDecoder(os.Stdin).Decode(&in); err != nil {
		fmt.Fprintln(os.Stderr, "claude-statusline:", err)
		os.Exit(1)
	}

	cwd := in.Workspace.CurrentDir
	if cwd == "" {
		cwd = in.CWD
	}

	model := in.Model.DisplayName
	if model == "" {
		model = "unknown"
	}

	dirDisplay := cwd
	if home, err := os.UserHomeDir(); err == nil && strings.HasPrefix(cwd, home) {
		dirDisplay = fmt.Sprintf("~%s", cwd[len(home):])
	}

	sl := &statusLine{}

	sl.firstSeg(24, fmt.Sprintf("%s %s", iconDir, dirDisplay))

	if branch := gitBranch(cwd); branch != "" {
		sl.seg(55, fmt.Sprintf("%s %s", iconBranch, branch))
	}

	sl.seg(23, fmt.Sprintf("%s %s", iconModel, model))

	if in.SessionName != "" {
		sl.seg(58, fmt.Sprintf("%s %s", iconSession, in.SessionName))
	}

	if in.ContextWindow.UsedPercentage != nil {
		pct := int(math.Round(*in.ContextWindow.UsedPercentage))
		ctxLabel := fmt.Sprintf("%d%%", pct)
		if in.ContextWindow.CurrentUsage != nil && in.ContextWindow.ContextWindowSize != nil {
			u := in.ContextWindow.CurrentUsage
			total := u.InputTokens + u.CacheCreationInputTokens + u.CacheReadInputTokens
			ctxLabel = fmt.Sprintf("%dk/%dk (%d%%)", total/1000, *in.ContextWindow.ContextWindowSize/1000, pct)
		}
		sl.seg(28, fmt.Sprintf("%s %s", iconCtx, ctxLabel))
	}

	// Plan limits are only present for claude.ai subscriber sessions, so
	// skip the usage endpoint for API key sessions.
	var fable *rateLimitInfo
	if in.RateLimits.FiveHour != nil || in.RateLimits.SevenDay != nil {
		fable = fableLimit()
	}

	if rl := in.RateLimits.FiveHour; rl != nil {
		sl.seg(88, fmt.Sprintf("%s %d%%%s", icon5h, usedPct(rl), resetSuffix(rl.ResetsAt, "15:04")))
	}

	if rl := in.RateLimits.SevenDay; rl != nil {
		label := fmt.Sprintf("%d%%", usedPct(rl))
		// The Fable limit normally resets with the weekly limit, so share
		// one segment and reset time when they line up.
		if fable != nil && sameReset(rl, fable) {
			label = fmt.Sprintf("%s (Fable %d%%)", label, usedPct(fable))
			fable = nil
		}
		sl.seg(54, fmt.Sprintf("%s %s%s", icon7d, label, resetSuffix(rl.ResetsAt, "Mon 15:04")))
	}

	if fable != nil {
		sl.seg(94, fmt.Sprintf("%s Fable %d%%%s", icon7d, usedPct(fable), resetSuffix(fable.ResetsAt, "Mon 15:04")))
	}

	sl.endCap()
	fmt.Print(sl.sb.String())
}
