package gomato

import (
	"gomato/pkg/common"
	"testing"
)

func TestNewSettingModelWithSettings(t *testing.T) {
	settings := common.Settings{
		Pomodoro:        25,
		ShortBreak:      5,
		LongBreak:       15,
		Cycle:           4,
		TimeDisplayMode: "ansi",
		Language:        "zh",
	}
	model := NewSettingModelWithSettings(settings)

	if model.Settings.Pomodoro != 25 {
		t.Errorf("Pomodoro expected 25, got %d", model.Settings.Pomodoro)
	}
	if model.timeDisplayIndex != 0 {
		t.Errorf("timeDisplayIndex expected 0, got %d", model.timeDisplayIndex)
	}
	if model.languageIndex != 0 {
		t.Errorf("languageIndex expected 0, got %d", model.languageIndex)
	}
}

func TestNewSettingModelWithSettingsEnglishNormal(t *testing.T) {
	settings := common.Settings{
		Pomodoro:        30,
		ShortBreak:      10,
		LongBreak:       20,
		Cycle:           5,
		TimeDisplayMode: "normal",
		Language:        "en",
	}
	model := NewSettingModelWithSettings(settings)

	if model.Settings.Pomodoro != 30 {
		t.Errorf("Pomodoro expected 30, got %d", model.Settings.Pomodoro)
	}
	if model.timeDisplayIndex != 1 {
		t.Errorf("timeDisplayIndex expected 1, got %d", model.timeDisplayIndex)
	}
	if model.languageIndex != 1 {
		t.Errorf("languageIndex expected 1, got %d", model.languageIndex)
	}
}
