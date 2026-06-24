//go:build darwin || linux

package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func appRuntimeDir() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("resolve config dir: %w", err)
	}

	return filepath.Join(configDir, "smart-card-reader"), nil
}

func processRunning(pid int) (bool, error) {
	process, err := os.FindProcess(pid)
	if err != nil {
		return false, fmt.Errorf("find process: %w", err)
	}

	if err := process.Signal(os.Signal(syscallSignalZero())); err != nil {
		return false, nil
	}

	return true, nil
}

func terminateProcess(process *os.Process) error {
	return process.Signal(os.Interrupt)
}

func ensureStartupRegistration(string) error {
	return nil
}

func removeStartupRegistration() error {
	return nil
}

func startupRegistrationStatus() (bool, string, error) {
	return false, "", nil
}
