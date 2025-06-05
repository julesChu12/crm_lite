package process

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	DefaultPIDFile = "server.pid"
)

// WritePIDFile 写入当前进程的 PID 到文件
func WritePIDFile(pidFile string) error {
	if pidFile == "" {
		pidFile = DefaultPIDFile
	}

	// 确保目录存在
	dir := filepath.Dir(pidFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory for PID file: %v", err)
	}

	// 写入 PID
	pid := os.Getpid()
	if err := os.WriteFile(pidFile, []byte(fmt.Sprintf("%d", pid)), 0644); err != nil {
		return fmt.Errorf("failed to write PID file: %v", err)
	}

	return nil
}

// ReadPIDFile 从文件读取 PID
func ReadPIDFile(pidFile string) (int, error) {
	if pidFile == "" {
		pidFile = DefaultPIDFile
	}

	// 读取 PID 文件
	content, err := os.ReadFile(pidFile)
	if err != nil {
		return 0, fmt.Errorf("failed to read PID file: %v", err)
	}

	// 解析 PID
	pidStr := strings.TrimSpace(string(content))
	pid, err := strconv.Atoi(pidStr)
	if err != nil {
		return 0, fmt.Errorf("invalid PID in file: %v", err)
	}

	return pid, nil
}

// RemovePIDFile 删除 PID 文件
func RemovePIDFile(pidFile string) error {
	if pidFile == "" {
		pidFile = DefaultPIDFile
	}

	if err := os.Remove(pidFile); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove PID file: %v", err)
	}
	return nil
}

// CleanupPIDFile 清理 PID 文件并记录日志
func CleanupPIDFile(pidFile string) {
	if err := RemovePIDFile(pidFile); err != nil {
		// 如果有日志实例，记录错误
		if log := logger.GetLogger(); log != nil {
			log.Errorf("Failed to remove PID file: %v", err)
		} else {
			fmt.Printf("Warning: Failed to remove PID file: %v\n", err)
		}
	}
}
