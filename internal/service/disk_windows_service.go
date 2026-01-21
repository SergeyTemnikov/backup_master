//go:build windows

package service

import (
	"fmt"

	"golang.org/x/sys/windows"
)

func GetDiskTotalBytes(path string) (int64, error) {
	p, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return 0, err
	}

	var freeBytes uint64
	var totalBytes uint64
	var totalFreeBytes uint64

	err = windows.GetDiskFreeSpaceEx(
		p,
		&freeBytes,
		&totalBytes,
		&totalFreeBytes,
	)
	if err != nil {
		return 0, fmt.Errorf("GetDiskFreeSpaceEx failed: %w", err)
	}

	return int64(totalBytes), nil
}
