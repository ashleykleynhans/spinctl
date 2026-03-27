package config

import (
	"fmt"
	"path/filepath"

	"github.com/gofrs/flock"
)

func AcquireLock(lockPath string) (*flock.Flock, error) {
	dir := filepath.Dir(lockPath)
	if err := ensureDir(dir); err != nil {
		return nil, err
	}
	fl := flock.New(lockPath)
	locked, err := fl.TryLock()
	if err != nil {
		return nil, fmt.Errorf("acquiring lock: %w", err)
	}
	if !locked {
		return nil, fmt.Errorf("another spinctl instance is running (lock: %s)", lockPath)
	}
	return fl, nil
}

func ReleaseLock(fl *flock.Flock) {
	if fl != nil {
		_ = fl.Unlock()
	}
}
