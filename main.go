package main

import (
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"slices"
	"strconv"
)

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

	goproc, err := GetProcessByPid(strconv.Itoa(os.Getpid()))
	if err != nil {
		panic(err)
	}

	fmt.Printf("Current User: %s (UID: %s)\n\n", goproc.User().Username, goproc.User().Uid)

	fmt.Printf(
		"Go Parent Process Info:\n  PID: %s\n  PPID: %s\n  Username: %s (%s)\n  Groupname: %s (%s)\n  Binary: %s\n  Cmdline: %v\n\n",
		goproc.Pid(),
		goproc.Ppid(),
		goproc.User().Username,
		goproc.Uid(),
		goproc.Group().Name,
		goproc.Gid(),
		goproc.Binary(),
		goproc.Cmdline(),
	)

	ppid := goproc.Ppid()
	for ppid > "1" {
		proc, err := GetProcessByPid(ppid)
		if err != nil {
			break
		}

		fmt.Printf(
			"Parent Process Info:\n  PID: %s\n  PPID: %s\n  UID: %s\n  GID: %s\n  Binary: %s\n  Cmdline: %v\n State: %v\n\n",
			proc.Pid(),
			proc.Ppid(),
			proc.Uid(),
			proc.Gid(),
			proc.Binary(),
			proc.Cmdline(),
			proc.State(),
		)

		if goproc.Uid() != "0" && proc.Uid() == "0" && proc.Ppid() != "0" /*&& proc.Binary() != "cron" && proc.Binary() != "crond"*/ {
			fmt.Println(fmt.Errorf("found root process in parent tree: %s (PID: %s)", proc.Binary(), proc.Pid()))
			// os.Exit(1)
		}

		ppid = proc.Ppid()
	}

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
				"    PID: %s\n    PPID: %s\n    UID: %s\n    Cmdline: %s\n    Is %v setuid? %v\n\n",
				proc.pid,
				proc.ppid,
				proc.uid,
				proc.cmdline,
				cmdPath,
				isSetuid,
			)

			if slices.Contains(proc.cmdline, cUser.Username) && proc.uid != cUser.Uid {
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
