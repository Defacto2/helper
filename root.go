package helper

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// Package root contains helper functions for file system operations
// that are constrained to the root directory.

// Duplicater copies the contents of the named file to a new named file with the root.
// The function returns an error if the newpath already exists.
func Duplicater(r *os.Root, name, newname string) (int64, error) {
	const createNoTruncate = os.O_CREATE | os.O_WRONLY | os.O_EXCL
	return duplicater(r, name, newname, createNoTruncate)
}

// DuplicaterOW copies the contents of the named file to a new file with the root.
// The function will truncate and overwrite the newpath if it already exists.
func DuplicaterOW(r *os.Root, name, newname string) (int64, error) {
	const createTruncate = os.O_CREATE | os.O_WRONLY | os.O_TRUNC
	return duplicater(r, name, newname, createTruncate)
}

func duplicater(r *os.Root, name, newname string, flag int) (written int64, err error) {
	const format = "duplicater %s %w"
	src, err := r.Open(name)
	if err != nil {
		return 0, fmt.Errorf(format, "open", err)
	}
	defer src.Close()

	dst, err := r.OpenFile(newname, flag, WriteWriteRead)
	if err != nil {
		return 0, fmt.Errorf(format, "create", err)
	}
	defer func() {
		if cErr := dst.Close(); cErr != nil {
			err = errors.Join(err, fmt.Errorf(format, "close destination", cErr))
		}
	}()
	n, err := io.Copy(dst, src)
	if err != nil {
		return 0, fmt.Errorf(format, "io copy", err)
	}
	return n, nil
}

// FileMatchR returns true if the two named files are the same.
// It returns false if the files are of different lengths or
// if an error occurs while reading the files.
// The read buffer size is 4096 bytes.
func FileMatchR(r *os.Root, name1, name2 string) (bool, error) {
	const format = "file match open %s: %w"
	f1, err := r.Open(name1)
	if err != nil {
		return false, fmt.Errorf(format, name1, err)
	}
	defer f1.Close()
	f2, err := r.Open(name2)
	if err != nil {
		return false, fmt.Errorf(format, name2, err)
	}
	defer f2.Close()
	return fileMatch(f1, f2)
}

// RenameRoot renames a file from oldname to newname..
// It returns an error if the oldname does not exist or is a directory,
// newname already exists, or the rename fails.
func RenameRoot(r *os.Root, oldname, newname string) error {
	const format = "rename file %s: %w"
	st, err := r.Stat(oldname)
	if err != nil {
		return fmt.Errorf(format, "stat", err)
	}
	if st.IsDir() {
		return fmt.Errorf(format+" %s", "oldname", ErrFilePath, oldname)
	}
	if _, err = r.Stat(newname); err == nil {
		return fmt.Errorf(format+" %s", "newname", ErrExistPath, newname)
	}
	oldpath := filepath.Join(r.Name(), oldname)
	newpath := filepath.Join(r.Name(), newname)
	if err := os.Rename(oldpath, newpath); err != nil {
		return fmt.Errorf(format, "os rename", err)
	}
	return nil
}

// RenameRootOW renames a file from oldname to newname..
// It returns an error if the oldname does not exist or is a directory
// or the rename fails.
func RenameRootOW(r *os.Root, oldname, newname string) error {
	st, err := r.Stat(newname)
	if err == nil && st.IsDir() {
		_ = r.Remove(newname)
	}
	return RenameRoot(r, oldname, newname)
}

// StrongIntegrityR returns the SHA-386 checksum value of the named file..
func StrongIntegrityR(r *os.Root, name string) (string, error) {
	const format = "strong integrity %s: %w"
	f, err := r.Open(name)
	if err != nil {
		return "", fmt.Errorf(format, "open", err)
	}
	defer f.Close()
	strong, err := Sum386(f)
	if err != nil {
		return "", fmt.Errorf(format, "sum", err)
	}
	return strong, nil
}

// TouchR creates a new, empty named file..
// If the file already exists, an error is returned.
func TouchR(r *os.Root, name string) error {
	const format = "touch r %s file %w"
	const flag = os.O_CREATE | os.O_EXCL
	f, err := r.OpenFile(name, flag, WriteWriteRead)
	if err != nil {
		return fmt.Errorf(format, "open", err)
	}
	if err := f.Close(); err != nil {
		return fmt.Errorf(format, "close", err)
	}
	return nil
}

// TouchWR creates a new named file with the given data..
// If the file already exists, an error is returned.
func TouchWR(r *os.Root, name string, data ...byte) (written int, err error) {
	const format = "touch wr %s file %w"
	const flag = os.O_CREATE | os.O_EXCL | os.O_WRONLY
	file, err := r.OpenFile(name, flag, WriteWriteRead)
	if err != nil {
		return 0, fmt.Errorf(format, "open", err)
	}
	defer func() {
		if cErr := file.Close(); cErr != nil {
			err = errors.Join(err, fmt.Errorf(format, "close", cErr))
		}
	}()

	if len(data) == 0 {
		return 0, nil
	}
	i, err := file.Write(data)
	if err != nil {
		return 0, fmt.Errorf(format, "write", err)
	}
	return i, nil
}
