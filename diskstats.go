package main

import (
	"bufio"
	"os"
	"regexp"
	"strings"
)

const PROC_DISKSTATS = "/proc/diskstats"
const PROC_PARTITIONS = "/proc/partitions"

const KERNEL26_DISKSTATS_INDEX = 2

type DiskStats struct {
	All       []DiskStat
	Total     IoStat
	TotalDiff IoStat
}

func (ds *DiskStats) Init() {
	ds.All = []DiskStat{}
	ds.Total = IoStat{}
	ds.TotalDiff = IoStat{}
	ds.Update()
}

func (ds *DiskStats) Update() {
	newStats := readDiskStats(ds.All)
	ds.All = nil
	ds.All = newStats
	var tot IoStat
	for _, d := range ds.All {
		tot.ReadsCompleted += d.Stat.ReadsCompleted
		tot.ReadsMerged += d.Stat.ReadsMerged
		tot.SectorsRead += d.Stat.SectorsRead
		tot.SpentReading += d.Stat.SpentReading
		tot.WritesCompleted += d.Stat.WritesCompleted
		tot.WritesMerged += d.Stat.WritesMerged
		tot.SectorsWritten += d.Stat.SectorsWritten
		tot.SpentWriting += d.Stat.SpentWriting
		tot.IoInProgress += d.Stat.IoInProgress
		tot.SpentIo += d.Stat.SpentIo
		tot.WeightedSpentIo += d.Stat.WeightedSpentIo
	}
	ds.TotalDiff = calcDiffIoStat(ds.Total, tot)
	ds.Total = tot
}

type DiskStat struct {
	Name string
	Stat IoStat
	Diff IoStat
}

type IoStat struct {
	ReadsCompleted  int64
	ReadsMerged     int64
	SectorsRead     int64
	SpentReading    int64
	WritesCompleted int64
	WritesMerged    int64
	SectorsWritten  int64
	SpentWriting    int64
	IoInProgress    int64
	SpentIo         int64
	WeightedSpentIo int64
}

func calcDiffIoStat(os IoStat, ns IoStat) IoStat {
	return IoStat{
		ReadsCompleted:  ns.ReadsCompleted - os.ReadsCompleted,
		ReadsMerged:     ns.ReadsMerged - os.ReadsMerged,
		SectorsRead:     ns.SectorsRead - os.SectorsRead,
		SpentReading:    ns.SpentReading - os.SpentReading,
		WritesCompleted: ns.WritesCompleted - os.WritesCompleted,
		WritesMerged:    ns.WritesMerged - os.WritesMerged,
		SectorsWritten:  ns.SectorsWritten - os.SectorsWritten,
		SpentWriting:    ns.SpentWriting - os.SpentWriting,
		SpentIo:         ns.SpentIo - os.SpentIo,
		WeightedSpentIo: ns.WeightedSpentIo - os.WeightedSpentIo,
	}
}

func getDisk(name string, disks []DiskStat) DiskStat {
	for _, d := range disks {
		if d.Name == name {
			return d
		}
	}
	return DiskStat{}
}

func readDiskStats(oldStats []DiskStat) []DiskStat {
	stats := []DiskStat{}
	inFile, _ := os.Open(PROC_DISKSTATS)
	defer inFile.Close()
	scanner := bufio.NewScanner(inFile)
	for scanner.Scan() {
		line := scanner.Text()
		name, stat := parseDiskStatLine(KERNEL26_DISKSTATS_INDEX, line)
		if !regexp.MustCompile(`ram.*|loop.*|sr.*`).MatchString(name) {
			stats = append(stats, DiskStat{Name: name, Stat: stat, Diff: calcDiffIoStat(getDisk(name, oldStats).Stat, stat)})
		}
	}
	return stats
}

func parseDiskStatLine(startIndex int, line string) (name string, stat IoStat) {
	statArray := regexp.MustCompile(`[\s\t ]+`).Split(strings.TrimSpace(line), -1)
	if len(statArray)-startIndex-1 != 11 {
		return "", IoStat{}
	}
	return statArray[startIndex+0], IoStat{
		ReadsCompleted:  parseInt64_(statArray[startIndex+1]),
		ReadsMerged:     parseInt64_(statArray[startIndex+2]),
		SectorsRead:     parseInt64_(statArray[startIndex+3]),
		SpentReading:    parseInt64_(statArray[startIndex+4]),
		WritesCompleted: parseInt64_(statArray[startIndex+5]),
		WritesMerged:    parseInt64_(statArray[startIndex+6]),
		SectorsWritten:  parseInt64_(statArray[startIndex+7]),
		SpentWriting:    parseInt64_(statArray[startIndex+8]),
		IoInProgress:    parseInt64_(statArray[startIndex+9]),
		SpentIo:         parseInt64_(statArray[startIndex+10]),
		WeightedSpentIo: parseInt64_(statArray[startIndex+11]),
	}
}
