// +build windows

package utils

import "golang.org/x/sys/windows"

var (
	Modkernel32 = windows.NewLazySystemDLL("kernel32.dll")
	ModPsapi    = windows.NewLazySystemDLL("psapi.dll")
)
