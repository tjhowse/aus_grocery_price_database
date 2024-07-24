package main

import "runtime"

// Returns the RAM utilisation of this process as a percentage.
func GetRAMUtilisationPercent() float64 {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return float64(m.Alloc) / float64(m.Sys) * 100
}
