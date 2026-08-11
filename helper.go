// Package helper has general and shared functions.
package helper

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"math"
	"net"
	"net/http"
	"os"
	"time"
)

const (
	// Eraseline is an ANSI escape control to erase the active line of the terminal.
	Eraseline = "\x1b[2K"
	// Timeout is the HTTP client timeout.
	Timeout = 5 * time.Second
	// UserAgent to send with the HTTP request.
	UserAgent = "Defacto2 website (thanks!)"

	byteUnits = "kMGTPE" // byteUnits is a list of units used for formatting byte sizes

	controlStart   = 0x00  // ASCII control character start
	controlEnd     = 0x1f  // ASCII control character end
	undefinedStart = 0x7f  // Latin-1 undefined characters start
	undefinedEnd   = 0x9f  // Latin-1 undefined characters end
	escape         = 0x1b  // ASCII escape control character
	unknownRune    = 65533 // Unicode replacement character (�)
	kcfAltEsc      = 0x9b  // Amiga had Keymap Qualifier Bits, which could be a typo to generate an Alt-Esc sequence?
	bell           = 0x07  // ASCII bell character that is sometimes found in Amiga ANSI files

	formFeed       = '\f'
	newline        = '\n'
	carriageReturn = '\r'
	tab            = '\t'
	verticalTab    = '\v'
)

var (
	ErrDiffLength = errors.New("files are of different lengths")
	ErrDirPath    = errors.New("directory path is a file")
	ErrExistPath  = errors.New("path ready exists and will not overwrite")
	ErrFileEmpty  = errors.New("file is empty with no content")
	ErrFilePath   = errors.New("file path is a directory")
	ErrKey        = errors.New("could not generate a random session key")
	ErrOSFile     = errors.New("os file is nil")
	ErrNoDir      = errors.New("not a directory")
	ErrRead       = errors.New("could not read files")
)

// Add1 returns the value of a + 1.
// The type of a must be a signed integer type or the result is 0.
func Add1(a any) int64 {
	switch val := a.(type) {
	case int:
		return int64(val) + 1
	case int8:
		return int64(val) + 1
	case int16:
		return int64(val) + 1
	case int32:
		return int64(val) + 1
	case int64:
		if val == math.MaxInt64 {
			return val // overflow wrap-around
		}
		return val + 1
	default:
		return 0
	}
}

// CookieStore generates a key for use with the sessions cookie store middleware.
// envKey is the value of an imported environment session key. But if it is empty,
// a 32-bit randomized value is generated that changes on every restart.
//
// The effect of using a randomized key will invalidate all existing sessions on every restart.
func CookieStore(envKey string) ([]byte, error) {
	if envKey != "" {
		key := []byte(envKey)
		return key, nil
	}

	const size = 32
	key := make([]byte, size)
	if _, err := rand.Read(key); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrKey, err)
	}

	return key, nil
}

// Day returns true if the i value can be used as a day time value.
func Day(i int) bool {
	const maxDay = 31
	return i > 0 && i <= maxDay
}

// Latency returns the stored, current local time.
func Latency() *time.Time {
	now := time.Now()
	return &now
}

// LocalIPs returns a list of local IP addresses.
//
// credit, [gosamples]
//
// [gosamples]: https://gosamples.dev/local-ip-address
func LocalIPs() ([]net.IP, error) {
	addresses, err := net.InterfaceAddrs()
	if err != nil {
		return nil, fmt.Errorf("net interface addresses: %w", err)
	}

	var ips []net.IP
	for _, addr := range addresses {
		ipnet, ok := addr.(*net.IPNet)
		if !ok || ipnet.IP.IsLoopback() {
			continue
		}

		// Ensure it is a valid IPv4 address and append the 4-byte slice
		if ip4 := ipnet.IP.To4(); ip4 != nil {
			ips = append(ips, ip4)
		}
	}

	return ips, nil
}

// LocalHosts returns a list of local hostnames.
func LocalHosts() ([]string, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return nil, fmt.Errorf("local hosts hostname: %w", err)
	}

	const size = 2
	hosts := make([]string, 0, size)
	hosts = append(hosts, hostname)

	ctx, cancel := context.WithTimeout(context.Background(), Timeout)
	defer cancel()

	const localhost = "localhost"
	resolver := net.Resolver{}
	if _, err := resolver.LookupHost(ctx, localhost); err == nil {
		if hostname != localhost {
			hosts = append(hosts, localhost)
		}
	}

	return hosts, nil
}

// Default HTTP client with connection reuse enabled.
var defaultPingClient = &http.Client{ //nolint:gochecknoglobals
	Timeout: Timeout,
}

// Ping sends a HTTP GET request to the provided URI, it returns the status code and response size.
func Ping(ctx context.Context, uri string) (int, int64, error) {
	const format = "helper ping %s %s: %w"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return 0, 0, fmt.Errorf(format, "new request", uri, err)
	}
	req.Header.Set("User-Agent", UserAgent)

	res, err := defaultPingClient.Do(req)
	if res != nil {
		defer res.Body.Close()
	}
	if err != nil {
		return 0, 0, fmt.Errorf(format, "client do", uri, err)
	}

	size, err := io.Copy(io.Discard, res.Body)
	if err != nil {
		return res.StatusCode, 0, fmt.Errorf(format, "body copy", uri, err)
	}

	return res.StatusCode, size, nil
}

// LocalHostPing sends a HTTP GET request to the provided URI on localhost.
func LocalHostPing(ctx context.Context, uri, proto string, port int) (int, int64, error) {
	url := fmt.Sprintf("%s://localhost:%d%s", proto, port, uri)
	return Ping(ctx, url)
}

// TimeDistance describes the difference between two time values.
// The seconds parameter determines if the string should handle less than a minute values.
func TimeDistance(from, to time.Time, seconds bool) string {
	// This function is a port of a CFWheels framework function programmed in ColdFusion (CFML).
	// https://github.com/cfwheels/cfwheels/blob/cf8e6da4b9a216b642862e7205345dd5fca34b54/wheels/global/misc.cfm#L112

	delta := to.Sub(from)
	if delta < 0 {
		delta = -delta
	}

	secs := int(delta.Seconds())
	mins := int(delta.Minutes())
	hrs := int(delta.Hours())

	const (
		hours    = 1440    // 1 day in minutes
		days     = 43200   // 30 days in minutes
		months   = 525600  // 365 days in minutes
		year     = 657000  // 456 days in minutes
		years    = 919800  // 638 days in minutes
		twoyears = 1051200 // 730 days in minutes
	)
	switch {
	case mins <= 1:
		return lessMin(secs, seconds)
	case mins < hours:
		return lessHours(mins, hrs)
	case mins < days:
		return lessDays(mins, hrs)
	case mins < months:
		return lessMonths(mins, hrs)
	case mins < year:
		return "about 1 year"
	case mins < years:
		return "over 1 year"
	case mins < twoyears:
		return "almost 2 years"
	default:
		y := mins / months
		return fmt.Sprintf("%d years", y)
	}
}

// lessMin returns a string describing the time difference in seconds or minutes.
func lessMin(secs int, seconds bool) string {
	if seconds {
		return lessMinAsSec(secs)
	}
	const minute = 60
	if secs < minute {
		return "less than a minute"
	}
	return "1 minute"
}

// lessMinAsSec returns a string describing the time difference in seconds.
func lessMinAsSec(secs int) string {
	const five, ten, twenty, forty = 5, 10, 20, 40
	switch {
	case secs < five:
		return "less than 5 seconds"
	case secs < ten:
		return "less than 10 seconds"
	case secs < twenty:
		return "less than 20 seconds"
	case secs < forty:
		return "half a minute"
	default:
		return "1 minute"
	}
}

// lessHours returns a string describing the time difference in hours.
func lessHours(mins, hrs int) string {
	const parthour, abouthour = 45, 90
	switch {
	case mins < parthour:
		return fmt.Sprintf("%d minutes", mins)
	case mins < abouthour:
		return "about 1 hour"
	default:
		if hrs == 0 {
			hrs = 1
		}
		return fmt.Sprintf("about %d hours", hrs)
	}
}

// lessDays returns a string describing the time difference in days.
func lessDays(mins, hrs int) string {
	const day = 2880
	if mins < day {
		return "1 day"
	}
	const hours = 24
	d := hrs / hours
	if d == 0 {
		d = 1
	}
	return fmt.Sprintf("%d days", d)
}

// lessMonths returns a string describing the time difference in months.
func lessMonths(mins, hrs int) string {
	const month = 86400
	if mins < month {
		return "about 1 month"
	}
	const hours = 730
	m := hrs / hours
	if m == 0 {
		m = 1
	}
	return fmt.Sprintf("%d months", m)
}

// Year returns true if the i value is between 1970 and (inclusive of) the current year.
func Year(i int) bool {
	const unix = 1970
	now := time.Now().Year()
	return i >= unix && i <= now
}

// LoggerKey is unused.
//
// Deprecated: As of release v1.5.
const LoggerKey string = "logger"

// Deprecated: As of release v1.5, this function returns nil.
func Logger(_ context.Context) any {
	return nil
}
