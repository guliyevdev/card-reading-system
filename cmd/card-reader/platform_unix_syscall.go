//go:build darwin || linux

package main

import "syscall"

func syscallSignalZero() syscall.Signal {
	return syscall.Signal(0)
}
