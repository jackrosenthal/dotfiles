package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/joshuarubin/go-sway"
)

const pipWidth = 480
const pipHeight = 270
const padding = 20

type handler struct {
	sway.EventHandler
}

func isPiP(node *sway.Node) bool {
	return node.AppID != nil && *node.AppID == "google-chrome" &&
		!strings.HasSuffix(node.Name, " - Google Chrome")
}

func (h handler) Window(ctx context.Context, e sway.WindowEvent) {
	if e.Change != sway.WindowFloating {
		return
	}
	if !isPiP(&e.Container) {
		return
	}
	if err := position(ctx, &e.Container); err != nil {
		log.Printf("position: %v", err)
	}
}

func position(ctx context.Context, pip *sway.Node) error {
	client, err := sway.New(ctx)
	if err != nil {
		return fmt.Errorf("sway connect: %w", err)
	}

	tree, err := client.GetTree(ctx)
	if err != nil {
		return fmt.Errorf("get tree: %w", err)
	}

	// Find the workspace containing the PiP to get usable area (excludes waybar).
	ws := findWorkspace(tree, pip)
	if ws == nil {
		return fmt.Errorf("no workspace found for PiP window")
	}

	x := ws.Rect.X + ws.Rect.Width - pipWidth - padding
	y := ws.Rect.Y + ws.Rect.Height - pipHeight - padding

	_, err = client.RunCommand(ctx, fmt.Sprintf(
		`[con_id=%d] floating enable, resize set %d %d, move absolute position %d %d`,
		pip.ID, pipWidth, pipHeight, x, y,
	))
	return err
}

func findWorkspace(node *sway.Node, pip *sway.Node) *sway.Node {
	if node.Type == sway.NodeWorkspace && node.Name != "__i3_scratch" {
		cx := pip.Rect.X + pip.Rect.Width/2
		cy := pip.Rect.Y + pip.Rect.Height/2
		if cx >= node.Rect.X && cx < node.Rect.X+node.Rect.Width &&
			cy >= node.Rect.Y && cy < node.Rect.Y+node.Rect.Height {
			return node
		}
	}
	for _, child := range node.Nodes {
		if result := findWorkspace(child, pip); result != nil {
			return result
		}
	}
	return nil
}

func main() {
	h := handler{EventHandler: sway.NoOpEventHandler()}
	ctx := context.Background()
	if err := sway.Subscribe(ctx, h, sway.EventTypeWindow); err != nil {
		log.Fatal(err)
	}
}
