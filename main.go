package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"syscall"

	"lesiw.io/cmdio"
	"lesiw.io/cmdio/sys"
)

var (
	fg    = flag.Bool("f", false, "start daemon in foreground mode")
	kill  = flag.Bool("k", false, "kill running daemon")
	usage = `zync - zsh history syncer

  zync -f  Start server in foreground mode.
  zync -k  Kill server.

environment variables

  ZYNCREPO  URL of the git repository for zsh history.
  ZYNCPASS  Password for encrypting and decrypting zsh history.`
	errUsage = errors.New(usage)
	rnr      = sys.Runner()
)

func main() {
	cmdio.Trace = io.Discard
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	if err := validate(); err != nil {
		return err
	}
	flag.Parse()
	if *fg {
		return runServer()
	}
	if *kill {
		return signalServer(syscall.SIGTERM)
	}
	if err := spawnServer(); err != nil {
		return err
	}
	return nil
}

func validate() error {
	var errs []error
	if os.Getenv("ZYNCREPO") == "" {
		errs = append(errs, errors.New("ZYNCREPO not set"))
	}
	if os.Getenv("ZYNCPASS") == "" {
		errs = append(errs, errors.New("ZYNCPASS not set"))
	}
	if _, err := rnr.Get("git", "--version"); err != nil {
		errs = append(errs, errors.New("git not found"))
	}
	if len(errs) > 0 {
		errs = append(errs, errUsage)
	}
	return errors.Join(errs...)
}

func spawnServer() error {
	if err := signalServer(syscall.SIGUSR1); err == nil {
		return nil // Server already running, so no need to start it.
	}
	cmd := exec.Command(os.Args[0], "-f")
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}
	pidPath, err := pidFile()
	if err != nil {
		return fmt.Errorf("could not get path to server pidfile: %w", err)
	}
	err = os.WriteFile(pidPath, []byte(strconv.Itoa(cmd.Process.Pid)), 0755)
	if err != nil {
		return fmt.Errorf("failed to write pidfile: %w", err)
	}
	return nil
}
