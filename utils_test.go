package main

import "testing"

func TestGetRAMUtilisationPercent(t *testing.T) {
	ram_start := GetRAMUtilisationPercent()
	if ram_start < 0 {
		t.Errorf("RAM utilisation is less than 0")
	}
	// Allocate a bunch of stuff that will use RAM
	_ = make([]byte, 1024*1024*1024)

	ram_end := GetRAMUtilisationPercent()
	if ram_end < ram_start {
		t.Errorf("RAM utilisation has not increased")
	}
}
