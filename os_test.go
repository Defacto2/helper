package helper_test

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/Defacto2/helper"
	"github.com/nalgeon/be"
)

const testDataFileCount = 3

func ExampleCountExts() {
	dir, _ := filepath.Abs("testdata")
	counts, _ := helper.CountExts(dir)
	fmt.Printf("%v", counts)
	// Output: [{.bmp 1} {.doc 1} {.txt 1}]
}

func ExampleLines() {
	name, _ := filepath.Abs(filepath.Join("testdata", "PKZ80A1.TXT"))
	lines, _ := helper.Lines(name)
	fmt.Printf("%v", lines)
	// Output: 175
}

func ExampleFiles() {
	dir, _ := filepath.Abs("testdata")
	files, _ := helper.Files(dir)
	fmt.Printf("%v", files)
	// Output: [PKZ80A1.TXT TEST.BMP TEST.DOC]
}

func ExampleDiskUsage() {
	dir, _ := filepath.Abs("testdata")
	du, _ := helper.DiskUsage(dir)
	fmt.Printf("%v", du)
	// Output: 755105
}

func ExampleCount() {
	dir, _ := filepath.Abs("testdata")
	count, _ := helper.Count(dir)
	fmt.Printf("%v", count)
	// Output: 3
}

func TestCount(t *testing.T) {
	t.Parallel()
	dir, err := filepath.Abs("testdata")
	be.Err(t, err, nil)
	i, err := helper.Count("")
	be.Err(t, err)
	be.Equal(t, i, 0)
	i, err = helper.Count("nosuchfile")
	be.Err(t, err)
	be.Equal(t, i, 0)
	i, err = helper.Count(dir)
	be.Err(t, err, nil)
	be.Equal(t, i, testDataFileCount)
}

func TestDuplicate(t *testing.T) {
	t.Parallel()
	dir, err := filepath.Abs("testdata")
	be.Err(t, err, nil)
	r, err := helper.Duplicate(dir, "")
	be.Err(t, err)
	be.Equal(t, r, 0)
	r, err = helper.Duplicate(dir, dir)
	be.Err(t, err)
	be.Equal(t, r, 0)
	file, err := filepath.Abs(filepath.Join("testdata", "TEST.DOC"))
	be.Err(t, err, nil)
	r, err = helper.Duplicate(file, "")
	be.Err(t, err)
	be.Equal(t, r, 0)
	r, err = helper.Duplicate("", file)
	be.Err(t, err)
	be.Equal(t, r, 0)
	r, err = helper.Duplicate(file, file)
	be.Err(t, err)
	be.Equal(t, r, 0)
	r, err = helper.Duplicate(file, t.TempDir())
	be.Err(t, err)
	be.Equal(t, r, 0)
	dest := filepath.Join(t.TempDir(), "TEST.NFO")
	written, err := helper.Duplicate(file, dest)
	be.Err(t, err, nil)
	be.Equal(t, int64(13), written)
}

func TestFiles(t *testing.T) {
	t.Parallel()
	dir, err := filepath.Abs("testdata")
	be.Err(t, err, nil)
	r, err := helper.Files("")
	be.Err(t, err)
	be.Equal(t, len(r), 0)
	r, err = helper.Files("nosuchfile")
	be.Err(t, err)
	be.Equal(t, len(r), 0)
	r, err = helper.Files(dir)
	be.Err(t, err, nil)
	be.Equal(t, len(r), testDataFileCount)
}

func TestLines(t *testing.T) {
	t.Parallel()
	i, err := helper.Lines("")
	be.Err(t, err)
	be.Equal(t, 0, i)
	i, err = helper.Lines("nosuchfile")
	be.Err(t, err)
	be.Equal(t, 0, i)
	i, err = helper.Lines("")
	be.Err(t, err)
	be.Equal(t, 0, i)
	dir, err := filepath.Abs("testdata")
	be.Err(t, err, nil)
	name := filepath.Join(dir, "TEST.BMP")
	i, err = helper.Lines(name)
	be.Err(t, err)
	be.Equal(t, 0, i)
	name = filepath.Join(dir, "PKZ80A1.TXT")
	i, err = helper.Lines(name)
	be.Err(t, err, nil)
	be.Equal(t, 175, i)
}

func TestRenameFile(t *testing.T) {
	t.Parallel()
	const name = "test_rename_file"
	err := helper.RenameFile("", "")
	be.Err(t, err)
	err = helper.RenameFile(t.TempDir(), "")
	be.True(t, errors.Is(err, helper.ErrFilePath))
	abs := filepath.Join(t.TempDir(), name)
	err = helper.Touch(abs)
	be.Err(t, err, nil)
	err = helper.RenameFile(abs, "")
	be.Err(t, err)
	err = helper.RenameFile(abs, abs)
	be.Err(t, err)
	err = helper.RenameFile(abs, abs+"~")
	be.Err(t, err, nil)
}

func TestRenameFileOW(t *testing.T) {
	t.Parallel()
	const name = "test_rename_file"
	err := helper.RenameFileOW("", "")
	be.Err(t, err)
	err = helper.RenameFileOW(t.TempDir(), "")
	be.True(t, errors.Is(err, helper.ErrFilePath))
	abs := filepath.Join(t.TempDir(), name)
	err = helper.Touch(abs)
	be.Err(t, err, nil)
	err = helper.RenameFileOW(abs, "")
	be.Err(t, err)
	err = helper.RenameFileOW(abs, abs)
	be.Err(t, err)
	err = helper.RenameFileOW(abs, abs+"~")
	be.Err(t, err, nil)
}

func TestRenameCrossDevice(t *testing.T) {
	t.Parallel()
	const name = "test_rename_file"
	err := helper.RenameCrossDevice("", "")
	be.Err(t, err)
	err = helper.RenameCrossDevice(t.TempDir(), "")
	be.Err(t, err)
	abs := filepath.Join(t.TempDir(), name)
	err = helper.Touch(abs)
	be.Err(t, err, nil)
	err = helper.RenameCrossDevice(abs, "")
	be.Err(t, err)
	err = helper.RenameCrossDevice(abs, abs+"~")
	be.Err(t, err)
}

func TestSize(t *testing.T) {
	t.Parallel()
	const name = "test_rename_file"
	const none = int64(-1)
	data := []byte("Hello, World!")
	i := helper.Size("")
	be.Equal(t, none, i)
	i = helper.Size("nosuchfile")
	be.Equal(t, none, i)
	abs := filepath.Join(t.TempDir(), name)
	x, err := helper.TouchW(abs, data...)
	be.Err(t, err, nil)
	i = helper.Size(abs)
	be.Equal(t, int64(x), i)
}

func TestStrongIntegrity(t *testing.T) {
	t.Parallel()
	const name = "test_strong_integrity"
	const expected = "5485cc9b3365b4305dfb4e8337e0a598a574f8242bf17289e0" +
		"dd6c20a3cd44a089de16ab4ab308f63e44b1170eb5f515"
	data := []byte("Hello, World!")

	s, err := helper.StrongIntegrity("")
	be.Err(t, err)
	be.True(t, s == "")

	s, err = helper.StrongIntegrity("nosuchfile")
	be.Err(t, err)
	be.True(t, s == "")

	abs := filepath.Join(t.TempDir(), name)
	_, err = helper.TouchW(abs, data...)
	be.Err(t, err, nil)

	s, err = helper.StrongIntegrity(abs)
	be.Err(t, err, nil)
	be.Equal(t, s, expected)

	r, err := os.OpenRoot(t.TempDir())
	be.Err(t, err, nil)
	defer r.Close()
	_, err = helper.TouchWR(r, name, data...)
	be.Err(t, err, nil)
	s, err = helper.StrongIntegrityR(r, name)
	be.Err(t, err, nil)
	be.Equal(t, s, expected)

	err = helper.TouchR(r, name)
	be.Err(t, err)
	_ = r.Remove(name)
	err = helper.TouchR(r, name)
	be.Err(t, err, nil)

	err = helper.RenameRootOW(r, name, name)
	be.Err(t, err)
	err = helper.RenameRootOW(r, name, name+"abc")
	be.Err(t, err, nil)

	ok, err := helper.FileMatchR(r, name, name+"abc")
	be.Err(t, err)
	be.True(t, !ok)

	err = helper.TouchR(r, name)
	be.Err(t, err, nil)
	ok, err = helper.FileMatchR(r, name, name+"abc")
	be.Err(t, err, nil)
	be.True(t, ok)
}

func TestOwner(t *testing.T) {
	t.Parallel()
	groups, username, err := helper.Owner()
	be.Err(t, err, nil)
	notEmpty := !reflect.ValueOf(groups).IsZero()
	be.True(t, notEmpty)
	notEmpty = !reflect.ValueOf(username).IsZero()
	be.True(t, notEmpty)
}
