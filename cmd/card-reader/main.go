package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"card-reading-system/internal/httpapi"
	"card-reading-system/internal/smartcard"
	"card-reading-system/internal/state"
)

func main() {
	command := "start"
	if len(os.Args) > 1 {
		command = os.Args[1]
	}

	switch command {
	case "serve":
		if err := runServer(os.Stdout); err != nil {
			log.Fatalf("server error: %v", err)
		}
	case "start":
		if err := startBackground(); err != nil {
			log.Fatalf("background start error: %v", err)
		}
	case "stop":
		if err := stopBackground(); err != nil {
			log.Fatalf("stop error: %v", err)
		}
	case "status":
		if err := printStatus(); err != nil {
			log.Fatalf("status error: %v", err)
		}
	default:
		log.Fatalf("unknown command: %s", command)
	}
}

func runServer(output *os.File) error {
	logger := log.New(output, "", log.LstdFlags)
	cardState := state.NewStore()
	service := smartcard.NewService(logger, cardState)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go service.Start(ctx)

	server := &http.Server{
		Addr:    ":4121",
		Handler: httpapi.NewHandler(cardState),
	}

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), smartcard.DefaultPollInterval)
		defer cancel()
		_ = server.Shutdown(shutdownCtx)
	}()

	logger.Println("Local smartcard server running on http://localhost:4121")

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("http server: %w", err)
	}

	return nil
}

func startBackground() error {
	pidPath, logPath, err := runtimePaths()
	if err != nil {
		return err
	}

	running, pid, err := existingPID(pidPath)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	if running {
		return fmt.Errorf("already running with PID %d", pid)
	}

	if err := os.MkdirAll(filepath.Dir(pidPath), 0o755); err != nil {
		return fmt.Errorf("create runtime dir: %w", err)
	}

	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("open log file: %w", err)
	}
	defer logFile.Close()

	devNull, err := os.Open(os.DevNull)
	if err != nil {
		return fmt.Errorf("open devnull: %w", err)
	}
	defer devNull.Close()

	executable, err := os.Executable()
	if err != nil {
		return fmt.Errorf("resolve executable: %w", err)
	}

	cmd := exec.Command(executable, "serve")
	cmd.Dir, _ = os.Getwd()
	cmd.Stdin = devNull
	cmd.Stdout = logFile
	cmd.Stderr = logFile
	cmd.SysProcAttr = backgroundSysProcAttr()

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start detached process: %w", err)
	}

	if err := os.WriteFile(pidPath, []byte(strconv.Itoa(cmd.Process.Pid)), 0o644); err != nil {
		_ = cmd.Process.Kill()
		return fmt.Errorf("write pid file: %w", err)
	}

	time.Sleep(150 * time.Millisecond)
	fmt.Printf("Started in background. PID=%d\n", cmd.Process.Pid)
	fmt.Printf("Log file: %s\n", logPath)
	return nil
}

func stopBackground() error {
	pidPath, _, err := runtimePaths()
	if err != nil {
		return err
	}

	running, pid, err := existingPID(pidPath)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	if !running {
		_ = os.Remove(pidPath)
		return fmt.Errorf("not running")
	}

	process, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("find process: %w", err)
	}

	if err := process.Signal(syscall.SIGTERM); err != nil {
		return fmt.Errorf("signal process: %w", err)
	}

	_ = os.Remove(pidPath)
	fmt.Printf("Stopped process PID=%d\n", pid)
	return nil
}

func printStatus() error {
	pidPath, logPath, err := runtimePaths()
	if err != nil {
		return err
	}

	running, pid, err := existingPID(pidPath)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	if !running {
		fmt.Println("Status: stopped")
		return nil
	}

	fmt.Printf("Status: running (PID=%d)\n", pid)
	fmt.Printf("Log file: %s\n", logPath)
	return nil
}

func runtimePaths() (pidPath string, logPath string, err error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", "", fmt.Errorf("get cwd: %w", err)
	}

	runtimeDir := filepath.Join(wd, "runtime")
	return filepath.Join(runtimeDir, "smart-card-reader.pid"), filepath.Join(runtimeDir, "smart-card-reader.log"), nil
}

func existingPID(pidPath string) (bool, int, error) {
	data, err := os.ReadFile(pidPath)
	if err != nil {
		return false, 0, err
	}

	pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		return false, 0, fmt.Errorf("parse pid file: %w", err)
	}

	process, err := os.FindProcess(pid)
	if err != nil {
		return false, pid, fmt.Errorf("find process: %w", err)
	}

	if err := process.Signal(syscall.Signal(0)); err != nil {
		return false, pid, nil
	}

	return true, pid, nil
}
