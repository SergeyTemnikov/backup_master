//go:build linux || darwin

package service

import "syscall"

func GetDiskTotalBytes(path string) (int64, error) {
	var stat syscall.Statfs_t
	if err := syscall.Statfs(path, &stat); err != nil {
		return 0, err
	}
	return int64(stat.Blocks) * int64(stat.Bsize), nil
}
