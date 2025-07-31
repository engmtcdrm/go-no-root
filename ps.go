package main

import (
	"fmt"
	"io"
	"os"
	"os/user"
	"path"
	"regexp"
	"strconv"
	"strings"
	"syscall"
)

type Process struct {
	pid     int
	ppid    int
	state   string
	uid     int
	gid     int
	binary  string
	cmdline []string
}

func (p *Process) Pid() int {
	return p.pid
}

func (p *Process) Ppid() int {
	return p.ppid
}

func (p *Process) Uid() int {
	return p.uid
}

func (p *Process) Gid() int {
	return p.gid
}

func (p *Process) Binary() string {
	return p.binary
}

func (p *Process) Cmdline() []string {
	return p.cmdline
}

func getUIDFromUser(username string) (int, error) {
	u, err := user.Lookup(username)
	if err != nil {
		return -1, fmt.Errorf("user not found: %s", username)
	}

	uid, err := strconv.Atoi(u.Uid)
	if err != nil {
		return -1, fmt.Errorf("invalid UID: %s", err)
	}

	return uid, nil
}

func GetProcessByPid(pid int) (*Process, error) {
	pidDir := fmt.Sprintf("/proc/%d", pid)

	procStat, err := os.Stat(pidDir)
	if err != nil {
		return nil, fmt.Errorf("pid %d: does not exist", pid)
	}

	if !procStat.IsDir() {
		return nil, fmt.Errorf("could not get pid information, pid %d", pid)
	}

	statPath := path.Join(pidDir, "stat")
	statBytes, err := os.ReadFile(statPath)
	if err != nil {
		return nil, fmt.Errorf("could not read stat file for PID %d: %v", pid, err)
	}

	stats := strings.TrimSpace(string(statBytes))
	binStart := strings.IndexRune(stats, '(') + 1
	binEnd := strings.IndexRune(stats[binStart:], ')')
	binary := stats[binStart : binStart+binEnd]
	statSlice := strings.Split(strings.TrimSpace(stats[binStart+binEnd+1:]), " ")

	if len(statSlice) < 2 {
		return nil, fmt.Errorf("invalid stat format for PID %d", pid)
	}

	state := statSlice[0]
	ppid, _ := strconv.Atoi(statSlice[1])

	cmdlinePath := path.Join(pidDir, "cmdline")
	cmdlineBytes, err := os.ReadFile(cmdlinePath)
	if err != nil {
		return nil, fmt.Errorf("could not read cmdline file for PID %d: %v", pid, err)
	}

	cmdline := strings.Split(string(cmdlineBytes), "\x00")

	dsys, err := os.Stat(pidDir)
	if err != nil {
		return nil, fmt.Errorf("could not stat process directory for PID %d: %v", pid, err)
	}
	stat, ok := dsys.Sys().(*syscall.Stat_t)
	if !ok {
		return nil, fmt.Errorf("could not get syscall stat for PID %d", pid)
	}

	return &Process{
		pid:     pid,
		uid:     int(stat.Uid),
		gid:     int(stat.Gid),
		binary:  binary,
		cmdline: cmdline,
		state:   state,
		ppid:    ppid,
	}, nil
}

func GetUserProcesses(username string) ([]Process, error) {
	uid, err := getUIDFromUser(username)
	if err != nil {
		return nil, err
	}

	procs, err := GetProcesses()
	if err != nil {
		return nil, err
	}

	var userProcs []Process
	for _, proc := range procs {
		if proc.uid == uid {
			userProcs = append(userProcs, proc)
		}
	}
	return userProcs, nil
}

func GetProcesses() ([]Process, error) {
	pdir, err := os.Open("/proc")
	if err != nil {
		return nil, err
	}
	defer pdir.Close()

	// Regex to match numeric directory names
	reDigit := regexp.MustCompile(`^\d+$`)

	procs := []Process{}
	for {
		dirs, err := pdir.Readdir(10)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		for _, dir := range dirs {
			if !reDigit.MatchString(dir.Name()) {
				continue
			}

			pid, err := strconv.Atoi(dir.Name())
			if err != nil {
				continue
			}

			proc, err := GetProcessByPid(pid)
			if err != nil {
				continue
			}

			procs = append(procs, *proc)
		}

	}
	return procs, nil
}
