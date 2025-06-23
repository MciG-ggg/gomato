/*
 * Config package for managing application settings.(only for the a config, not for the app)
 * This package provides a structure to hold configuration options
 * such as work duration, break duration, sound settings, and task name.
 * It also includes methods to toggle sound and set durations.
 */
package config

import "time"

type TaskConfig struct {
	WorkDuration  time.Duration
	BreakDuration time.Duration
	TaskName      string
}

var DefaultTaskConfig = TaskConfig{
	WorkDuration:  25 * time.Minute,
	BreakDuration: 5 * time.Minute,
	TaskName:      "默认任务",
}

func (c *TaskConfig) SetWork(d time.Duration) {
	c.WorkDuration = d
}

func (c *TaskConfig) SetBreak(d time.Duration) {
	c.BreakDuration = d
}
