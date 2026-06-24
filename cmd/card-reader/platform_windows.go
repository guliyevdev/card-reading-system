//go:build windows

package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
)

const startupRegistryName = "SmartCardReader"
const stillActiveExitCode = 259

func appRuntimeDir() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("resolve config dir: %w", err)
	}

	return filepath.Join(configDir, "SmartCardReader"), nil
}

func processRunning(pid int) (bool, error) {
	handle, err := windows.OpenProcess(windows.PROCESS_QUERY_LIMITED_INFORMATION, false, uint32(pid))
	if err != nil {
		if errors.Is(err, windows.ERROR_INVALID_PARAMETER) {
			return false, nil
		}
		return false, fmt.Errorf("open process: %w", err)
	}
	defer windows.CloseHandle(handle)

	var exitCode uint32
	if err := windows.GetExitCodeProcess(handle, &exitCode); err != nil {
		return false, fmt.Errorf("get exit code: %w", err)
	}

	return exitCode == stillActiveExitCode, nil
}

func terminateProcess(process *os.Process) error {
	return process.Kill()
}

func ensureStartupRegistration(executable string) error {
	key, _, err := registry.CreateKey(registry.CURRENT_USER, `Software\Microsoft\Windows\CurrentVersion\Run`, registry.SET_VALUE|registry.QUERY_VALUE)
	if err != nil {
		return fmt.Errorf("open startup registry key: %w", err)
	}
	defer key.Close()

	command := fmt.Sprintf(`"%s"`, executable)
	current, _, err := key.GetStringValue(startupRegistryName)
	if err == nil && current == command {
		return nil
	}

	if err := key.SetStringValue(startupRegistryName, command); err != nil {
		return fmt.Errorf("set startup registry value: %w", err)
	}

	return nil
}

func removeStartupRegistration() error {
	key, err := registry.OpenKey(registry.CURRENT_USER, `Software\Microsoft\Windows\CurrentVersion\Run`, registry.SET_VALUE)
	if err != nil {
		if errors.Is(err, registry.ErrNotExist) {
			return nil
		}
		return fmt.Errorf("open startup registry key: %w", err)
	}
	defer key.Close()

	err = key.DeleteValue(startupRegistryName)
	if err != nil && !errors.Is(err, registry.ErrNotExist) {
		return fmt.Errorf("delete startup registry value: %w", err)
	}

	return nil
}

func startupRegistrationStatus() (bool, string, error) {
	key, err := registry.OpenKey(registry.CURRENT_USER, `Software\Microsoft\Windows\CurrentVersion\Run`, registry.QUERY_VALUE)
	if err != nil {
		if errors.Is(err, registry.ErrNotExist) {
			return false, "", nil
		}
		return false, "", fmt.Errorf("open startup registry key: %w", err)
	}
	defer key.Close()

	command, _, err := key.GetStringValue(startupRegistryName)
	if err != nil {
		if errors.Is(err, registry.ErrNotExist) {
			return false, "", nil
		}
		return false, "", fmt.Errorf("read startup registry value: %w", err)
	}

	return true, command, nil
}
