package utils

import (
	"fmt"
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
