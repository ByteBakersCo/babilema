package utils

import (
	"path/filepath"
	"runtime"
)

func RootDir() string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(file), "../..")
}
