package test

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCng(t *testing.T) {
	tests := []struct {
		name, stdout, stderr, exclude, pattern string
		verbose, kill, init, skip              bool
		delay                                  int
		steps                                  func(write func(string))
	}{
		{
			name:    "adds newly created files",
			pattern: "*.txt",
			steps: func(write func(string)) {
				for range 3 {
					write("*.txt")
				}
			},
			stdout: "hello\nhello\nhello\n",
		},
		{
			name:    "init flag: should run on startup",
			pattern: "*.txt",
			stdout:  "hello\n",
			init:    true,
		},
		{
			name:    "exclude flag: ignores excluded files",
			pattern: "*.txt",
			exclude: "*.md",
			steps: func(write func(string)) {
				write("*.txt")

				// should not be picked up by the watcher
				write("*.md")
			},
			stdout: "hello\n",
		},
		// todo: should report helpful error if missing pattern
		// {
		// 	name: "adds renamed files",
		// },
		// {
		// 	name: "flag: delay execution",
		// },
		// {
		// 	name: "stops watching deleted files",
		// },
		// todo: can ignore based on glob
		// todo: ignore files in node_modules / ,git by default
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			if test.skip {
				t.Skip()
			}

			dir := t.TempDir()

			var stdoutBuf, stderrBuf bytes.Buffer
			conf := conf{
				dir:     dir,
				pattern: test.pattern,
				exclude: test.exclude,
				init:    test.init,
				verbose: test.verbose,
				kill:    test.kill,
				delay:   test.delay,
			}
			// t.Logf("CONF: %+v", conf)
			cmd := command(t, &stdoutBuf, &stderrBuf, conf)
			err := cmd.Start()
			assert.NoError(t, err)

			// wait for the process to start
			// anything less than about 100ms and the process won't have time to start
			time.Sleep(200 * time.Millisecond)

			if test.steps != nil {
				test.steps(func(path string) {
					write(t, dir, path, "content", 500)
				})
			}

			time.Sleep(100 * time.Millisecond)

			// Send SIGINT to the process
			err = cmd.Process.Signal(os.Interrupt)
			assert.NoError(t, err)

			// Wait for the process to exit
			if err := cmd.Wait(); err != nil {
				exiterr, ok := err.(*exec.ExitError)
				// assert.NoError(t, err)
				assert.True(t, ok)
				assert.NotNil(t, exiterr)

				status, ok := exiterr.Sys().(syscall.WaitStatus)
				assert.NoError(t, err)
				assert.True(t, ok && status.Signaled() && status.Signal() == syscall.SIGINT)
			}

			// Read and assert on the process's output
			stdout := stdoutBuf.String()
			stderr := stderrBuf.String()

			assert.Equal(t, test.stdout, stdout)
			assert.Equal(t, test.stderr, stderr)
		})
	}
}

type conf struct {
	dir, exclude, pattern string
	verbose, kill, init   bool
	delay                 int
}

func write(t *testing.T, dir, path, content string, waitMs int) /**os.File*/ {
	t.Helper()
	var f *os.File

	// Create the file if it doesn't exist yet:
	_, err := os.Stat(filepath.Join(dir, path))
	if os.IsNotExist(err) {
		f, err = os.CreateTemp(dir, path)
		assert.NoError(t, err)
	} else {
		f, err = os.OpenFile(filepath.Join(dir, path), os.O_RDWR, 0o644)
		assert.NoError(t, err)
	}

	// write to the file:
	_, err = f.Write([]byte(content))
	assert.NoError(t, err)

	f.Close()

	wait(waitMs)

	// return f
}

// command returns a new exec.Cmd for running cng with the given configuration.
func command(t *testing.T, stdout, stderr io.Writer, conf conf) *exec.Cmd {
	// t.Helper()
	parts := []string{}
	if conf.init {
		parts = append(parts, "-i")
	}
	if conf.verbose {
		parts = append(parts, "-v")
	}
	if conf.kill {
		parts = append(parts, "-k")
	}
	if conf.exclude != "" {
		parts = append(parts, "-e", fmt.Sprintf("%s/%s", conf.dir, conf.exclude))
	}
	parts = append(parts, fmt.Sprintf("%s/%s", conf.dir, conf.pattern))
	parts = append(parts, "--", "echo", "hello")
	cmd := exec.Command("cng", parts...)
	t.Log("CMD:", cmd.String())
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	return cmd
}

func wait(ms int) {
	time.Sleep(time.Duration(ms) * time.Millisecond)
}
