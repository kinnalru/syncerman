package version

import (
	"testing"
)

func TestVersionVariables(t *testing.T) {
	if Version == "" {
		t.Skip("Version not set (test run without build)")
	}
}

func TestDefaultValues(t *testing.T) {
	tests := []struct {
		name  string
		value string
	}{
		{"GitCommit", GitCommit},
		{"BuildTime", BuildTime},
		{"GoVersion", GoVersion},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.value == "" {
				t.Errorf("%s should have default value", tt.name)
			}
			if tt.value == "unknown" {
				t.Logf("%s has default value 'unknown' (built without ldflags)", tt.name)
			}
		})
	}
}
