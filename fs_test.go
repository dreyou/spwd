package main

import (
	"fmt"
	"testing"
)

func TestReadFsSize(t *testing.T) {
	size := readFsSize("/")
	if size.Total < 1 {
		t.Error("Wrong filesystem syze!")
	}
	if size.Used < 1 {
		t.Error("Wrong used syze!")
	}
	if size.BlockSize < 1 {
		t.Error("Wrong block syze!")
	}
	fmt.Printf("Total: %v, Used: %v, Avail: %v\n", size.Total/1024, size.Used/1024, size.Avail/1024)
}

func TestReadValidFs(t *testing.T) {
	fs := readValidFs()
	if len(fs) < 1 {
		t.Error("Empty filesystem list!")
	}
	for _, val := range fs {
		fmt.Printf("Dev: %v, Mount: %v, Type: %v\n", val.Dev, val.Mount, val.Type)
	}
}
func TestAllFs(t *testing.T) {
	allFs := AllFs{}
	allFs.Init()
	if len(allFs.All) < 1 {
		t.Error("Empty filesystem list!")
	}
	for _, val := range allFs.All {
		if val.Mount == "/" {
			if val.Size.Total == 0 {
				t.Error("Zero fs total!")
			}
			fmt.Printf("Mount: %v, Total: %v\n", val.Mount, val.Size.Total/1024)
		}
	}
}
