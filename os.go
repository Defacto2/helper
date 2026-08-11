package helper

// Package os contains helper functions for file system operations.

import (
	"bytes"
	"cmp"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/user"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"syscall"
	"unicode/utf8"
)

const (
	DSStore  = ".DS_Store"       // DSStore is the macOS directory service store file.
	TempBase = "defacto2-server" // TempBase is the base subdirectory for temporary files.
)

const (
	// WriteWriteRead is the file mode for read and write access.
	// The file owner and group has read and write access, and others have read access.
	WriteWriteRead   fs.FileMode = 0o664 // WriteWriteRead is the file mode for read and write access.
	DirWriteReadRead fs.FileMode = 0o755 // DirWriteReadRead sets directory permissions for read, write, and execute.
)

// Extension is a file extension with a count of files.
type Extension struct {
	Name  string // Name is the file extension.
	Count int64  // Count is the number of files with the extension.
}

// CountExts returns the file extensions and the number of files in the given directory.
func CountExts(dir string) ([]Extension, error) {
	exts := make(map[string]int64)
	files, err := os.ReadDir(dir)
	if err != nil {
		const format = "count extensions read directory: %w"
		return nil, fmt.Errorf(format, err)
	}
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		if file.Name() == DSStore {
			continue
		}
		ext := strings.ToLower(filepath.Ext(file.Name()))
		exts[ext]++
	}
	extensions := make([]Extension, 0, len(exts))
	for k, v := range exts {
		if k == "" {
			k = "uuid"
		}
		extensions = append(extensions, Extension{Name: k, Count: v})
	}
	sort.Slice(extensions, func(i, j int) bool {
		if extensions[i].Count == extensions[j].Count {
			return extensions[i].Name < extensions[j].Name
		}
		return extensions[i].Count > extensions[j].Count
	})
	return extensions, nil
}

// Count returns the number of files in the given directory.
func Count(dir string) (int, error) {
	const format = "count directory files %s %w"
	entries, err := os.ReadDir(dir)
	if err != nil {
		return 0, fmt.Errorf(format, dir, err)
	}

	count := 0
	for _, entry := range entries {
		if skipCount(entry) {
			continue
		}
		count++
	}

	return count, nil
}

func skipCount(entry os.DirEntry) bool {
	return entry == nil || entry.IsDir() || !entry.Type().IsRegular() || entry.Name() == DSStore
}

// CountStream returns the file count using chunked directory method.
// It is more memory efficient with directories containing many items (10,000+)
// however, it is slightly slower.
func CountStream(dir string) (int, error) {
	const format = "steam count directory files %s %w"
	f, err := os.Open(dir)
	if err != nil {
		return 0, fmt.Errorf(format, dir, err)
	}
	defer f.Close()

	// read entries in batches of n to prevent memory spikes
	const n = 1024
	count := 0
	for {
		entries, err := f.ReadDir(n)
		for _, entry := range entries {
			if skipCount(entry) {
				continue
			}
			count++
		}
		if err != nil {
			break
		}
	}

	return count, nil
}

// DiskUsage returns the total size of the files in the given directory.
func DiskUsage(path string) (int64, error) {
	const format = "disk usage %s: %w"
	var total int64

	err := filepath.WalkDir(path, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf(format, "access "+p, err)
		}

		// skip directories and non-regular files such as symlinks, devices, sockets
		if !d.IsDir() && d.Type().IsRegular() {
			info, err := d.Info()
			if err != nil {
				return fmt.Errorf(format, "info "+p, err)
			}
			total += info.Size()
		}

		return nil
	})
	if err != nil {
		return 0, fmt.Errorf(format, "walk", err)
	}

	return total, nil
}

// Duplicate is a workaround for renaming files across different devices.
// A cross device can also be a different file system such as a Docker volume.
// It returns the number of bytes written to the new file.
// The function returns an error if the newpath already exists.
func Duplicate(oldpath, newpath string) (int64, error) {
	const createNoTruncate = os.O_CREATE | os.O_WRONLY | os.O_EXCL
	return duplicate(oldpath, newpath, createNoTruncate)
}

// DuplicateOW is a workaround for renaming files across different devices.
// A cross device can also be a different file system such as a Docker volume.
// It returns the number of bytes written to the new file.
// The function will truncate and overwrite the newpath if it already exists.
func DuplicateOW(oldpath, newpath string) (int64, error) {
	const createTruncate = os.O_CREATE | os.O_WRONLY | os.O_TRUNC
	return duplicate(oldpath, newpath, createTruncate)
}

func duplicate(oldpath, newpath string, flag int) ( //nolint:nonamedreturns
	written int64, err error,
) {
	const format = "duplicate %s %w"
	src, err := os.Open(oldpath)
	if err != nil {
		return 0, fmt.Errorf(format, "open", err)
	}
	defer src.Close()

	dst, err := os.OpenFile(newpath, flag, WriteWriteRead)
	if err != nil {
		return 0, fmt.Errorf(format, "create", err)
	}
	defer func() {
		if cErr := dst.Close(); cErr != nil {
			err = errors.Join(err, fmt.Errorf(format, "close", cErr))
		}
	}()
	n, err := io.Copy(dst, src)
	if err != nil {
		return 0, fmt.Errorf(format, "copy buffer", err)
	}
	return n, nil
}

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

func duplicater(r *os.Root, name, newname string, flag int) ( //nolint:nonamedreturns
	written int64, err error,
) {
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
			err = errors.Join(err, fmt.Errorf(format, "close", cErr))
		}
	}()
	n, err := io.Copy(dst, src)
	if err != nil {
		return 0, fmt.Errorf(format, "copy buffer", err)
	}
	return n, nil
}

// File returns true if the named file exists on the system.
func File(name string) bool {
	s, err := os.Stat(name)
	if err != nil {
		if os.IsNotExist(err) {
			return false
		}
		return false
	}
	if s.IsDir() {
		return false
	}
	return true
}

// Files returns the filenames in the given directory.
func Files(dir string) ([]string, error) {
	const format = "files %s %w"
	st, err := os.Stat(dir)
	if err != nil {
		return nil, fmt.Errorf(format, "stat", err)
	}
	if !st.IsDir() {
		return nil, fmt.Errorf("%w: %s", ErrDirPath, dir)
	}
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf(format, "readdir", err)
	}
	names := []string{}
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		if file.Name() == DSStore {
			continue
		}
		names = append(names, file.Name())
	}
	return names, nil
}

// FileMatch returns true if the two named files are the same.
// It returns false if the files are of different lengths or
// if an error occurs while reading the files.
// The read buffer size is 4096 bytes.
func FileMatch(name1, name2 string) (bool, error) {
	const format = "file match open %s: %w"
	f1, err := os.Open(name1)
	if err != nil {
		return false, fmt.Errorf(format, name1, err)
	}
	defer f1.Close()

	f2, err := os.Open(name2)
	if err != nil {
		return false, fmt.Errorf(format, name2, err)
	}
	defer f2.Close()

	if ok, result, err := infoMatch(f1, f2); ok {
		return result, err
	}

	return ReaderMatch(f1, f2)
}

func infoMatch(f1, f2 *os.File) (bool, bool, error) {
	const format = "file match open %s: %w"

	fi1, err := f1.Stat()
	if err != nil {
		return true, false, fmt.Errorf(format, "name1", err)
	}
	fi2, err := f2.Stat()
	if err != nil {
		return true, false, fmt.Errorf(format, "name2", err)
	}

	if os.SameFile(fi1, fi2) {
		return true, true, nil
	}
	if fi1.Size() != fi2.Size() {
		return true, false, nil
	}

	return false, false, nil
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

	if ok, result, err := infoMatch(f1, f2); ok {
		return result, err
	}

	return ReaderMatch(f1, f2)
}

// ReaderMatch returns true if the content of the two readers are the same.
// The read buffer size is 4096 bytes.
func ReaderMatch(r1, r2 io.Reader) (bool, error) { //nolint:cyclop
	const format = "file match chunk: %w"
	const bufSize = 4096
	buf1 := make([]byte, bufSize)
	buf2 := make([]byte, bufSize)

	for {
		n1, err1 := io.ReadFull(r1, buf1)
		n2, err2 := io.ReadFull(r2, buf2)

		// compare bytes read in these 4096B chunks first
		if n1 != n2 || !bytes.Equal(buf1[:n1], buf2[:n2]) {
			return false, nil
		}

		// handle completion and errors
		if err1 != nil || err2 != nil {
			if errors.Is(err1, io.EOF) || errors.Is(err1, io.ErrUnexpectedEOF) {
				if errors.Is(err2, io.EOF) || errors.Is(err2, io.ErrUnexpectedEOF) {
					return true, nil // both reached EOF cleanly together
				}
			}
			if err1 != nil && !errors.Is(err1, io.EOF) {
				return false, fmt.Errorf(format, err1)
			}
			return false, fmt.Errorf(format, err2)
		}
	}
}

// Finds returns true if the name is found in the collection of names.
func Finds(name string, names ...string) bool {
	return slices.Contains(names, name)
}

// Integrity returns the sha384 hash of the named embed file.
// This is intended to be used for Subresource Integrity (SRI)
// verification with integrity attributes in HTML script and link tags.
func Integrity(name string, fsys fs.FS) (string, error) {
	f, err := fsys.Open(name)
	if err != nil {
		const format = "integrity fs open %s: %w"
		return "", fmt.Errorf(format, name, err)
	}
	defer f.Close()

	return IntegrityReader(f)
}

// IntegrityFile returns the sha384 hash of the named file.
// This can be used as a link cache buster.
func IntegrityFile(name string) (string, error) {
	f, err := os.Open(name)
	if err != nil {
		const format = "integrity os open %s: %w"
		return "", fmt.Errorf(format, name, err)
	}
	defer f.Close()

	return IntegrityReader(f)
}

// IntegrityReader calculates the sha384 hash of an io.Reader stream.
func IntegrityReader(r io.Reader) (string, error) {
	h := sha512.New384()
	if _, err := io.Copy(h, r); err != nil {
		const format = "integrity read stream: %w"
		return "", fmt.Errorf(format, err)
	}

	b64 := base64.StdEncoding.EncodeToString(h.Sum(nil))
	return "sha384-" + b64, nil
}

// IntegrityBytes returns the sha384 hash of the given byte slice.
func IntegrityBytes(b []byte) string {
	sum := sha512.Sum384(b)
	b64 := base64.StdEncoding.EncodeToString(sum[:])
	return "sha384-" + b64
}

// Lines returns the number of lines in the named file.
func Lines(name string) (int, error) {
	file, err := os.Open(name)
	if err != nil {
		return 0, fmt.Errorf("lines open %s: %w", name, err)
	}
	defer file.Close()

	return CountLines(file)
}

// CountLines counts newline bytes in an io.Reader without line length limits.
func CountLines(r io.Reader) (int, error) {
	const size = 32 * 1024
	buf := make([]byte, size)
	count := 0
	lineCounted := false

	for {
		n, err := r.Read(buf)
		if n > 0 {
			count += bytes.Count(buf[:n], []byte{'\n'})
			lineCounted = true
		}
		if err != nil {
			if err == io.EOF {
				break
			}
			return 0, fmt.Errorf("lines read: %w", err)
		}
	}

	if !lineCounted {
		return 0, nil
	}

	return count, nil
}

// MkContent returns the destination directory for the extracted archive content.
// The directory is created if it does not exist. The directory is named after the source file.
func MkContent(src string) (string, error) {
	const format = "make content %s: %w"

	if !filepath.IsLocal(src) {
		return "", fmt.Errorf(format, "invalid path "+src, filepath.ErrBadPattern)
	}

	base := filepath.Base(src)
	base = strings.TrimSpace(strings.ToLower(base))
	pattern := "artifact-content-" + base
	dst := filepath.Join(TmpDir(), pattern)

	if err := os.MkdirAll(dst, DirWriteReadRead); err != nil {
		return "", fmt.Errorf(format, "mkdir all", err)
	}

	st, err := os.Stat(dst)
	if err != nil {
		return "", fmt.Errorf(format, "stat", err)
	}
	if !st.IsDir() {
		return "", fmt.Errorf(format, dst, ErrNoDir)
	}

	return dst, nil
}

// Owner returns the running user and group of the web application.
// The function returns the group names and the username of the owner.
func Owner() (groups []string, username string, err error) { //nolint:nonamedreturns
	const format = "owner %s: %w"
	curr, err := user.Current()
	if err != nil {
		return nil, "", fmt.Errorf(format, "current user", err)
	}

	ids, err := curr.GroupIds()
	if err != nil {
		return nil, "", fmt.Errorf(format, "group ids for user "+curr.Username, err)
	}

	s := make([]string, 0, len(ids))
	for _, id := range ids {
		grp, err := user.LookupGroupId(id)
		if err != nil || grp == nil {
			// fallback to numeric ID if group name lookup fails
			s = append(s, id)
			continue
		}

		if grp.Name != "" {
			s = append(s, grp.Name)
		} else {
			s = append(s, grp.Gid)
		}
	}

	return s, curr.Username, nil
}

// RenameFile renames a file from oldpath to newpath.
// An error is returned if oldpath does not exist,
// or if oldpath is a directory, or newpath already exists, or the rename fails.
func RenameFile(oldpath, newpath string) error {
	const format = "rename file %s %s: %w"
	// check old path
	st, err := os.Stat(oldpath)
	if err != nil {
		return fmt.Errorf(format, "stat", oldpath, err)
	}
	if st.IsDir() {
		return fmt.Errorf(format, "is dir", oldpath, ErrFilePath)
	}
	// check new path
	if _, err := os.Stat(newpath); err == nil {
		return fmt.Errorf(format, "newpath", newpath, ErrExistPath)
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf(format, "stat", newpath, err)
	}
	// rename paths
	err = os.Rename(oldpath, newpath)
	if err != nil {
		// portable cross-device link error check
		if errors.Is(err, syscall.EXDEV) {
			return RenameCrossDevice(oldpath, newpath)
		}

		// cross-device errors across operating systems check
		var linkErr *os.LinkError
		if errors.As(err, &linkErr) && errors.Is(linkErr.Err, syscall.EXDEV) {
			return RenameCrossDevice(oldpath, newpath)
		}

		return fmt.Errorf("rename file %s to %s: %w", oldpath, newpath, err)
	}

	return nil
}

// RenameFileOW renames a file from oldpath to newpath.
// If newpath is an existing directory, it is removed.
// An error is returned if the oldpath does not exist,
// the oldpath is a directory, or the rename fails.
func RenameFileOW(oldpath, newpath string) error {
	st, err := os.Stat(newpath)
	if err == nil && st.IsDir() {
		if rErr := os.RemoveAll(newpath); rErr != nil {
			const format = "rename file ow remove all newpath %s: %w"
			return fmt.Errorf(format, newpath, rErr)
		}
	}
	return RenameFile(oldpath, newpath)
}

// RenameCrossDevice copies oldpath to newpath and removes oldpath when crossing file system boundaries.
// It can be used as a workaround for renaming files across different devices and file systems,
// such as a mounted Docker volume.
func RenameCrossDevice(oldpath, newpath string) error {
	const format = "cross-device copy %s: %w"

	src, err := os.Open(oldpath)
	if err != nil {
		return fmt.Errorf(format, "open src", err)
	}
	st, err := src.Stat()
	if err != nil {
		return fmt.Errorf(format, "stat src", err)
	}
	if st.IsDir() {
		return fmt.Errorf(format, "src is a directory", ErrFilePath)
	}

	// O_EXCL is to ensure newpath is created atomically without overwriting
	const flag = os.O_WRONLY | os.O_CREATE | os.O_EXCL
	const perm = WriteWriteRead
	dst, err := os.OpenFile(newpath, flag, perm)
	if err != nil {
		src.Close()
		if errors.Is(err, os.ErrExist) {
			return fmt.Errorf(format, newpath, ErrExistPath)
		}
		return fmt.Errorf(format, "create dst", err)
	}

	// copy stream from src to dst
	_, copyErr := io.Copy(dst, src)
	closeErr := dst.Close()

	// clean up the destination file if there were any errors
	if copyErr != nil {
		src.Close()
		_ = os.Remove(newpath)
		return fmt.Errorf(format, "content", copyErr)
	}
	if closeErr != nil {
		src.Close()
		_ = os.Remove(newpath)
		return fmt.Errorf(format, "close dst", closeErr)
	}

	// close src before removing oldpath (required on Windows)
	if err := src.Close(); err != nil {
		return fmt.Errorf(format, "close src", err)
	}

	if err := os.Remove(oldpath); err != nil {
		return fmt.Errorf(format, "remove src", err)
	}

	return nil
}

// RenameRoot renames the file oldname to newname.
// An error is returned if oldname does not exist,
// or if oldname is a directory, or newname already exists, or the rename fails.
func RenameRoot(r *os.Root, oldname, newname string) error {
	const format = "rename file %s %s: %w"
	// check old path
	st, err := r.Stat(oldname)
	if err != nil {
		return fmt.Errorf(format, "stat", oldname, err)
	}
	if st.IsDir() {
		return fmt.Errorf(format, "is dir", oldname, ErrFilePath)
	}
	// check new path
	if _, err = r.Stat(newname); err == nil {
		return fmt.Errorf(format, "newname", newname, ErrExistPath)
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf(format, "stat", newname, err)
	}
	// rename files within the limitations of root
	if err := r.Rename(oldname, newname); err != nil {
		return fmt.Errorf(format, "rename to", newname, err)
	}
	return nil
}

// RenameRootOW renames a file from oldname to newname.
// An error is returned if the oldname does not exist,
// the oldname is a directory, or the rename fails.
func RenameRootOW(r *os.Root, oldname, newname string) error {
	st, err := r.Stat(newname)
	if err == nil && st.IsDir() {
		if rErr := r.RemoveAll(newname); rErr != nil {
			const format = "rename file remove all newname %s: %w"
			return fmt.Errorf(format, newname, rErr)
		}
	}
	return RenameRoot(r, oldname, newname)
}

// Size returns the size of the named file.
// If the path name does not exist, is inaccessible or is a directory, it returns -1.
func Size(name string) int64 {
	st, err := os.Stat(name)
	if err != nil || st.IsDir() {
		return -1
	}
	return st.Size()
}

// Stat stats the named file or directory to confirm it exists on the system.
func Stat(name string) bool {
	_, err := os.Stat(name)
	if err == nil {
		return true
	}
	if errors.Is(err, os.ErrNotExist) {
		// the file or directory does not exist
		return false
	}
	// it may exist, however there could be an file system or permissions error
	return false
}

// SortNames sorts the names using the filepath separator, where the root files are preferred.
//
// Usually the sep value is a forward slash (/) or a Windows backslash (\).
func SortNames(sep string, names []string) []string {
	if len(names) <= 1 {
		return names
	}

	slices.SortFunc(names, func(a, b string) int {
		x := strings.Count(a, sep)
		y := strings.Count(b, sep)
		if x != y {
			return cmp.Compare(x, y)
		}
		// else, use more memory to sort lexicographically
		return cmp.Compare(strings.ToLower(a), strings.ToLower(b))
	})
	return names
}

// StrongIntegrity returns the SHA-384 checksum value of the named file.
func StrongIntegrity(name string) (string, error) {
	const format = "strong integrity %s: %w"

	f, err := os.Open(name)
	if err != nil {
		return "", fmt.Errorf(format, "open", err)
	}
	defer f.Close()

	strong, err := Sum384(f)
	if err != nil {
		return "", fmt.Errorf(format, "sum", err)
	}
	return strong, nil
}

// StrongIntegrityR returns the SHA-386 checksum value of the named file..
func StrongIntegrityR(r *os.Root, name string) (string, error) {
	const format = "strong integrity %s: %w"

	f, err := r.Open(name)
	if err != nil {
		return "", fmt.Errorf(format, "open", err)
	}
	defer f.Close()

	strong, err := Sum384(f)
	if err != nil {
		return "", fmt.Errorf(format, "sum", err)
	}
	return strong, nil
}

// Sum384 returns the hex-encoded SHA-384 checksum value of an io.Reader stream.
func Sum384(r io.Reader) (string, error) {
	if r == nil {
		return "", nil
	}

	strong := sha512.New384()
	if _, err := io.Copy(strong, r); err != nil {
		return "", fmt.Errorf("sha384 checksum read: %w", err)
	}

	return hex.EncodeToString(strong.Sum(nil)), nil
}

// Deprecated: Use [Sum384] instead.
func Sum386(f *os.File) (string, error) {
	if f == nil {
		return "", nil
	}
	return Sum384(f)
}

// TmpDir returns the temporary directory for the server,
// which is a subdirectory of the system temp directory.
func TmpDir() string {
	path := filepath.Join(os.TempDir(), TempBase)
	if err := os.MkdirAll(path, DirWriteReadRead); err != nil {
		return os.TempDir()
	}
	return path
}

// Touch creates a new, empty named file.
// If the file already exists, an error is returned.
func Touch(name string) error {
	_, err := TouchW(name)
	return err
}

// TouchR creates a new, empty named file.
// If the file already exists, an error is returned.
func TouchR(r *os.Root, name string) error {
	_, err := TouchWR(r, name)
	return err
}

// TouchW creates a new named file with the given data.
// If the file already exists, an error is returned.
func TouchW(name string, data ...byte) (written int, err error) { //nolint:nonamedreturns
	const flag = os.O_CREATE | os.O_EXCL | os.O_WRONLY
	const format = "touch file %s: %w"

	file, err := os.OpenFile(name, flag, WriteWriteRead)
	if err != nil {
		return 0, fmt.Errorf(format, "open", err)
	}
	defer func() {
		if cErr := file.Close(); cErr != nil {
			err = errors.Join(err, fmt.Errorf(format, "close", cErr))
		}
	}()
	return touch(file, data...)
}

// TouchWR creates a new named file with the given data.
// If the file already exists, an error is returned.
func TouchWR(r *os.Root, name string, data ...byte) ( //nolint:nonamedreturns
	written int, err error,
) {
	const flag = os.O_CREATE | os.O_EXCL | os.O_WRONLY
	const format = "touch file %s: %w"

	file, err := r.OpenFile(name, flag, WriteWriteRead)
	if err != nil {
		return 0, fmt.Errorf(format, "open", err)
	}
	defer func() {
		if cErr := file.Close(); cErr != nil {
			err = errors.Join(err, fmt.Errorf(format, "close", cErr))
		}
	}()
	return touch(file, data...)
}

func touch(file *os.File, data ...byte) (int, error) {
	const format = "touch file %s: %w"
	if len(data) == 0 {
		return 0, nil
	}

	n, err := file.Write(data)
	if err != nil {
		return 0, fmt.Errorf(format, "write", err)
	}
	if err := file.Sync(); err != nil {
		return n, fmt.Errorf(format, "sync", err)
	}
	return n, nil
}

// UTF8 returns true if the named file is a valid UTF-8 encoded file.
// The function reads the first 512 bytes of the file to determine the encoding.
func UTF8(name string) (bool, error) {
	const format = "utf8 %s: %w"
	r, err := os.Open(name)
	if err != nil {
		return false, fmt.Errorf(format, "open", err)
	}
	defer r.Close()

	const size = 512
	buf := make([]byte, size)

	// short read gracefully until EOF
	n, err := io.ReadAtLeast(r, buf, 1)
	if err != nil {
		if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
			if n == 0 {
				return false, fmt.Errorf(format, name, ErrFileEmpty)
			}
		} else {
			return false, fmt.Errorf(format, name, err)
		}
	}

	p := buf[:n]
	if n == size {
		// reading exactly 512 bytes might truncate a multi-byte rune
		p = trimRuneTrail(p)
	}

	return utf8.Valid(p), nil
}

// trimRuneTrail strips up to 3 bytes from the end of p,
// if it forms incomplete Unicode.
func trimRuneTrail(p []byte) []byte {
	for i := 1; i <= 3 && i <= len(p); i++ {
		// is b a leading byte of a rune?
		b := p[len(p)-i]
		if utf8.RuneStart(b) {
			r, _ := utf8.DecodeRune(p[len(p)-i:])
			if r == utf8.RuneError {
				// trim incomplete sequence
				return p[:len(p)-i]
			}
			break
		}
	}
	return p
}
