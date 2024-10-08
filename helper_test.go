package helper_test

import (
	"bytes"
	"embed"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/Defacto2/helper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/encoding/unicode"
)

//go:embed testdata
var testdataFS embed.FS

func TestDetermineEncoding(t *testing.T) {
	e := helper.Determine(nil)
	assert.Nil(t, e)

	sr := strings.NewReader("Hello world!")
	e = helper.Determine(sr)
	assert.Equal(t, charmap.ISO8859_1, e)

	sr = strings.NewReader("Hello world 👾!!!")
	e = helper.Determine(sr)
	assert.Equal(t, unicode.UTF8, e, "wanted UTF-8 but got %s", e)

	p := []byte("")
	p = append(p, 0x1b)
	p = append(p, []byte("[31mHelloWorld")...)
	br := bytes.NewReader(p)
	e = helper.Determine(br)
	assert.Equal(t, charmap.ISO8859_1, e, "wanted ISO-8859-1 but got %s", e)

	sr = strings.NewReader("\nHello world!\n")
	e = helper.Determine(sr)
	assert.Equal(t, charmap.ISO8859_1, e, "wanted ISO-8859-1 but got %s", e)

	p = []byte("")
	p = append(p, 0xb2)
	p = append(p, []byte(" Hello world! ")...)
	p = append(p, 0xb2)
	br = bytes.NewReader(p)
	e = helper.Determine(br)
	assert.Equal(t, charmap.ISO8859_1, e, "wanted ISO-8859-1 but got %s", e)

	p = []byte("")
	p = append(p, 0x0D, 0x0E) // CP437 ♪ ♫
	p = append(p, []byte(" aah bah cah")...)
	br = bytes.NewReader(p)
	e = helper.Determine(br)
	assert.Equal(t, charmap.CodePage437, e, "wanted CP-437 but got %s", e)

	const house = 0x7f
	p = []byte("")
	p = append(p, house)
	p = append(p, []byte(" a DOS house glyph ")...)
	br = bytes.NewReader(p)
	e = helper.Determine(br)
	assert.Equal(t, charmap.CodePage437, e, "wanted CP-437 but got %s", e)

	const line = 0xc4
	p = []byte("")
	p = append(p, line, line, line, line, line, line)
	p = append(p, []byte(" a DOS line glyph ")...)
	br = bytes.NewReader(p)
	e = helper.Determine(br)
	assert.Equal(t, charmap.CodePage437, e, "wanted CP-437 but got %s", e)
}

func TestCookieStore(t *testing.T) {
	t.Parallel()
	b, err := helper.CookieStore("")
	require.NoError(t, err)
	l := utf8.RuneCount(b)
	assert.Equal(t, 32, l)

	const key = "my-secret-key"
	b, err = helper.CookieStore(key)
	require.NoError(t, err)
	assert.Len(t, b, len(key))
}

func TestLocalIPs(t *testing.T) {
	t.Parallel()
	ips, err := helper.LocalIPs()
	require.NoError(t, err)
	assert.NotEmpty(t, ips)
	// we can't test the actual IP addresses as they will be different on each machine.
}

func TestLocalHosts(t *testing.T) {
	t.Parallel()
	hosts, err := helper.LocalHosts()
	require.NoError(t, err)
	assert.NotEmpty(t, hosts)
	// we can't test the actual host names as they will be different on each machine.
}

func TestIntegrity(t *testing.T) {
	t.Parallel()
	x, err := helper.Integrity("", embed.FS{})
	require.Error(t, err)
	assert.Empty(t, x)
	x, err = helper.Integrity("nosuchfile", testdataFS)
	require.Error(t, err)
	assert.Empty(t, x)
	x, err = helper.Integrity("testdata/TEST.DOC", testdataFS)
	require.NoError(t, err)
	assert.Equal(t, "sha384-5X6isqmILTavQSao9DigKt3O8fX1Hd6hrGJ7pUROFPYWmkKRnFuWwTnjO3h9QkWP", x)
}

func TestIntegrityFile(t *testing.T) {
	t.Parallel()
	x, err := helper.IntegrityFile("")
	require.Error(t, err)
	assert.Empty(t, x)
	x, err = helper.IntegrityFile("nosuchfile")
	require.Error(t, err)
	assert.Empty(t, x)
	x, err = helper.IntegrityFile("testdata/TEST.DOC")
	require.NoError(t, err)
	assert.Equal(t, "sha384-5X6isqmILTavQSao9DigKt3O8fX1Hd6hrGJ7pUROFPYWmkKRnFuWwTnjO3h9QkWP", x)
}

func TestIntegrityBytes(t *testing.T) {
	t.Parallel()
	x := helper.IntegrityBytes(nil)
	assert.Equal(t, "sha384-OLBgp1GsljhM2TJ+sbHjaiH9txEUvgdDTAzHv2P24donTt6/529l+9Ua0vFImLlb", x)
	x = helper.IntegrityBytes([]byte("hello world"))
	assert.Equal(t, "sha384-/b2OdaZ/KfcBpOBAOF4uI5hjA+oQI5IRr5B/y7g1eLPkF8txzmRu/QgZ3YwIjeG9", x)
}

func TestLatency(t *testing.T) {
	result := helper.Latency()
	now := time.Now()
	assert.Less(t, *result, now)
}

func TestTimeDistance(t *testing.T) {
	now := time.Now()
	s := helper.TimeDistance(now, now, false)
	assert.Equal(t, "less than a minute", s)
	s = helper.TimeDistance(now, now.Add(time.Minute+time.Second), false)
	assert.Equal(t, "1 minute", s)
	s = helper.TimeDistance(now, now.Add(time.Second*2), true)
	assert.Equal(t, "less than 5 seconds", s)
	s = helper.TimeDistance(now, now.Add(time.Second*9), true)
	assert.Equal(t, "less than 10 seconds", s)
	s = helper.TimeDistance(now, now.Add(time.Second*19), true)
	assert.Equal(t, "less than 20 seconds", s)
	s = helper.TimeDistance(now, now.Add(time.Second*35), true)
	assert.Equal(t, "half a minute", s)
	s = helper.TimeDistance(now, now.Add(time.Second*60), true)
	assert.Equal(t, "1 minute", s)
	s = helper.TimeDistance(now, now.Add(time.Hour), true)
	assert.Equal(t, "about 1 hour", s)
	s = helper.TimeDistance(now, now.Add(time.Hour*24), true)
	assert.Equal(t, "1 day", s)
	s = helper.TimeDistance(now, now.Add(time.Hour*24*2), true)
	assert.Equal(t, "2 days", s)
	s = helper.TimeDistance(now, now.Add(time.Hour*24*30), true)
	assert.Equal(t, "about 1 month", s)
	s = helper.TimeDistance(now, now.Add(time.Hour*24*365), true)
	assert.Equal(t, "about 1 year", s)
	s = helper.TimeDistance(now, now.Add(time.Hour*24*500), true)
	assert.Equal(t, "over 1 year", s)
	s = helper.TimeDistance(now, now.Add(time.Hour*24*700), true)
	assert.Equal(t, "almost 2 years", s)
	s = helper.TimeDistance(now, now.Add(time.Hour*24*365*10), true)
	assert.Equal(t, "10 years", s)
}

func TestAdd1(t *testing.T) {
	tests := []struct {
		a         any
		expect    int64
		assertion assert.ComparisonAssertionFunc
	}{
		{0, 1, assert.Equal},
		{"xyz", 0, assert.Equal},
		{123, 124, assert.Equal},
		{1234567890, 1234567891, assert.Equal},
		{1234567890123456789, 1234567890123456790, assert.Equal},
	}
	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			assert.Equal(t, tt.expect, helper.Add1(tt.a))
		})
	}
}

func TestFileMatch(t *testing.T) {
	_, err := helper.FileMatch("", "")
	require.Error(t, err)
	v, err := helper.FileMatch("helper.go", "helper.go")
	require.NoError(t, err)
	assert.True(t, v)
	v, err = helper.FileMatch("helper_test.go", "helper.go")
	require.NoError(t, err)
	assert.False(t, v)
}

func TestFinds(t *testing.T) {
	s := []string{"abc", "def", "ghi"}
	type args struct {
		name  string
		names []string
	}
	tests := []struct {
		args      args
		expect    bool
		assertion assert.ComparisonAssertionFunc
	}{
		{args{"", nil}, false, assert.Equal},
		{args{"", []string{}}, false, assert.Equal},
		{args{"xyz", s}, false, assert.Equal},
		{args{"def", s}, true, assert.Equal},
	}
	for _, tt := range tests {
		t.Run(tt.args.name, func(t *testing.T) {
			tt.assertion(t, tt.expect, helper.Finds(tt.args.name, tt.args.names...))
		})
	}
}

func TestIsFile(t *testing.T) {
	self := filepath.Join(".", "helper_test.go")
	tests := []struct {
		name      string
		expect    bool
		assertion assert.ComparisonAssertionFunc
	}{
		{self, true, assert.Equal},
		{"^&%#$%@#", false, assert.Equal},
		{"testdata/", false, assert.Equal},
		{"testdata/TEST.DOC", true, assert.Equal},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.assertion(t, tt.expect, helper.File(tt.name))
		})
	}
}

func TestIsStat(t *testing.T) {
	self := filepath.Join(".", "helper_test.go")
	tests := []struct {
		name      string
		expect    bool
		assertion assert.ComparisonAssertionFunc
	}{
		{self, true, assert.Equal},
		{"^&%#$%@#", false, assert.Equal},
		{"testdata/", true, assert.Equal},
		{"testdata/TEST.DOC", true, assert.Equal},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.assertion(t, tt.expect, helper.Stat(tt.name))
		})
	}
}

func TestBools(t *testing.T) {
	assert.False(t, helper.Day(-1))
	assert.False(t, helper.Day(32))
	assert.True(t, helper.Day(1))
	assert.False(t, helper.Year(-1))
	assert.True(t, helper.Year(1970))
	assert.False(t, helper.Year(time.Now().Year()+1))
	// assert.False(t, ext.IsApp("myapp"))
	// assert.True(t, ext.IsApp("myapp.exe"))
	// assert.True(t, ext.IsArchive("stuff.zip"))
	// assert.True(t, ext.IsDocument("readme.doc"))
	// assert.True(t, ext.IsImage("cat.jpeg"))
	// assert.True(t, ext.IsHTML("index.html"))
	// assert.True(t, ext.IsAudio("song.wav"))
	// assert.True(t, ext.IsTune("song.mod"))
	// assert.True(t, ext.IsVideo("cat.divx"))
}

func TestDetermineFile(t *testing.T) {
	t.Parallel()

	r, err := os.Open("testdata/INFINITY.NFO")
	require.NoError(t, err)
	defer r.Close()

	e := helper.Determine(r)
	assert.Equal(t, charmap.CodePage437, e)
}
