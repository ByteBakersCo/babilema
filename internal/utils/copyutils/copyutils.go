package copyutils

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"sync"

	"golang.org/x/sync/errgroup"
)

type rollbackAction struct {
	Action      func() error
	Description string // for logging
}

// copyRollbacker is a type that defines a rollbacker for copy operations.
// It is thread-safe and should be passed around as a pointer
type copyRollbacker struct {
	mu              sync.Mutex
	rollbackActions []rollbackAction
	backupDir       string
}

// NewCopyRollbacker creates a new copyRollbacker with a backup directory
// for rollback actions.
// If the backup directory is empty, it will return nil.
func NewCopyRollbacker(backupDir string) (*copyRollbacker, error) {
	if backupDir == "" {
		return nil, fmt.Errorf(
			"NewCopyRollbacker(%q): backup directory not set",
			backupDir,
		)
	}

	return &copyRollbacker{
		rollbackActions: make([]rollbackAction, 0),
		backupDir:       backupDir,
	}, nil
}

func (cr *copyRollbacker) AddRollbackAction(
	description string,
	action func() error,
) {
	if cr == nil {
		return
	}

	cr.mu.Lock()
	defer cr.mu.Unlock()

	cr.rollbackActions = append(cr.rollbackActions, rollbackAction{
		Description: description,
		Action:      action,
	})
}

// AddCopyFileRollbackAction is a utility method to add a rollback action to the
// copyRollbacker for file copy operations.
// It will move the `src` to the `dest`.
func (cr *copyRollbacker) AddCopyFileRollbackAction(
	src string,
	dest string,
) error {
	action := func() error {
		info, err := os.Stat(src)
		if err != nil {
			return fmt.Errorf(
				"AddCopyRollbackAction(%q, %q): %w",
				src,
				dest,
				err,
			)
		}

		if info.IsDir() {
			return fmt.Errorf(
				"AddCopyRollbackAction(%q, %q): source is a directory",
				src,
				dest,
			)
		}

		info, err = os.Stat(dest)
		if err != nil && !errors.Is(err, fs.ErrNotExist) {
			return fmt.Errorf(
				"AddCopyRollbackAction(%q, %q): %w",
				src,
				dest,
				err,
			)
		}

		if err == nil && info.IsDir() {
			return fmt.Errorf(
				"AddCopyRollbackAction(%q, %q): destination is a directory",
				src,
				dest,
			)
		}

		if err == nil {
			err = os.Remove(dest)
			if err != nil {
				return fmt.Errorf(
					"AddCopyRollbackAction(%q, %q): %w",
					src,
					dest,
					err,
				)
			}
		}

		err = CopyFile(src, dest)
		if err != nil {
			return fmt.Errorf(
				"AddCopyRollbackAction(%q, %q): %w",
				src,
				dest,
				err,
			)
		}

		return nil
	}

	cr.AddRollbackAction(
		fmt.Sprintf("Rollback %q --> %q", src, dest),
		action,
	)

	return nil
}

func (cr *copyRollbacker) Rollback() error {
	cr.mu.Lock()
	defer cr.mu.Unlock()

	var err error
	for _, action := range cr.rollbackActions {
		if aerr := action.Action(); aerr != nil {
			err = errors.Join(err, fmt.Errorf(
				"\nexecuteRollback(): %s: %w",
				action.Description,
				aerr,
			))
		} else {
			log.Printf("[SUCCESS] Rollback: %s\n", action.Description)
		}
	}

	cr.rollbackActions = make([]rollbackAction, 0)
	return err
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
	defer sourceFile.Close()

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
	defer destFile.Close()

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
	if src == dest {
		return nil
	}

	if src == "" {
		return fmt.Errorf(
			"copyDir(%q, %q): source directory: %w",
			src,
			dest,
			fs.ErrNotExist,
		)
	}

	if dest == "" {
		return fmt.Errorf(
			"copyDir(%q, %q): destination directory: %w",
			src,
			dest,
			fs.ErrNotExist,
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

	backupDir, err := os.MkdirTemp("", "*-babilema-bak")
	if err != nil {
		return fmt.Errorf("CopyDir(%q, %q): %w", src, dest, err)
	}
	defer os.RemoveAll(backupDir)

	rollbacker, err := NewCopyRollbacker(backupDir)
	if err != nil {
		return fmt.Errorf("CopyDir(%q, %q): %w", src, dest, err)
	}

	return copyDir(context.Background(), src, dest, backupDir, rollbacker)
}

// copyDir copies a directory from src to dest. If backupDir is not empty, it
// will backup the duplicate files in the destination directory to the backup
// directory before copying the source directory to the destination directory.
// If the source and destination directories are the same, it will return nil.
// If the backup directory is not empty, it will rollback the changes on error.
func copyDir(
	ctx context.Context,
	src string,
	dest string,
	backupDir string,
	rollbacker *copyRollbacker,
) error {
	srcDirContents, err := os.ReadDir(src)
	if err != nil {
		return fmt.Errorf("copyDir(%q, %q): %w", src, dest, err)
	}

	// WARN: recusion might cause stack overflow on deeply nested directories
	errGroup, ctx := errgroup.WithContext(ctx)
	for _, entry := range srcDirContents {
		// entry := entry
		errGroup.Go(func() error {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
				srcPath := filepath.Join(src, entry.Name())
				destPath := filepath.Join(dest, entry.Name())

				if entry.IsDir() {
					return copyDir(
						ctx,
						srcPath,
						destPath,
						backupDir,
						rollbacker,
					)
				} else {
					if backupDir != "" {
						newFileInfo, err := entry.Info()
						if err != nil {
							return err
						}

						oldFileInfo, err := os.Stat(filepath.Join(destPath, entry.Name()))
						if err != nil {
							return err
						}

						isModified := newFileInfo.ModTime().After(oldFileInfo.ModTime())
						if isModified {
							backupPath := filepath.Join(backupDir, entry.Name())
							if err = CopyFile(destPath, backupPath); err != nil && !errors.Is(err, fs.ErrNotExist) {
								return err
							}

							err = rollbacker.AddCopyFileRollbackAction(backupPath, destPath)
							if err != nil {
								return fmt.Errorf("copyDir(%q, %q): %w", src, dest, err)
							}
						}
					}

					err = CopyFile(srcPath, destPath)
					if err != nil {
						return fmt.Errorf("copyDir(%q, %q): %w", src, dest, err)
					}

					if backupDir == "" {
						err = rollbacker.AddCopyFileRollbackAction(destPath, srcPath)
						if err != nil {
							return fmt.Errorf("copyDir(%q, %q): %w", src, dest, err)
						}
					}

					return nil
				}
			}
		})
	}

	if err := errGroup.Wait(); err != nil {
		return fmt.Errorf(
			"copyDir(ctx, %q, %q, %q) [Goroutine]: %w",
			src,
			dest,
			backupDir,
			err,
		)
	}

	return nil
}
