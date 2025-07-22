package helper_test

import (
	"bytes"
	"embed"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/Defacto2/helper"
	"github.com/nalgeon/be"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/charmap"
)

//go:embed testdata
var testdataFS embed.FS

var (
	latin1 encoding.Encoding = charmap.ISO8859_1   //nolint:gochecknoglobals
	cp437  encoding.Encoding = charmap.CodePage437 //nolint:gochecknoglobals
)

func TestDetermineEncoding_Unicode(t *testing.T) {
	t.Parallel()
	sr := strings.NewReader("Hello world 👾!!!")
	e := helper.Determine(sr)
	be.Equal(t, e, cp437)
}

func TestDetermineEncoding(t *testing.T) {
	t.Parallel()
	e := helper.Determine(nil)
	be.Equal(t, e, nil)
	sr := strings.NewReader("Hello world!")
	e = helper.Determine(sr)
	be.Equal(t, e, latin1)
	p := []byte("")
	p = append(p, 0x1b)
	p = append(p, []byte("[31mHelloWorld")...)
	br := bytes.NewReader(p)
	e = helper.Determine(br)
	be.Equal(t, e, latin1)
	sr = strings.NewReader("\nHello world!\n")
	e = helper.Determine(sr)
	be.Equal(t, e, latin1)
	p = []byte("")
	p = append(p, 0xb2)
	p = append(p, []byte(" Hello world! ")...)
	p = append(p, 0xb2)
	br = bytes.NewReader(p)
	e = helper.Determine(br)
	be.Equal(t, e, latin1)
	p = []byte("")
	p = append(p, 0x0D, 0x0E) // CP437 ♪ ♫
	p = append(p, []byte(" aah bah cah")...)
	br = bytes.NewReader(p)
	e = helper.Determine(br)
	be.Equal(t, e, cp437)
	const house = 0x7f
	p = []byte("")
	p = append(p, house)
	p = append(p, []byte(" a DOS house glyph ")...)
	br = bytes.NewReader(p)
	e = helper.Determine(br)
	be.Equal(t, e, cp437)
	const line = 0xc4
	p = []byte("")
	p = append(p, line, line, line, line, line, line)
	p = append(p, []byte(" a DOS line glyph ")...)
	br = bytes.NewReader(p)
	e = helper.Determine(br)
	be.Equal(t, e, cp437)
}

func TestCookieStore(t *testing.T) {
	t.Parallel()
	b, err := helper.CookieStore("")
	be.Err(t, err, nil)
	l := utf8.RuneCount(b)
	be.Equal(t, l, 32)

	const key = "my-secret-key"
	b, err = helper.CookieStore(key)
	be.Err(t, err, nil)
	be.Equal(t, len(key), len(b))
}

func TestLocalIPs(t *testing.T) {
	t.Parallel()
	ips, err := helper.LocalIPs()
	// we can't test the actual IP addresses as they will be different on each machine.
	be.Err(t, err, nil)
	be.True(t, len(ips) > 0)
}

func TestLocalHosts(t *testing.T) {
	t.Parallel()
	hosts, err := helper.LocalHosts()
	be.Err(t, err, nil)
	be.True(t, len(hosts) > 0)
	// we can't test the actual host names as they will be different on each machine.
}

func TestIntegrity(t *testing.T) {
	t.Parallel()
	s, err := helper.Integrity("", embed.FS{})
	be.Err(t, err)
	be.Equal(t, s, "")
	s, err = helper.Integrity("nosuchfile", testdataFS)
	be.Err(t, err)
	be.Equal(t, s, "")
	s, err = helper.Integrity("testdata/TEST.DOC", testdataFS)
	be.Err(t, err, nil)
	be.Equal(t, s, "sha384-5X6isqmILTavQSao9DigKt3O8fX1Hd6hrGJ7pUROFPYWmkKRnFuWwTnjO3h9QkWP")
}

func TestIntegrityFile(t *testing.T) {
	t.Parallel()
	s, err := helper.IntegrityFile("")
	be.Err(t, err)
	be.Equal(t, s, "")
	s, err = helper.IntegrityFile("nosuchfile")
	be.Err(t, err)
	be.Equal(t, s, "")
	s, err = helper.IntegrityFile("testdata/TEST.DOC")
	be.Err(t, err, nil)
	be.Equal(t, s, "sha384-5X6isqmILTavQSao9DigKt3O8fX1Hd6hrGJ7pUROFPYWmkKRnFuWwTnjO3h9QkWP")
}

func TestIntegrityBytes(t *testing.T) {
	t.Parallel()
	x := helper.IntegrityBytes(nil)
	be.Equal(t, x, "sha384-OLBgp1GsljhM2TJ+sbHjaiH9txEUvgdDTAzHv2P24donTt6/529l+9Ua0vFImLlb")
	x = helper.IntegrityBytes([]byte("hello world"))
	be.Equal(t, x, "sha384-/b2OdaZ/KfcBpOBAOF4uI5hjA+oQI5IRr5B/y7g1eLPkF8txzmRu/QgZ3YwIjeG9")
}

func TestLatency(t *testing.T) {
	t.Parallel()
	result := helper.Latency()
	older := result.Before(time.Now())
	be.True(t, older)
}

func TestTimeDistance(t *testing.T) {
	t.Parallel()
	now := time.Now()
	s := helper.TimeDistance(now, now, false)
	be.Equal(t, s, "less than a minute")
	s = helper.TimeDistance(now, now.Add(time.Minute+time.Second), false)
	be.Equal(t, s, "1 minute")
	s = helper.TimeDistance(now, now.Add(time.Second*2), true)
	be.Equal(t, s, "less than 5 seconds")
	s = helper.TimeDistance(now, now.Add(time.Second*9), true)
	be.Equal(t, s, "less than 10 seconds")
	s = helper.TimeDistance(now, now.Add(time.Second*19), true)
	be.Equal(t, s, "less than 20 seconds")
	s = helper.TimeDistance(now, now.Add(time.Second*35), true)
	be.Equal(t, s, "half a minute")
	s = helper.TimeDistance(now, now.Add(time.Second*60), true)
	be.Equal(t, s, "1 minute")
	s = helper.TimeDistance(now, now.Add(time.Hour), true)
	be.Equal(t, s, "about 1 hour")
	s = helper.TimeDistance(now, now.Add(time.Hour*24), true)
	be.Equal(t, s, "1 day")
	s = helper.TimeDistance(now, now.Add(time.Hour*24*2), true)
	be.Equal(t, s, "2 days")
	s = helper.TimeDistance(now, now.Add(time.Hour*24*30), true)
	be.Equal(t, s, "about 1 month")
	s = helper.TimeDistance(now, now.Add(time.Hour*24*365), true)
	be.Equal(t, s, "about 1 year")
	s = helper.TimeDistance(now, now.Add(time.Hour*24*500), true)
	be.Equal(t, s, "over 1 year")
	s = helper.TimeDistance(now, now.Add(time.Hour*24*700), true)
	be.Equal(t, s, "almost 2 years")
	s = helper.TimeDistance(now, now.Add(time.Hour*24*365*10), true)
	be.Equal(t, s, "10 years")
}

func TestAdd1(t *testing.T) {
	t.Parallel()
	tests := []struct {
		a      any
		expect int64
	}{
		{0, 1},
		{"xyz", 0},
		{123, 124},
		{1234567890, 1234567891},
		{1234567890123456789, 1234567890123456790},
	}
	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			t.Parallel()
			be.Equal(t, helper.Add1(tt.a), tt.expect)
		})
	}
}

func TestFileMatch(t *testing.T) {
	t.Parallel()
	_, err := helper.FileMatch("", "")
	be.Err(t, err)
	v, err := helper.FileMatch("helper.go", "helper.go")
	be.Err(t, err, nil)
	be.True(t, v)
	v, err = helper.FileMatch("helper_test.go", "helper.go")
	be.Err(t, err, nil)
	be.True(t, !v)
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
			b := helper.Finds(tt.args.name, tt.args.names...)
			be.Equal(t, b, tt.expect)
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

func TestIsStat(t *testing.T) {
	t.Parallel()
	self := filepath.Join(".", "helper_test.go")
	tests := []struct {
		name   string
		expect bool
	}{
		{self, true},
		{"^&%#$%@#", false},
		{"testdata/", true},
		{"testdata/TEST.DOC", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			be.Equal(t, helper.Stat(tt.name), tt.expect)
		})
	}
}

func TestBools(t *testing.T) {
	t.Parallel()
	be.True(t, !helper.Day(-1))
	be.True(t, !helper.Day(32))
	be.True(t, helper.Day(1))
	be.True(t, !helper.Year(-1))
	be.True(t, helper.Year(1970))
	be.True(t, !helper.Year(time.Now().Year()+1))
}

func TestDetermineFile(t *testing.T) {
	t.Parallel()

	// This is a CP-437 file that can also be read as ISO-8859-1.
	r, err := os.Open("testdata/PKZ80A1.TXT")
	be.Err(t, err, nil)
	defer r.Close()
	e := helper.Determine(r)
	be.Equal(t, e, latin1)
}

func TestLocalHostPing(t *testing.T) {
	t.Parallel()
	tests := []struct {
		uri       string
		proto     string
		port      int
		expect    int
		expectErr bool
	}{
		{"/", "http", 80, http.StatusInternalServerError, true},
		{"/", "http", 7654, http.StatusInternalServerError, true},
		{"/", "https", 443, http.StatusInternalServerError, true},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s://localhost:%d%s", tt.proto, tt.port, tt.uri), func(t *testing.T) {
			t.Parallel()
			status, size, err := helper.LocalHostPing(tt.uri, tt.proto, tt.port)
			if tt.expectErr {
				be.Err(t, err)
			} else {
				be.Err(t, err, nil)
			}
			be.Equal(t, status, tt.expect)
			be.True(t, size >= int64(0))
		})
	}
}
