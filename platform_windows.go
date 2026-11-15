//go:build windows

package main

import (
	"errors"
	"fmt"
	"os"
	"syscall"
	"unsafe"
)

var (
	shell32  = syscall.NewLazyDLL("shell32.dll")
	advapi32 = syscall.NewLazyDLL("advapi32.dll")
	user32   = syscall.NewLazyDLL("user32.dll")
	kernel32 = syscall.NewLazyDLL("kernel32.dll")
	gdi32    = syscall.NewLazyDLL("gdi32.dll")

	procShellExecuteW      = shell32.NewProc("ShellExecuteW")
	procRegOpenKeyExW      = advapi32.NewProc("RegOpenKeyExW")
	procRegSetValueExW     = advapi32.NewProc("RegSetValueExW")
	procRegDeleteValueW    = advapi32.NewProc("RegDeleteValueW")
	procRegCloseKey        = advapi32.NewProc("RegCloseKey")
	procMessageBoxW        = user32.NewProc("MessageBoxW")
	procPostMessageW       = user32.NewProc("PostMessageW")
	procSetProcessDPIAware = user32.NewProc("SetProcessDPIAware")
	procGetConsoleWindow   = kernel32.NewProc("GetConsoleWindow")
	procAttachConsole      = kernel32.NewProc("AttachConsole")
	procCreateMutexW       = kernel32.NewProc("CreateMutexW")
	procCloseHandle        = kernel32.NewProc("CloseHandle")
	procGetModuleHandleW   = kernel32.NewProc("GetModuleHandleW")
	procLoadIconW          = user32.NewProc("LoadIconW")
)

const (
	hkeyCurrentUser = 0x80000001
	keySetValue     = 0x0002
	keyQueryValue   = 0x0001
	regSZ           = 1

	mbOK              = 0x00000000
	mbIconError       = 0x00000010
	mbIconInformation = 0x00000040
	mbSetForeground   = 0x00010000
	mbTopMost         = 0x00040000

	idiApplication = 32512
)

func platformName() string { return "windows" }

func startAtLoginSupported() bool { return true }

func openBrowser(rawURL string) error {
	verb, _ := syscall.UTF16PtrFromString("open")
	target, _ := syscall.UTF16PtrFromString(rawURL)
	result, _, _ := procShellExecuteW.Call(
		0,
		uintptr(unsafe.Pointer(verb)),
		uintptr(unsafe.Pointer(target)),
		0, 0, 1,
	)
	if result <= 32 {
		return errors.New("could not open the default browser")
	}
	return nil
}

func openFolder(path string) error {
	verb, _ := syscall.UTF16PtrFromString("explore")
	target, _ := syscall.UTF16PtrFromString(path)
	result, _, _ := procShellExecuteW.Call(
		0,
		uintptr(unsafe.Pointer(verb)),
		uintptr(unsafe.Pointer(target)),
		0, 0, 1,
	)
	if result <= 32 {
		return errors.New("could not open the data folder")
	}
	return nil
}

func setAutoStart(enabled bool) error {
	subkey, _ := syscall.UTF16PtrFromString(`Software\Microsoft\Windows\CurrentVersion\Run`)
	valueName, _ := syscall.UTF16PtrFromString("TwitchLiveNotifier")

	var key syscall.Handle
	opened, _, openErr := procRegOpenKeyExW.Call(
		hkeyCurrentUser,
		uintptr(unsafe.Pointer(subkey)),
		0,
		keySetValue|keyQueryValue,
		uintptr(unsafe.Pointer(&key)),
	)
	if opened != 0 {
		return fmt.Errorf("could not open the Windows startup registry key: %v", openErr)
	}
	defer procRegCloseKey.Call(uintptr(key))

	if !enabled {
		deleted, _, deleteErr := procRegDeleteValueW.Call(uintptr(key), uintptr(unsafe.Pointer(valueName)))
		if deleted != 0 && !errors.Is(deleteErr, syscall.ERROR_FILE_NOT_FOUND) {
			return fmt.Errorf("could not remove the Windows startup entry: %v", deleteErr)
		}
		return nil
	}

	executable, err := os.Executable()
	if err != nil {
		return errors.New("could not find the application path")
	}
	command := `"` + executable + `" --background`
	data, err := syscall.UTF16FromString(command)
	if err != nil {
		return errors.New("could not build the Windows startup entry")
	}
	written, _, writeErr := procRegSetValueExW.Call(
		uintptr(key),
		uintptr(unsafe.Pointer(valueName)),
		0,
		regSZ,
		uintptr(unsafe.Pointer(&data[0])),
		uintptr(len(data)*2),
	)
	if written != 0 {
		return fmt.Errorf("could not save the Windows startup entry: %v", writeErr)
	}
	return nil
}

func showMessage(title, text string, isError bool) {
	procSetProcessDPIAware.Call()

	titlePtr, _ := syscall.UTF16PtrFromString(title)
	textPtr, _ := syscall.UTF16PtrFromString(text)
	flags := uintptr(mbOK | mbSetForeground | mbTopMost)
	if isError {
		flags |= mbIconError
	} else {
		flags |= mbIconInformation
	}
	procMessageBoxW.Call(0, uintptr(unsafe.Pointer(textPtr)), uintptr(unsafe.Pointer(titlePtr)), flags)
}

func prepareConsole() {
	if window, _, _ := procGetConsoleWindow.Call(); window != 0 {
		return
	}
	const attachParentProcess = ^uintptr(0)
	procAttachConsole.Call(attachParentProcess)

	console, err := os.OpenFile("CONOUT$", os.O_WRONLY, 0)
	if err != nil {
		return
	}
	os.Stdout = console
	os.Stderr = console
}
