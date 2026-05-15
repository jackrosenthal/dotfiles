// ssha connects to a host and attaches to (or creates) a tmux session
// there. Before connecting, it cleans up stale ControlMaster sockets that
// would otherwise hang the next ssh invocation.
package main

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"syscall"

	"github.com/alecthomas/kong"
)

type CLI struct {
	Host    string `arg:"" help:"Hostname (matches an entry in ~/.ssh/config)."`
	Session string `arg:"" optional:"" default:"main" help:"tmux session name on the remote."`
}

// safeSession bounds what can be interpolated into the remote tmux command.
var safeSession = regexp.MustCompile(`^[A-Za-z0-9._-]+$`)

// resolvedControlPath returns the ControlPath ssh would use for host, or "" if
// none is configured.
func resolvedControlPath(host string) (string, error) {
	out, err := exec.Command("ssh", "-G", host).Output()
	if err != nil {
		return "", fmt.Errorf("ssh -G %s: %w", host, err)
	}
	for line := range strings.SplitSeq(string(out), "\n") {
		fields := strings.Fields(line)
		if len(fields) >= 2 && strings.EqualFold(fields[0], "controlpath") {
			path := fields[1]
			if path == "none" {
				return "", nil
			}
			return path, nil
		}
	}
	return "", nil
}

// reapStaleMaster removes a leftover ControlMaster socket whose master process
// is no longer responsive. The next ssh would otherwise block trying to
// multiplex over a dead socket.
func reapStaleMaster(host string) error {
	path, err := resolvedControlPath(host)
	if err != nil {
		return err
	}
	if path == "" {
		return nil
	}
	if _, err := os.Stat(path); errors.Is(err, fs.ErrNotExist) {
		return nil
	} else if err != nil {
		return fmt.Errorf("stat %s: %w", path, err)
	}
	if exec.Command("ssh", "-O", "check", host).Run() == nil {
		return nil
	}
	_ = exec.Command("ssh", "-O", "exit", host).Run()
	if err := os.Remove(path); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("remove stale socket %s: %w", path, err)
	}
	fmt.Fprintf(os.Stderr, "ssha: removed stale control socket %s\n", path)
	return nil
}

func run(cli *CLI) error {
	if !safeSession.MatchString(cli.Session) {
		return fmt.Errorf("session name %q must match %s", cli.Session, safeSession)
	}
	if err := reapStaleMaster(cli.Host); err != nil {
		return err
	}
	sshPath, err := exec.LookPath("ssh")
	if err != nil {
		return fmt.Errorf("ssh not found in PATH: %w", err)
	}
	remote := fmt.Sprintf("tmux new-session -A -s %s", cli.Session)
	args := []string{"ssh", "-t", cli.Host, remote}
	return syscall.Exec(sshPath, args, os.Environ())
}

func main() {
	var cli CLI
	kong.Parse(&cli,
		kong.Name("ssha"),
		kong.Description("SSH to a host and attach to a tmux session, reaping stale ControlMaster sockets first."),
	)
	if err := run(&cli); err != nil {
		fmt.Fprintln(os.Stderr, "ssha:", err)
		os.Exit(1)
	}
}
