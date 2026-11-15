//go:build windows

package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"runtime"
	"sync"
	"syscall"
	"time"
	"unsafe"
)

var (
	procRegisterClassExW       = user32.NewProc("RegisterClassExW")
	procCreateWindowExW        = user32.NewProc("CreateWindowExW")
	procDefWindowProcW         = user32.NewProc("DefWindowProcW")
	procGetMessageW            = user32.NewProc("GetMessageW")
	procTranslateMessage       = user32.NewProc("TranslateMessage")
	procDispatchMessageW       = user32.NewProc("DispatchMessageW")
	procPostQuitMessage        = user32.NewProc("PostQuitMessage")
	procDestroyWindow          = user32.NewProc("DestroyWindow")
	procCreatePopupMenu        = user32.NewProc("CreatePopupMenu")
	procAppendMenuW            = user32.NewProc("AppendMenuW")
	procTrackPopupMenu         = user32.NewProc("TrackPopupMenu")
	procDestroyMenu            = user32.NewProc("DestroyMenu")
	procSetForegroundWindow    = user32.NewProc("SetForegroundWindow")
	procGetCursorPos           = user32.NewProc("GetCursorPos")
	procRegisterWindowMessageW = user32.NewProc("RegisterWindowMessageW")
	procLoadCursorW            = user32.NewProc("LoadCursorW")
	procShellNotifyIconW       = shell32.NewProc("Shell_NotifyIconW")
)

const (
	wmDestroy       = 0x0002
	wmClose         = 0x0010
	wmCommand       = 0x0111
	wmNull          = 0x0000
	wmContextMenu   = 0x007B
	wmRButtonUp     = 0x0205
	wmLButtonDblClk = 0x0203
	wmApp           = 0x8000
	wmAppTray       = wmApp + 1
	wmAppQuit       = wmApp + 2

	mfString    = 0x00000000
	mfSeparator = 0x00000800
	mfDefault   = 0x00000000

	tpmRightButton = 0x0002
	tpmBottomAlign = 0x0020

	nimAdd         = 0
	nimDelete      = 2
	nimSetVersion  = 4
	nifMessage     = 0x00000001
	nifIcon        = 0x00000002
	nifTip         = 0x00000004
	nifShowTip     = 0x00000080
	notifyVersion4 = 4

	idTrayOpen   = 2001
	idTrayCheck  = 2002
	idTrayPause  = 2003
	idTrayFolder = 2004
	idTrayQuit   = 2005

	csHRedraw = 0x0002
	csVRedraw = 0x0001
	idcArrow  = 32512
)

type winPoint struct {
	X, Y int32
}

type winMsg struct {
	Hwnd     uintptr
	Message  uint32
	WParam   uintptr
	LParam   uintptr
	Time     uint32
	Pt       winPoint
	LPrivate uint32
}

type wndClassEx struct {
	CbSize     uint32
	Style      uint32
	WndProc    uintptr
	ClsExtra   int32
	WndExtra   int32
	Instance   uintptr
	Icon       uintptr
	Cursor     uintptr
	Background uintptr
	MenuName   *uint16
	ClassName  *uint16
	IconSm     uintptr
}

type notifyIconData struct {
	CbSize           uint32
	HWnd             uintptr
	UID              uint32
	UFlags           uint32
	UCallbackMessage uint32
	HIcon            uintptr
	SzTip            [128]uint16
	DwState          uint32
	DwStateMask      uint32
	SzInfo           [256]uint16
	UVersion         uint32
	SzInfoTitle      [64]uint16
	DwInfoFlags      uint32
	GuidItem         [16]byte
	HBalloonIcon     uintptr
}

type trayState struct {
	app       *App
	hwnd      uintptr
	hInstance uintptr
	icon      uintptr
	dashboard *dashboardServer
	portFile  string
	quitOnce  sync.Once
}

var (
	activeTray            *trayState
	trayWndProcPtr        = syscall.NewCallback(trayWndProc)
	taskbarCreatedMessage uintptr
)

func init() {
	name, _ := syscall.UTF16PtrFromString("TaskbarCreated")
	message, _, _ := procRegisterWindowMessageW.Call(uintptr(unsafe.Pointer(name)))
	taskbarCreatedMessage = message
}

func runDesktop(configPath string, pathErr error, open bool) int {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	if pathErr != nil {
		showMessage(appName, "❌ "+pathErr.Error(), true)
		return 1
	}

	mutex, alreadyRunning, err := acquireSingleInstance()
	if err != nil {
		showMessage(appName, "❌ Could not start the app: "+err.Error(), true)
		return 1
	}
	if alreadyRunning {
		url := waitForDashboard(configPath)
		if url != "" {
			_ = openBrowser(url)
		} else {
			showMessage(appName, "Twitch Live to Discord is already running. Open it from the tray icon.", false)
		}
		procCloseHandle.Call(uintptr(mutex))
		return 0
	}
	defer procCloseHandle.Call(uintptr(mutex))

	app, err := NewApp(configPath)
	if err != nil {
		showMessage(appName, "❌ "+err.Error(), true)
		return 1
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	app.Start(ctx)

	quit := make(chan struct{}, 1)
	dashboard, err := startDashboard(app, func() {
		select {
		case quit <- struct{}{}:
		default:
		}
	}, "")
	if err != nil {
		showMessage(appName, "❌ "+err.Error(), true)
		return 1
	}
	defer func() {
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer shutdownCancel()
		dashboard.Shutdown(shutdownCtx)
	}()

	portFile := serverPortPath(configPath)
	if err := writePortFile(portFile, dashboard.port); err != nil {
		showMessage(appName, "❌ Could not save the dashboard port: "+err.Error(), true)
		return 1
	}
	defer func() { _ = removeFile(portFile) }()

	tray := &trayState{
		app:       app,
		dashboard: dashboard,
		portFile:  portFile,
	}
	if err := tray.createWindow(); err != nil {
		showMessage(appName, "❌ Could not create the tray icon: "+err.Error(), true)
		return 1
	}
	activeTray = tray

	go func() {
		select {
		case <-quit:
			if tray.hwnd != 0 {
				procPostMessageW.Call(tray.hwnd, wmAppQuit, 0, 0)
			}
		case <-ctx.Done():
		}
	}()

	if open {
		if err := openBrowser(dashboard.URL()); err != nil {
			showMessage(appName, "Open this address in your browser:\n\n"+dashboard.URL(), false)
		}
	}

	var msg winMsg
	for {
		ret, _, _ := procGetMessageW.Call(uintptr(unsafe.Pointer(&msg)), 0, 0, 0)
		if int32(ret) == -1 {
			break
		}
		if ret == 0 {
			break
		}
		procTranslateMessage.Call(uintptr(unsafe.Pointer(&msg)))
		procDispatchMessageW.Call(uintptr(unsafe.Pointer(&msg)))
	}

	app.Shutdown()
	activeTray = nil
	return 0
}

func acquireSingleInstance() (syscall.Handle, bool, error) {
	name, _ := syscall.UTF16PtrFromString(`Local\TwitchLiveNotifier.SingleInstance`)
	handle, _, err := procCreateMutexW.Call(0, 0, uintptr(unsafe.Pointer(name)))
	if handle == 0 {
		return 0, false, err
	}
	if errors.Is(err, syscall.ERROR_ALREADY_EXISTS) {
		return syscall.Handle(handle), true, nil
	}
	return syscall.Handle(handle), false, nil
}

func waitForDashboard(configPath string) string {
	path := serverPortPath(configPath)
	for i := 0; i < 30; i++ {
		if port, err := readPortFile(path); err == nil {
			return fmt.Sprintf("http://127.0.0.1:%d/", port)
		}
		time.Sleep(100 * time.Millisecond)
	}
	return ""
}

func (t *trayState) createWindow() error {
	instance, _, _ := procGetModuleHandleW.Call(0)
	t.hInstance = instance
	className, _ := syscall.UTF16PtrFromString("TwitchLiveNotifierTrayWindow")
	title, _ := syscall.UTF16PtrFromString("TwitchLiveNotifierTray")
	iconName, _ := syscall.UTF16PtrFromString("APP")
	icon, _, _ := procLoadIconW.Call(instance, uintptr(unsafe.Pointer(iconName)))
	if icon == 0 {
		icon, _, _ = procLoadIconW.Call(0, idiApplication)
	}
	cursor, _, _ := procLoadCursorW.Call(0, idcArrow)

	wc := wndClassEx{
		CbSize:    uint32(unsafe.Sizeof(wndClassEx{})),
		Style:     csHRedraw | csVRedraw,
		WndProc:   trayWndProcPtr,
		Instance:  instance,
		Icon:      icon,
		Cursor:    cursor,
		ClassName: className,
		IconSm:    icon,
	}
	if atom, _, _ := procRegisterClassExW.Call(uintptr(unsafe.Pointer(&wc))); atom == 0 {
		return errors.New("could not register the tray window")
	}

	hwnd, _, _ := procCreateWindowExW.Call(
		0,
		uintptr(unsafe.Pointer(className)),
		uintptr(unsafe.Pointer(title)),
		0,
		0, 0, 0, 0,
		0, 0, instance, 0,
	)
	if hwnd == 0 {
		return errors.New("could not create the tray window")
	}
	t.hwnd = hwnd
	t.icon = icon
	return t.addIcon()
}

func (t *trayState) addIcon() error {
	data := notifyIconData{
		CbSize:           uint32(unsafe.Sizeof(notifyIconData{})),
		HWnd:             t.hwnd,
		UID:              1,
		UFlags:           nifMessage | nifIcon | nifTip | nifShowTip,
		UCallbackMessage: wmAppTray,
		HIcon:            t.icon,
		UVersion:         notifyVersion4,
	}
	copyUTF16(data.SzTip[:], appName)
	if result, _, err := procShellNotifyIconW.Call(nimAdd, uintptr(unsafe.Pointer(&data))); result == 0 {
		return fmt.Errorf("could not add the notification icon: %v", err)
	}
	procShellNotifyIconW.Call(nimSetVersion, uintptr(unsafe.Pointer(&data)))
	return nil
}

func (t *trayState) removeIcon() {
	data := notifyIconData{
		CbSize: uint32(unsafe.Sizeof(notifyIconData{})),
		HWnd:   t.hwnd,
		UID:    1,
	}
	procShellNotifyIconW.Call(nimDelete, uintptr(unsafe.Pointer(&data)))
}

func (t *trayState) openDashboard() {
	if t.dashboard != nil {
		_ = openBrowser(t.dashboard.URL())
	}
}

func (t *trayState) showMenu() {
	menu, _, _ := procCreatePopupMenu.Call()
	if menu == 0 {
		return
	}
	defer procDestroyMenu.Call(menu)

	appendMenu(menu, mfString|mfDefault, idTrayOpen, "Open dashboard")
	appendMenu(menu, mfSeparator, 0, "")
	appendMenu(menu, mfString, idTrayCheck, "Check all now")
	if t.app.IsPaused() {
		appendMenu(menu, mfString, idTrayPause, "Resume monitoring")
	} else {
		appendMenu(menu, mfString, idTrayPause, "Pause monitoring")
	}
	appendMenu(menu, mfString, idTrayFolder, "Open data folder")
	appendMenu(menu, mfSeparator, 0, "")
	appendMenu(menu, mfString, idTrayQuit, "Quit")

	var point winPoint
	procGetCursorPos.Call(uintptr(unsafe.Pointer(&point)))
	procSetForegroundWindow.Call(t.hwnd)
	procTrackPopupMenu.Call(menu, tpmRightButton|tpmBottomAlign, uintptr(point.X), uintptr(point.Y), 0, t.hwnd, 0)
	procPostMessageW.Call(t.hwnd, wmNull, 0, 0)
}

func appendMenu(menu uintptr, flags uintptr, id int, text string) {
	ptr, _ := syscall.UTF16PtrFromString(text)
	procAppendMenuW.Call(menu, flags, uintptr(id), uintptr(unsafe.Pointer(ptr)))
}

func (t *trayState) quit() {
	t.quitOnce.Do(func() {
		t.app.Shutdown()
		if t.dashboard != nil {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			t.dashboard.Shutdown(ctx)
			cancel()
		}
		if t.portFile != "" {
			_ = removeFile(t.portFile)
		}
		t.removeIcon()
		procDestroyWindow.Call(t.hwnd)
	})
}

func trayWndProc(hwnd, msg, wparam, lparam uintptr) uintptr {
	t := activeTray
	switch msg {
	case wmCommand:
		switch int(wparam & 0xFFFF) {
		case idTrayOpen:
			if t != nil {
				t.openDashboard()
			}
		case idTrayCheck:
			if t != nil {
				t.app.TriggerCheck()
			}
		case idTrayPause:
			if t != nil {
				_, _ = t.app.TogglePause()
			}
		case idTrayFolder:
			if t != nil {
				_ = openFolder(t.app.DataDir())
			}
		case idTrayQuit:
			if t != nil {
				t.quit()
			}
		}
		return 0
	case wmAppTray:
		if t != nil {
			event := uint32(lparam & 0xFFFF)
			switch event {
			case wmContextMenu, wmRButtonUp:
				t.showMenu()
			case wmLButtonDblClk:
				t.openDashboard()
			}
		}
		return 0
	case taskbarCreatedMessage:
		if t != nil {
			_ = t.addIcon()
		}
		return 0
	case wmAppQuit:
		if t != nil {
			t.quit()
		}
		return 0
	case wmClose:
		if t != nil {
			t.quit()
		} else {
			procDestroyWindow.Call(hwnd)
		}
		return 0
	case wmDestroy:
		if t != nil {
			t.removeIcon()
		}
		procPostQuitMessage.Call(0)
		return 0
	}
	result, _, _ := procDefWindowProcW.Call(hwnd, msg, wparam, lparam)
	return result
}

func copyUTF16(destination []uint16, value string) {
	encoded, err := syscall.UTF16FromString(value)
	if err != nil {
		return
	}
	copy(destination, encoded)
}

func removeFile(path string) error {
	return os.Remove(path)
}
