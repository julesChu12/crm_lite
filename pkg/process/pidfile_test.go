package process

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestWritePIDFile(t *testing.T) {
	// 创建临时目录
	tempDir := t.TempDir()

	t.Run("写入PID文件成功", func(t *testing.T) {
		pidFile := filepath.Join(tempDir, "test.pid")

		err := WritePIDFile(pidFile)
		if err != nil {
			t.Fatalf("WritePIDFile() error = %v", err)
		}

		// 验证文件是否创建
		if _, err := os.Stat(pidFile); os.IsNotExist(err) {
			t.Errorf("PID file was not created")
		}

		// 验证文件内容
		content, err := os.ReadFile(pidFile)
		if err != nil {
			t.Fatalf("Failed to read PID file: %v", err)
		}

		pidStr := strings.TrimSpace(string(content))
		pid, err := strconv.Atoi(pidStr)
		if err != nil {
			t.Errorf("Invalid PID in file: %v", err)
		}

		expectedPID := os.Getpid()
		if pid != expectedPID {
			t.Errorf("Expected PID %d, got %d", expectedPID, pid)
		}
	})

	t.Run("使用默认PID文件名", func(t *testing.T) {
		// 更改工作目录到临时目录
		originalDir, _ := os.Getwd()
		defer os.Chdir(originalDir)
		os.Chdir(tempDir)

		err := WritePIDFile("")
		if err != nil {
			t.Fatalf("WritePIDFile() error = %v", err)
		}

		// 验证默认文件是否创建
		if _, err := os.Stat(DefaultPIDFile); os.IsNotExist(err) {
			t.Errorf("Default PID file was not created")
		}

		// 清理
		os.Remove(DefaultPIDFile)
	})

	t.Run("创建嵌套目录", func(t *testing.T) {
		pidFile := filepath.Join(tempDir, "nested", "dir", "test.pid")

		err := WritePIDFile(pidFile)
		if err != nil {
			t.Fatalf("WritePIDFile() error = %v", err)
		}

		// 验证文件是否创建
		if _, err := os.Stat(pidFile); os.IsNotExist(err) {
			t.Errorf("PID file was not created in nested directory")
		}
	})
}

func TestReadPIDFile(t *testing.T) {
	tempDir := t.TempDir()

	t.Run("读取PID文件成功", func(t *testing.T) {
		pidFile := filepath.Join(tempDir, "test.pid")
		expectedPID := 12345

		// 先写入测试数据
		err := os.WriteFile(pidFile, []byte(strconv.Itoa(expectedPID)), 0644)
		if err != nil {
			t.Fatalf("Failed to write test PID file: %v", err)
		}

		pid, err := ReadPIDFile(pidFile)
		if err != nil {
			t.Fatalf("ReadPIDFile() error = %v", err)
		}

		if pid != expectedPID {
			t.Errorf("Expected PID %d, got %d", expectedPID, pid)
		}
	})

	t.Run("使用默认PID文件名", func(t *testing.T) {
		// 更改工作目录到临时目录
		originalDir, _ := os.Getwd()
		defer os.Chdir(originalDir)
		os.Chdir(tempDir)

		expectedPID := 54321
		err := os.WriteFile(DefaultPIDFile, []byte(strconv.Itoa(expectedPID)), 0644)
		if err != nil {
			t.Fatalf("Failed to write test PID file: %v", err)
		}

		pid, err := ReadPIDFile("")
		if err != nil {
			t.Fatalf("ReadPIDFile() error = %v", err)
		}

		if pid != expectedPID {
			t.Errorf("Expected PID %d, got %d", expectedPID, pid)
		}
	})

	t.Run("读取不存在的PID文件", func(t *testing.T) {
		pidFile := filepath.Join(tempDir, "nonexistent.pid")

		_, err := ReadPIDFile(pidFile)
		if err == nil {
			t.Errorf("Expected error when reading nonexistent PID file")
		}
	})

	t.Run("读取无效PID内容", func(t *testing.T) {
		pidFile := filepath.Join(tempDir, "invalid.pid")

		// 写入无效内容
		err := os.WriteFile(pidFile, []byte("invalid_pid"), 0644)
		if err != nil {
			t.Fatalf("Failed to write invalid PID file: %v", err)
		}

		_, err = ReadPIDFile(pidFile)
		if err == nil {
			t.Errorf("Expected error when reading invalid PID content")
		}
	})

	t.Run("读取带空白字符的PID", func(t *testing.T) {
		pidFile := filepath.Join(tempDir, "whitespace.pid")
		expectedPID := 99999

		// 写入带空白字符的PID
		err := os.WriteFile(pidFile, []byte(fmt.Sprintf("  %d  \n", expectedPID)), 0644)
		if err != nil {
			t.Fatalf("Failed to write PID file with whitespace: %v", err)
		}

		pid, err := ReadPIDFile(pidFile)
		if err != nil {
			t.Fatalf("ReadPIDFile() error = %v", err)
		}

		if pid != expectedPID {
			t.Errorf("Expected PID %d, got %d", expectedPID, pid)
		}
	})
}

func TestRemovePIDFile(t *testing.T) {
	tempDir := t.TempDir()

	t.Run("删除存在的PID文件", func(t *testing.T) {
		pidFile := filepath.Join(tempDir, "test.pid")

		// 先创建PID文件
		err := os.WriteFile(pidFile, []byte("12345"), 0644)
		if err != nil {
			t.Fatalf("Failed to create test PID file: %v", err)
		}

		err = RemovePIDFile(pidFile)
		if err != nil {
			t.Fatalf("RemovePIDFile() error = %v", err)
		}

		// 验证文件是否被删除
		if _, err := os.Stat(pidFile); !os.IsNotExist(err) {
			t.Errorf("PID file was not removed")
		}
	})

	t.Run("删除不存在的PID文件", func(t *testing.T) {
		pidFile := filepath.Join(tempDir, "nonexistent.pid")

		// 删除不存在的文件应该不报错
		err := RemovePIDFile(pidFile)
		if err != nil {
			t.Errorf("RemovePIDFile() should not error when file doesn't exist, got: %v", err)
		}
	})

	t.Run("使用默认PID文件名", func(t *testing.T) {
		// 更改工作目录到临时目录
		originalDir, _ := os.Getwd()
		defer os.Chdir(originalDir)
		os.Chdir(tempDir)

		// 创建默认PID文件
		err := os.WriteFile(DefaultPIDFile, []byte("12345"), 0644)
		if err != nil {
			t.Fatalf("Failed to create default PID file: %v", err)
		}

		err = RemovePIDFile("")
		if err != nil {
			t.Fatalf("RemovePIDFile() error = %v", err)
		}

		// 验证文件是否被删除
		if _, err := os.Stat(DefaultPIDFile); !os.IsNotExist(err) {
			t.Errorf("Default PID file was not removed")
		}
	})
}

func TestCleanupPIDFile(t *testing.T) {
	tempDir := t.TempDir()

	t.Run("清理存在的PID文件", func(t *testing.T) {
		pidFile := filepath.Join(tempDir, "test.pid")

		// 先创建PID文件
		err := os.WriteFile(pidFile, []byte("12345"), 0644)
		if err != nil {
			t.Fatalf("Failed to create test PID file: %v", err)
		}

		// CleanupPIDFile 不返回错误，只是清理
		CleanupPIDFile(pidFile)

		// 验证文件是否被删除
		if _, err := os.Stat(pidFile); !os.IsNotExist(err) {
			t.Errorf("PID file was not cleaned up")
		}
	})

	t.Run("清理不存在的PID文件", func(t *testing.T) {
		pidFile := filepath.Join(tempDir, "nonexistent.pid")

		// 清理不存在的文件应该不崩溃
		CleanupPIDFile(pidFile)
		// 没有断言，只要不崩溃就算通过
	})
}

func TestWriteAndReadPIDFile(t *testing.T) {
	tempDir := t.TempDir()
	pidFile := filepath.Join(tempDir, "integration.pid")

	t.Run("写入后读取PID文件", func(t *testing.T) {
		// 写入PID文件
		err := WritePIDFile(pidFile)
		if err != nil {
			t.Fatalf("WritePIDFile() error = %v", err)
		}

		// 读取PID文件
		pid, err := ReadPIDFile(pidFile)
		if err != nil {
			t.Fatalf("ReadPIDFile() error = %v", err)
		}

		// 验证读取的PID是当前进程PID
		expectedPID := os.Getpid()
		if pid != expectedPID {
			t.Errorf("Expected PID %d, got %d", expectedPID, pid)
		}

		// 清理
		CleanupPIDFile(pidFile)

		// 验证文件被清理
		if _, err := os.Stat(pidFile); !os.IsNotExist(err) {
			t.Errorf("PID file was not cleaned up")
		}
	})
}
