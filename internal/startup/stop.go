package startup

import (
	"crm_lite/internal/core/config"
	"crm_lite/pkg/process"
	"fmt"
	"os"
	"syscall"
)

func Stop() {
	opts := config.GetInstance()
	pidFile := opts.Server.PidFile
	if pidFile == "" {
		pidFile = process.DefaultPIDFile
	}

	pid, err := process.ReadPIDFile(pidFile)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Printf("PID file not found, server may not be running: %s\n", pidFile)
			return
		}
		fmt.Printf("Failed to read PID file: %v\n", err)
		os.Exit(1)
	}

	proc, err := os.FindProcess(pid)
	if err != nil {
		fmt.Printf("Failed to find process with PID %d: %v\n", pid, err)
		// Maybe the process is already dead, so we just remove the pidfile
		process.CleanupPIDFile(pidFile)
		os.Exit(1)
	}

	// 向进程发送终止信号
	if err := proc.Signal(syscall.SIGTERM); err != nil {
		fmt.Printf("Failed to send SIGTERM to process %d: %v\n", pid, err)
		// If the process doesn't exist, signal will return an error.
		// In that case, we should probably just clean up the PID file.
		if err.Error() == "os: process not found" {
			fmt.Println("Process not found, removing stale PID file.")
			process.CleanupPIDFile(pidFile)
		}
		os.Exit(1)
	}

	fmt.Printf("Sent SIGTERM to process %d. It may take a few seconds to shut down.\n", pid)
}
