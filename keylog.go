package main

import (
	"fmt"
	"log"
	"syscall"
	"time"
	"unsafe"
)

var (
	user32                  = syscall.NewLazyDLL("user32.dll")
	procEnumWindows         = user32.NewProc("EnumWindows")
	procGetWindowTextW      = user32.NewProc("GetWindowTextW")
	procGetForegroundWindow = user32.NewProc("GetForegroundWindow")
	procGetAsyncKeyState    = user32.NewProc("GetAsyncKeyState")
)

func EnumWindows(enumFunc uintptr, lparam uintptr) (err error) {
	r1, _, e1 := syscall.Syscall(procEnumWindows.Addr(), 2, uintptr(enumFunc), uintptr(lparam), 0)
	if r1 == 0 {
		if e1 != 0 {
			err = error(e1)
		} else {
			err = syscall.EINVAL
		}
	}
	return
}

func GetWindowText(hwnd syscall.Handle, str *uint16, maxCount int32) (len int32, err error) {
	r0, _, e1 := syscall.Syscall(procGetWindowTextW.Addr(), 3, uintptr(hwnd), uintptr(unsafe.Pointer(str)), uintptr(maxCount))
	len = int32(r0)
	if len == 0 {
		if e1 != 0 {
			err = error(e1)
		} else {
			err = syscall.EINVAL
		}
	}
	return
}

func GetForegroundWindow() syscall.Handle {
	ret, _, _ := syscall.Syscall(procGetForegroundWindow.Addr(), 0,
		0, 0, 0)
	return syscall.Handle(ret)
}

func GetAsyncKeyState(vKey int) uint16 {
	ret, _, _ := procGetAsyncKeyState.Call(uintptr(vKey))
	return uint16(ret)
}

func FindWindow(title string) (syscall.Handle, error) {
	var hwnd syscall.Handle
	cb := syscall.NewCallback(func(h syscall.Handle, p uintptr) uintptr {
		b := make([]uint16, 200)
		_, err := GetWindowText(h, &b[0], int32(len(b)))
		if err != nil {
			// ignore the error
			return 1 // continue enumeration
		}
		if syscall.UTF16ToString(b) == title {
			// note the window
			hwnd = h
			return 0 // stop enumeration
		}
		return 1 // continue enumeration
	})
	EnumWindows(cb, 0)
	if hwnd == 0 {
		return 0, fmt.Errorf("no window with title '%s' found", title)
	}
	return hwnd, nil
}

func main() {
	const title = "SoF2 MP Test"
	const ftitle = "SoF2 MP"
	sofHandle, err := FindWindow(title)

	if err != nil {
		log.Fatal(err)
	}

	VK_ONE := 0x31
	VK_TWO := 0x32

	for {
		time.Sleep(1 * time.Second)

		if sofHandle == GetForegroundWindow() {
			for {
				oneKeyPressed, _, _ := procGetAsyncKeyState.Call(uintptr(VK_ONE))
				if oneKeyPressed != 0 {
					CopyTextLastChatter()
					break
				}

				twoKeyPressed, _, _ := procGetAsyncKeyState.Call(uintptr(VK_TWO))
				if twoKeyPressed != 0 {
					ImpersonateLastChatter()
					break
				}
			}

		}

	}
}
