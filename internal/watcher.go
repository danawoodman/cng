package internal

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

// todo: inject logging using WithLogger
var logger = log.NewWithOptions(os.Stderr, log.Options{
	Prefix: "cng",
})

type WatcherConfig struct {
	ExcludePaths []string
	Command      []string
	Initial      bool
	Verbose      bool
	Kill         bool
	Exclude      []string
	Delay        int
}

type Watcher struct {
	config       WatcherConfig
	cmd          *exec.Cmd
	lastCmdStart time.Time
	log          func(msg string, args ...interface{})
	skipper      Skipper
	workDir      string
}

func NewWatcher(config WatcherConfig) Watcher {
	if config.Verbose {
		logger.Info("Starting watcher with config:", "config", config)
	}

	workDir, err := os.Getwd()
	if err != nil {
		panic("Could not get working directory: " + err.Error())
	}

	return Watcher{
		config: config,
		// todo: make this injectable
		workDir: workDir,
		skipper: NewSkipper(workDir, config.Exclude),
		log: func(msg string, args ...interface{}) {
			if config.Verbose {
				logger.Info(msg, args...)
			}
		},
	}
}

func (w Watcher) Start() {
	w.log("Command to run:", "cmd", w.config.Command)
	w.log("Watched paths:", "paths", w.config.ExcludePaths)

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		w.exit("Could not start fsnotify process", err)
	}
	defer watcher.Close()

	w.log("Adding watched paths:", "paths", w.config.ExcludePaths)
	for _, pattern := range w.config.ExcludePaths {
		expandedPath := filepath.Join(w.workDir, pattern)
		rootDir, _ := doublestar.SplitPattern(expandedPath)

		if err := watcher.Add(rootDir); err != nil {
			w.exit("Could not watch root directory:", "dir", rootDir, " error:", err)
		}

		matches, err := doublestar.FilepathGlob(expandedPath)
		if err != nil {
			w.exit("Could not watch glob pattern", pattern, " error: ", err)
		}
		w.log("Glob matches", "pattern", pattern, "matches", matches)

		for _, path := range matches {
			w.log("Watching", "path", path)

			if err := watcher.Add(path); err != nil {
				w.exit("Could not watch path", path, " error:", err)
			}
			// get root dir of path:
			dir := filepath.Dir(path)
			if dir != "" {
				w.addFiles(watcher, dir)
			}

			// fileInfo, err := os.Stat(path)
			// if err != nil {
			// 	w.exit("Could not watch path", path, " error:", err)
			// }
			// if fileInfo.IsDir() {
			// 	w.addFiles(watcher, path)
			// }
		}
	}

	if w.config.Initial {
		w.log("Starting initial run...")
		w.runCmd()
	}

	// Watch for a SIGINT signal and call .kill on the current command
	// process if we receive one:
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)

	w.log("Current watchlist:", "watchlist", watcher.WatchList())

	done := make(chan bool)
	go func() {
		for {
			select {
			case event := <-watcher.Events:
				w.log("Detected fsnotify event", "op", event.Op, "name", event.Name)

				if w.skipper.ShouldExclude(event.Name) {
					w.log("File in exclude path, skipping", "path", event.Name)
					continue
				}

				// todo: make this configurable using -f
				if event.Op&fsnotify.Chmod == fsnotify.Chmod {
					w.log("CHMOD change detected, skipping")
					continue
				}

				if time.Since(w.lastCmdStart) < time.Duration(w.config.Delay)*time.Millisecond {
					w.log("Last command started less than the configured delay, skipping", "delay", w.config.Delay)
					continue
				}

				if event.Op&fsnotify.Create == fsnotify.Create { // || event.Op&fsnotify.Remove == fsnotify.Remove {
					// Attempt to remove in case it's deleted
					// watcher.Remove(event.Name)

					// Check if the created item is a directory; if so, add it and its contents to the watcher
					w.log("File created, adding to watcher", "path", event.Name)
					fileInfo, err := os.Stat(event.Name)
					if err == nil {
						if fileInfo.IsDir() {
							w.addFiles(watcher, event.Name) // Add the new directory and its contents
						} else {
							// It's a file, add directly
							watcher.Add(event.Name)
						}
					}
				}

				if w.cmd != nil && w.config.Kill {
					w.kill()
				}

				w.runCmd()

			case <-sig:
				w.log("Received SIGINT, exiting...")
				w.kill()
				os.Exit(0)
			}
		}
	}()

	<-done
}

func (w Watcher) runCmd() {
	w.log("Running command...")
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
func (w Watcher) kill() {
	if w.cmd == nil {
		return
	}
	pgid, err := syscall.Getpgid(w.cmd.Process.Pid)
	if err == nil {
		syscall.Kill(-pgid, syscall.SIGINT)
	}
	w.log("Killing current command process:", "cmd", pgid)
	w.cmd.Wait()
}

// exit logs a fatal message and exits the program because of
// some invalid condition.
func (w Watcher) exit(msg string, args ...interface{}) {
	logger.Fatal(msg, args...)
}

func (w Watcher) addFiles(watcher *fsnotify.Watcher, rootPath string) {
	w.log("Adding files in directory to watcher", "path", rootPath)
	err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err // Handle the error
		}
		if !info.IsDir() {
			return nil // Skip directories
		}
		// Add the directory to the watcher
		err = watcher.Add(path)
		if err != nil {
			w.exit("Could not watch path", path, " error:", err) // Adjust to use the watcher's logging method
		}
		return nil
	})
	if err != nil {
		w.exit("Error walking the path", rootPath, " error:", err) // Adjust to use the watcher's logging method
	}
}
