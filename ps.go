package main

import (
	"fmt"
	"io"
	"os"
	"os/user"
	"regexp"
	"strconv"
	"strings"
	"syscall"
)

type Process struct {
	pid     int
	uid     int
	gid     int
	binary  string
	cmdline string
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

	// Regular expression to match numeric directory names
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
			// syscall.Stat_t
			dsys := dir.Sys()

			stat, ok := dsys.(*syscall.Stat_t)

			// Type assert to *syscall.Stat_t
			if !ok {
				continue
			}

			// Only care about root processes
			// if stat.Uid != 0 {
			// 	continue
			// }

			pid, err := strconv.ParseInt(dir.Name(), 10, 0)
			if err != nil {
				continue
			}

			statPath := fmt.Sprintf("/proc/%d/stat", pid)
			dataBytes, err := os.ReadFile(statPath)
			if err != nil {
				continue
			}

			data := string(dataBytes)
			binStart := strings.IndexRune(data, '(') + 1
			binEnd := strings.IndexRune(data[binStart:], ')')
			binary := data[binStart : binStart+binEnd]

			cmdlinePath := fmt.Sprintf("/proc/%d/cmdline", pid)
			cmdlineBytes, err := os.ReadFile(cmdlinePath)
			if err != nil {
				continue
			}

			cmdline := strings.Replace(string(cmdlineBytes), "\x00", " ", -1)

			procs = append(procs, Process{pid: int(pid), uid: int(stat.Uid), gid: int(stat.Gid), binary: binary, cmdline: cmdline})
		}

	}
	return procs, nil
}
