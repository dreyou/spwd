package main

import (
	"fmt"
	"testing"
)

func TestReadNet(t *testing.T) {
	net := Net{}
	net.Init()
	if len(net.All) < 1 {
		t.Error("No net devices!")
	}
}

func TestReadNetDevs(t *testing.T) {
	devs := readNetDevs()
	if len(devs) < 1 {
		t.Error("No net devices!")
	}
	for _, dev := range devs {
		fmt.Printf("Name: %v, BtRc: %v, BtTr: %v\n", dev.Name, dev.Receive.Bytes, dev.Transmit.Bytes)
		if dev.Name == "lo" {
			if dev.Receive.Bytes == 0 {
				t.Error("0 bytes receive over lo!")
			}
			if dev.Transmit.Bytes == 0 {
				t.Error("0 bytes transmit over lo!")
			}
		}
	}
}
