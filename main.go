package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"strconv"
	"strings"
)

func isSetUid(filename string) (bool, error) {
	info, err := os.Stat(filename)
	if err != nil {
		return false, err
	}

	return info.Mode()&os.ModeSetuid != 0, nil
}

func main() {
	rprocs, err := GetUserProcesses("root")
	if err != nil {
		panic(err)
	}

	cUser, err := user.Current()
	if err != nil {
		panic(err)
	}
	fmt.Printf("Current User: %s (UID: %s)\n", cUser.Username, cUser.Uid)

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
			fmt.Printf("UID: %d\tIs %s setuid? %v\tCmdline: %s\n", proc.uid, cmdPath, isSetuid, proc.cmdline)

			cUid, _ := strconv.Atoi(cUser.Uid)

			if strings.Contains(proc.cmdline, cUser.Username) && proc.uid != cUid {
				fmt.Printf("Found process with current user in cmdline: %s\n", proc.cmdline)
				panic(errors.New("found process with current user in cmdline"))
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
