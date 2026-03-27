package config

import (
	"path/filepath"
	"testing"
)

func TestAcquireLock(t *testing.T) {
	tmpDir := t.TempDir()
	lockPath := filepath.Join(tmpDir, ".lock")
	lock, err := AcquireLock(lockPath)
	if err != nil {
		t.Fatalf("AcquireLock() error = %v", err)
	}
	defer ReleaseLock(lock)
	if lock == nil {
		t.Fatal("lock should not be nil")
	}
}

func TestAcquireLockConflict(t *testing.T) {
	tmpDir := t.TempDir()
	lockPath := filepath.Join(tmpDir, ".lock")
	lock1, err := AcquireLock(lockPath)
	if err != nil {
		t.Fatalf("first AcquireLock() error = %v", err)
	}
	defer ReleaseLock(lock1)
	_, err = AcquireLock(lockPath)
	if err == nil {
		t.Error("expected error when lock is already held")
	}
}

func TestReleaseLock(t *testing.T) {
	tmpDir := t.TempDir()
	lockPath := filepath.Join(tmpDir, ".lock")
	lock1, err := AcquireLock(lockPath)
	if err != nil {
		t.Fatalf("AcquireLock() error = %v", err)
	}
	ReleaseLock(lock1)
	lock2, err := AcquireLock(lockPath)
	if err != nil {
		t.Fatalf("second AcquireLock() error = %v", err)
	}
	ReleaseLock(lock2)
}
