package helper_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Defacto2/helper"
	"github.com/nalgeon/be"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/charmap"
	uni "golang.org/x/text/encoding/unicode"
)

func TestDetermineFile0(t *testing.T) {
	t.Parallel()

	// This is a CP-437 file that can also be read as ISO-8859-1.
	name := filepath.Join(testdata, "PKZ80A1.TXT")
	r, err := os.Open(name)
	be.Err(t, err, nil)
	defer r.Close()

	be.Equal(t, helper.Determine(r), latin1)
}

func TestDetermineFile1(t *testing.T) {
	t.Parallel()

	// This is a basic ASCII text file.
	name := filepath.Join(testdata, "TEST.DOC")
	r, err := os.Open(name)
	be.Err(t, err, nil)
	defer r.Close()

	be.Equal(t, helper.Determine(r), latin1)
}

func TestDetermineFile2(t *testing.T) {
	t.Parallel()

	// Binary files use most or all of the 256 characters in an 8-bit codepage map.
	// Without detecting magic bytes, it will be confused for a CP-437 text.
	name := filepath.Join(testdata, "TEST.BMP")
	r, err := os.Open(name)
	be.Err(t, err, nil)
	defer r.Close()

	be.Equal(t, helper.Determine(r), cp437)
}

func TestDetermineEncoding_Unicode(t *testing.T) {
	t.Parallel()

	// Like binary files, without looking for a Byte-Order-Mark,
	// it is hard to determine the difference between a codepage
	// character map text and a UTF-8 encoded document.
	sr := strings.NewReader("Hello world 👾!!!")
	got := helper.Determine(sr)
	be.Equal(t, got, cp437)
}

// Test the fast BOM path specifically.
func TestDetermineEncoding_BOM(t *testing.T) {
	t.Parallel()

	// UTF-8 BOM should be detected instantly.
	withBOM := strings.NewReader("\xEF\xBB\xBFHello World")
	got := helper.Determine(withBOM)
	be.Equal(t, got, uni.UTF8)
}

func TestDetermineEncoding(t *testing.T) {
	t.Parallel()

	be.Equal(t, helper.Determine(nil), nil)

	sr := strings.NewReader("Hello world!")
	be.Equal(t, helper.Determine(sr), latin1)

	p := []byte("")
	p = append(p, 0x1b)
	p = append(p, []byte("[31mHelloWorld")...)
	r := bytes.NewReader(p)
	be.Equal(t, helper.Determine(r), latin1)

	p = []byte("\nHello world!\n")
	r = bytes.NewReader(p)
	be.Equal(t, helper.Determine(r), latin1)

	p = []byte("")
	p = append(p, 0xb2)
	p = append(p, []byte(" Hello world! ")...)
	p = append(p, 0xb2)
	r = bytes.NewReader(p)
	be.Equal(t, helper.Determine(r), latin1)

	p = []byte("")
	p = append(p, 0x0D, 0x0E) // CP437 ♪ ♫
	p = append(p, []byte(" aah bah cah")...)
	r = bytes.NewReader(p)

	be.Equal(t, helper.Determine(r), cp437)
	const house = 0x7f
	p = []byte("")
	p = append(p, house)
	p = append(p, []byte(" a DOS house glyph ")...)
	r = bytes.NewReader(p)
	be.Equal(t, helper.Determine(r), cp437)

	const line = 0xc4
	p = []byte("")
	p = append(p, line, line, line, line, line, line)
	p = append(p, []byte(" a DOS line glyph ")...)
	r = bytes.NewReader(p)
	be.Equal(t, helper.Determine(r), cp437)
}

func TestDetermineEncoding_falsepositives(t *testing.T) {
	t.Parallel()

	const arabic = "ڿ"
	p := []byte(arabic)
	p = append(p, []byte(" a DOS/Unicode false-positive glyph ")...)
	r := bytes.NewReader(p)
	be.Equal(t, helper.Determine(r), cp437)

	const amiga = 0x9b
	p = []byte("")
	p = append(p, amiga)
	p = append(p, []byte(" an Amiga false-positive glyph ")...)
	r = bytes.NewReader(p)
	be.Equal(t, helper.Determine(r), latin1)
}

func TestSequences(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input []byte
		want  encoding.Encoding
	}{
		{
			name:  "detect double horizontal bar (4 bytes)",
			input: []byte{0xcd, 0xcd, 0xcd, 0xcd},
			want:  charmap.CodePage437,
		},
		{
			name:  "detect guillemets pair",
			input: []byte{0xae, 0xaf},
			want:  charmap.CodePage437,
		},
		{
			name:  "standard ASCII text returns nil",
			input: []byte("Hello, World!"),
			want:  nil,
		},
		{
			name:  "single CP-437 char (below threshold) returns nil",
			input: []byte{0xdb, 0xdb}, // fullBlock needs 4, 2 should fail
			want:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Test with nil logger (ensures no panics)
			got := helper.DetermineSequences(nil, tt.input)
			be.Equal(t, got, tt.want)
		})
	}
}

func TestChars(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input []byte
		want  encoding.Encoding
	}{
		{
			name:  "standard plain ASCII returns nil",
			input: []byte("Hello World!\n\r\t"),
			want:  nil,
		},
		{
			name:  "bullet character triggers CP-437",
			input: []byte{'H', 'e', 'l', 'l', 'o', 0xf9},
			want:  charmap.CodePage437,
		},
		{
			name:  "interpunct character triggers CP-437",
			input: []byte{0xfa},
			want:  charmap.CodePage437,
		},
		{
			name:  "control character glyph triggers CP-437",
			input: []byte{0x00},
			want:  charmap.CodePage437,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Test with nil logger to guarantee no nil pointer panics
			got := helper.DetermineChar(nil, tt.input)
			be.Equal(t, got, tt.want)
		})
	}
}
