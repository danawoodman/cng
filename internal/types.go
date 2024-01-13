package internal

import (
	"os/exec"
	"time"
)

type WatcherConfig struct {
	Paths   []string
	Command []string
	Initial bool
	Verbose bool
	Kill    bool
	Exclude []string
	Delay   int
}

type Watcher struct {
	config       *WatcherConfig
	cmd          *exec.Cmd
	lastCmdStart time.Time
}
