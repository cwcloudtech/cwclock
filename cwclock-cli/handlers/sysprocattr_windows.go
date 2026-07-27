//go:build windows

package handlers

import "syscall"

func helmInstallSysProcAttr() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{}
}
