package helper_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/Defacto2/helper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	require.NoError(t, err)

	i, err := helper.Count("")
	require.Error(t, err)
	assert.Equal(t, 0, i)

	i, err = helper.Count("nosuchfile")
	require.Error(t, err)
	assert.Equal(t, 0, i)

	i, err = helper.Count(dir)
	require.NoError(t, err)
	assert.Equal(t, testDataFileCount, i)
}

func TestDuplicate(t *testing.T) {
	t.Parallel()
	dir, err := filepath.Abs("testdata")
	require.NoError(t, err)

	r, err := helper.Duplicate(dir, "")
	require.Error(t, err)
	assert.Empty(t, r)
	r, err = helper.Duplicate(dir, dir)
	require.Error(t, err)
	assert.Empty(t, r)

	file, err := filepath.Abs(filepath.Join("testdata", "TEST.DOC"))
	require.NoError(t, err)

	r, err = helper.Duplicate(file, "")
	require.Error(t, err)
	assert.Empty(t, r)

	r, err = helper.Duplicate("", file)
	require.Error(t, err)
	assert.Empty(t, r)

	r, err = helper.Duplicate(file, file)
	require.Error(t, err)
	assert.Empty(t, r)

	r, err = helper.Duplicate(file, t.TempDir())
	require.Error(t, err)
	assert.Empty(t, r)

	dest := filepath.Join(t.TempDir(), "TEST.NFO")
	written, err := helper.Duplicate(file, dest)
	require.NoError(t, err)
	assert.Equal(t, int64(13), written)
}

func TestFiles(t *testing.T) {
	t.Parallel()
	dir, err := filepath.Abs("testdata")
	require.NoError(t, err)

	r, err := helper.Files("")
	require.Error(t, err)
	assert.Empty(t, r)

	r, err = helper.Files("nosuchfile")
	require.Error(t, err)
	assert.Empty(t, r)

	r, err = helper.Files(dir)
	require.NoError(t, err)
	assert.Len(t, r, testDataFileCount)
}

func TestLines(t *testing.T) {
	t.Parallel()
	i, err := helper.Lines("")
	require.Error(t, err)
	assert.Equal(t, 0, i)

	i, err = helper.Lines("nosuchfile")
	require.Error(t, err)
	assert.Equal(t, 0, i)

	i, err = helper.Lines("")
	require.Error(t, err)
	assert.Equal(t, 0, i)

	dir, err := filepath.Abs("testdata")
	require.NoError(t, err)
	name := filepath.Join(dir, "TEST.BMP")
	i, err = helper.Lines(name)
	require.Error(t, err)
	assert.Equal(t, 0, i)

	name = filepath.Join(dir, "PKZ80A1.TXT")
	i, err = helper.Lines(name)
	require.NoError(t, err)
	assert.Equal(t, 175, i)
}

func TestRenameFile(t *testing.T) {
	t.Parallel()
	const name = "test_rename_file"

	err := helper.RenameFile("", "")
	require.Error(t, err)

	err = helper.RenameFile(t.TempDir(), "")
	require.ErrorIs(t, err, helper.ErrFilePath)

	abs := filepath.Join(t.TempDir(), name)
	err = helper.Touch(abs)
	require.NoError(t, err)

	err = helper.RenameFile(abs, "")
	require.Error(t, err)

	err = helper.RenameFile(abs, abs)
	require.Error(t, err)

	err = helper.RenameFile(abs, abs+"~")
	require.NoError(t, err)
}

func TestRenameFileOW(t *testing.T) {
	t.Parallel()
	const name = "test_rename_file"

	err := helper.RenameFileOW("", "")
	require.Error(t, err)

	err = helper.RenameFileOW(t.TempDir(), "")
	require.ErrorIs(t, err, helper.ErrFilePath)

	abs := filepath.Join(t.TempDir(), name)
	err = helper.Touch(abs)
	require.NoError(t, err)

	err = helper.RenameFileOW(abs, "")
	require.Error(t, err)

	err = helper.RenameFileOW(abs, abs)
	require.Error(t, err)

	err = helper.RenameFileOW(abs, abs+"~")
	require.NoError(t, err)
}

func TestRenameCrossDevice(t *testing.T) {
	t.Parallel()
	const name = "test_rename_file"

	err := helper.RenameCrossDevice("", "")
	require.Error(t, err)

	err = helper.RenameCrossDevice(t.TempDir(), "")
	require.Error(t, err)

	abs := filepath.Join(t.TempDir(), name)
	err = helper.Touch(abs)
	require.NoError(t, err)

	err = helper.RenameCrossDevice(abs, "")
	require.Error(t, err)

	err = helper.RenameCrossDevice(abs, abs+"~")
	require.Error(t, err)
}

func TestSize(t *testing.T) {
	t.Parallel()
	const name = "test_rename_file"
	const none = int64(-1)
	data := []byte("Hello, World!")

	i := helper.Size("")
	assert.Equal(t, none, i)

	i = helper.Size("nosuchfile")
	assert.Equal(t, none, i)

	abs := filepath.Join(t.TempDir(), name)
	x, err := helper.TouchW(abs, data...)
	require.NoError(t, err)

	i = helper.Size(abs)
	assert.Equal(t, int64(x), i)
}

func TestStrongIntegrity(t *testing.T) {
	t.Parallel()
	const name = "test_strong_integrity"
	const expected = "5485cc9b3365b4305dfb4e8337e0a598a574f8242bf17289e0" +
		"dd6c20a3cd44a089de16ab4ab308f63e44b1170eb5f515"
	data := []byte("Hello, World!")

	s, err := helper.StrongIntegrity("")
	require.Error(t, err)
	assert.Empty(t, s)

	s, err = helper.StrongIntegrity("nosuchfile")
	require.Error(t, err)
	assert.Empty(t, s)

	abs := filepath.Join(t.TempDir(), name)
	_, err = helper.TouchW(abs, data...)
	require.NoError(t, err)

	s, err = helper.StrongIntegrity(abs)
	require.NoError(t, err)
	assert.Equal(t, expected, s)

	r, err := os.OpenRoot(t.TempDir())
	require.NoError(t, err)
	defer r.Close()
	_, err = helper.TouchWR(r, name, data...)
	require.NoError(t, err)
	s, err = helper.StrongIntegrityR(r, name)
	require.NoError(t, err)
	assert.Equal(t, expected, s)

	err = helper.TouchR(r, name)
	require.Error(t, err)
	_ = r.Remove(name)
	err = helper.TouchR(r, name)
	require.NoError(t, err)

	err = helper.RenameRootOW(r, name, name)
	require.Error(t, err)
	err = helper.RenameRootOW(r, name, name+"abc")
	require.NoError(t, err)

	ok, err := helper.FileMatchR(r, name, name+"abc")
	require.Error(t, err)
	assert.False(t, ok)

	err = helper.TouchR(r, name)
	require.NoError(t, err)
	ok, err = helper.FileMatchR(r, name, name+"abc")
	require.NoError(t, err)
	assert.True(t, ok)
}

func TestOwner(t *testing.T) {
	t.Parallel()

	groups, username, err := helper.Owner()
	require.NoError(t, err)
	assert.NotEmpty(t, groups)
	assert.NotEmpty(t, username)
}
