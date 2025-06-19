package config

import "time"

type Config struct {
	WorkDuration  time.Duration
	BreakDuration time.Duration
	SoundEnabled  bool
	TaskName      string
}

var DefaultConfig = Config{
	WorkDuration:  25 * time.Minute,
	BreakDuration: 5 * time.Minute,
	SoundEnabled:  true,
	TaskName:      "默认任务",
}

func (c *Config) ToggleSound() {
	c.SoundEnabled = !c.SoundEnabled
}

func (c *Config) SetWork(d time.Duration) {
	c.WorkDuration = d
}

func (c *Config) SetBreak(d time.Duration) {
	c.BreakDuration = d
}
