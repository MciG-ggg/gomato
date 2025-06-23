package setting

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestGetConfigPath 使用表驱动测试测试获取配置文件路径的函数
func TestGetConfigPath(t *testing.T) {
	tests := []struct {
		name           string
		expectedSuffix string
		expectError    bool
	}{
		{
			name:           "正常获取配置文件路径",
			expectedSuffix: filepath.Join(".gomato", "settings", "appsetting.yaml"),
			expectError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path, err := getConfigPath()

			if tt.expectError && err == nil {
				t.Errorf("期望错误但未得到错误")
				return
			}

			if !tt.expectError && err != nil {
				t.Fatalf("getConfigPath() 返回错误: %v", err)
			}

			if !tt.expectError {
				// 检查路径是否包含预期的目录结构
				if !strings.HasSuffix(path, tt.expectedSuffix) {
					t.Errorf("配置文件路径不正确，期望包含 %s，实际得到 %s", tt.expectedSuffix, path)
				}
				t.Logf("配置文件路径: %s", path)
			}
		})
	}
}

// TestEnsureConfigDir 使用表驱动测试测试确保配置目录存在的函数
func TestEnsureConfigDir(t *testing.T) {
	tests := []struct {
		name        string
		expectError bool
	}{
		{
			name:        "正常创建配置目录",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ensureConfigDir()

			if tt.expectError && err == nil {
				t.Errorf("期望错误但未得到错误")
				return
			}

			if !tt.expectError && err != nil {
				t.Fatalf("ensureConfigDir() 返回错误: %v", err)
			}

			if !tt.expectError {
				// 验证目录是否真的被创建了
				homeDir, err := os.UserHomeDir()
				if err != nil {
					t.Fatalf("获取用户主目录失败: %v", err)
				}

				settingsPath := filepath.Join(homeDir, configDir, settingsDir)
				if _, err := os.Stat(settingsPath); os.IsNotExist(err) {
					t.Errorf("配置目录应该存在但未找到: %s", settingsPath)
				}

				t.Logf("配置目录已创建: %s", settingsPath)
			}
		})
	}
}

// TestCreateDefaultSettings 使用表驱动测试测试创建默认设置的函数
func TestCreateDefaultSettings(t *testing.T) {
	tests := []struct {
		name               string
		expectedTimeFormat string
		expectError        bool
	}{
		{
			name:               "创建默认设置",
			expectedTimeFormat: "normal",
			expectError:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 保存原始设置
			originalSetting := appSetting

			// 测试创建默认设置
			err := createDefaultSettings()

			if tt.expectError && err == nil {
				t.Errorf("期望错误但未得到错误")
				return
			}

			if !tt.expectError && err != nil {
				t.Fatalf("createDefaultSettings() 返回错误: %v", err)
			}

			if !tt.expectError {
				// 验证默认设置是否正确
				if appSetting.TimeDisplayFormat != tt.expectedTimeFormat {
					t.Errorf("默认时间显示格式不正确，期望 '%s'，实际得到 '%s'", tt.expectedTimeFormat, appSetting.TimeDisplayFormat)
				}
			}

			// 恢复原始设置
			appSetting = originalSetting

			if !tt.expectError {
				t.Logf("默认设置创建成功: %+v", appSetting)
			}
		})
	}
}

// TestSaveAndLoad 使用表驱动测试测试保存和加载配置的功能
func TestSaveAndLoad(t *testing.T) {
	tests := []struct {
		name               string
		inputSetting       Setting
		expectedTimeFormat string
		expectError        bool
	}{
		{
			name: "保存和加载测试格式",
			inputSetting: Setting{
				TimeDisplayFormat: "test_format",
			},
			expectedTimeFormat: "test_format",
			expectError:        false,
		},
		{
			name: "保存和加载自定义格式",
			inputSetting: Setting{
				TimeDisplayFormat: "custom_format",
			},
			expectedTimeFormat: "custom_format",
			expectError:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 保存原始设置
			originalSetting := appSetting

			// 设置测试数据
			appSetting = tt.inputSetting

			// 测试保存
			err := Save()
			if tt.expectError && err == nil {
				t.Errorf("期望错误但未得到错误")
				return
			}
			if !tt.expectError && err != nil {
				t.Fatalf("Save() 返回错误: %v", err)
			}

			if !tt.expectError {
				// 重置设置
				appSetting = Setting{}

				// 测试加载
				err = Load()
				if err != nil {
					t.Fatalf("Load() 返回错误: %v", err)
				}

				// 验证加载的设置是否正确
				if appSetting.TimeDisplayFormat != tt.expectedTimeFormat {
					t.Errorf("加载的设置不正确，期望 '%s'，实际得到 '%s'", tt.expectedTimeFormat, appSetting.TimeDisplayFormat)
				}
			}

			// 恢复原始设置
			appSetting = originalSetting

			if !tt.expectError {
				t.Logf("保存和加载测试成功")
			}
		})
	}
}

// TestLoadWithNonExistentFile 使用表驱动测试测试加载不存在的配置文件
func TestLoadWithNonExistentFile(t *testing.T) {
	tests := []struct {
		name               string
		expectedTimeFormat string
		expectError        bool
	}{
		{
			name:               "处理不存在的配置文件",
			expectedTimeFormat: "normal",
			expectError:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 保存原始设置
			originalSetting := appSetting

			// 临时备份并删除现有配置文件
			configPath, err := getConfigPath()
			if err != nil {
				t.Fatalf("获取配置文件路径失败: %v", err)
			}

			// 如果配置文件存在，临时重命名它
			if _, err := os.Stat(configPath); err == nil {
				tempPath := configPath + ".backup"
				if err := os.Rename(configPath, tempPath); err != nil {
					t.Fatalf("备份配置文件失败: %v", err)
				}
				defer os.Rename(tempPath, configPath) // 测试结束后恢复
			}

			// 测试加载不存在的文件
			err = Load()
			if tt.expectError && err == nil {
				t.Errorf("期望错误但未得到错误")
				return
			}
			if !tt.expectError && err != nil {
				t.Fatalf("Load() 应该处理不存在的文件，但返回错误: %v", err)
			}

			if !tt.expectError {
				// 验证是否创建了默认设置
				if appSetting.TimeDisplayFormat != tt.expectedTimeFormat {
					t.Errorf("应该创建默认设置，但时间显示格式不正确: %s", appSetting.TimeDisplayFormat)
				}
			}

			// 恢复原始设置
			appSetting = originalSetting

			if !tt.expectError {
				t.Logf("处理不存在配置文件的测试成功")
			}
		})
	}
}

// TestSettingValidation 使用表驱动测试测试设置验证
func TestSettingValidation(t *testing.T) {
	tests := []struct {
		name        string
		setting     Setting
		expectValid bool
	}{
		{
			name: "有效的时间显示格式",
			setting: Setting{
				TimeDisplayFormat: "normal",
			},
			expectValid: true,
		},
		{
			name: "空的时间显示格式",
			setting: Setting{
				TimeDisplayFormat: "",
			},
			expectValid: true, // 空值也是有效的，会使用默认值
		},
		{
			name: "自定义时间显示格式",
			setting: Setting{
				TimeDisplayFormat: "custom_format",
			},
			expectValid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 保存原始设置
			originalSetting := appSetting

			// 设置测试数据
			appSetting = tt.setting

			// 尝试保存设置
			err := Save()
			isValid := err == nil

			if tt.expectValid && !isValid {
				t.Errorf("期望设置有效但保存失败: %v", err)
			}

			if !tt.expectValid && isValid {
				t.Errorf("期望设置无效但保存成功")
			}

			// 恢复原始设置
			appSetting = originalSetting

			if tt.expectValid {
				t.Logf("设置验证通过: %+v", tt.setting)
			}
		})
	}
}

// BenchmarkSave 性能测试：保存配置
func BenchmarkSave(b *testing.B) {
	// 设置测试数据
	appSetting = Setting{
		TimeDisplayFormat: "benchmark_format",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := Save()
		if err != nil {
			b.Fatalf("Save() 返回错误: %v", err)
		}
	}
}

// BenchmarkLoad 性能测试：加载配置
func BenchmarkLoad(b *testing.B) {
	// 确保有配置文件可以加载
	appSetting = Setting{
		TimeDisplayFormat: "benchmark_format",
	}
	if err := Save(); err != nil {
		b.Fatalf("准备测试数据失败: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := Load()
		if err != nil {
			b.Fatalf("Load() 返回错误: %v", err)
		}
	}
}

// TestResetToDefaultSettings 使用表驱动测试测试重置为默认设置的功能
func TestResetToDefaultSettings(t *testing.T) {
	tests := []struct {
		name               string
		initialSetting     Setting
		expectedTimeFormat string
		expectError        bool
	}{
		{
			name: "从自定义设置重置为默认",
			initialSetting: Setting{
				TimeDisplayFormat: "custom_format",
			},
			expectedTimeFormat: "normal",
			expectError:        false,
		},
		{
			name: "从空设置重置为默认",
			initialSetting: Setting{
				TimeDisplayFormat: "",
			},
			expectedTimeFormat: "normal",
			expectError:        false,
		},
		{
			name: "从默认设置重置为默认",
			initialSetting: Setting{
				TimeDisplayFormat: "normal",
			},
			expectedTimeFormat: "normal",
			expectError:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 保存原始设置
			originalSetting := appSetting

			// 设置初始测试数据
			appSetting = tt.initialSetting

			// 测试重置功能
			err := ResetToDefaultSettings()

			if tt.expectError && err == nil {
				t.Errorf("期望错误但未得到错误")
				return
			}

			if !tt.expectError && err != nil {
				t.Fatalf("ResetToDefaultSettings() 返回错误: %v", err)
			}

			if !tt.expectError {
				// 验证设置是否已重置为默认值
				if appSetting.TimeDisplayFormat != tt.expectedTimeFormat {
					t.Errorf("重置后的时间显示格式不正确，期望 '%s'，实际得到 '%s'",
						tt.expectedTimeFormat, appSetting.TimeDisplayFormat)
				}

				// 验证配置文件是否已保存
				configPath, err := getConfigPath()
				if err != nil {
					t.Fatalf("获取配置文件路径失败: %v", err)
				}

				if _, err := os.Stat(configPath); os.IsNotExist(err) {
					t.Errorf("重置后配置文件应该存在但未找到: %s", configPath)
				}
			}

			// 恢复原始设置
			appSetting = originalSetting

			if !tt.expectError {
				t.Logf("重置为默认设置测试成功，重置后设置: %+v", appSetting)
			}
		})
	}
}

// TestResetToDefaultSettingsIntegration 集成测试：测试重置功能与保存/加载的集成
func TestResetToDefaultSettingsIntegration(t *testing.T) {
	// 保存原始设置
	originalSetting := appSetting

	// 设置一个自定义设置
	appSetting = Setting{
		TimeDisplayFormat: "integration_test_format",
	}

	// 保存自定义设置
	if err := Save(); err != nil {
		t.Fatalf("保存自定义设置失败: %v", err)
	}

	// 重置为默认设置
	if err := ResetToDefaultSettings(); err != nil {
		t.Fatalf("重置为默认设置失败: %v", err)
	}

	// 验证内存中的设置已重置
	if appSetting.TimeDisplayFormat != "normal" {
		t.Errorf("内存中的设置未正确重置，期望 'normal'，实际得到 '%s'", appSetting.TimeDisplayFormat)
	}

	// 清空内存中的设置
	appSetting = Setting{}

	// 重新加载设置，验证文件中的设置也已重置
	if err := Load(); err != nil {
		t.Fatalf("重新加载设置失败: %v", err)
	}

	if appSetting.TimeDisplayFormat != "normal" {
		t.Errorf("文件中的设置未正确重置，期望 'normal'，实际得到 '%s'", appSetting.TimeDisplayFormat)
	}

	// 恢复原始设置
	appSetting = originalSetting

	t.Logf("重置功能集成测试成功")
}
