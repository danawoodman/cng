package domain

import (
	"fmt"
	"path/filepath"

	"github.com/bmatcuk/doublestar/v4"
)

type Skipper interface {
	ShouldExclude(path string) bool
}

type skipper struct {
	workDir string
	exclude []string
}

func NewSkipper(workDir string, exclude []string) Skipper {
	return skipper{workDir: workDir, exclude: exclude}
}

// shouldExclude returns true if the given path should be excluded from
// triggering a command run based on the `-e, --exclude` flag.
func (s skipper) ShouldExclude(path string) bool {
	// skip common things like .git and node_modules dirs
	// todo: make this configurable
	ignores := []string{".git", "node_modules"}
	for _, ignore := range ignores {
		if matches, _ := doublestar.PathMatch(fmt.Sprintf("**/%s/**", ignore), path); matches {
			return true
		}
	}

	// fmt.Println("EXCLUDE:", s.exclude)
	// fmt.Println("PATH:", path)

	// check if the path matches any of the exclude patterns
	for _, pattern := range s.exclude {
		// s.log("Checking exclude pattern", "pattern", pattern, "path", path)
		expandedPattern := filepath.Join(s.workDir, pattern)
		// fmt.Println("EXPANDED:", expandedPattern)
		if matches, _ := doublestar.PathMatch(expandedPattern, path); matches {
			// s.log("File in exclude path, skipping", "exclude", pattern)
			return true
		}
	}

	return false
}
