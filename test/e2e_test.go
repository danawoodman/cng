package test

import (
	"bytes"
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
		{
			name:    "ignores default excluded dirs",
			pattern: "*.txt",
			exclude: "*.md",
			steps: func(write func(string)) {
				write(".git/foo.txt")
				write("node_modules/foo.txt")
			},
			stdout: "",
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
			// t.Parallel()
			if test.skip {
				t.Skip()
			}

			dir := t.TempDir()
			// fmt.Println("TEMP DIR:", dir)
			err := os.Chdir(dir)
			assert.NoError(t, err)
			// wd, err := os.Getwd()
			// assert.NoError(t, err)
			// fmt.Println("WORK DIR:", wd)

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
			err = cmd.Start()
			assert.NoError(t, err)

			// wait for the process to start
			// anything less than about 100ms and the process won't have time to start
			time.Sleep(300 * time.Millisecond)

			if test.steps != nil {
				test.steps(func(path string) {
					write(t, dir, path, "content", 500)
				})
			}

			// List all files in dir:
			// files, err := os.ReadDir(dir)
			// assert.NoError(t, err)
			// for _, file := range files {
			// 	fmt.Println("FILE:", file.Name())
			// }

			time.Sleep(300 * time.Millisecond)

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

	fullPath := filepath.Join(dir, path)
	fullDir := filepath.Dir(fullPath)
	fileName := filepath.Base(fullPath)

	// Create the file if it doesn't exist yet:
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		// recursively create the file's parent directories:
		err = os.MkdirAll(fullDir, 0o755)
		assert.NoError(t, err)
		f, err = os.CreateTemp(fullDir, fileName)
		assert.NoError(t, err)
	} else {
		f, err = os.OpenFile(fullPath, os.O_RDWR, 0o644)
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
		parts = append(parts, "-e", conf.exclude)
	}
	parts = append(parts, conf.pattern)
	parts = append(parts, "--", "echo", "hello")
	cmd := exec.Command("cng", parts...)
	// t.Log("CMD:", cmd.String())
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	return cmd
}

func wait(ms int) {
	time.Sleep(time.Duration(ms) * time.Millisecond)
}
