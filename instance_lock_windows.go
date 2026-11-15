//go:build windows

package main

// acquireInstanceLock reuses the named mutex that the tray app already uses.
// A second --serve, --check, or desktop launch therefore shares one instance.
func acquireInstanceLock(path string) (func(), bool, error) {
	handle, already, err := acquireSingleInstance()
	if err != nil {
		return nil, false, err
	}
	release := func() {
		if handle != 0 {
			procCloseHandle.Call(uintptr(handle))
		}
	}
	return release, already, nil
}
