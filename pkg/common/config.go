package common

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Settings struct {
	Pomodoro        uint   `json:"pomodoro"`
	ShortBreak      uint   `json:"shortBreak"`
	LongBreak       uint   `json:"longBreak"`
	Cycle           uint   `json:"cycle"`
	TimeDisplayMode string `json:"timeDisplayMode"` // "normal" 或 "ansi"
	Language        string `json:"language"`        // "zh" 或 "en"
}

var defaultSettings = Settings{
	Pomodoro:        25,
	ShortBreak:      5,
	LongBreak:       15,
	Cycle:           4,
	TimeDisplayMode: "ansi", // 默认使用ANSI艺术显示
	Language:        "zh",   // 默认中文
}

func getSettingsPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	configDir := filepath.Join(home, ".gomato")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return "", err
	}
	return filepath.Join(configDir, "setting.json"), nil
}

func LoadSettings() (Settings, error) {
	path, err := getSettingsPath()
	if err != nil {
		return defaultSettings, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return defaultSettings, nil // Return defaults if file doesn't exist
		}
		return defaultSettings, err
	}

	if len(data) == 0 {
		return defaultSettings, nil // Return defaults if file is empty
	}

	var s Settings
	if err := json.Unmarshal(data, &s); err != nil {
		return defaultSettings, err
	}
	return s, nil
}

func (s *Settings) Save() error {
	path, err := getSettingsPath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}
