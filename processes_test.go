package main

import (
	"fmt"
	"regexp"
	"testing"
)

func TestPidListAndProcessStatsRead(t *testing.T) {
	procs := Processes{}
	procs.Init()
	if len(procs.All) < 1 {
		t.Error("Empty process List!")
	}
	pid := procs.All[15].Pid
	stats := readProcessStatMap(pid)
	if len(stats) < 1 {
		t.Error("Empty stats map!")
	}
	if !regexp.MustCompile("[1234567890]+").MatchString(stats["pid"]) {
		t.Error("pid is empty!")
	}
	if len(stats["comm"]) < 3 {
		t.Error("comm is empty!")
	}
	fmt.Printf("Process pid: %v, name: %v\n", stats["pid"], stats["comm"])
	statsm := readProcessStatmMap(pid)
	if len(statsm) < 1 {
		t.Error("Empty statsm (memory) map!")
	}
	if !regexp.MustCompile("[1234567890]+").MatchString(statsm["size"]) {
		t.Error("size is empty!")
	}
}
