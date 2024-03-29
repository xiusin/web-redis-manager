package windows

import (
	"os"
	"path/filepath"
	"strings"
)

func GetStorePath(path ...string) string {
	userhome, _ := os.UserHomeDir()
	dir := filepath.Join(userhome, "OneDrive")
	if _, err := os.Stat(dir); err != nil {
		dir = filepath.Join(userhome, ".rdm")
		os.MkdirAll(dir, os.ModePerm)
	}
	if len(path) > 0 {
		dir = filepath.Join(dir, strings.Join(path, "/"))
	}
	return dir
}
