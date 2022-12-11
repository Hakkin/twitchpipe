//go:build windows

package main

import (
	"flag"
	"syscall"
	"unsafe"

	"rsc.io/getopt"
)

var (
	user32   = syscall.NewLazyDLL("user32.dll")
	kernel32 = syscall.NewLazyDLL("kernel32.dll")
)

var (
	getWindowThreadProcessID = user32.NewProc("GetWindowThreadProcessId")
	showWindowAsync          = user32.NewProc("ShowWindowAsync")
	getConsoleWindow         = kernel32.NewProc("GetConsoleWindow")
	getCurrentProcessID      = kernel32.NewProc("GetCurrentProcessId")
)

func init() {
	flag.BoolVar(&hideConsole, "h", hideConsoleDefault, "Hide own console window")
	getopt.Aliases(
		"h", "hide-console",
	)
}

func hideWindow() {
	console, _, _ := getConsoleWindow.Call()
	if console == 0 {
		return
	}

	var consolePID uint32
	getWindowThreadProcessID.Call(
		console,
		uintptr(unsafe.Pointer(&consolePID)),
	)

	selfPIDPtr, _, _ := getCurrentProcessID.Call()
	selfPID := uint32(selfPIDPtr)

	if selfPID == consolePID {
		showWindowAsync.Call(console, 0)
	}
}
