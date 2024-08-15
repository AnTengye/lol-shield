package admin

import (
	"fmt"
	"os"
	"strings"
	"syscall"

	"golang.org/x/sys/windows"
)

func MustRunWithAdmin() {
	if !IsAdmin() {
		RunMeElevated()
	}
}

// RunMeElevated runs the current executable with elevated privileges.
//
// It uses the Windows ShellExecute function to launch the executable with the "runas" verb,
// which prompts the user for administrator privileges. If the user cancels the elevation
// prompt, the function exits with a status code of -2.
func RunMeElevated() {
	verb := "runas"
	exe, _ := os.Executable()
	cwd, _ := os.Getwd()
	args := strings.Join(os.Args[1:], " ")
	verbPtr, _ := syscall.UTF16PtrFromString(verb)
	exePtr, _ := syscall.UTF16PtrFromString(exe)
	cwdPtr, _ := syscall.UTF16PtrFromString(cwd)
	argPtr, _ := syscall.UTF16PtrFromString(args)
	var showCmd int32 = 1 // SW_NORMAL
	err := windows.ShellExecute(0, verbPtr, exePtr, argPtr, cwdPtr, showCmd)
	if err != nil {
		fmt.Println(err)
	}
	os.Exit(-2)
}

func IsAdmin() bool {
	_, err := os.Open("\\\\.\\PHYSICALDRIVE0")
	return err == nil
}

func RunAsAdmin(cmd string, args string) error {
	verb, _ := syscall.UTF16PtrFromString("runas")
	exe, _ := syscall.UTF16PtrFromString(cmd)
	arg, _ := syscall.UTF16PtrFromString(args)
	cwd, _ := syscall.Getwd()
	dir, _ := syscall.UTF16PtrFromString(cwd)

	// 调用ShellExecuteEx启动程序
	return windows.ShellExecute(0, verb, exe, arg, dir, syscall.SW_NORMAL)
}
