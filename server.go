package main

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"syscall"
)

func runServer() error {
	signal.Ignore(syscall.SIGHUP)
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM, syscall.SIGUSR1)
	if err := sync(); err != nil {
		// Always sync on startup.
		slog.Error(err.Error())
	}
	for s := range sig {
		switch s {
		default:
			return nil
		case syscall.SIGUSR1:
			if err := sync(); err != nil {
				slog.Error(err.Error())
			}
		}
	}
	pidPath, err := pidFile()
	if err != nil {
		slog.Info("could not find pid file", "error", err)
		return nil
	}
	if err := os.RemoveAll(pidPath); err != nil {
		slog.Info("failed to clean up pid file", "error", err)
	}
	return nil
}

func signalServer(sig os.Signal) error {
	pidPath, err := pidFile()
	if err != nil {
		return fmt.Errorf("could not get path to server pidfile: %w", err)
	}
	buf, err := os.ReadFile(pidPath)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("could not read pidfile: %w", err)
	}
	var p *os.Process
	pid, err := strconv.Atoi(string(buf))
	if err != nil {
		return fmt.Errorf("invalid pid %q: %w", string(buf), err)
	}
	if p, err = os.FindProcess(pid); err != nil {
		return fmt.Errorf("could not find process: %w", err)
	}
	return p.Signal(sig)
}

func pidFile() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("could not get user home dir: %w", err)
	}
	return filepath.Join(home, ".zync.pid"), nil
}
