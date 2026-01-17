package noroot

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
	pid     string
	ppid    string
	state   string
	uid     string
	gid     string
	binary  string
	cmdline []string
	user    *user.User
	group   *user.Group
}

func (p *Process) Pid() string {
	return p.pid
}

func (p *Process) Ppid() string {
	return p.ppid
}

func (p *Process) Uid() string {
	return p.uid
}

func (p *Process) Gid() string {
	return p.gid
}

func (p *Process) Binary() string {
	return p.binary
}

func (p *Process) Cmdline() []string {
	return p.cmdline
}

func (p *Process) User() *user.User {
	if p.user != nil {
		return p.user
	}
	return nil
}

func (p *Process) Group() *user.Group {
	if p.group != nil {
		return p.group
	}
	return nil
}

func (p *Process) State() string {
	return p.state
}

func getUIDFromUser(username string) (string, error) {
	u, err := user.Lookup(username)
	if err != nil {
		return "-1", fmt.Errorf("user not found: %s", username)
	}

	return u.Uid, nil
}

func GetCurrentGoProcess() (*Process, error) {
	return GetProcessByPid(strconv.Itoa(os.Getpid()))
}

func GetProcessByPid(pid string) (*Process, error) {
	pidDir := fmt.Sprintf("/proc/%s", pid)

	procStat, err := os.Stat(pidDir)
	if err != nil {
		return nil, fmt.Errorf("pid %s: does not exist", pid)
	}

	if !procStat.IsDir() {
		return nil, fmt.Errorf("could not get pid information, pid %s", pid)
	}

	commPath := path.Join(pidDir, "comm")
	commBytes, _ := os.ReadFile(commPath)

	binary := strings.TrimSpace(string(commBytes))

	statPath := path.Join(pidDir, "stat")
	statBytes, err := os.ReadFile(statPath)
	if err != nil {
		return nil, fmt.Errorf("could not read stat file for PID %s: %v", pid, err)
	}

	stats := strings.TrimSpace(string(statBytes))

	matches := regexp.MustCompile(`^\d+ \((.*)\)`).FindStringSubmatch(stats)
	if len(matches) < 2 {
		return nil, fmt.Errorf("could not parse binary name from stat file for PID %s", pid)
	}
	statSlice := strings.Split(strings.TrimSpace(stats[len(matches[0]):]), " ")

	// Fallback in case /proc/<pid>/comm file had an issue
	if binary == "" {
		binary = matches[1]
	}

	if len(statSlice) < 2 {
		return nil, fmt.Errorf("invalid stat format for PID %s", pid)
	}

	state := statSlice[0]
	ppid := statSlice[1]

	cmdlinePath := path.Join(pidDir, "cmdline")
	cmdlineBytes, err := os.ReadFile(cmdlinePath)
	if err != nil {
		return nil, fmt.Errorf("could not read cmdline file for PID %s: %v", pid, err)
	}

	cmdline := strings.Split(strings.TrimSpace(string(cmdlineBytes)), "\x00")

	dsys, err := os.Stat(pidDir)
	if err != nil {
		return nil, fmt.Errorf("could not stat process directory for PID %s: %v", pid, err)
	}
	stat, ok := dsys.Sys().(*syscall.Stat_t)
	if !ok {
		return nil, fmt.Errorf("could not get syscall stat for PID %s", pid)
	}

	username, err := user.LookupId(strconv.Itoa(int(stat.Uid)))
	if err != nil {
		return nil, fmt.Errorf("could not lookup user for UID %d: %v", stat.Uid, err)
	}

	group, err := user.LookupGroupId(strconv.Itoa(int(stat.Gid)))
	if err != nil {
		return nil, fmt.Errorf("could not lookup group for GID %d: %v", stat.Gid, err)
	}

	return &Process{
		pid:     pid,
		uid:     strconv.Itoa(int(stat.Uid)),
		gid:     strconv.Itoa(int(stat.Gid)),
		binary:  binary,
		cmdline: cmdline,
		state:   state,
		ppid:    ppid,
		user:    username,
		group:   group,
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

			pid := dir.Name()

			proc, err := GetProcessByPid(pid)
			if err != nil {
				continue
			}

			procs = append(procs, *proc)
		}

	}
	return procs, nil
}
