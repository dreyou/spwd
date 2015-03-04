package main

import (
	"bufio"
	"os"
	"regexp"
	"sort"
	"strings"
	"syscall"
)

const PROC_MOUNT = "/proc/mounts"

var VALID_FS_LIST = []string{"ext2", "ext3", "ext4", "xfs", "gfs", "ntfs", "vfat"}

type AllFs struct {
	All []FsInfo
}

func (fs *AllFs) Init() {
	fs.All = []FsInfo{}
	fs.Update()
}

func (fs *AllFs) Update() {
	fs.All = nil
	fs.All = []FsInfo{}
	fs.All = readValidFs()
}

type FsInfo struct {
	Dev   string
	Mount string
	Type  string
	Size  FsSize
}

type FsSize struct {
	BlockSize uint64
	Avail     uint64
	Used      uint64
	Total     uint64
}

func readFsSize(path string) FsSize {
	var stat syscall.Statfs_t
	_ = syscall.Statfs(path, &stat)
	fsSize := FsSize{}
	fsSize.BlockSize = uint64(stat.Bsize)
	fsSize.Total = uint64(stat.Blocks) * uint64(stat.Bsize)
	fsSize.Used = (uint64(stat.Blocks) - uint64(stat.Bfree)) * uint64(stat.Bsize)
	fsSize.Avail = uint64(stat.Bavail) * uint64(stat.Bsize)
	return fsSize
}

func readValidFs() []FsInfo {
	all := []FsInfo{}
	sort.Strings(VALID_FS_LIST)
	inFile, _ := os.Open(PROC_MOUNT)
	defer inFile.Close()
	scanner := bufio.NewScanner(inFile)
	for scanner.Scan() {
		line := scanner.Text()
		res := regexp.MustCompile(`[ \t]+`).Split(line, -1)
		fsMount := res[1]
		fsType := res[2]
		s := sort.SearchStrings(VALID_FS_LIST, fsType)
		if s < len(VALID_FS_LIST) && VALID_FS_LIST[s] == fsType && !strings.Contains(fsMount, "docker") {
			all = append(all, FsInfo{Dev: res[0], Mount: res[1], Type: fsType, Size: readFsSize(fsMount)})
		}
	}
	return all
}
