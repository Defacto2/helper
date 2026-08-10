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
	got, err := helper.Count("")
	be.Err(t, err)
	be.Equal(t, got, 0)
	got, err = helper.Count("nosuchfile")
	be.Err(t, err)
	be.Equal(t, got, 0)
	got, err = helper.Count(dir)
	be.Err(t, err, nil)
	be.Equal(t, got, testDataFileCount)
}

func TestDuplicate(t *testing.T) {
	t.Parallel()
	dir, err := filepath.Abs("testdata")
	be.Err(t, err, nil)
	got, err := helper.Duplicate(dir, "")
	be.Err(t, err)
	be.Equal(t, got, 0)
	got, err = helper.Duplicate(dir, dir)
	be.Err(t, err)
	be.Equal(t, got, 0)
	file, err := filepath.Abs(filepath.Join("testdata", "TEST.DOC"))
	be.Err(t, err, nil)
	got, err = helper.Duplicate(file, "")
	be.Err(t, err)
	be.Equal(t, got, 0)
	got, err = helper.Duplicate("", file)
	be.Err(t, err)
	be.Equal(t, got, 0)
	got, err = helper.Duplicate(file, file)
	be.Err(t, err)
	be.Equal(t, got, 0)
	got, err = helper.Duplicate(file, t.TempDir())
	be.Err(t, err)
	be.Equal(t, got, 0)
	dest := filepath.Join(t.TempDir(), "TEST.NFO")
	written, err := helper.Duplicate(file, dest)
	be.Err(t, err, nil)
	be.Equal(t, written, int64(13))
}

func TestFiles(t *testing.T) {
	t.Parallel()
	dir, err := filepath.Abs("testdata")
	be.Err(t, err, nil)
	got, err := helper.Files("")
	be.Err(t, err)
	be.Equal(t, len(got), 0)
	got, err = helper.Files("nosuchfile")
	be.Err(t, err)
	be.Equal(t, len(got), 0)
	got, err = helper.Files(dir)
	be.Err(t, err, nil)
	be.Equal(t, len(got), testDataFileCount)
}

func TestLines(t *testing.T) {
	t.Parallel()
	got, err := helper.Lines("")
	be.Err(t, err)
	be.Equal(t, got, 0)
	got, err = helper.Lines("nosuchfile")
	be.Err(t, err)
	be.Equal(t, got, 0)
	got, err = helper.Lines("")
	be.Err(t, err)
	be.Equal(t, got, 0)
	dir, err := filepath.Abs("testdata")
	be.Err(t, err, nil)
	name := filepath.Join(dir, "TEST.BMP")
	got, err = helper.Lines(name)
	be.Err(t, err, nil)
	be.Equal(t, got, 0)
	name = filepath.Join(dir, "PKZ80A1.TXT")
	got, err = helper.Lines(name)
	be.Err(t, err, nil)
	be.Equal(t, got, 175)
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
	got := helper.Size("")
	be.Equal(t, got, none)
	got = helper.Size("nosuchfile")
	be.Equal(t, got, none)
	abs := filepath.Join(t.TempDir(), name)
	x, err := helper.TouchW(abs, data...)
	be.Err(t, err, nil)
	got = helper.Size(abs)
	be.Equal(t, got, int64(x))
}

func TestStrongIntegrity(t *testing.T) {
	t.Parallel()
	const name = "test_strong_integrity"
	const expected = "5485cc9b3365b4305dfb4e8337e0a598a574f8242bf17289e0" +
		"dd6c20a3cd44a089de16ab4ab308f63e44b1170eb5f515"
	data := []byte("Hello, World!")

	got, err := helper.StrongIntegrity("")
	be.Err(t, err)
	be.True(t, got == "")

	got, err = helper.StrongIntegrity("nosuchfile")
	be.Err(t, err)
	be.True(t, got == "")

	abs := filepath.Join(t.TempDir(), name)
	_, err = helper.TouchW(abs, data...)
	be.Err(t, err, nil)

	got, err = helper.StrongIntegrity(abs)
	be.Err(t, err, nil)
	be.Equal(t, got, expected)

	r, err := os.OpenRoot(t.TempDir())
	be.Err(t, err, nil)
	defer r.Close()
	_, err = helper.TouchWR(r, name, data...)
	be.Err(t, err, nil)
	got, err = helper.StrongIntegrityR(r, name)
	be.Err(t, err, nil)
	be.Equal(t, got, expected)

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
	got := !reflect.ValueOf(groups).IsZero()
	be.True(t, got)
	got = !reflect.ValueOf(username).IsZero()
	be.True(t, got)
}
