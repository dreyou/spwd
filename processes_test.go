package main

import (
	"fmt"
	"testing"
)

func TestPidListAndProcessStatsRead(t *testing.T) {
	procs := Processes{}
	procs.Init()
	if len(procs.All) < 1 {
		t.Error("Empty process List!")
	}
	pid := procs.All[len(procs.All)-2].Pid
	stats := readProcessStatMap(pid)
	if len(stats) < 1 {
		t.Error("Empty stats map!")
	}
	if stats["pid"] <= 0 {
		t.Error("invalid pid !")
	}
	fmt.Printf("Process pid: %v, name: %v\n", stats["pid"])
	statsm := readProcessStatmMap(pid)
	if len(statsm) < 1 {
		t.Error("Empty statsm (memory) map!")
	}
	if statsm["size"] <= 0 {
		t.Error("memory size is not valid!")
	}
}
