package main

import (
	"os"
	"testing"

	"gitlab.com/kinnalru/syncerman/internal/cmd"
)

func TestRunIntegration(t *testing.T) {
	t.Run("help command", func(t *testing.T) {
		oldArgs := os.Args
		defer func() { os.Args = oldArgs }()
		os.Args = []string{"syncerman", "--help"}

		err := run()
		if err != nil {
			t.Errorf("run() error = %v", err)
		}
	})
}

func TestMainIntegration(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	os.Args = []string{"syncerman", "--help"}

	err := cmd.Execute()
	if err != nil {
		t.Errorf("cmd.Execute() error = %v", err)
	}
}
