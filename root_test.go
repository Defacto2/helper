package helper_test

import (
	"crypto/rand"
	"os"
	"path/filepath"
	"testing"

	"github.com/Defacto2/helper"
	"github.com/stretchr/testify/assert"
)

func TestDuplicater(t *testing.T) {
	t.Parallel()
	const oldf = "oldfile.txt"
	const newf = "newfile.txt"
	tests := []struct {
		name    string
		oldPath string
		newPath string
		want    int64
		wantErr bool
	}{
		{"missing", "", newf, 0, true},
		{"missing", oldf, "", 0, true},
		{"ok", oldf, newf, 26, false},
		{"readonly", oldf, oldf, 0, true},
	}
	for _, tt := range tests {
		r, err := os.OpenRoot(t.TempDir())
		if err != nil {
			t.Fatalf("Failed to open root: %v", err)
		}
		f, err := r.OpenFile(oldf, os.O_CREATE|os.O_RDWR, 0o777)
		if err != nil {
			t.Fatalf("Failed to create file: %v", err)
		}
		text := rand.Text()
		n, err := f.WriteString(text)
		if err != nil {
			t.Fatalf("Failed to write file: %v", err)
		}
		if n != len(text) {
			t.Fatalf("Failed to write file: %v", err)
		}
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			i, err := helper.Duplicater(r, tt.oldPath, tt.newPath)
			if (err != nil) != tt.wantErr {
				t.Errorf("Duplicater() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.want, i)
		})
		t.Cleanup(func() {
			f.Close()
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
			assert.EqualValues(t, tt.want, got)
		})
	}
}

func TestUTF8(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		input   []byte
		want    bool
		wantErr bool
	}{
		{"empty", nil, false, true},
		{"ascii", []byte("hello"), true, false},
		{"latin1", []byte("fractions \xbc \xbd \xbe"), false, false},
		{"utf8", []byte("fractions \u00bc \u00bd \u00be"), true, false},
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
			assert.Equal(t, tt.want, got)
		})
	}
}
