package main

import (
	"regexp"
	"strconv"
	"strings"
)

const PROC_NET_DATA = "/proc/net/dev"

type NetData struct {
	Bytes      int64
	Packets    int64
	Errs       int64
	Drop       int64
	Fifo       int64
	Frame      int64
	Compressed int64
	Multicast  int64
}

type NetDev struct {
	Name     string
	Receive  NetData
	Transmit NetData
}

type Net struct {
	All []NetDev
}

func (n *Net) Init() {
	n.All = []NetDev{}
	n.Update()
}

func (n *Net) Update() {
	n.All = readNetDevs()
}

func readNetDevs() []NetDev {
	netDevs := []NetDev{}
	devMap := readFileMap([]string{`[\w]+`}, PROC_NET_DATA, `:`)
	for key, value := range devMap {
		vals := regexp.MustCompile(`[\s]+`).Split(strings.TrimSpace(value), -1)
		if len(vals) >= 16 {
			netDevs = append(netDevs, NetDev{Name: key, Receive: readNetData(vals, 0), Transmit: readNetData(vals, 8)})
		}
	}
	return netDevs
}
func readNetData(in []string, off int) NetData {
	netData := NetData{}
	netData.Bytes, _ = strconv.ParseInt(in[0+off], 0, 64)
	netData.Packets, _ = strconv.ParseInt(in[1+off], 0, 64)
	netData.Errs, _ = strconv.ParseInt(in[2+off], 0, 64)
	netData.Drop, _ = strconv.ParseInt(in[3+off], 0, 64)
	netData.Fifo, _ = strconv.ParseInt(in[4+off], 0, 64)
	netData.Frame, _ = strconv.ParseInt(in[5+off], 0, 64)
	netData.Compressed, _ = strconv.ParseInt(in[6+off], 0, 64)
	netData.Multicast, _ = strconv.ParseInt(in[7+off], 0, 64)
	return netData
}
