package main

import (
	"bufio"
	"fmt"
	ioutil "io/ioutil"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

/*
#include <unistd.h>
#include <sys/types.h>
#include <pwd.h>
#include <stdlib.h>
*/
import "C"

const PROC_DIR = "/proc"
const PROC_STAT = "/proc/stat"
const PROC_CPUINFO = "/proc/cpuinfo"
const PROC_MEMNFO = "/proc/meminfo"
const PROC_AVG = "/proc/loadavg"
const PROC_UPTIME = "/proc/uptime"
const PROC_SYS_KERNEL_HOSTNAME = "/proc/sys/kernel/hostname"

type ProcCommon interface {
	Init()
	Update()
}

type ProcAll struct {
	TimeStamp string
	HostId    string
	all       []ProcCommon
	Stat      Stat
	Misc      Misc
	Meminfo   Mem
	Kernel    Kernel
	Processes Processes
	AllFs     AllFs
	Net       Net
}

func (pa *ProcAll) Init() {
	pa.TimeStamp = time.Now().Format(time.RFC3339)
	pa.all = []ProcCommon{
		&pa.Stat,
		&pa.Meminfo,
		&pa.Kernel,
		&pa.Processes,
		&pa.AllFs,
		&pa.Net,
		&pa.Misc,
	}
	for _, a := range pa.all {
		a.Init()
	}
}

func (pa *ProcAll) Update() {
	for _, a := range pa.all {
		a.Update()
	}
	for _, p := range pa.Processes.All {
		p.updateLoadInfo(pa.Stat, pa.Meminfo, (pa.Processes.Time - pa.Processes.TimePrev))
	}
}

type Misc struct {
	Avg    string
	Uptime int64
}

func (m *Misc) Update() {
	m.Avg = readOneLine(PROC_AVG)
	res := regexp.MustCompile(`\.`).Split(readOneLine(PROC_UPTIME), -1)
	m.Uptime, _ = strconv.ParseInt(res[0], 0, 64)
}

func (m *Misc) Init() {
	m.Update()
}

type Kernel struct {
	Hostname string
}

func (k *Kernel) Update() {
	k.Hostname = readOneLine(PROC_SYS_KERNEL_HOSTNAME)
}

func (k *Kernel) Init() {
	k.Update()
}

type Mem struct {
	Info map[string]int64
}

func readFileToMap(fileName string, splitMatch string) map[string]int64 {
	Map := make(map[string]int64)
	inFile, _ := os.Open(fileName)
	defer inFile.Close()
	scanner := bufio.NewScanner(inFile)
	for scanner.Scan() {
		line := scanner.Text()
		res := regexp.MustCompile(splitMatch).Split(line, -1)
		if len(res) > 1 {
			Map[res[0]], _ = strconv.ParseInt(res[1], 0, 64)
		}
	}
	return Map
}

func (m *Mem) Update() {
	m.Info = readFileToMap(PROC_MEMNFO, "[: ]+")
}

func (m *Mem) Init() {
	m.Update()
}

type Stat struct {
	Cpu        *Proc
	Cpus       []*Proc
	Stats      map[string]string
	Sc_clk_tck C.long
	Pagesize   int
}

func createCpusArray() []*Proc {
	Cpus := []*Proc{
		&Proc{Match: Match{match: "cpu +", split: "[ ]+"}},
	}
	for i := 0; i < countWords(PROC_CPUINFO, "processor"); i++ {
		Cpus = append(Cpus, &Proc{Match: Match{match: fmt.Sprintf("cpu%v +", i), split: "[ ]+"}})
		Cpus[i+1].parent = Cpus[0]
	}
	return Cpus
}

func (s *Stat) Init() {
	s.Cpus = createCpusArray()
	s.Cpu = s.Cpus[0]
	s.Update()
}

func (s *Stat) Update() {
	readStatLines(PROC_STAT, proc2proc(s.Cpus))
	s.Stats = readFileMap(initProcStatNames(), PROC_STAT, `[: \t]+`)
	s.Sc_clk_tck = C.sysconf(C._SC_CLK_TCK)
	s.Pagesize = os.Getpagesize()
}

func initProcStatNames() []string {
	Names := []string{
		"btime",
		"procs_running",
		"procs_blocked",
	}
	return Names
}

func readFileMap(names []string, fileName string, splitMatch string) map[string]string {
	Map := map[string]string{}
	inFile, _ := os.Open(fileName)
	defer inFile.Close()
	scanner := bufio.NewScanner(inFile)
	for scanner.Scan() {
		line := scanner.Text()
		values := regexp.MustCompile(splitMatch).Split(line, -1)
		if len(values) < 2 {
			continue
		}
		for _, name := range names {
			if regexp.MustCompile(name).MatchString(strings.TrimSpace(values[0])) {
				Map[strings.TrimSpace(values[0])] = values[1]
			}
		}
	}
	return Map
}

type ProcMatch interface {
	isMatch(in string) bool
	doSplit(in string) []string
	update(line string)
}

func (p Proc) isMatch(in string) bool {
	return regexp.MustCompile(p.match).MatchString(in)
}

func (p Proc) doSplit(in string) []string {
	return regexp.MustCompile(p.split).Split(in, -1)
}

type ProcLoad struct {
	User   float32
	Nice   float32
	System float32
	Idle   float32
	All    float32
	Div    float32
}

type ProcStat struct {
	User   int64
	Nice   int64
	System int64
	Idle   int64
}

type Proc struct {
	Match
	Stat   ProcStat
	Diff   ProcStat
	Load   ProcLoad
	parent *Proc
}

type Match struct {
	Name  string
	match string
	split string
}

func parseInt64(in string) (i int64, err error) {
	return strconv.ParseInt(strings.TrimSpace(in), 0, 64)
}

func (p *Proc) update(line string) {
	procStats := p.doSplit(line)
	p.Name = procStats[0]
	var nstat ProcStat
	nstat.User, _ = parseInt64(procStats[1])
	nstat.Nice, _ = parseInt64(procStats[2])
	nstat.System, _ = parseInt64(procStats[3])
	nstat.Idle, _ = parseInt64(procStats[4])
	p.Load = p.calcLoad(p.Stat, nstat)
	p.Stat = nstat
}

func (p *Proc) calcLoad(stat ProcStat, nstat ProcStat) ProcLoad {
	p.Diff.User = nstat.User - stat.User
	p.Diff.Nice = nstat.Nice - stat.Nice
	p.Diff.System = nstat.System - stat.System
	p.Diff.Idle = nstat.Idle - stat.Idle
	dUser := float32(p.Diff.User)
	dNice := float32(p.Diff.Nice)
	dSystem := float32(p.Diff.System)
	dIdle := float32(p.Diff.Idle)
	var div float32
	if p.parent != nil {
		div = p.parent.Load.Div
	} else {
		div = float32((dIdle + dNice + dSystem + dUser) / 100)
	}
	return ProcLoad{trun(dUser / div), trun(dNice / div), trun(dSystem / div), trun(dIdle / div), trun((dUser + dNice + dSystem) / div), div}
}

func trun(in float32) float32 {
	return float32(int(in*100)) / 100
}

func readOneLine(fileName string) string {
	inFile, _ := os.Open(fileName)
	defer inFile.Close()
	reader := bufio.NewReader(inFile)
	line, _, _ := reader.ReadLine()
	return string(line)
}

func countWords(fileName, word string) int {
	fileContent, _ := ioutil.ReadFile(fileName)
	return strings.Count(string(fileContent), word)
}

func readStatLines(fileName string, matches []ProcMatch) {
	inFile, _ := os.Open(fileName)
	defer inFile.Close()
	scanner := bufio.NewScanner(inFile)
	for scanner.Scan() {
		line := scanner.Text()
		for _, match := range matches {
			if match.isMatch(line) {
				match.update(line)
			}
		}
	}
}

func proc2proc(proc []*Proc) []ProcMatch {
	procMatch := make([]ProcMatch, len(proc))
	for i, v := range proc {
		procMatch[i] = v
	}
	return procMatch
}
