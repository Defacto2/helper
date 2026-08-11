// Package helper has general and shared functions.
package helper

import (
	"bytes"
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"maps"
	"math/big"
	"net"
	"net/http"
	"os"
	"time"
	"unicode"
	"unicode/utf8"

	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/charmap"
	uni "golang.org/x/text/encoding/unicode"
	"golang.org/x/text/unicode/runenames"
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
	house          = 0x7f  // CP-437 house character that displays a unique glyph in the Amiga Topaz font

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
// The type of a must be an integer type or the result is 0.
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

	const letters = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	const length = 32
	key := make([]byte, length)
	n, err := rand.Read(key)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrKey, err.Error())
	}
	if n != length {
		return nil, ErrKey
	}

	for i, b := range key {
		key[i] = letters[b%byte(len(letters))]
	}

	return key, nil
}

// Day returns true if the i value can be used as a day time value.
func Day(i int) bool {
	const maxDay = 31
	if i > 0 && i <= maxDay {
		return true
	}
	return false
}

// Determine returns the encoding of the plain text byte slice, either
// the charmap.ISO8859_1 or charmap.CodePage437 encoding is returned.
//
// Without false-positives, is difficult to determine the encoding of a text slice without
// a BOM or other metadata, especially a legacy, 8-bit code page encoding vs UTF-8 encoding.
// For example, the 👾 (alien monster) emoji in UTF-8 is comprised
// of the bytes 240, 159, 145, 190, which are all valid CP-437 characters.
//
//	"👾"	// [240 159 145 190] unicode.UTF8
//	"≡ƒæ╛"	// [240 159 145 190] charmap.CodePage437
func Determine(r io.Reader) encoding.Encoding { //nolint:ireturn
	sl := slog.New(slog.DiscardHandler)
	return determine(sl, r)
}

// DetermineS functions the same as Determine, however you can provide a
// slog logger to track character or sequence matches for false-positive
// discoveries and other possible problems.
func DetermineS(sl *slog.Logger, r io.Reader) encoding.Encoding { //nolint:ireturn
	if sl == nil {
		sl = slog.New(slog.DiscardHandler)
	}
	return determine(sl, r)
}

func determine(sl *slog.Logger, r io.Reader) encoding.Encoding { //nolint:ireturn,cyclop
	const msg = "helper determine r encoding"
	if sl == nil {
		sl = slog.New(slog.DiscardHandler)
	}
	if r == nil {
		sl.Info(msg, slog.Bool("empty reader", true))
		return nil
	}
	p, err := io.ReadAll(r)
	if err != nil {
		sl.Info(msg, slog.Any("readall rune", err))
		return nil
	}
	sl.Info(msg, slog.Int("bytes", len(p)))
	if e := suppliment(sl, p); e != nil {
		return e
	}
	if e := chars(sl, p); e != nil {
		return e
	}
	if e := sequences(sl, p); e != nil {
		return e
	}
	// Check for Unicode multi-byte characters
	// If an unknown rune is encountered then assume the encoding is
	// using a legacy 8-bit code page encoding, such as CP-437.
	tick := time.Now()
	for _, r := range bytes.Runes(p) {
		if utf8.RuneLen(r) > 1 {
			switch {
			// we use this switch to handle any obvious false-positives
			case unicode.Is(unicode.Arabic, r):
				sl.Info(msg, slog.Bool("arabic rune", true),
					slog.Duration("time", time.Since(tick)))
				// '┌┐' cp437 char sequence gets mistaken as a multi-byte Arabic ڿ script
				return charmap.CodePage437
			case r == unknownRune:
				sl.Info(msg, slog.Bool("unknown rune", true),
					slog.Duration("time", time.Since(tick)))
				return charmap.ISO8859_1
			}
			sl.Info(msg, slog.Bool("multi-byte rune", true),
				slog.Duration("time", time.Since(tick)))
			return uni.UTF8
		}
	}
	sl.Info(msg, slog.Bool("latin-1", true),
		slog.Duration("time", time.Since(tick)))
	return charmap.ISO8859_1
}

// Suppliment returns the encoding of UTF8 if p contains the following
// Unicode characters.
//
//	•─█
func suppliment(sl *slog.Logger, p []byte) encoding.Encoding { //nolint:ireturn
	const msg = "helper determine p unicode suppliments"
	const bullet = '\u2022'    // •
	const lightHoz = '\u2500'  // ─
	const fullBlock = '\u2588' // █
	f := func(r rune) bool {
		return r == lightHoz || r == bullet || r == fullBlock
	}
	if bytes.ContainsFunc(p, f) {
		sl.Info(msg, slog.Bool("found match", true))
		return uni.UTF8
	}
	return nil
}

// Chars returns the encoding based on the presence of common CP-437 or ISO-8859-1 characters.
// A nil encoding is returned if no encoding is determined.
//
// This should be done before checking for multi-byte characters, which could be misinterpreted as UTF-8 runes.
func chars(sl *slog.Logger, p []byte) encoding.Encoding { //nolint:ireturn
	const msg = "helper determine p chars"
	const bullet, interpunct = 0xf9, 0xfa
	tick := time.Now()
	for i, char := range p {
		switch {
		case char == escape:
			// escape control character commonly used in ANSI escaped sequences
			continue
		case // oddball control characters that are sometimes found in Amiga ANSI files
			char == kcfAltEsc,
			char == bell:
			continue
		case // common whitespace control characters
			char == formFeed,
			char == newline,
			char == carriageReturn,
			char == tab,
			char == verticalTab:
			continue
		// case r == unknownRune:
		// 	// when an unknown extended-ASCII character (128-255) is encountered
		// 	continue
		case char >= undefinedStart && char <= undefinedEnd:
			// unused ASCII, which we can probably assumed to be CP-437
			logChar(sl, msg, tick, i, char)
			return charmap.CodePage437
		case char >= controlStart && char <= controlEnd:
			// ASCII control characters, which we can probably assumed to be CP-437 glyphs
			logChar(sl, msg, tick, i, char)
			return charmap.CodePage437
		case char == interpunct, char == bullet:
			return charmap.CodePage437
		}
	}
	sl.Info(msg, slog.Duration("time", time.Since(tick)))
	return nil
}

func logChar(sl *slog.Logger, msg string, tick time.Time, i int, char byte) {
	r := rune(char)
	sl.Info(msg, slog.String("character",
		fmt.Sprintf("%d> %s 0x%X %d %s %s", i,
			string(char), char, char, string(r), runenames.Name(r))),
		slog.Duration("time", time.Since(tick)))
}

// Sequences returns the encoding based on the presence of common CP-437 or ISO-8859-1 character sequences.
// Full block, medium shade, horizontal bars and half blocks are sequences of characters that are often
// unique to the CP-437 encoding.
//
// This should be done before checking for multi-byte characters, which could be misinterpreted as UTF-8 runes.
func sequences(sl *slog.Logger, p []byte) encoding.Encoding { //nolint:ireturn
	const msg = "helper determine p seqs"
	const (
		shadeLight     = 0xb0 // ░ ~ °
		shadeMedium    = 0xb1 // ▒ ~ ±
		shadeDark      = 0xb2 // ▓ ~ ²
		singleHorizBar = 0xc4 // ─ ~ Ä
		doubleHorizBar = 0xcd // ═ ~ Í
		fullBlock      = 0xdb // █ ~ Û
		lowerHalfBlock = 0xdc // ▄ ~ Ü
		upperHalfBlock = 0xdf // ▀ ~ ß
		interpunct     = 0xfa // · ~ ú
		bulletpoint    = 0xf9 // • ~ ù
	)
	// four and pairs are counts of unlikely sequences of characters
	// for example, in normal use the degree sign is appended to digits
	// and wouldn't be used as a pair "20°°"
	const four, pair = 4, 2
	matches := map[uint8]int{
		shadeLight:     pair,
		shadeMedium:    pair,
		shadeDark:      pair,
		singleHorizBar: pair,
		doubleHorizBar: four,
		fullBlock:      four,
		lowerHalfBlock: four,
		upperHalfBlock: pair,
		interpunct:     pair,
		bulletpoint:    pair,
	}
	tick := time.Now()
	for char, count := range maps.All(matches) {
		subslice := bytes.Repeat([]byte{char}, count)
		if bytes.Contains(p, subslice) {
			sl.Info(msg,
				slog.String("matched subslice", string(subslice)),
				slog.Duration("time", time.Since(tick)))
			return charmap.CodePage437
		}
	}
	guillemets := []byte{0xae, 0xaf} // «»
	if bytes.Contains(p, guillemets) {
		sl.Info(msg,
			slog.Bool("guillements pair", true),
			slog.Duration("time", time.Since(tick)))
		return charmap.CodePage437
	}
	sl.Info(msg,
		slog.Duration("time", time.Since(tick)))
	return nil
}

// Latency returns the stored, current local time.
func Latency() *time.Time {
	start := time.Now()
	r := new(big.Int)
	const n, k = 1000, 10
	r.Binomial(n, k)
	return &start
}

// LocalIPs returns a list of local IP addresses.
//
// credit, [gosamples]
//
// [gosamples]: https://gosamples.dev/local-ip-address
func LocalIPs() ([]net.IP, error) {
	var ips []net.IP
	addresses, err := net.InterfaceAddrs()
	if err != nil {
		const format = "net interface addresses: %w"
		return nil, fmt.Errorf(format, err)
	}

	for _, addr := range addresses {
		if ipnet, ipnetExists := addr.(*net.IPNet); ipnetExists && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				ips = append(ips, ipnet.IP)
			}
		}
	}
	return ips, nil
}

// LocalHosts returns a list of local hostnames.
func LocalHosts() ([]string, error) {
	const format = "local hosts %s: %w"
	hostname, err := os.Hostname()
	if err != nil {
		return nil, fmt.Errorf(format, "hostname", err)
	}
	hosts := []string{}
	hosts = append(hosts, hostname)
	resolve := net.Resolver{}
	_, err = resolve.LookupHost(context.Background(), "localhost")
	if err != nil {
		return nil, fmt.Errorf(format, "net lookup host", err)
	}
	hosts = append(hosts, "localhost")
	return hosts, nil
}

// Ping sends a HTTP GET request to the provided URI and returns the status code and size of the response.
func Ping(ctx context.Context, uri string) (int, int64, error) {
	const format = "helper ping %s %w: %s"
	client := http.Client{
		Timeout: Timeout,
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return 0, 0, fmt.Errorf(format, "new request", err, uri)
	}
	req.Header.Set("User-Agent", UserAgent)
	res, err := client.Do(req)
	if err != nil || res == nil {
		return 0, 0, fmt.Errorf(format, "client do", err, uri)
	}
	defer res.Body.Close()

	size, err := io.Copy(io.Discard, res.Body)
	if err != nil {
		return http.StatusInternalServerError, 0, fmt.Errorf(format, "body copy", err, uri)
	}
	return res.StatusCode, size, nil
}

// LocalHostPing sends a HTTP GET request to the provided URI on the localhost
// and returns the status code and size of the response.
func LocalHostPing(ctx context.Context, uri, proto string, port int) (int, int64, error) {
	const local = "localhost"
	resolve := net.Resolver{}
	_, err := resolve.LookupHost(ctx, local)
	if err != nil {
		const format = "helper localhost ping lookup: %w"
		return http.StatusInternalServerError, 0, fmt.Errorf(format, err)
	}
	url := fmt.Sprintf("%s://%s:%d%s", proto, local, port, uri)
	return Ping(ctx, url)
}

// TimeDistance describes the difference between two time values.
// The seconds parameter determines if the string should handle less than a minute values.
func TimeDistance(from, to time.Time, seconds bool) string {
	// This function is a port of a CFWheels framework function programmed in ColdFusion (CFML).
	// https://github.com/cfwheels/cfwheels/blob/cf8e6da4b9a216b642862e7205345dd5fca34b54/wheels/global/misc.cfm#L112

	delta := to.Sub(from)
	secs, mins, hrs := int(delta.Seconds()),
		int(delta.Minutes()),
		int(delta.Hours())

	const hours, days, months, year, years, twoyears = 1440, 43200, 525600, 657000, 919800, 1051200
	switch {
	case mins <= 1:
		if !seconds {
			return lessMin(secs)
		}
		return lessMinAsSec(secs)
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
func lessMin(secs int) string {
	const minute = 60
	switch {
	case secs < minute:
		return "less than a minute"
	default:
		return "1 minute"
	}
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
	const parthour, abouthour, hours = 45, 90, 1440
	switch {
	case mins < parthour:
		return fmt.Sprintf("%d minutes", mins)
	case mins < abouthour:
		return "about 1 hour"
	case mins < hours:
		if hrs == 0 {
			hrs = 1
		}
		return fmt.Sprintf("about %d hours", hrs)
	default:
		return ""
	}
}

// lessDays returns a string describing the time difference in days.
func lessDays(mins, hrs int) string {
	const day, days = 2880, 43200
	switch {
	case mins < day:
		return "1 day"
	case mins < days:
		const hoursinaday = 24
		d := hrs / hoursinaday
		if d == 0 {
			d = 1
		}
		return fmt.Sprintf("%d days", d)
	default:
		return ""
	}
}

// lessMonths returns a string describing the time difference in months.
func lessMonths(mins, hrs int) string {
	const month, months = 86400, 525600
	switch {
	case mins < month:
		return "about 1 month"
	case mins < months:
		const hoursinamonth = 730
		m := hrs / hoursinamonth
		if m == 0 {
			m = 1
		}
		return fmt.Sprintf("%d months", m)
	default:
		return ""
	}
}

// Year returns true if the i value is greater than 1969
// or equal to the current year.
func Year(i int) bool {
	const unix = 1970
	now := time.Now().Year()
	if i >= unix && i <= now {
		return true
	}
	return false
}

// Deprecated: As of release v1.5.
const LoggerKey string = "logger"

// Deprecated: As of release v1.5, this function returns nil.
func Logger(_ context.Context) any {
	return nil
}
