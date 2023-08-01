//go:build !windows

package windows

import (
	"os"
	"path/filepath"
	"strings"
)

func GetStorePath(path ...string) string {
	home, _ := os.UserHomeDir()
	dir := filepath.Join(home, "icloud", "个人保管库")
	if _, err := os.Stat(dir); err != nil {
		dir = filepath.Join(home, ".rdm")
		_ = os.MkdirAll(dir, os.ModePerm)
	}
	if len(path) > 0 {
		dir = filepath.Join(dir, strings.Join(path, "/"))
	}
	return dir
}
