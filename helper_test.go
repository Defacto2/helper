//nolint:gochecknoglobals,nonamedreturns
package helper_test

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Defacto2/helper"
	"github.com/nalgeon/be"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/charmap"
)

//go:embed testdata
var testdataFS embed.FS

var (
	latin1 encoding.Encoding = charmap.ISO8859_1
	cp437  encoding.Encoding = charmap.CodePage437
)

const (
	testdataBad    = "no_such-path!" // intended to be used as an invalid file or directory path
	testdataCount  = 3               // the expected number of files found in the helper/testdata directory
	testdataBMP    = 750_054         // the byte count of the textdata file helper/textdata/TEST.BMP
	testdataBMP384 = `cfa5f5f91417786fd4d63d79e82e613e355621dc8759` +
		`b616f53a1be738880d2ea6da25ae1fef13de3174903d1818f3a2` // the result of sha384hmac -u TEST.BMP
	testUNID = "00000000-0000-0000-0000-000000000000" // common universal unique identifier example
	testCUID = "00000000-0000-0000-0000000000000000"  // coldfusion uuid example
)

// testdata returns the absolute path to the helper/testdata directory.
var testdata = func() string {
	const format = "testdata %s: %v"
	dir, err := filepath.Abs("testdata")
	if err != nil {
		panic(fmt.Sprintf(format, "absolute path failed", err))
	}
	st, err := os.Stat(dir)
	if err != nil {
		panic(fmt.Sprintf(format, "missing or unreadable "+dir, err))
	}
	if !st.IsDir() {
		panic("testdata is not a directory " + dir)
	}
	return dir
}()

// createCombo creates a temporary directory containing a single text file.
// The directory gets cleaned up after use.
//
//   - The returned size is the file size of the created text file.
//   - The dir is the absolute path to the temporary directory.
//   - The src is the absolute path to the src.txt file in the temporary directory.
//   - The dst is the absolute path to a possible target destination file, if warranted.
func createCombo(tb testing.TB) (size int, dir, src, dst string) {
	tb.Helper()

	dir = tb.TempDir()
	src = filepath.Join(dir, "src.txt")
	dst = filepath.Join(dir, "dst.txt")

	const s = "Hello world!"
	data := []byte(s)
	size, err := helper.TouchW(src, data...)
	be.Err(tb, err, nil)
	be.Equal(tb, len(data), size)

	return
}

// cleanup removes the path and its content and logs any errors.
func cleanup(tb testing.TB, path string) {
	tb.Helper()

	if err := os.RemoveAll(path); err != nil {
		tb.Logf("could not remove the path %s: %v", path, err)
	}
}

func TestCookieStore(t *testing.T) {
	t.Parallel()
	b, err := helper.CookieStore("")
	be.Err(t, err, nil)
	be.Equal(t, len(b), 32)

	const key = "my-secret-key"
	b, err = helper.CookieStore(key)
	be.Err(t, err, nil)
	be.Equal(t, len(key), len(b))
}

func TestLocalIPs(t *testing.T) {
	t.Parallel()

	ips, err := helper.LocalIPs()
	be.Err(t, err, nil)

	for _, ip := range ips {
		be.True(t, ip.To4() != nil)
		be.True(t, !ip.IsLoopback())
		be.True(t, !ip.IsUnspecified())
	}
}

func TestLocalHosts(t *testing.T) {
	t.Parallel()

	hosts, err := helper.LocalHosts()
	be.Err(t, err, nil)
	be.True(t, len(hosts) > 0)

	// ensure hostnames are not empty
	for _, h := range hosts {
		be.True(t, len(h) > 0)
	}
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
		{-99, -98},
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

func TestBools(t *testing.T) {
	t.Parallel()

	be.True(t, helper.Day(1))
	be.True(t, helper.Year(1970))
	be.True(t, !helper.Day(-1))                   // negative day
	be.True(t, !helper.Day(32))                   // invalid day
	be.True(t, !helper.Year(-1))                  // negative year
	be.True(t, !helper.Year(time.Now().Year()+1)) // one year into the future
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

// TestByteCountEdgeCases tests extremely large numbers that could overflow.
func TestByteCountEdgeCases(t *testing.T) {
	t.Parallel()

	huge := int64(1) * 1024 * 1024 * 1024 * 1024 * 1024 * 1024 // 1 exabyte

	got := helper.ByteCount(huge)
	be.True(t, len(got) > 0)
	be.True(t, strings.ContainsAny(got, "KMGTPE"))
}

// TestTimeDistanceZeroHoursFix verifies that 2 hours doesn't return "0 hours".
func TestTimeDistanceZeroHoursFix(t *testing.T) {
	t.Parallel()

	base := time.Date(2020, 1, 1, 12, 0, 0, 0, time.UTC)
	got := helper.TimeDistance(base, base.Add(2*time.Hour), false)

	// Must not contain "0 hours"
	be.True(t, !strings.Contains(got, "0 hours"))
	// Must contain "2 hours"
	be.True(t, strings.Contains(got, "2 hours"))
}
