package pathutils

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

func RootDir() (string, error) {
	isRunningTest := flag.Lookup("test.v") != nil
	if isRunningTest {
		_, file, _, _ := runtime.Caller(0)
		return filepath.Join(filepath.Dir(file), "..", "..", ".."), nil
	}

	executable, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("RootDir(): %w", err)
	}

	return filepath.Dir(executable), nil
}

func RelativeFilePath(path string) (string, error) {
	rootDir, err := RootDir()
	if err != nil {
		return "", fmt.Errorf("relativeFilePath(%q): %w", path, err)
	}

	relativePath, err := filepath.Rel(rootDir, path)
	if err != nil {
		return "", fmt.Errorf("relativeFilePath(%q): %w", path, err)
	}

	return "/" + relativePath, nil
}
