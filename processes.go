package main

import (
	"fmt"
	ioutil "io/ioutil"
	osuser "os/user"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type Processes struct {
	all      map[int64]*Process
	All      []*Process
	Time     int64
	TimePrev int64
}

func (pra *Processes) Init() {
	pra.all = map[int64]*Process{}
	pra.All = []*Process{}
	pra.Update()
}

func (pra *Processes) Update() {
	pra.All = nil
	pra.All = []*Process{}
	pra.initProcessPidList()
	pra.TimePrev = pra.Time
	pra.Time = int64(time.Now().Unix())
	for _, value := range pra.all {
		pra.All = append(pra.All, value)
	}
}

type Process struct {
	Pid       int64
	ParentPid int64
	Cmdline   string
	Stat      map[string]string
	Statm     map[string]string
	Diff      map[string]int64
	User      User
	Group     User
	live      bool
}

type User struct {
	Real      Auth
	Effective Auth
}

type Auth struct {
	Name string
	Id   int
}

func (pr *Process) Init() {
	pr.live = false
	pr.Update()
}

func (pr *Process) Update() {
	newStat := readProcessStatMap(pr.Pid)
	if !regexp.MustCompile("[RSDZTW]+").MatchString(newStat["state"]) {
		return
	}
	pr.Diff = calcDiffLoad(extractInts(pr.Stat), extractInts(newStat))
	pr.Stat = nil
	pr.Stat = newStat
	pr.Statm = nil
	pr.Statm = readProcessStatmMap(pr.Pid)
	pr.Cmdline = readOneLine(fmt.Sprintf("%v/%v/cmdline", PROC_DIR, pr.Pid))
	pr.updateAuthInfo()
	pr.live = true
}

func (pr *Process) updateAuthInfo() {
	fileName := fmt.Sprintf("%v/%v/status", PROC_DIR, pr.Pid)
	auths := readFileMap([]string{"Uid", "Gid"}, fileName, `:`)
	for key, value := range auths {
		auths[key] = strings.TrimSpace(value)
	}
	uid := regexp.MustCompile(`[\s]+`).Split(auths["Uid"], -1)
	gid := regexp.MustCompile(`[\s]+`).Split(auths["Gid"], -1)

	userR, err := osuser.LookupId(uid[0])
	if err == nil {
		pr.User.Real.Name = userR.Username
		pr.User.Real.Id, _ = strconv.Atoi(userR.Uid)
	}

	userE, _ := osuser.LookupId(uid[1])
	if err == nil {
		pr.User.Effective.Name = userE.Username
		pr.User.Effective.Id, _ = strconv.Atoi(userE.Uid)
	}

	pr.Group.Real.Id, _ = strconv.Atoi(gid[0])
	pr.Group.Effective.Id, _ = strconv.Atoi(gid[1])
}

func calcDiffLoad(times []int64, ntimes []int64) map[string]int64 {
	Diff := map[string]int64{
		"utime":  int64(ntimes[0] - times[0]),
		"stime":  int64(ntimes[1] - times[1]),
		"cutime": int64(ntimes[2] - times[2]),
		"cstime": int64(ntimes[3] - times[3]),
	}
	return Diff
}

func extractInts(stat map[string]string) []int64 {
	utime, _ := strconv.ParseInt(stat["utime"], 0, 64)
	stime, _ := strconv.ParseInt(stat["stime"], 0, 64)
	cutime, _ := strconv.ParseInt(stat["cutime"], 0, 64)
	cstime, _ := strconv.ParseInt(stat["cstime"], 0, 64)
	return []int64{utime, stime, cutime, cstime}
}

func readLineMap(names []string, valuesLine string) map[string]string {
	Map := map[string]string{}
	values := regexp.MustCompile(`[\t ]+`).Split(valuesLine, -1)
	if len(values) != len(names) {
		return Map
	}
	for i := 0; i < len(names); i++ {
		Map[names[i]] = values[i]
	}
	return Map
}

func readProcessStatMap(pid int64) map[string]string {
	return readLineMap(initProcessStatNames(), readOneLine(fmt.Sprintf("%v/%v/stat", PROC_DIR, pid)))
}

func readProcessStatmMap(pid int64) map[string]string {
	return readLineMap(initProcessStatmNames(), readOneLine(fmt.Sprintf("%v/%v/statm", PROC_DIR, pid)))
}

func initProcessStatmNames() []string {
	Names := []string{
		"size",
		"resident",
		"share",
		"text",
		"lib",
		"data",
		"dt",
	}
	return Names
}

func initProcessStatNames() []string {
	Names := []string{
		"pid",
		"comm",
		"state",
		"ppid",
		"pgrp",
		"session",
		"tty_nr",
		"tpgid",
		"flags",
		"minflt",
		"cminflt",
		"majflt",
		"cmajflt",
		"utime",
		"stime",
		"cutime",
		"cstime",
		"priority",
		"nice",
		"num_threads",
		"itrealvalue",
		"starttime",
		"vsize",
		"rss",
		"rsslim",
		"startcode",
		"endcode",
		"startstack %lu",
		"kstkesp",
		"kstkeip",
		"signal",
		"blocked",
		"sigignore",
		"sigcatch",
		"wchan",
		"nswap",
		"cnswap",
		"exit_signal",
		"processor",
		"rt_priority",
		"policy",
		"delayacct_blkio_ticks",
		"guest_time",
		"cguest_time",
	}
	return Names
}

func (pra *Processes) initProcessPidList() {
	for _, value := range pra.all {
		value.live = false
	}
	dir, _ := ioutil.ReadDir(PROC_DIR)
	for _, file := range dir {
		if file.IsDir() && regexp.MustCompile("[0123456789]+").MatchString(file.Name()) {
			pid, _ := strconv.ParseInt(file.Name(), 0, 64)
			if proc, ok := pra.all[pid]; ok {
				proc.Init()
			} else {
				pra.all[pid] = &Process{Pid: pid}
				pra.all[pid].Init()
			}
		}
	}
	for key, value := range pra.all {
		if !value.live {
			delete(pra.all, key)
		}
	}
}
