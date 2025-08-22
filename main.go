package main

import (
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"slices"
	"strconv"
	"strings"
)

const width = 25

func isSetUid(filename string) (bool, error) {
	info, err := os.Stat(filename)
	if err != nil {
		return false, err
	}

	return info.Mode()&os.ModeSetuid != 0, nil
}

func logProcess(icon, title string, proc *Process) {
	if title == "" {
		title = "Process Info:"
	}

	fmt.Println(strings.Repeat(icon, width))
	fmt.Printf(
		"%s %s\n",
		icon,
		title,
	)
	fmt.Println(strings.Repeat(icon, width))
	fmt.Printf(
		"\n   PID: %s\n   PPID: %s\n   Username: %s (%s)\n   Groupname: %s (%s)\n   Binary: %s\n   Cmdline: %s\n   State: %v\n",
		proc.Pid(),
		proc.Ppid(),
		proc.Uid(),
		proc.User().Username,
		proc.Gid(),
		proc.Group().Name,
		proc.Binary(),
		proc.Cmdline(),
		proc.State(),
	)

	fmt.Println()
}

func main() {
	cUser, err := user.Current()
	if err != nil {
		panic(err)
	}

	goProcess, err := GetProcessByPid(strconv.Itoa(os.Getpid()))
	if err != nil {
		panic(err)
	}

	fmt.Println(strings.Repeat("🧑", width))
	fmt.Println("🧑 Current User:")
	fmt.Println(strings.Repeat("🧑", width))
	fmt.Printf(
		"   Username: %s\n   UID: %s\n   GID: %s\n\n",
		goProcess.User().Username,
		goProcess.User().Uid,
		goProcess.User().Gid,
	)

	logProcess("📺", "", goProcess)

	ppid := goProcess.Ppid()
	for ppid > "1" {
		proc, err := GetProcessByPid(ppid)
		if err != nil {
			break
		}

		logProcess("👻", "Parent Process Info:", proc)

		if goProcess.Uid() != "0" && proc.Uid() == "0" && proc.Ppid() != "0" && proc.Cmdline()[0] != "/init" /*&& proc.Binary() != "cron" && proc.Binary() != "crond"*/ {
			fmt.Println(fmt.Errorf("found root process in parent tree: %s (PID: %s)", proc.Binary(), proc.Pid()))
			os.Exit(1)
		}

		ppid = proc.Ppid()
	}

	rprocs, err := GetUserProcesses("root")
	if err != nil {
		panic(err)
	}

	fmt.Println(strings.Repeat("🧙‍♂️", width))
	fmt.Println("🧙‍♂️ Root Processes:")
	fmt.Println(strings.Repeat("🧙‍♂️", width))
	fmt.Println()

	for _, proc := range rprocs {
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
			logProcess("👻", "Root Process Info:", &proc)

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
