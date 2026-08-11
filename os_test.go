package helper_test

import (
	"bytes"
	"embed"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/Defacto2/helper"
	"github.com/nalgeon/be"
)

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

	got, err := helper.Count("")
	be.Err(t, err)
	be.Equal(t, got, 0)

	got, err = helper.Count(testdataBad)
	be.Err(t, err)
	be.Equal(t, got, 0)

	got, err = helper.Count(testdata)
	be.Err(t, err, nil)
	be.Equal(t, got, testdataCount)
}

func TestCountStream(t *testing.T) {
	t.Parallel()

	got, err := helper.CountStream("")
	be.Err(t, err)
	be.Equal(t, got, 0)

	got, err = helper.CountStream(testdataBad)
	be.Err(t, err)
	be.Equal(t, got, 0)

	got, err = helper.CountStream(testdata)
	be.Err(t, err, nil)
	be.Equal(t, got, testdataCount)
}

func TestDuplicate(t *testing.T) {
	t.Parallel()

	got, err := helper.Duplicate(testdata, "")
	be.Err(t, err)
	be.Equal(t, got, 0)

	got, err = helper.Duplicate(testdata, testdata)
	be.Err(t, err)
	be.Equal(t, got, 0)

	n, dir, src, dst := createCombo(t)

	got, err = helper.Duplicate(src, "")
	be.Err(t, err)
	be.Equal(t, got, 0)

	got, err = helper.Duplicate("", dst)
	be.Err(t, err)
	be.Equal(t, got, 0)

	got, err = helper.Duplicate(src, src)
	be.Err(t, err)
	be.Equal(t, got, 0)

	got, err = helper.Duplicate(src, dir)
	be.Err(t, err)
	be.Equal(t, got, 0)

	got, err = helper.Duplicate(src, dst)
	be.Err(t, err, nil)
	be.Equal(t, got, int64(n))

	// running duplicate again should fail as dst now exists
	_, err = helper.Duplicate(src, dst)
	be.Err(t, err)

	// but this should overwrite dst
	got, err = helper.DuplicateOW(src, dst)
	be.Err(t, err, nil)
	be.Equal(t, got, int64(n))
}

func TestFiles(t *testing.T) {
	t.Parallel()

	got, err := helper.Files("")
	be.Err(t, err)
	be.Equal(t, len(got), 0)

	got, err = helper.Files(testdataBad)
	be.Err(t, err)
	be.Equal(t, len(got), 0)

	got, err = helper.Files(testdata)
	be.Err(t, err, nil)
	be.Equal(t, len(got), testdataCount)
}

func TestLines(t *testing.T) {
	t.Parallel()

	got, err := helper.Lines("")
	be.Err(t, err)
	be.Equal(t, got, 0)

	got, err = helper.Lines(testdataBad)
	be.Err(t, err)
	be.Equal(t, got, 0)

	got, err = helper.Lines(testdata)
	be.Err(t, err)
	be.Equal(t, got, 0)

	binary := filepath.Join(testdata, "TEST.BMP")
	got, err = helper.Lines(binary)
	be.Err(t, err, nil)
	be.Equal(t, got, 0)

	textfile := filepath.Join(testdata, "PKZ80A1.TXT")
	const want = 175
	got, err = helper.Lines(textfile)
	be.Err(t, err, nil)
	be.Equal(t, got, want)
}

func TestRenameFile(t *testing.T) {
	t.Parallel()

	err := helper.RenameFile("", "")
	be.Err(t, err)

	_, dir, src, dst := createCombo(t)

	err = helper.RenameFile(dir, "")
	be.True(t, errors.Is(err, helper.ErrFilePath))

	err = helper.RenameFile(src, "")
	be.Err(t, err)

	err = helper.RenameFile(src, src)
	be.Err(t, err)

	err = helper.RenameFile(src, dst)
	be.Err(t, err, nil)

	// if we run the same command again we should get an error
	// as the src should not exist
	err = helper.RenameFile(src, dst)
	be.Err(t, err)
}

func TestRenameFileOW(t *testing.T) {
	t.Parallel()

	_, dir, src, dst := createCombo(t)

	err := helper.RenameFileOW(dir, "")
	be.True(t, errors.Is(err, helper.ErrFilePath))

	be.Err(t, helper.RenameFileOW("", ""))
	be.Err(t, helper.RenameFileOW(src, ""))
	be.Err(t, helper.RenameFileOW("", dst))
	be.Err(t, helper.RenameFileOW(dst, dst))
	be.Err(t, helper.RenameFileOW(src, src))

	err = helper.RenameFileOW(src, dst)
	be.Err(t, err, nil)

	// if we run the same command again we should get an error
	// as the src should not exist
	err = helper.RenameFileOW(src, dst)
	be.Err(t, err)

	// test the overwrite directory ability
	err = os.MkdirAll(src, helper.DirWriteReadRead)
	be.Err(t, err, nil)

	err = helper.RenameFileOW(dst, src)
	be.Err(t, err, nil)
}

func TestRenameCrossDevice(t *testing.T) {
	t.Parallel()

	n, dir, src, dst := createCombo(t)

	be.Err(t, helper.RenameCrossDevice("", ""))
	be.Err(t, helper.RenameCrossDevice(dir, dst))
	be.Err(t, helper.RenameCrossDevice(src, ""))
	be.Err(t, helper.RenameCrossDevice("", dst))
	be.Err(t, helper.RenameCrossDevice(testdataBad, dst))

	// rename src to dst and confirm the expected dst file size.
	got := helper.RenameCrossDevice(src, dst)
	be.Err(t, got, nil)
	st, err := os.Stat(dst)
	be.Err(t, err, nil)
	be.Equal(t, st.Size(), int64(n))
}

func TestSize(t *testing.T) {
	t.Parallel()

	const none = int64(-1)
	n, dir, src, dst := createCombo(t)
	want := int64(n)

	be.Equal(t, helper.Size(""), none)
	be.Equal(t, helper.Size(testdataBad), none)
	be.Equal(t, helper.Size(dir), none)
	be.Equal(t, helper.Size(dst), none)

	be.Equal(t, helper.Size(src), want)
	bmp := filepath.Join(testdata, "TEST.BMP")
	be.Equal(t, helper.Size(bmp), testdataBMP)
}

func TestStrongIntegrity(t *testing.T) {
	t.Parallel()

	got, err := helper.StrongIntegrity("")
	be.Err(t, err)
	be.True(t, got == "")

	got, err = helper.StrongIntegrity(testdataBad)
	be.Err(t, err)
	be.True(t, got == "")

	name := filepath.Join(testdata, "TEST.BMP")
	got, err = helper.StrongIntegrity(name)
	be.Err(t, err, nil)
	be.Equal(t, got, testdataBMP384)
}

func TestOwner(t *testing.T) {
	t.Parallel()

	groups, username, err := helper.Owner()
	be.Err(t, err, nil)

	// as the results will be different on every system,
	// test to confirm the groups and username are not empty
	be.True(t, !reflect.ValueOf(groups).IsZero())
	be.True(t, !reflect.ValueOf(username).IsZero())
}

func TestTmpDir_Touch(t *testing.T) {
	t.Parallel()

	dir := helper.TmpDir()
	t.Cleanup(func() {
		if err := os.RemoveAll(dir); err != nil {
			t.Logf("could not remove dir %s: %v", dir, err)
		}
	})

	// confirm we can write a file to the new tmp dir
	data := []byte(testdataBMP384)
	name := filepath.Join(dir, "TEST.DATA")
	n, err := helper.TouchW(name, data...)
	be.Err(t, err, nil)
	be.Equal(t, n, len(testdataBMP384))

	// trying again should return an error
	// because the named file already exists
	_, err = helper.TouchW(name, data...)
	be.Err(t, err)

	// also test the touch func to confirm it returns an error
	be.Err(t, helper.Touch(name))
}

func TestMkContent(t *testing.T) {
	t.Parallel()

	dir, err := helper.MkContent("")
	be.Err(t, err)
	be.Equal(t, dir, "")

	dir, err = helper.MkContent("..")
	be.Err(t, err)
	be.Equal(t, dir, "")

	dir, err = helper.MkContent("TEST.DATA")
	be.Err(t, err, nil)
	be.True(t, strings.Contains(dir, helper.TempBase))
	be.True(t, strings.Contains(dir, "test.data"))
	be.True(t, strings.Contains(dir, "artifact-content"))

	if dir != "" {
		t.Cleanup(func() { cleanup(t, dir) })
	}
}

func TestFileMatch(t *testing.T) {
	t.Parallel()

	_, err := helper.FileMatch("", "")
	be.Err(t, err)
	_, err = helper.FileMatch(testdataBad, "")
	be.Err(t, err)
	_, err = helper.FileMatch("", testdataBad)
	be.Err(t, err)
	_, err = helper.FileMatch(testdataBad, testdataBad)
	be.Err(t, err)

	bmp := filepath.Join(testdata, "TEST.BMP")
	doc := filepath.Join(testdata, "TEST.DOC")

	got, err := helper.FileMatch(bmp, doc)
	be.Err(t, err, nil)
	be.True(t, !got)
	got, err = helper.FileMatch(doc, bmp)
	be.Err(t, err, nil)
	be.True(t, !got)

	got, err = helper.FileMatch(bmp, bmp)
	be.Err(t, err, nil)
	be.True(t, got)

	// create a duplicate of the bmp in a different dir
	dupe := filepath.Join(t.TempDir(), "TEST.DUMP")
	n, err := helper.Duplicate(bmp, dupe)
	be.Err(t, err, nil)
	be.Equal(t, n, testdataBMP)

	got, err = helper.FileMatch(bmp, dupe)
	be.Err(t, err, nil)
	be.True(t, got)
}

func TestUTF8(t *testing.T) {
	t.Parallel()

	grin := []byte("😄")
	padded := bytes.Repeat([]byte("a"), 510)
	padded = append(padded, grin...)

	tests := []struct {
		name    string
		input   []byte
		want    bool
		wantErr bool
	}{
		{
			"empty",
			nil, false, true,
		},
		{
			"ascii",
			[]byte("hello"), true, false,
		},
		{
			"latin1",
			[]byte("fractions \xbc \xbd \xbe"), false, false,
		},
		{
			"utf8",
			[]byte("fractions \u00bc \u00bd \u00be"), true, false,
		},
		{
			"utf8 grinning face",
			grin, true, false,
		},
		{
			"test trimRuneTrail with broken emoji",
			padded, true, false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			name := filepath.Join(t.TempDir(), "file.txt")
			_, _ = helper.TouchW(name, tt.input...)
			got, err := helper.UTF8(name)
			if (err != nil) != tt.wantErr {
				t.Errorf("UTF8() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			be.Equal(t, got, tt.want)
		})
	}
}

func TestSortName(t *testing.T) {
	t.Parallel()
	names := []string{
		"file1.txt",
		"dir1/file2.txt",
		"dir2/file3.txt",
		"file4.txt",
		"dir1/subdir1/file5.txt",
		"dir1/file.txt",
	}
	sorted := []string{
		"file1.txt",
		"file4.txt",
		"dir1/file.txt",
		"dir1/file2.txt",
		"dir2/file3.txt",
		"dir1/subdir1/file5.txt",
	}

	tests := []struct {
		name string
		in   []string
		want []string
	}{
		{"empty", []string{}, []string{}},
		{"one", []string{"b", "a", "2", "0"}, []string{"0", "2", "a", "b"}},
		{"two", names, sorted},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := helper.SortNames("/", tt.in)
			be.Equal(t, got, tt.want)
		})
	}
}

func TestStat(t *testing.T) {
	t.Parallel()
	self := filepath.Join(".", "helper_test.go")
	tests := []struct {
		name   string
		expect bool
	}{
		{self, true},
		{"^&%#$%@#", false},
		{testdataBad, false},
		{testdata, true},
		{"testdata/TEST.DOC", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			be.Equal(t, helper.Stat(tt.name), tt.expect)
		})
	}
}

func TestIntegrityFile(t *testing.T) {
	t.Parallel()
	got, err := helper.IntegrityFile("")
	be.Err(t, err)
	be.Equal(t, got, "")
	got, err = helper.IntegrityFile("nosuchfile")
	be.Err(t, err)
	be.Equal(t, got, "")
	got, err = helper.IntegrityFile("testdata/TEST.DOC")
	be.Err(t, err, nil)
	be.Equal(t, got, "sha384-5X6isqmILTavQSao9DigKt3O8fX1Hd6hrGJ7pUROFPYWmkKRnFuWwTnjO3h9QkWP")
}

func TestIntegrityBytes(t *testing.T) {
	t.Parallel()
	got := helper.IntegrityBytes(nil)
	be.Equal(t, got, "sha384-OLBgp1GsljhM2TJ+sbHjaiH9txEUvgdDTAzHv2P24donTt6/529l+9Ua0vFImLlb")
	got = helper.IntegrityBytes([]byte("hello world"))
	be.Equal(t, got, "sha384-/b2OdaZ/KfcBpOBAOF4uI5hjA+oQI5IRr5B/y7g1eLPkF8txzmRu/QgZ3YwIjeG9")
}

func TestIntegrity(t *testing.T) {
	t.Parallel()
	got, err := helper.Integrity("", embed.FS{})
	be.Err(t, err)
	be.Equal(t, got, "")
	got, err = helper.Integrity("nosuchfile", testdataFS)
	be.Err(t, err)
	be.Equal(t, got, "")
	got, err = helper.Integrity("testdata/TEST.DOC", testdataFS)
	be.Err(t, err, nil)
	be.Equal(t, got, "sha384-5X6isqmILTavQSao9DigKt3O8fX1Hd6hrGJ7pUROFPYWmkKRnFuWwTnjO3h9QkWP")
}

func TestFinds(t *testing.T) {
	t.Parallel()
	s := []string{"abc", "def", "ghi"}
	type args struct {
		name  string
		names []string
	}
	tests := []struct {
		args   args
		expect bool
	}{
		{args{"", nil}, false},
		{args{"", []string{}}, false},
		{args{"xyz", s}, false},
		{args{"def", s}, true},
	}
	for _, tt := range tests {
		t.Run(tt.args.name, func(t *testing.T) {
			t.Parallel()
			got := helper.Finds(tt.args.name, tt.args.names...)
			be.Equal(t, got, tt.expect)
		})
	}
}

func TestIsFile(t *testing.T) {
	t.Parallel()
	self := filepath.Join(".", "helper_test.go")
	tests := []struct {
		name   string
		expect bool
	}{
		{self, true},
		{"^&%#$%@#", false},
		{"testdata/", false},
		{"testdata/TEST.DOC", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			be.Equal(t, helper.File(tt.name), tt.expect)
		})
	}
}
