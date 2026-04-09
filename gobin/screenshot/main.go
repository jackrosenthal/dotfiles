package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/joshuarubin/go-sway"
)

func focusedOutput(ctx context.Context, client sway.Client) (string, error) {
	workspaces, err := client.GetWorkspaces(ctx)
	if err != nil {
		return "", fmt.Errorf("get workspaces: %w", err)
	}
	for _, ws := range workspaces {
		if ws.Focused {
			return ws.Output, nil
		}
	}
	return "", fmt.Errorf("no focused workspace found")
}

func run() error {
	ctx := context.Background()

	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Join(home, "Pictures", "Screenshots"), 0o755); err != nil {
		return err
	}

	client, err := sway.New(ctx)
	if err != nil {
		return fmt.Errorf("sway connect: %w", err)
	}

	output, err := focusedOutput(ctx, client)
	if err != nil {
		return err
	}

	tmp, err := os.CreateTemp("", "screenshot-*.png")
	if err != nil {
		return err
	}
	tmp.Close()
	defer os.Remove(tmp.Name())

	if err := exec.Command("grim", "-o", output, tmp.Name()).Run(); err != nil {
		return fmt.Errorf("grim: %w", err)
	}

	cmd := exec.Command("satty", "-f", tmp.Name())
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("satty: %w", err)
	}

	return nil
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}
