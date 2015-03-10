package main

import (
	"fmt"
	"testing"
)

func TestReadDiskStatLine(t *testing.T) {
	line := "  3    0   hda 446216 784926 9550688 4382310 424847 312726 5922052 19310380 0 3376340 23705160"
	name, stat := parseDiskStatLine(2, line)
	if name != "hda" {
		t.Error("wrong disk name!")
	}
	if stat.ReadsCompleted != 446216 {
		t.Error("wrong ReadsCompleted!")
	}
}

func TestReadDiskStats(t *testing.T) {
	disks := readDiskStats([]DiskStat{})
	if len(disks) < 1 {
		t.Error("No disks!")
	}
	for _, disk := range disks {
		fmt.Printf("Name: %v, ReadsC: %v, WritesC: %v\n", disk.Name, disk.Stat.ReadsCompleted, disk.Stat.WritesCompleted)
		if disk.Name == "" {
			t.Error("No disk name!")
		}
	}
}
