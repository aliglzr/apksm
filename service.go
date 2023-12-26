package apksm

import (
	"apksm/logger"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

type Services []*Service

type Service struct {
	Name            string `json:"name"`
	SpecificPattern string `json:"specificPattern"`
	CheckInterval   int    `json:"checkInterval"`
	RestartIfDown   bool   `json:"restartIfDown"`
	SaveLogsOnStop  bool   `json:"saveLogsOnStop"`
}

func (s *Service) String() string {
	return fmt.Sprintf("%s", s.Name)
}

func (s *Service) IsRunning() bool {
	cmd := exec.Command("service", s.Name, "status")

	// Create a buffer to store the output of your process
	var out bytes.Buffer

	// Define the process standard output
	cmd.Stdout = &out

	// Run the command
	err := cmd.Run()

	if err != nil {
		// error case : status code of command is different from 0
		logger.Logln("error in ", s.Name, " service check:", err)
	}
	serviceStatus := out.String()
	isActive := strings.Contains(serviceStatus, "Active: active (running)")
	return isActive
}

func (s *Service) Restart() bool {
	cmd := exec.Command("service", s.Name, "restart")

	// Create a buffer to store the output of your process
	var out bytes.Buffer

	// Define the process standard output
	cmd.Stdout = &out

	// Run the command
	err := cmd.Run()
	if err != nil {
		// error case : status code of command is different from 0
		logger.Logln("error in restarting", s.Name, " service:", err)
		return false
	}
	// Restarted successfully
	return true
}

func (s *Service) Logs() string {
	var out bytes.Buffer
	cmd1 := exec.Command("journalctl", "-u", s.Name)
	cmd2 := exec.Command("cat")

	cmd2.Stdin, _ = cmd1.StdoutPipe()
	cmd2.Stdout = &out

	err1 := cmd1.Start()
	if err1 != nil {
		return ""
	}
	err2 := cmd2.Start()
	if err2 != nil {
		return ""
	}

	err3 := cmd1.Wait()
	if err3 != nil {
		return ""
	}
	err4 := cmd2.Wait()
	if err4 != nil {
		return ""
	}

	serviceLogs := out.String()
	return serviceLogs
}

func (s *Service) MarshalText() (text []byte, err error) {
	return []byte(s.String()), nil
}

func (s *Service) SaveLogs() {
	directory := "/var/log/apksm/services/" + s.Name + "/"
	err := os.MkdirAll(directory, os.ModePerm)
	if err != nil {
		return
	}
	file, err := os.Create(directory + s.Name + "-" + time.Now().Format("2006-01-02 15:04:05") + ".log")
	if err != nil {
		return
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {

		}
	}(file)

	_, err = file.WriteString(s.Logs())
	if err != nil {
		return
	}
}
