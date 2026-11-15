//go:build windows

package main

import (
	"errors"
	"syscall"
	"unsafe"
)

var (
	crypt32DLL             = syscall.NewLazyDLL("crypt32.dll")
	kernel32DLL            = syscall.NewLazyDLL("kernel32.dll")
	procCryptProtectData   = crypt32DLL.NewProc("CryptProtectData")
	procCryptUnprotectData = crypt32DLL.NewProc("CryptUnprotectData")
	procLocalFree          = kernel32DLL.NewProc("LocalFree")
)

type dataBlob struct {
	size uint32
	data *byte
}

func newDataBlob(value []byte) dataBlob {
	if len(value) == 0 {
		return dataBlob{}
	}
	return dataBlob{size: uint32(len(value)), data: &value[0]}
}

func (b dataBlob) bytes() []byte {
	if b.data == nil || b.size == 0 {
		return nil
	}
	return unsafe.Slice(b.data, b.size)
}

func dpapiProtect(data []byte) ([]byte, error) {
	input := newDataBlob(data)
	var output dataBlob
	ok, _, callErr := procCryptProtectData.Call(
		uintptr(unsafe.Pointer(&input)),
		0,
		0,
		0,
		0,
		0,
		uintptr(unsafe.Pointer(&output)),
	)
	if ok == 0 {
		if callErr != nil && callErr != syscall.Errno(0) {
			return nil, errors.New("DPAPI could not encrypt the secret: " + callErr.Error())
		}
		return nil, errors.New("DPAPI could not encrypt the secret")
	}
	defer procLocalFree.Call(uintptr(unsafe.Pointer(output.data)))
	result := make([]byte, output.size)
	copy(result, output.bytes())
	return result, nil
}

func dpapiUnprotect(data []byte) ([]byte, error) {
	input := newDataBlob(data)
	var output dataBlob
	ok, _, callErr := procCryptUnprotectData.Call(
		uintptr(unsafe.Pointer(&input)),
		0,
		0,
		0,
		0,
		0,
		uintptr(unsafe.Pointer(&output)),
	)
	if ok == 0 {
		if callErr != nil && callErr != syscall.Errno(0) {
			return nil, errors.New("DPAPI could not decrypt the secret: " + callErr.Error())
		}
		return nil, errors.New("DPAPI could not decrypt the secret")
	}
	defer procLocalFree.Call(uintptr(unsafe.Pointer(output.data)))
	result := make([]byte, output.size)
	copy(result, output.bytes())
	return result, nil
}

type dpapiProtector struct{}

func (dpapiProtector) Protect(plaintext []byte) ([]byte, error) {
	return dpapiProtect(plaintext)
}

func (dpapiProtector) Unprotect(ciphertext []byte) ([]byte, error) {
	return dpapiUnprotect(ciphertext)
}

func (dpapiProtector) Description() string {
	return "dpapi"
}

func newPlatformSecretProtector(dir string) (secretProtector, error) {
	return dpapiProtector{}, nil
}
