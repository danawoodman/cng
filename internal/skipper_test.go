package internal_test

import (
	"path/filepath"
	"testing"

	"github.com/danawoodman/cng/internal"
	"github.com/stretchr/testify/assert"
)

func TestSkipper(t *testing.T) {
	tests := []struct {
		name       string
		workDir    string
		exclusions []string
		expected   map[string]bool
	}{
		{
			name:       "ignore common directories",
			workDir:    "/foo",
			exclusions: []string{"**/*.txt"},
			expected: map[string]bool{
				"/foo/.git/test.txt":             true,
				"/foo/bar/.git/test.txt":         true,
				"/foo/node_modules/test.txt":     true,
				"/foo/bar/node_modules/test.txt": true,
			},
		},
		{
			name:       "absolute paths",
			workDir:    "/foo",
			exclusions: []string{"**/*.txt"},
			expected: map[string]bool{
				"/foo/bar/baz/test.txt": true,
				"/biz/bang/bop/test.md": false,
			},
		},
		{
			name:       "relative paths",
			workDir:    "/foo/bar",
			exclusions: []string{"**/*.txt"},
			expected: map[string]bool{
				"/test.txt":         false,
				"/foo/test.txt":     false,
				"/foo/bar/test.txt": true,
				"/test.md":          false,
				"/foo/test.md":      false,
				"/foo/bar/test.md":  false,
			},
		},
		{
			name:       "current dir",
			workDir:    "/foo",
			exclusions: []string{"*.txt"},
			expected: map[string]bool{
				"/foo/test.txt": true,
				"/foo/foo.txt":  true,
				"/test.txt":     false,
				"/test.md":      false,
				"/foo/test.md":  false,
			},
		},
		{
			name:       "fragment pattern",
			workDir:    "/",
			exclusions: []string{"foo_*.txt"},
			expected: map[string]bool{
				"/test.txt":    false,
				"/foo.txt":     false,
				"/foo_bar.txt": true,
				"/foo_1.txt":   true,
			},
		},
		{
			name:       "optional ending",
			workDir:    "/foo",
			exclusions: []string{"*.{js,jsx}"},
			expected: map[string]bool{
				"/foo/a.js":   true,
				"/foo/b.jsx":  true,
				"/foo/c.j":    false,
				"/foo/c.jsxx": false,
			},
		},
		// todo: test windows \ paths
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			var exclusions []string
			for _, e := range test.exclusions {
				exclusions = append(exclusions, filepath.Join(test.workDir, e))
			}
			s := internal.NewSkipper(test.workDir, exclusions)
			for path, val := range test.expected {
				p := filepath.Join(test.workDir, path)
				assert.Equal(t, val, s.ShouldExclude(p), "exclude patterns %v should skip path '%s' but didn't", test.exclusions, path)
			}
		})
	}
}
