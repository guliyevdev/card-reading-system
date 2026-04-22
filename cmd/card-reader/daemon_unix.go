//go:build darwin || linux

package main

import "syscall"

func backgroundSysProcAttr() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{
		Setsid: true,
	}
}
