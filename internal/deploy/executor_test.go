package deploy

import (
	"context"
	"testing"
)

func TestRealExecutorRunsCommand(t *testing.T) {
	exec := &RealExecutor{}
	err := exec.Run(context.Background(), "echo", "hello")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

func TestRealExecutorFailsOnBadCommand(t *testing.T) {
	exec := &RealExecutor{}
	err := exec.Run(context.Background(), "this-command-does-not-exist-xyz")
	if err == nil {
		t.Fatal("expected error for non-existent command, got nil")
	}
}
