package main

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
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
	iconCaveman = "\uf490" // nf-mdi-fire
)

type statusInput struct {
	Model struct {
		DisplayName string `json:"display_name"`
	} `json:"model"`
	Workspace struct {
		CurrentDir string `json:"current_dir"`
	} `json:"workspace"`
	CWD         string `json:"cwd"`
	SessionName string `json:"session_name"`
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

func parseResetTime(raw json.RawMessage, layout string) string {
	if len(raw) == 0 {
		return ""
	}
	// Unix timestamp (number)
	var ts float64
	if err := json.Unmarshal(raw, &ts); err == nil {
		return fmt.Sprintf(" @ %s", time.Unix(int64(ts), 0).Format(layout))
	}
	// ISO 8601 / RFC3339 string
	var str string
	if err := json.Unmarshal(raw, &str); err == nil {
		if t, err := time.Parse(time.RFC3339, str); err == nil {
			return fmt.Sprintf(" @ %s", t.Format(layout))
		}
	}
	return ""
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

	if flagFile := os.Getenv("HOME") + "/.claude/.caveman-active"; func() bool {
		_, err := os.Stat(flagFile)
		return err == nil
	}() {
		mode, _ := os.ReadFile(flagFile)
		modeStr := strings.TrimSpace(string(mode))
		label := fmt.Sprintf("%s %s", iconCaveman, modeStr)
		if modeStr == "" {
			label = iconCaveman
		}
		sl.seg(172, label)
	}

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

	if rl := in.RateLimits.FiveHour; rl != nil {
		reset := parseResetTime(rl.ResetsAt, "15:04")
		sl.seg(88, fmt.Sprintf("%s %d%%%s", icon5h, int(math.Round(rl.UsedPercentage)), reset))
	}

	if rl := in.RateLimits.SevenDay; rl != nil {
		reset := parseResetTime(rl.ResetsAt, "Mon 15:04")
		sl.seg(54, fmt.Sprintf("%s %d%%%s", icon7d, int(math.Round(rl.UsedPercentage)), reset))
	}

	sl.endCap()
	fmt.Print(sl.sb.String())
}
