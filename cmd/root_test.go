package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestFindConfigInProject 测试 findConfigInProject 方法
func TestFindConfigInProject(t *testing.T) {
	// 保存当前工作目录
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current working directory: %v", err)
	}
	defer os.Chdir(originalWd)

	t.Run("在真实项目中查找存在的配置文件", func(t *testing.T) {
		// 切换到项目根目录
		if err := os.Chdir(".."); err != nil {
			t.Fatalf("Failed to change to parent directory: %v", err)
		}

		// 测试查找存在的配置文件
		result, err := findConfigInProject("app.dev.yaml")
		if err != nil {
			t.Errorf("Expected no error, but got: %v", err)
		}

		// 验证返回的路径包含正确的配置文件名
		expectedSuffix := filepath.Join("config", "app.dev.yaml")
		if !filepath.IsAbs(result) {
			t.Errorf("Expected absolute path, but got: %s", result)
		}
		if !strings.HasSuffix(result, expectedSuffix) {
			t.Errorf("Expected path to end with %s, but got: %s", expectedSuffix, result)
		}

		// 验证文件确实存在
		if _, err := os.Stat(result); os.IsNotExist(err) {
			t.Errorf("Config file should exist at: %s", result)
		}
	})

	t.Run("查找不存在的配置文件", func(t *testing.T) {
		// 切换到项目根目录
		if err := os.Chdir(".."); err != nil {
			t.Fatalf("Failed to change to parent directory: %v", err)
		}

		// 测试查找不存在的配置文件
		_, err := findConfigInProject("nonexistent.yaml")
		if err == nil {
			t.Error("Expected error for nonexistent config file, but got none")
		}
	})

	t.Run("在临时目录中测试错误情况", func(t *testing.T) {
		// 创建临时目录并切换到其中（没有 go.mod 文件）
		tempDir := t.TempDir()
		if err := os.Chdir(tempDir); err != nil {
			t.Fatalf("Failed to change to temp directory: %v", err)
		}

		// 测试找不到 go.mod 的情况
		_, err := findConfigInProject("app.dev.yaml")
		if err == nil {
			t.Error("Expected error when go.mod not found, but got none")
		}

		expectedErrMsg := "cannot find go.mod in current directory or any parent"
		if err.Error() != expectedErrMsg {
			t.Errorf("Expected error message '%s', but got: %s", expectedErrMsg, err.Error())
		}
	})

	t.Run("空配置文件名", func(t *testing.T) {
		// 切换到项目根目录
		if err := os.Chdir(originalWd); err != nil {
			t.Fatalf("Failed to change to original directory: %v", err)
		}
		if err := os.Chdir(".."); err != nil {
			t.Fatalf("Failed to change to parent directory: %v", err)
		}

		// 测试空配置文件名 - 空字符串实际上会匹配到 config 目录本身
		// 这不是一个有效的配置文件，但由于目录存在，不会在 os.Stat 阶段失败
		// 这个测试更多是验证函数的行为而不是期望错误
		result, err := findConfigInProject("")
		if err != nil {
			// 如果返回错误，也是可以接受的
			t.Logf("Empty config name returned error (expected): %v", err)
		} else {
			// 如果没有错误，验证返回的路径是否指向 config 目录
			if !strings.HasSuffix(result, "config") {
				t.Errorf("Expected empty config name to resolve to config directory, but got: %s", result)
			}
		}
	})
}

// TestFindConfigInProjectWithMockFileSystem 测试模拟文件系统场景
func TestFindConfigInProjectWithMockFileSystem(t *testing.T) {
	// 保存当前工作目录
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current working directory: %v", err)
	}
	defer os.Chdir(originalWd)

	t.Run("模拟项目结构测试", func(t *testing.T) {
		// 创建临时目录结构
		tempDir := t.TempDir()
		projectRoot := filepath.Join(tempDir, "test_project")
		configDir := filepath.Join(projectRoot, "config")
		subDir := filepath.Join(projectRoot, "subdir")

		// 创建目录
		if err := os.MkdirAll(configDir, 0755); err != nil {
			t.Fatalf("Failed to create config directory: %v", err)
		}
		if err := os.MkdirAll(subDir, 0755); err != nil {
			t.Fatalf("Failed to create sub directory: %v", err)
		}

		// 创建 go.mod 文件
		goModPath := filepath.Join(projectRoot, "go.mod")
		if err := os.WriteFile(goModPath, []byte("module test_project\n\ngo 1.21\n"), 0644); err != nil {
			t.Fatalf("Failed to create go.mod: %v", err)
		}

		// 创建配置文件
		configPath := filepath.Join(configDir, "app.test.yaml")
		configContent := `
app:
  name: test
  debug: true
`
		if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
			t.Fatalf("Failed to create config file: %v", err)
		}

		// 切换到子目录进行测试
		if err := os.Chdir(subDir); err != nil {
			t.Fatalf("Failed to change directory: %v", err)
		}

		// 测试从子目录查找配置文件
		result, err := findConfigInProject("app.test.yaml")
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		// 使用 filepath.EvalSymlinks 解析符号链接，然后比较路径
		expectedPath, err := filepath.EvalSymlinks(configPath)
		if err != nil {
			expectedPath = configPath
		}
		resultPath, err := filepath.EvalSymlinks(result)
		if err != nil {
			resultPath = result
		}

		if resultPath != expectedPath {
			t.Errorf("Expected path %s, but got %s", expectedPath, resultPath)
		}

		// 验证文件确实存在
		if _, err := os.Stat(result); os.IsNotExist(err) {
			t.Errorf("Config file should exist at: %s", result)
		}
	})
}

// BenchmarkFindConfigInProject 性能基准测试
func BenchmarkFindConfigInProject(b *testing.B) {
	// 保存当前工作目录
	originalWd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get current working directory: %v", err)
	}
	defer os.Chdir(originalWd)

	// 切换到项目根目录
	if err := os.Chdir(".."); err != nil {
		b.Fatalf("Failed to change to parent directory: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := findConfigInProject("app.dev.yaml")
		if err != nil {
			b.Errorf("Unexpected error: %v", err)
		}
	}
}
