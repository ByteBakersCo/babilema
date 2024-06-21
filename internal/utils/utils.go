package utils

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

func RootDir() (string, error) {
	isRunningTest := flag.Lookup("test.v") != nil

	if isRunningTest {
		_, file, _, _ := runtime.Caller(0)
		return filepath.Join(filepath.Dir(file), "..", ".."), nil

	}

	executable, err := os.Executable()
	if err != nil {
		return "", err
	}

	executablePath := filepath.Dir(executable)

	return executablePath, nil
}

// Pretty format struct
func FormatStruct(s interface{}, msg ...string) string {
	return strings.Join(
		msg,
		" ",
	) + "\n" + strings.ReplaceAll(
		fmt.Sprintf("%+v", s),
		" ",
		"\n",
	)
}

func RelativeFilePath(path string) (string, error) {
	rootDir, err := RootDir()
	if err != nil {
		return "", err
	}

	relativePath, err := filepath.Rel(rootDir, path)
	if err != nil {
		return "", err
	}

	return "/" + relativePath, nil
}

func IsCommandAvailable(cmd string) bool {
	command := exec.Command("/bin/sh", "-c", cmd)
	return command.Run() == nil
}

func IsFileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
