package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"syscall"

	"github.com/joshuarubin/go-sway"
)

func maxScale(ctx context.Context) float64 {
	client, err := sway.New(ctx)
	if err != nil {
		log.Printf("warning: sway connect: %v; using scale 1", err)
		return 1.0
	}
	outputs, err := client.GetOutputs(ctx)
	if err != nil {
		log.Printf("warning: get outputs: %v; using scale 1", err)
		return 1.0
	}
	scale := 1.0
	for _, o := range outputs {
		if o.Scale > scale {
			scale = o.Scale
		}
	}
	return scale
}

func main() {
	ctx := context.Background()
	fontSize := int(maxScale(ctx) * 24)

	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("home dir: %v", err)
	}
	wrapper := fmt.Sprintf("%s/.local/bin/xwayland-root-wrapper", home)
	toasters := "/usr/lib/xscreensaver/flyingtoasters"

	swaylockPath, err := exec.LookPath("swaylock-plugin")
	if err != nil {
		log.Fatalf("swaylock-plugin not found: %v", err)
	}

	args := []string{
		"swaylock-plugin", "-f",
		"--font", "JetBrains Mono Nerd Font",
		"--font-size", fmt.Sprintf("%d", fontSize),
		"--indicator-idle-visible",
		"--indicator-radius", "96",
		"--indicator-thickness", "6",
		"--inside-color", "000000b8",
		"--inside-clear-color", "000000b8",
		"--inside-ver-color", "000000d0",
		"--inside-wrong-color", "000000d0",
		"--line-color", "00000000",
		"--line-clear-color", "00000000",
		"--line-ver-color", "00000000",
		"--line-wrong-color", "00000000",
		"--ring-color", "ffffffcc",
		"--ring-clear-color", "ffffffcc",
		"--ring-ver-color", "e0e0e0ff",
		"--ring-wrong-color", "ffffffff",
		"--key-hl-color", "ffffffff",
		"--bs-hl-color", "c0c0c0ff",
		"--separator-color", "00000000",
		"--text-color", "ffffffff",
		"--text-clear-color", "ffffffff",
		"--text-ver-color", "e0e0e0ff",
		"--text-wrong-color", "ffffffff",
		"--command-each", fmt.Sprintf("windowtolayer %s %s -root", wrapper, toasters),
	}

	if err := syscall.Exec(swaylockPath, args, os.Environ()); err != nil {
		log.Fatalf("exec swaylock-plugin: %v", err)
	}
}
