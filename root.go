package helper

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// Package file root.go contains the helper functions for file system
// that are constrained to the root directory.

// Duplicater copies the contents of the named file to a new named file with the root.
// The function returns an error if the newpath already exists.
func Duplicater(r *os.Root, name, newname string) (int64, error) {
	const createNoTruncate = os.O_CREATE | os.O_WRONLY | os.O_EXCL
	return duplicater(r, name, newname, createNoTruncate)
}

// Duplicater copies the contents of the named file to a new file with the root.
// The function will truncate and overwrite the newpath if it already exists.
func DuplicaterOW(r *os.Root, name, newname string) (int64, error) {
	const createTruncate = os.O_CREATE | os.O_WRONLY | os.O_TRUNC
	return duplicater(r, name, newname, createTruncate)
}

func duplicater(r *os.Root, name, newname string, flag int) (int64, error) {
	src, err := r.Open(name)
	if err != nil {
		return 0, fmt.Errorf("r duplicate os.open %w", err)
	}
	defer src.Close()

	dst, err := r.OpenFile(newname, flag, WriteWriteRead)
	if err != nil {
		return 0, fmt.Errorf("r duplicate os.create %w", err)
	}
	defer dst.Close()

	written, err := io.Copy(dst, src)
	if err != nil {
		return 0, fmt.Errorf("r duplicate io.copy %w", err)
	}
	return written, nil
}

// FileMatchR returns true if the two named files are the same.
// It returns false if the files are of different lengths or
// if an error occurs while reading the files.
// The read buffer size is 4096 bytes.
func FileMatchR(r *os.Root, name1, name2 string) (bool, error) {
	f1, err := r.Open(name1)
	if err != nil {
		return false, fmt.Errorf("file match os.open %s: %w", name1, err)
	}
	defer f1.Close()
	f2, err := r.Open(name2)
	if err != nil {
		return false, fmt.Errorf("file match os.open %s: %w", name2, err)
	}
	defer f2.Close()
	return fileMatch(f1, f2)
}

// RenameRoot renames a file from oldname to newname.
// It returns an error if the oldname does not exist or is a directory,
// newname already exists, or the rename fails.
func RenameRoot(r *os.Root, oldname, newname string) error {
	st, err := r.Stat(oldname)
	if err != nil {
		return fmt.Errorf("rename file r.stat %w", err)
	}
	if st.IsDir() {
		return fmt.Errorf("rename file oldname %w: %s", ErrFilePath, oldname)
	}
	if _, err = r.Stat(newname); err == nil {
		return fmt.Errorf("rename file newname %w: %s", ErrExistPath, newname)
	}
	oldpath := filepath.Join(r.Name(), oldname)
	newpath := filepath.Join(r.Name(), newname)
	if err := os.Rename(oldpath, newpath); err != nil {
		return fmt.Errorf("rename file os.rename %w", err)
	}
	return nil
}

// RenameFRenameRootOWileOW renames a file from oldname to newname.
// It returns an error if the oldname does not exist or is a directory
// or the rename fails.
func RenameRootOW(r *os.Root, oldname, newname string) error {
	st, err := r.Stat(newname)
	if err == nil && st.IsDir() {
		_ = r.Remove(newname)
	}
	return RenameRoot(r, oldname, newname)
}

// StrongIntegrityR returns the SHA-386 checksum value of the named file.
func StrongIntegrityR(r *os.Root, name string) (string, error) {
	f, err := r.Open(name)
	if err != nil {
		return "", fmt.Errorf("strong integrity os.open %s: %w", name, err)
	}
	defer f.Close()
	strong, err := Sum386(f)
	if err != nil {
		return "", fmt.Errorf("strong integrity %w", err)
	}
	return strong, nil
}

// TouchR creates a new, empty named file.
// If the file already exists, an error is returned.
func TouchR(r *os.Root, name string) error {
	f, err := r.OpenFile(name, os.O_CREATE|os.O_EXCL, WriteWriteRead)
	if err != nil {
		return fmt.Errorf("touch open file %w", err)
	}
	if err := f.Close(); err != nil {
		return fmt.Errorf("touch file close %w", err)
	}
	return nil
}

// TouchWR creates a new named file with the given data.
// If the file already exists, an error is returned.
func TouchWR(r *os.Root, name string, data ...byte) (int, error) {
	file, err := r.OpenFile(name, os.O_CREATE|os.O_WRONLY, WriteWriteRead)
	if err != nil {
		return 0, fmt.Errorf("touch write open file %w", err)
	}
	if len(data) == 0 {
		if err := file.Close(); err != nil {
			return 0, fmt.Errorf("touch write open file close %w", err)
		}
		return 0, nil
	}
	i, err := file.Write(data)
	if err != nil {
		return 0, fmt.Errorf("touch write file write %w", err)
	}
	if err := file.Close(); err != nil {
		return 0, fmt.Errorf("touch write file write close %w", err)
	}
	return i, nil
}
