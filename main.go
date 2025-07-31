package main

import (
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"slices"
	"strconv"
)

func traverseProcessTree(startPid int, currentUser *user.User) error {
	bannedCommands := []string{"sudo", "su", "doas", "super", "calife", "pkexec", "systemd-run", "machinectl", "runuser"}

	for pid := startPid; pid > 1; {
		proc, err := GetProcessByPid(pid)
		if err != nil {
			break
		}

		cUid, err := strconv.Atoi(currentUser.Uid)
		// If same user, continue up the tree
		if proc.Uid() == cUid {
			pid = proc.Ppid()
			continue
		}

		// If root process, check for banned commands
		if proc.Uid() == 0 {
			if slices.Contains(bannedCommands, proc.Binary()) {
				return fmt.Errorf("detected privilege escalation via %s", proc.Binary())
			}
		}

		pid = proc.Ppid()
	}
	return nil
}

func isSetUid(filename string) (bool, error) {
	info, err := os.Stat(filename)
	if err != nil {
		return false, err
	}

	return info.Mode()&os.ModeSetuid != 0, nil
}

func main() {
	cUser, err := user.Current()
	if err != nil {
		panic(err)
	}

	fmt.Printf("Current User: %s (UID: %s)\n\n", cUser.Username, cUser.Uid)

	goproc, err := GetProcessByPid(os.Getpid())
	if err != nil {
		panic(err)
	}

	fmt.Printf(
		"Go Process Info:\n  PID: %d\n  PPID: %d\n  UID: %d\n  GID: %d\n  Binary: %s\n  Cmdline: %v\n\n",
		goproc.Pid(),
		goproc.Ppid(),
		goproc.Uid(),
		goproc.Gid(),
		goproc.Binary(),
		goproc.Cmdline(),
	)

	rprocs, err := GetUserProcesses("root")
	if err != nil {
		panic(err)
	}

	fmt.Print("Root Processes:\n\n")

	for _, proc := range rprocs {
		// fmt.Println("PID:", proc.pid, " \tUID:", proc.uid, " \tGID:", proc.gid, " \tBinary:", proc.binary, "\tCmdline:", proc.cmdline)
		cmd := proc.binary
		cmdPath, err := exec.LookPath(cmd)
		if err != nil {
			continue
		}

		isSetuid, err := isSetUid(cmdPath)
		if err != nil {
			panic(err)
		}

		if isSetuid {
			fmt.Printf("  Command: %s\n", cmdPath)
			fmt.Printf(
				"    PID: %d\n    PPID: %d\n    UID: %d\n    Cmdline: %s\n    Is %s setuid? %v\n\n",
				proc.pid,
				proc.ppid,
				proc.uid,
				cmdPath,
				isSetuid,
				proc.cmdline,
			)

			cUid, _ := strconv.Atoi(cUser.Uid)
			if slices.Contains(proc.cmdline, cUser.Username) && proc.uid != cUid {
				panic(fmt.Errorf("found process with current user in it %v", proc))
			}
		}

	}

	// cmd := "sudo"

	// cmdPath, err := whichCommand(cmd)
	// if err != nil {
	// 	fmt.Println("Error finding command:", err)
	// 	return
	// }

	// isSetuid, err := fileSetUid(cmdPath)
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Printf("Is %s setuid? %v\n", cmdPath, isSetuid)
}
