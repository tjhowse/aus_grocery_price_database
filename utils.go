package main

import (
	"runtime"

	"golang.org/x/sys/unix"
)

// Returns the RAM utilisation of this process as a percentage.
func GetRAMUtilisationPercent() float64 {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return float64(m.Alloc) / float64(m.Sys) * 100
}

func GetHDDBytesFree() (int, error) {
	var stat unix.Statfs_t
	if err := unix.Statfs("/data", &stat); err != nil {
		return 0, err
	}
	return int(stat.Bavail) * int(stat.Bsize), nil
}
