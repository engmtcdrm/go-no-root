package noroot

import "strconv"

func IsRoot(goProcess *Process) bool {
	// ppid := goProcess.Ppid()
	ppid, err := strconv.Atoi(goProcess.Ppid())
	if err != nil {
		return false
	}

	for ppid > 1 {
		proc, err := GetProcessByPid(strconv.Itoa(ppid))
		if err != nil {
			break
		}

		if goProcess.Uid() != "0" && proc.Uid() == "0" && proc.Ppid() != "0" && proc.Cmdline()[0] != "/init" /*&& proc.Binary() != "cron" && proc.Binary() != "crond"*/ {
			return true
		}

		ppid, err = strconv.Atoi(proc.Ppid())
		if err != nil {
			break
		}
	}

	return false
}
