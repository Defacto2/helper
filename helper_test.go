package helper_test

import (
	"bytes"
	"embed"
	"fmt"
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
	uni "golang.org/x/text/encoding/unicode"
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
	got := helper.Determine(nil)
	be.Equal(t, got, nil)
	sr := strings.NewReader("Hello world!")
	got = helper.Determine(sr)
	be.Equal(t, got, latin1)
	p := []byte("")
	p = append(p, 0x1b)
	p = append(p, []byte("[31mHelloWorld")...)
	br := bytes.NewReader(p)
	got = helper.Determine(br)
	be.Equal(t, got, latin1)
	sr = strings.NewReader("\nHello world!\n")
	got = helper.Determine(sr)
	be.Equal(t, got, latin1)
	p = []byte("")
	p = append(p, 0xb2)
	p = append(p, []byte(" Hello world! ")...)
	p = append(p, 0xb2)
	br = bytes.NewReader(p)
	got = helper.Determine(br)
	be.Equal(t, got, latin1)
	p = []byte("")
	p = append(p, 0x0D, 0x0E) // CP437 ♪ ♫
	p = append(p, []byte(" aah bah cah")...)
	br = bytes.NewReader(p)
	got = helper.Determine(br)
	be.Equal(t, got, cp437)
	const house = 0x7f
	p = []byte("")
	p = append(p, house)
	p = append(p, []byte(" a DOS house glyph ")...)
	br = bytes.NewReader(p)
	got = helper.Determine(br)
	be.Equal(t, got, cp437)
	const line = 0xc4
	p = []byte("")
	p = append(p, line, line, line, line, line, line)
	p = append(p, []byte(" a DOS line glyph ")...)
	br = bytes.NewReader(p)
	got = helper.Determine(br)
	be.Equal(t, got, cp437)
}

func TestCookieStore(t *testing.T) {
	t.Parallel()
	b, err := helper.CookieStore("")
	be.Err(t, err, nil)
	got := utf8.RuneCount(b)
	be.Equal(t, got, 32)

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

func TestLatency(t *testing.T) {
	t.Parallel()
	result := helper.Latency()
	got := result.Before(time.Now())
	be.True(t, got)
}

func TestTimeDistance(t *testing.T) {
	t.Parallel()
	now := time.Now()
	got := helper.TimeDistance(now, now, false)
	be.Equal(t, got, "less than a minute")
	got = helper.TimeDistance(now, now.Add(time.Minute+time.Second), false)
	be.Equal(t, got, "1 minute")
	got = helper.TimeDistance(now, now.Add(time.Second*2), true)
	be.Equal(t, got, "less than 5 seconds")
	got = helper.TimeDistance(now, now.Add(time.Second*9), true)
	be.Equal(t, got, "less than 10 seconds")
	got = helper.TimeDistance(now, now.Add(time.Second*19), true)
	be.Equal(t, got, "less than 20 seconds")
	got = helper.TimeDistance(now, now.Add(time.Second*35), true)
	be.Equal(t, got, "half a minute")
	got = helper.TimeDistance(now, now.Add(time.Second*60), true)
	be.Equal(t, got, "1 minute")
	got = helper.TimeDistance(now, now.Add(time.Hour), true)
	be.Equal(t, got, "about 1 hour")
	got = helper.TimeDistance(now, now.Add(time.Hour*24), true)
	be.Equal(t, got, "1 day")
	got = helper.TimeDistance(now, now.Add(time.Hour*24*2), true)
	be.Equal(t, got, "2 days")
	got = helper.TimeDistance(now, now.Add(time.Hour*24*30), true)
	be.Equal(t, got, "about 1 month")
	got = helper.TimeDistance(now, now.Add(time.Hour*24*365), true)
	be.Equal(t, got, "about 1 year")
	got = helper.TimeDistance(now, now.Add(time.Hour*24*500), true)
	be.Equal(t, got, "over 1 year")
	got = helper.TimeDistance(now, now.Add(time.Hour*24*700), true)
	be.Equal(t, got, "almost 2 years")
	got = helper.TimeDistance(now, now.Add(time.Hour*24*365*10), true)
	be.Equal(t, got, "10 years")
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
	got, err := helper.FileMatch("helper.go", "helper.go")
	be.Err(t, err, nil)
	be.True(t, got)
	got, err = helper.FileMatch("helper_test.go", "helper.go")
	be.Err(t, err, nil)
	be.True(t, !got)
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

	// This is a CP-437 file that can also be read as ISO-8859-1..
	r, err := os.Open("testdata/PKZ80A1.TXT")
	be.Err(t, err, nil)
	defer r.Close()
	got := helper.Determine(r)
	be.Equal(t, got, latin1)
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
		{"/", "http", 80, 0, true},
		{"/", "http", 7654, 0, true},
		{"/", "https", 443, 0, true},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s://localhost:%d%s", tt.proto, tt.port, tt.uri), func(t *testing.T) {
			t.Parallel()
			got, size, err := helper.LocalHostPing(t.Context(), tt.uri, tt.proto, tt.port)
			if tt.expectErr {
				be.Err(t, err)
			} else {
				be.Err(t, err, nil)
			}
			be.Equal(t, got, tt.expect)
			be.True(t, size >= int64(0))
		})
	}
}

// TestByteCountEdgeCases tests extremely large numbers that would overflow byteUnits..
func TestByteCountEdgeCases(t *testing.T) {
	t.Parallel()
	huge := int64(1) * 1024 * 1024 * 1024 * 1024 * 1024 * 1024 // 1 exabyte
	got := helper.ByteCount(huge)
	be.True(t, len(got) > 0)
	// Verify it doesn't panic or produce invalid output
	be.True(t, strings.ContainsAny(got, "KMGTPE"))
}

// TestTimeDistanceZeroHoursFix verifies that 2 hours doesn't return "0 hours"..
func TestTimeDistanceZeroHoursFix(t *testing.T) {
	t.Parallel()
	base := time.Date(2020, 1, 1, 12, 0, 0, 0, time.UTC)
	got := helper.TimeDistance(base, base.Add(2*time.Hour), false)
	// Should NOT contain "0 hours"
	be.True(t, !strings.Contains(got, "0 hours"))
	// Should contain "2 hours"
	be.True(t, strings.Contains(got, "2 hours"))
}
