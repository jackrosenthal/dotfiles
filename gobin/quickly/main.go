// Command quickly hands a one-off task to a coding agent running
// non-interactively with automatic permission approval.
package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"syscall"

	"github.com/alecthomas/kong"
	"github.com/jackrosenthal/fay"
)

const (
	defaultClaudeModel = "sonnet"
	defaultEffort      = "medium"

	// GLM Flash has no medium variant, so OpenCode defaults to high.
	defaultOpencodeEffort = "high"
)

var (
	codexTerraRe    = regexp.MustCompile(`^gpt-(\d+(?:\.\d+)*)-terra$`)
	opencodeFlashRe = regexp.MustCompile(`^\S+/glm-(\d+(?:\.\d+)*)-flash$`)
)

type cli struct {
	Model  string   `short:"m" placeholder:"MODEL" help:"Override the model (default: sonnet for Claude Code, the latest Terra for Codex, the latest GLM Flash for OpenCode)."`
	Effort string   `short:"e" placeholder:"LEVEL" help:"Override the reasoning effort (default: medium for Claude Code and Codex, high for OpenCode)."`
	Codex  bool     `xor:"agent" help:"Run the task in Codex instead of Claude Code."`
	GLM    bool     `name:"glm" xor:"agent" help:"Run the task with GLM Flash in OpenCode instead of Claude Code."`
	Task   []string `arg:"" passthrough:"" help:"The task to perform, in plain words."`
}

// parseVersion converts a dotted version such as "6.1" into its numeric
// components.
func parseVersion(s string) ([]int, error) {
	parts := strings.Split(s, ".")
	version := make([]int, 0, len(parts))
	for _, part := range parts {
		n, err := strconv.Atoi(part)
		if err != nil {
			return nil, err
		}
		version = append(version, n)
	}
	return version, nil
}

// latest returns the candidate with the highest version, where the version
// is the first capture group of re.  Candidates that do not match re are
// ignored, and the earliest candidate wins a tie.
func latest(candidates []string, re *regexp.Regexp) (string, error) {
	var best string
	var bestVersion []int
	for _, candidate := range candidates {
		match := re.FindStringSubmatch(candidate)
		if match == nil {
			continue
		}
		version, err := parseVersion(match[1])
		if err != nil {
			continue
		}
		if best == "" || slices.Compare(version, bestVersion) > 0 {
			best, bestVersion = candidate, version
		}
	}
	if best == "" {
		return "", fmt.Errorf("no model matching %s is available", re)
	}
	return best, nil
}

// latestCodexTerra asks Codex for its model catalog and returns the newest
// Terra model.
func latestCodexTerra() (string, error) {
	out, err := exec.Command("codex", "debug", "models").Output()
	if err != nil {
		return "", fmt.Errorf("listing codex models: %w", err)
	}
	var catalog struct {
		Models []struct {
			Slug string `json:"slug"`
		} `json:"models"`
	}
	if err := json.Unmarshal(out, &catalog); err != nil {
		return "", fmt.Errorf("parsing codex model catalog: %w", err)
	}
	slugs := make([]string, 0, len(catalog.Models))
	for _, m := range catalog.Models {
		slugs = append(slugs, m.Slug)
	}
	return latest(slugs, codexTerraRe)
}

// latestOpencodeFlash asks OpenCode for its model list and returns the
// newest GLM Flash model.
func latestOpencodeFlash() (string, error) {
	out, err := exec.Command("opencode", "models").Output()
	if err != nil {
		return "", fmt.Errorf("listing opencode models: %w", err)
	}
	var models []string
	scanner := bufio.NewScanner(bytes.NewReader(out))
	for scanner.Scan() {
		models = append(models, strings.TrimSpace(scanner.Text()))
	}
	return latest(models, opencodeFlashRe)
}

// command returns the argv of the agent invocation for the parsed CLI.
func (c *cli) command() ([]string, error) {
	prompt := fmt.Sprintf("quickly, %s", strings.Join(c.Task, " "))

	model := func(fallback func() (string, error)) (string, error) {
		if c.Model != "" {
			return c.Model, nil
		}
		return fallback()
	}

	effort := c.Effort
	if effort == "" {
		effort = defaultEffort
	}

	switch {
	case c.Codex:
		m, err := model(latestCodexTerra)
		if err != nil {
			return nil, err
		}
		return []string{
			"codex", "exec",
			"--model", m,
			"--config", fmt.Sprintf("model_reasoning_effort=%q", effort),
			"--approve-for-me",
			prompt,
		}, nil
	case c.GLM:
		m, err := model(latestOpencodeFlash)
		if err != nil {
			return nil, err
		}
		// OpenCode expresses effort as a model variant.  A variant given
		// as part of --model takes the place of the default effort.
		switch {
		case !strings.Contains(m, "#"):
			variant := c.Effort
			if variant == "" {
				variant = defaultOpencodeEffort
			}
			m = fmt.Sprintf("%s#%s", m, variant)
		case c.Effort != "":
			return nil, fmt.Errorf("--effort conflicts with the variant already in model %q", m)
		}
		return []string{"opencode", "run", "--model", m, "--auto", prompt}, nil
	default:
		m, _ := model(func() (string, error) { return defaultClaudeModel, nil })
		return []string{
			"claude",
			"--model", m,
			"--effort", effort,
			"--permission-mode", "auto",
			"-p", prompt,
		}, nil
	}
}

func main() {
	var c cli
	ctx := kong.Parse(
		&c,
		kong.Name("quickly"),
		kong.Description("Quickly hand a one-off task to a coding agent."),
		fay.Register(),
	)

	argv, err := c.command()
	ctx.FatalIfErrorf(err)
	path, err := exec.LookPath(argv[0])
	ctx.FatalIfErrorf(err)
	ctx.FatalIfErrorf(syscall.Exec(path, argv, os.Environ()))
}
