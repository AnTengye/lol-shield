package admin

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"unsafe"

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
	y, err := isAdminWithProcessToken()
	if err != nil {
		fmt.Println(err)
		return false
	}
	return y
}

func isAdminOpenPHY() bool {
	_, err := os.Open("\\\\.\\PHYSICALDRIVE0")
	return err == nil
}

func isAdminNetSession() bool {
	cmd := exec.Command("net", "session")
	err := cmd.Run()
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

const (
	SECURITY_MANDATORY_HIGH_RID = 0x00003000
)

func isAdminWithProcessToken() (bool, error) {
	var token windows.Token
	// 获取当前进程的访问令牌
	err := windows.OpenProcessToken(windows.CurrentProcess(), windows.TOKEN_QUERY, &token)
	if err != nil {
		return false, err
	}
	defer token.Close()

	// 获取令牌中的完整性级别
	var tokenInfoLength uint32
	windows.GetTokenInformation(token, windows.TokenIntegrityLevel, nil, 0, &tokenInfoLength)
	tokenInfo := make([]byte, tokenInfoLength)
	err = windows.GetTokenInformation(token, windows.TokenIntegrityLevel, &tokenInfo[0], tokenInfoLength, &tokenInfoLength)
	if err != nil {
		return false, err
	}

	// 解析完整性级别
	tokenIL := (*windows.Tokenmandatorylabel)(unsafe.Pointer(&tokenInfo[0]))
	subAuthority := *(*uint32)(unsafe.Pointer(uintptr(unsafe.Pointer(tokenIL.Label.Sid)) + uintptr(8)))

	// 判断是否为高完整性级别
	isElevated := subAuthority >= SECURITY_MANDATORY_HIGH_RID
	return isElevated, nil
}
