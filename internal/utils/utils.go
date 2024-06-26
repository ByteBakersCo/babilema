package utils

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
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
func FormatStruct(s any, msg ...string) string {
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

func IsFileAndExists(filename string) bool {
	info, err := os.Stat(filename)
	if errors.Is(err, fs.ErrNotExist) {
		return false
	}
	return !info.IsDir()
}

// CopyFile copies a file from src to dest.
//
// If the source and destination files are the same, it will return nil.
// If the destination file already exists, it will return an error of type *PathError.
func CopyFile(src, dest string) error {
	if src == dest {
		return nil
	}

	if src == "" {
		return fmt.Errorf("CopyFile(%q, %q): source file not set", src, dest)
	}

	if dest == "" {
		return fmt.Errorf(
			"CopyFile(%q, %q): destination file not set",
			src,
			dest,
		)
	}

	info, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("CopyFile(%q, %q): %w", src, dest, err)
	}

	if info.IsDir() {
		return fmt.Errorf("CopyFile(%q, %q): source is a directory", src, dest)
	}

	sourceFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("CopyFile(%q, %q): %w", src, dest, err)
	}

	err = os.MkdirAll(filepath.Dir(dest), 0755)
	if err != nil {
		return fmt.Errorf("CopyFile(%q, %q): %w", src, dest, err)
	}

	destFile, err := os.OpenFile(
		dest,
		os.O_RDWR|os.O_CREATE|os.O_EXCL,
		info.Mode(),
	)
	if err != nil {
		return fmt.Errorf("CopyFile(%q, %q): %w", src, dest, err)
	}

	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return fmt.Errorf("CopyFile(%q, %q): %w", src, dest, err)
	}

	err = sourceFile.Close()
	if err != nil {
		return fmt.Errorf("CopyFile(%q, %q): %w", src, dest, err)
	}

	err = destFile.Close()
	if err != nil {
		return fmt.Errorf("CopyFile(%q, %q): %w", src, dest, err)
	}

	return nil
}

// CopyDir copies a directory from src to dest.
//
// If the destination directory already exists, it will backup the duplicate
// files in the destination directory to a temporary directory before copying
// the source directory to the destination directory.
// If the source and destination directories are the same, it will return nil.
// It will rollback the changes on error.
func CopyDir(src string, dest string) error {
	backupDir, err := os.MkdirTemp("", "*-babilema-bak")
	if err != nil {
		return fmt.Errorf("CopyDir(%q, %q): %w", src, dest, err)
	}
	defer os.RemoveAll(backupDir)

	rollback := func() {
		if err != nil && dest != "" {
			derr := copyDir(backupDir, dest, "")
			if derr != nil {
				err = fmt.Errorf(
					"CopyDir(%q, %q): failed cleanup: %v - original error: %w",
					src,
					dest,
					derr,
					err,
				)
			}
		}
	}
	defer rollback()

	err = copyDir(src, dest, backupDir)
	if err != nil {
		return err
	}

	return nil
}

// copyDir copies a directory from src to dest. If backupDir is not empty, it
// will backup the duplicate files in the destination directory to the backup
// directory before copying the source directory to the destination directory.
// If the source and destination directories are the same, it will return nil.
// If the backup directory is not empty, it will rollback the changes on error.
func copyDir(src, dest, backupDir string) error {
	if src == dest {
		return nil
	}

	if src == "" {
		return fmt.Errorf(
			"copyDir(%q, %q): source directory not set",
			src,
			dest,
		)
	}

	if dest == "" {
		return fmt.Errorf(
			"copyDir(%q, %q): destination directory not set",
			src,
			dest,
		)
	}

	info, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("copyDir(%q, %q): %w", src, dest, err)
	}

	if !info.IsDir() {
		return fmt.Errorf(
			"copyDir(%q, %q): source is not a directory",
			src,
			dest,
		)
	}

	err = os.MkdirAll(dest, 0755)
	if err != nil {
		return fmt.Errorf("copyDir(%q, %q): %w", src, dest, err)
	}

	info, err = os.Stat(dest)
	if err != nil {
		return fmt.Errorf("copyDir(%q, %q): %w", src, dest, err)
	}

	if !info.IsDir() {
		return fmt.Errorf(
			"copyDir(%q, %q): destination is not a directory",
			src,
			dest,
		)
	}

	srcDirContents, err := os.ReadDir(src)
	if err != nil {
		return fmt.Errorf("copyDir(%q, %q): %w", src, dest, err)
	}

	// WARN: recusion might cause stack overflow on deeply nested directories
	for _, entry := range srcDirContents {
		srcPath := filepath.Join(src, entry.Name())
		destPath := filepath.Join(dest, entry.Name())

		if entry.IsDir() {
			err = copyDir(srcPath, destPath, backupDir)
			if err != nil {
				return fmt.Errorf("copyDir(%q, %q): %w", src, dest, err)
			}
		} else {
			if backupDir != "" {
				backupPath := filepath.Join(backupDir, entry.Name())
				err = CopyFile(destPath, backupPath)
				if err != nil && !errors.Is(err, fs.ErrNotExist) {
					return err
				}
			}

			err = CopyFile(srcPath, destPath)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
