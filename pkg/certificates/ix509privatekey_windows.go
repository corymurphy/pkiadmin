//go:build windows

package certificates

import (
	"syscall"
)

func (v *IX509PrivateKey) Create() (err error) {
	// hr, _, _ := syscall.Syscall(
	// 	uintptr(v.VTable().Create),
	// 	2,
	// 	uintptr(unsafe.Pointer(v)),
	// 	uintptr(unsafe.Pointer(&disabled)),
	// 	0)
	hr, _, _ := syscall.SyscallN(
		uintptr(v.VTable().Create))
	if hr != 0 {
		return convertHresultToError(hr)
	}
	return
}
