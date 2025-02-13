package helper_test

import (
	"crypto/rand"
	"os"
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
