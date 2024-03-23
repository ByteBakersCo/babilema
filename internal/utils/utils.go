package utils

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

func RootDir() string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(file), "..", "..")
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
	relativePath, err := filepath.Rel(RootDir(), path)
	if err != nil {
		return "", err
	}

	return "/" + relativePath, nil
}

func CommitAndPushGeneratedFiles(commitMsg string) error {
	gitCommands := [][]string{
		{"config", "--global", "user.name", "'GitHub Actions'"},
		{"config", "--global", "user.email", "'github-actions@github.com'"},
		{"add", "-A"},
		{"commit", "-m", commitMsg},
		{"push"},
	}

	for _, gitCommand := range gitCommands {
		cmd := exec.Command("git", gitCommand...)
		err := cmd.Run()
		if err != nil {
			return err
		}
	}

	return nil
}
