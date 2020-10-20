package titan

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

type Process struct {
	pid int
	cpu float64
}

var pid int

func init() {
	pid = os.Getpid()
}

func GetCpuUsage() (*Process, error) {
	cmd := exec.Command("ps", "aux", strconv.Itoa(pid))
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return nil, err
	}

	for {
		line, err := out.ReadString('\n')
		if err != nil {
			break
		}
		tokens := strings.Split(line, " ")
		ft := make([]string, 0)
		for _, t := range tokens {
			if t != "" && t != "\t" {
				ft = append(ft, t)
			}
		}
		id, err := strconv.Atoi(ft[1])
		if err != nil {
			continue
		}
		cpu, err := strconv.ParseFloat(ft[2], 64)
		if err != nil {
			return nil, err
		}
		if id == pid {
			return &Process{pid, cpu}, nil
		}
	}
	return nil, fmt.Errorf("process not found  %d", pid)
}
