package main

import (
	"fmt"
	"regexp"
	"testing"
	"time"
)

func TestSysVersionRegexp(t *testing.T) {
	verString := "2.6.234-wer.wer"
	verValue := regexp.MustCompile(`^(\d+)\.(\d+)[^\d].*$`).ReplaceAllString(verString, "$1$2")
	if verValue != "26" {
		t.Error("Vrong version detection!")
	}
}

func TestSysKernel(t *testing.T) {
	kernel := new(Kernel)
	kernel.Init()
	if len(kernel.Hostname) < 1 {
		t.Error("Hostname is empty!")
	}
}

func TestMemInfo(t *testing.T) {
	mem := new(Mem)
	mem.Init()
	if len(mem.Info) < 1 {
		t.Error("Memory info must contain at least one string!")
	}
}

func TestReadFileToMap(t *testing.T) {
	Info := readFileToMap(PROC_MEMNFO, "[: ]+")
	if len(Info) < 1 {
		t.Error("Memory info must contain at least one string!")
	}
	if _, ok := Info["MemTotal"]; !ok {
		t.Error("Memory info must contains MemTotal value!")
	}
	if Info["MemTotal"] == 0 {
		t.Error("Memory info MemTotal must be not zero!")
	}
}

func TestCreateCpusArray(t *testing.T) {
	Cpus := createCpusArray()
	if len(Cpus) < 1 {
		t.Error("Cpu count's must be greater then 1!")
	}
}

func TestCountWords(t *testing.T) {
	count := countWords(PROC_CPUINFO, "processor")
	fmt.Printf("count: %v\n", count)
	if count == 0 {
		t.Error("Cpu count's must not be 0!")
	}
}
func TestProcLoad(t *testing.T) {
	stat := new(Stat)
	stat.Init()
	for i := 0; i < 1; i++ {
		time.Sleep(1000 * time.Millisecond)
		stat.Update()
		for _, p := range stat.Cpus {
			fmt.Printf("%v - %.2f (%.2f)\n", p.Name, p.Load.User, p.Load.All)
		}
	}
	if stat.Cpus[0].Load.All == 0 {
		t.Error("Cpu Load must not be 0!")
	}
}
func TestProcAll(t *testing.T) {
	all := new(ProcAll)
	all.Init()
	time.Sleep(1000 * time.Millisecond)
	all.Update()
	if all.Stat.Cpus[0].Load.All == 0 {
		t.Error("Cpu Load must not be 0!")
	}
	if all.Stat.Stats["btime"] == 0 {
		t.Error("btime must not be zero")
	}
	if len(all.Meminfo.Info) < 1 {
		t.Error("Memory info must contain at least one string!")
	}
	if len(all.Kernel.Hostname) < 1 {
		t.Error("Hostname is empty!")
	}
	if len(all.Processes.All) < 1 {
		t.Error("Empty processes list!")
	}
	if len(all.AllFs.All) < 1 {
		t.Error("Empty processes list!")
	}
	if len(all.Net.All) < 1 {
		t.Error("Empty net dev list!")
	}
	if len(all.DiskStats.All) < 1 {
		t.Error("Empty disk io list!")
	}
}
