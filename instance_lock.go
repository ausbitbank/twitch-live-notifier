package main

import (
	"path/filepath"
)

// instanceLockPath is stored next to the config so two processes using the
// same settings file share one lock, while separate configs can run side by
// side.
func instanceLockPath(configPath string) string {
	return filepath.Join(filepath.Dir(configPath), "instance.lock")
}
