package main

import (
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/charmbracelet/log"
	"github.com/fsnotify/fsnotify"
)

var logger = log.NewWithOptions(os.Stderr, log.Options{
	Prefix: "gochange",
})

func NewWatcher(config *WatcherConfig) *Watcher {
	if config.Verbose {
		logger.Info("Starting watcher with config:", "config", config)
	}
	return &Watcher{
		config: config,
	}
}

func (w *Watcher) Start() {

	w.log("command to run:", "cmd", w.config.Command)
	w.log("watched paths:", "paths", w.config.Paths)

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		w.exit("Could not start fsnotify process", err)
	}
	defer watcher.Close()

	for _, pattern := range w.config.Paths {
		matches, err := doublestar.FilepathGlob(pattern)
		if err != nil {
			w.exit("Could not watch glob pattern", pattern, " error: ", err)
		}
		for _, path := range matches {
			w.log("Watching", "path", path)
			if err := watcher.Add(path); err != nil {
				w.exit("Could not watch path", path, " error:", err)
			}
		}
	}

	if w.config.Initial {
		w.log("Starting initial run...")
		w.runCmd()
	}

	// Watch for a SIGINT signal and call .kill on the current command
	// process if we receive one:
	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, os.Interrupt)
		<-sig
		w.log("Received SIGINT, exiting...")
		if w.cmd != nil {
			w.kill()
		}
		os.Exit(0)
	}()

	for {
		select {
		case event := <-watcher.Events:
			w.log("Detected fsnotify event", "op", event.Op, "name", event.Name)

			// TODO: make this configurable using -f
			if event.Op&fsnotify.Chmod == fsnotify.Chmod {
				w.log("CHMOD change detected, skipping")
				continue
			}

			if w.shouldExclude(event.Name) {
				continue
			}

			if time.Since(w.lastCmdStart) < time.Duration(w.config.Delay)*time.Millisecond {
				w.log("Last command started less than the configured delay, skipping", "delay", w.config.Delay)
				continue
			}

			if w.cmd != nil && w.config.Kill {
				w.kill()
			}

			w.runCmd()
		}
	}

}

func (w *Watcher) runCmd() {
	cmd := exec.Command(w.config.Command[0], w.config.Command[1:]...)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	// Attach to current process so we can get color output:
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	w.log("Running command process", "cmd", cmd)
	cmd.Start()

	w.lastCmdStart = time.Now()

	w.cmd = cmd
}

// kill kills the current command process.
// It sends a SIGINT signal to the process group of cmd.
// We cannot simply call cmd.Process.Kill() because it will not kill
// the child processes of cmd which, in the case of something like a
// web server, would mean that we can't re-bind to the given port.
// We then wait for the task to exit cleanly before continuing.
func (w *Watcher) kill() {
	pgid, err := syscall.Getpgid(w.cmd.Process.Pid)
	if err == nil {
		syscall.Kill(-pgid, syscall.SIGINT)
	}

	w.cmd.Wait()
}

// shouldExclude returns true if the given path should be excluded from
// triggering a command run based on the `-e, --exclude` flag.
func (w *Watcher) shouldExclude(path string) bool {
	skip := false

	for _, excl := range w.config.Exclude {
		if matched, _ := filepath.Match(excl, "./"+path); matched {
			w.log("File in exclude path, skipping", "exclude", excl)
			skip = true
			break
		}
	}

	return skip
}

// exit logs a fatal message and exits the program because of
// some invalid condition.
func (w *Watcher) exit(msg string, args ...interface{}) {
	logger.Fatal(msg, args...)
}

// log logs a message if verbose mode is enabled.
func (w *Watcher) log(msg string, args ...interface{}) {
	if w.config.Verbose {
		logger.Info(msg, args...)
	}
}
