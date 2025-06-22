package setting

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Setting struct {
	TimeDisplayFormat string `yaml:"time_display_format"` // 时间显示格式
	// Add other settings as needed
}

var (
	// appSetting 是一个全局变量，持有当前应用的配置
	// 程序其他部分可以通过 setting.GetAppSetting 来访问
	appSetting Setting
)

// 配置文件路径常量
const (
	configDir   = ".gomato"
	settingsDir = "settings"
	configFile  = "appsetting.yaml"
)

// getConfigPath 获取配置文件的完整路径
func getConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("获取用户主目录失败: %w", err)
	}

	// 构建完整路径: ~/.gomato/settings/appsetting.yaml
	configPath := filepath.Join(homeDir, configDir, settingsDir, configFile)
	return configPath, nil
}

// ensureConfigDir 确保配置目录存在
func ensureConfigDir() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("获取用户主目录失败: %w", err)
	}

	// 创建 .gomato/settings 目录
	settingsPath := filepath.Join(homeDir, configDir, settingsDir)
	if err := os.MkdirAll(settingsPath, 0755); err != nil {
		return fmt.Errorf("创建配置目录失败: %w", err)
	}

	return nil
}

// GetAppSetting 返回当前应用的设置
func GetAppSetting() Setting {
	// 返回当前应用的设置
	// load first
	// 如果加载失败，返回当前的 appSetting
	err := Load()
	if err != nil {
		fmt.Printf("加载配置失败: %v\n", err)
		return appSetting
	}
	return appSetting
}

// SetAppSetting 更新当前应用的设置
// 这个函数可以在程序运行时修改设置
func SetAppSetting(newSetting Setting) {
	appSetting = newSetting
	// 并且可以在需要时调用 Save() 来保存更改
	if err := Save(); err != nil {
		fmt.Printf("保存配置失败: %v\n", err)
	}
}

func Save() error {
	// 确保配置目录存在
	if err := ensureConfigDir(); err != nil {
		return err
	}

	// 获取配置文件路径
	configPath, err := getConfigPath()
	if err != nil {
		return err
	}

	// 使用 yaml.Marshal 将 AppSetting 结构体转换为 YAML 格式的 []byte
	data, err := yaml.Marshal(&appSetting)
	if err != nil {
		return fmt.Errorf("转换配置到 YAML 格式失败: %w", err)
	}

	// 将数据写入文件，0644 是一个标准的文件权限
	err = os.WriteFile(configPath, data, 0644)
	if err != nil {
		return fmt.Errorf("保存配置文件失败: %w", err)
	}

	return nil
}

func Load() error {
	// 获取配置文件路径
	configPath, err := getConfigPath()
	if err != nil {
		return err
	}

	// 尝试读取配置文件
	data, err := os.ReadFile(configPath)

	// 检查读取时发生的错误
	if err != nil {
		// 如果错误是因为文件不存在
		if os.IsNotExist(err) {
			fmt.Println("配置文件未找到，将创建默认配置。")
			// 调用 createDefaultSettings 来初始化并保存一份默认配置
			return createDefaultSettings()
		}
		// 如果是其他读取错误，则直接返回错误
		return fmt.Errorf("读取配置文件失败: %w", err)
	}

	// 如果文件成功读取，使用 yaml.Unmarshal 将内容解析到 AppSetting 变量中
	err = yaml.Unmarshal(data, &appSetting)
	if err != nil {
		return fmt.Errorf("解析配置文件失败: %w", err)
	}

	return nil
}

func createDefaultSettings() error {
	// 在这里定义你的默认设置
	appSetting = Setting{
		TimeDisplayFormat: "normal",
		// 可以在这里添加更多默认设置
		// Language: "zh-cn",
	}

	// 调用 Save() 将这些默认设置写入文件
	return Save()
}

func ResetToDefaultSettings() error {
	// 创建默认设置
	if err := createDefaultSettings(); err != nil {
		return fmt.Errorf("重置为默认设置失败: %w", err)
	}

	fmt.Println("已重置为默认设置。")
	return nil
}
