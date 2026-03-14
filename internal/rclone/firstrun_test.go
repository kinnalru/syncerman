package rclone

import (
	"testing"
)

func TestIsFirstRunError(t *testing.T) {
	tests := []struct {
		name     string
		stderr   string
		expected bool
	}{
		{
			name:     "standard first-run error",
			stderr:   "ERROR : Bisync critical error: cannot find prior Path1 or Path2 listings, likely due to critical error on prior run\nTip: here are the filenames we were looking for. Do they exist?\n",
			expected: true,
		},
		{
			name:     "overall.md example",
			stderr:   "2026/03/14 20:14:03 ERROR : Bisync critical error: cannot find prior Path1 or Path2 listings, likely due to critical error on prior run \nTip: here are the filenames we were looking for. Do they exist? \nPath1: /home/jerry/.cache/rclone/bisync/tmp_file.path1.lst\nPath2: /home/jerry/.cache/rclone/bisync/tmp_file.path2.lst\n",
			expected: true,
		},
		{
			name:     "case insensitive match",
			stderr:   "Cannot find prior PATH1 or PATH2 listings... HERE ARE THE FILENAMES",
			expected: true,
		},
		{
			name:     "partial match - no filenames",
			stderr:   "ERROR: cannot find prior Path1 or Path2 listings",
			expected: false,
		},
		{
			name:     "partial match - no listings message",
			stderr:   "ERROR: here are the filenames we were looking for",
			expected: false,
		},
		{
			name:     "other error",
			stderr:   "ERROR: permission denied",
			expected: false,
		},
		{
			name:     "empty stderr",
			stderr:   "",
			expected: false,
		},
		{
			name:     "network error",
			stderr:   "ERROR: network timeout",
			expected: false,
		},
		{
			name:     "missing but not first-run",
			stderr:   "ERROR: cannot find prior files",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsFirstRunError(tt.stderr)
			if result != tt.expected {
				t.Errorf("IsFirstRunError() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestExtractFirstRunErrorPaths(t *testing.T) {
	tests := []struct {
		name     string
		stderr   string
		expected []string
	}{
		{
			name:   "full first-run error",
			stderr: "ERROR: cannot find prior Path1 or Path2 listings...here are the filenames we were looking for. Do they exist? \nPath1: /home/user/.cache/rclone/bisync/tmp_file.path1.lst\nPath2: /home/user/.cache/rclone/bisync/tmp_file.path2.lst\n",
			expected: []string{
				"/home/user/.cache/rclone/bisync/tmp_file.path1.lst",
				"/home/user/.cache/rclone/bisync/tmp_file.path2.lst",
			},
		},
		{
			name:   "overall.md example",
			stderr: "2026/03/14 20:14:03 ERROR : Bisync critical error: cannot find prior Path1 or Path2 listings, likely due to critical error on prior run \nTip: here are the filenames we were looking for. Do they exist? \nPath1: /home/jerry/.cache/rclone/bisync/tmp_tools..kinnalru@yandex.ru_tools.path1.lst\nPath2: /home/jerry/.cache/rclone/bisync/tmp_tools..kinnalru@yandex.ru_tools.path2.lst\n",
			expected: []string{
				"/home/jerry/.cache/rclone/bisync/tmp_tools..kinnalru@yandex.ru_tools.path1.lst",
				"/home/jerry/.cache/rclone/bisync/tmp_tools..kinnalru@yandex.ru_tools.path2.lst",
			},
		},
		{
			name:     "without pattern match",
			stderr:   "ERROR: permission denied",
			expected: []string{},
		},
		{
			name:     "empty stderr",
			stderr:   "",
			expected: []string{},
		},
		{
			name:   "only path1",
			stderr: "ERROR: cannot find prior Path1 or Path2 listings...here are the filenames... Path1: /tmp/path1.lst\n",
			expected: []string{
				"/tmp/path1.lst",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			paths := ExtractFirstRunErrorPaths(tt.stderr)

			if len(paths) != len(tt.expected) {
				t.Errorf("ExtractFirstRunErrorPaths() len = %d, want %d", len(paths), len(tt.expected))
				return
			}

			for i, path := range paths {
				if i >= len(tt.expected) || path != tt.expected[i] {
					t.Errorf("ExtractFirstRunErrorPaths()[%d] = %q, want %q", i, path, tt.expected[i])
				}
			}
		})
	}
}

func TestParseFirstRunError(t *testing.T) {
	tests := []struct {
		name      string
		stderr    string
		wantNil   bool
		wantPaths []string
	}{
		{
			name:      "valid first-run error",
			stderr:    "ERROR: cannot find prior Path1 or Path2 listings...here are the filenames...",
			wantNil:   false,
			wantPaths: []string{},
		},
		{
			name:      "other error",
			stderr:    "ERROR: permission denied",
			wantNil:   true,
			wantPaths: []string{},
		},
		{
			name:      "empty",
			stderr:    "",
			wantNil:   true,
			wantPaths: []string{},
		},
		{
			name:    "full first-run error with paths",
			stderr:  "ERROR: cannot find prior Path1 or Path2 listings...here are the filenames... Path1: /tmp/p1.lst\nPath2: /tmp/p2.lst\n",
			wantNil: false,
			wantPaths: []string{
				"/tmp/p1.lst",
				"/tmp/p2.lst",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ParseFirstRunError(tt.stderr)

			if tt.wantNil {
				if err != nil {
					t.Errorf("ParseFirstRunError() = %v, want nil", err)
				}
				return
			}

			if err == nil {
				t.Errorf("ParseFirstRunError() = nil, want non-nil")
				return
			}

			if err.Message != tt.stderr {
				t.Errorf("ParseFirstRunError().Message = %q, want %q", err.Message, tt.stderr)
			}

			if len(err.Paths) != len(tt.wantPaths) {
				t.Errorf("ParseFirstRunError().Paths len = %d, want %d", len(err.Paths), len(tt.wantPaths))
				return
			}

			for i, path := range err.Paths {
				if path != tt.wantPaths[i] {
					t.Errorf("ParseFirstRunError().Paths[%d] = %q, want %q", i, path, tt.wantPaths[i])
				}
			}
		})
	}
}

func TestFirstRunError_Pattern(t *testing.T) {
	tests := []struct {
		name      string
		text      string
		listings  bool
		filenames bool
	}{
		{
			name:      "standard format",
			text:      "cannot find prior Path1 or Path2 listings, here are the filenames",
			listings:  true,
			filenames: true,
		},
		{
			name:      "case variations",
			text:      "CANNOT FIND PRIOR path1 OR path2 listings... HERE ARE THE FILENAMES",
			listings:  true,
			filenames: true,
		},
		{
			name:      "missing first part",
			text:      "here are the filenames only",
			listings:  false,
			filenames: true,
		},
		{
			name:      "missing second part",
			text:      "cannot find prior Path1 or Path2 listings only",
			listings:  true,
			filenames: false,
		},
		{
			name:      "completely different",
			text:      "some random error message",
			listings:  false,
			filenames: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			listings := FirstRunErrorPatternListings.MatchString(tt.text)
			filenames := FirstRunErrorPatternFilenames.MatchString(tt.text)

			if listings != tt.listings {
				t.Errorf("FirstRunErrorPatternListings.MatchString(%q) = %v, want %v", tt.text, listings, tt.listings)
			}

			if filenames != tt.filenames {
				t.Errorf("FirstRunErrorPatternFilenames.MatchString(%q) = %v, want %v", tt.text, filenames, tt.filenames)
			}
		})
	}
}

func TestIsFirstRunRegression(t *testing.T) {
	regressionTests := []struct {
		name   string
		stderr string
		want   bool
	}{
		{
			name:   "path1 only",
			stderr: "ERROR: cannot find prior Path1 listings",
			want:   false,
		},
		{
			name:   "path2 only",
			stderr: "ERROR: cannot find prior Path2 listings",
			want:   false,
		},
		{
			name:   "paths reversed",
			stderr: "ERROR: cannot find prior Path2 or Path1 listings...here are the filenames",
			want:   false,
		},
		{
			name:   "typo - pat1h instead of path1",
			stderr: "ERROR: cannot find prior pat1h or path2 listings...here are the filenames",
			want:   false,
		},
		{
			name:   "typo - filenamez instead of filenames",
			stderr: "ERROR: cannot find prior path1 or path2 listings...here are the filenamez",
			want:   false,
		},
	}

	for _, tt := range regressionTests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsFirstRunError(tt.stderr)
			if result != tt.want {
				t.Errorf("IsFirstRunError() = %v, want %v (regression test)", result, tt.want)
			}
		})
	}
}
