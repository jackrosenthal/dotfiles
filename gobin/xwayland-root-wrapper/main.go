package main

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
)

const displaySearchLimit = 1024

type displayAllocation struct {
	number       int
	lockPath     string
	socketPath   string
	listener     *net.UnixListener
	listenerFile *os.File
}

func allocateDisplay() (*displayAllocation, error) {
	if err := os.MkdirAll("/tmp/.X11-unix", 0o777); err != nil {
		return nil, err
	}

	for candidate := 0; candidate < displaySearchLimit; candidate++ {
		lockPath := fmt.Sprintf("/tmp/.X%d-lock", candidate)
		socketPath := fmt.Sprintf("/tmp/.X11-unix/X%d", candidate)

		lockFile, err := os.OpenFile(lockPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
		if err != nil {
			if errors.Is(err, os.ErrExist) {
				continue
			}
			return nil, err
		}
		if _, err := fmt.Fprintf(lockFile, "%10d\n", os.Getpid()); err != nil {
			lockFile.Close()
			os.Remove(lockPath)
			return nil, err
		}
		lockFile.Close()

		addr := &net.UnixAddr{Name: socketPath, Net: "unix"}
		listener, err := net.ListenUnix("unix", addr)
		if err != nil {
			os.Remove(lockPath)
			continue
		}

		listenerFile, err := listener.File()
		if err != nil {
			listener.Close()
			os.Remove(lockPath)
			os.Remove(socketPath)
			return nil, err
		}

		return &displayAllocation{
			number:       candidate,
			lockPath:     lockPath,
			socketPath:   socketPath,
			listener:     listener,
			listenerFile: listenerFile,
		}, nil
	}

	return nil, errors.New("could not allocate an Xwayland display")
}

func (d *displayAllocation) cleanup() {
	if d.listenerFile != nil {
		d.listenerFile.Close()
	}
	if d.listener != nil {
		d.listener.Close()
	}
	os.Remove(d.lockPath)
	os.Remove(d.socketPath)
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: xwayland-root-wrapper <command> [args...]")
		os.Exit(2)
	}

	display, err := allocateDisplay()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer display.cleanup()

	readPipe, writePipe, err := os.Pipe()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer readPipe.Close()

	xCmd := exec.Command(
		"Xwayland",
		fmt.Sprintf(":%d", display.number),
		"-listenfd", "3",
		"-displayfd", "4",
		"-noreset",
		"-core",
	)

	extraFiles := []*os.File{display.listenerFile, writePipe}
	env := os.Environ()

	if waylandSocket := os.Getenv("WAYLAND_SOCKET"); waylandSocket != "" {
		fd, err := strconv.Atoi(waylandSocket)
		if err == nil {
			extraFiles = append(extraFiles, os.NewFile(uintptr(fd), "wayland-socket"))
			env = appendOrReplaceEnv(env, "WAYLAND_SOCKET", strconv.Itoa(3+len(extraFiles)-1))
		}
	}

	xCmd.ExtraFiles = extraFiles
	xCmd.Env = env
	xCmd.Stdout = os.Stdout
	xCmd.Stderr = os.Stderr

	if err := xCmd.Start(); err != nil {
		writePipe.Close()
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	writePipe.Close()

	xDone := make(chan error, 1)
	go func() {
		xDone <- xCmd.Wait()
	}()

	displayReady := make(chan error, 1)
	go func() {
		line, err := bufio.NewReader(readPipe).ReadString('\n')
		if err != nil {
			displayReady <- err
			return
		}
		if strings.TrimSpace(line) != strconv.Itoa(display.number) {
			displayReady <- fmt.Errorf("Xwayland reported unexpected display %q", strings.TrimSpace(line))
			return
		}
		displayReady <- nil
	}()

	select {
	case err := <-displayReady:
		if err != nil {
			xCmd.Process.Signal(syscall.SIGTERM)
			<-xDone
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	case err := <-xDone:
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	case <-time.After(5 * time.Second):
		xCmd.Process.Signal(syscall.SIGTERM)
		<-xDone
		fmt.Fprintln(os.Stderr, "timed out waiting for Xwayland")
		os.Exit(1)
	}

	child := exec.Command(os.Args[1], os.Args[2:]...)
	child.Env = appendOrReplaceEnv(os.Environ(), "DISPLAY", fmt.Sprintf(":%d", display.number))
	child.Env = removeEnv(child.Env, "WAYLAND_DISPLAY")
	child.Env = removeEnv(child.Env, "WAYLAND_SOCKET")
	child.Stdin = os.Stdin
	child.Stdout = os.Stdout
	child.Stderr = os.Stderr

	interrupts := make(chan os.Signal, 1)
	signal.Notify(interrupts, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(interrupts)

	if err := child.Start(); err != nil {
		xCmd.Process.Signal(syscall.SIGTERM)
		<-xDone
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	childDone := make(chan error, 1)
	go func() {
		childDone <- child.Wait()
	}()

	select {
	case sig := <-interrupts:
		if child.Process != nil {
			child.Process.Signal(sig)
		}
		<-childDone
	case err := <-childDone:
		xCmd.Process.Signal(syscall.SIGTERM)
		<-xDone
		if err == nil {
			return
		}
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			os.Exit(exitErr.ExitCode())
		}
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	xCmd.Process.Signal(syscall.SIGTERM)
	<-xDone
}

func appendOrReplaceEnv(env []string, key, value string) []string {
	prefix := key + "="
	for i, item := range env {
		if strings.HasPrefix(item, prefix) {
			env[i] = prefix + value
			return env
		}
	}
	return append(env, prefix+value)
}

func removeEnv(env []string, key string) []string {
	prefix := key + "="
	filtered := env[:0]
	for _, item := range env {
		if !strings.HasPrefix(item, prefix) {
			filtered = append(filtered, item)
		}
	}
	return filtered
}
