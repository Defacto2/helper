//nolint:cyclop,ireturn
package helper

import (
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"time"
	"unicode"
	"unicode/utf8"

	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/charmap"
	uni "golang.org/x/text/encoding/unicode"
	"golang.org/x/text/unicode/runenames"
)

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
func Determine(r io.Reader) encoding.Encoding {
	sl := slog.New(slog.DiscardHandler)
	return determine(sl, r)
}

// DetermineWithLogger functions the same as Determine, however you can provide a
// slog logger to track character or sequence matches for false-positive
// discoveries and other possible problems.
func DetermineWithLogger(sl *slog.Logger, r io.Reader) encoding.Encoding {
	if sl == nil {
		sl = slog.New(slog.DiscardHandler)
	}
	return determine(sl, r)
}

// Deprecated: Use [DetermineWithLogger] instead.
func DetermineS(sl *slog.Logger, r io.Reader) encoding.Encoding {
	return DetermineWithLogger(sl, r)
}

func determine(sl *slog.Logger, r io.Reader) encoding.Encoding {
	const msg = "helper determine r encoding"
	if sl == nil {
		sl = slog.New(slog.DiscardHandler)
	}
	if r == nil {
		sl.Info(msg, slog.Bool("empty_reader", true))
		return nil
	}

	p, err := io.ReadAll(r)
	if err != nil {
		sl.Info(msg, slog.Any("readall_rune", err))
		return nil
	}

	sl.Info(msg, slog.Int("bytes", len(p)))
	if e := DetermineSupplement(sl, p); e != nil {
		return e
	}

	if e := DetermineChar(sl, p); e != nil {
		return e
	}
	if e := DetermineSequences(sl, p); e != nil {
		return e
	}

	// Check for Unicode multi-byte characters
	// If an unknown rune is encountered then assume the encoding is
	// using a legacy 8-bit code page encoding, such as CP-437.
	tick := time.Now()

	logT := func(s string) {
		sl.Info(msg+" "+s, slog.Duration("time", time.Since(tick)))
	}

	for _, r := range bytes.Runes(p) {
		if utf8.RuneLen(r) > 1 {
			switch {
			// we use this switch to handle any obvious false-positives
			case unicode.Is(unicode.Arabic, r):
				logT(msg + "arabic rune")
				// '┌┐' cp437 char sequence gets mistaken as a multi-byte Arabic ڿ script
				return charmap.CodePage437
			case r == unknownRune:
				logT(msg + "unknown rune")
				return charmap.ISO8859_1
			}
			logT(msg + "multi-byte rune")
			return uni.UTF8
		}
	}

	logT("latin-1")
	return charmap.ISO8859_1
}

// DetermineSupplement returns unicode.UTF8 if p contains common UTF-8 block/symbol characters (•, ─, █).
func DetermineSupplement(sl *slog.Logger, p []byte) encoding.Encoding {
	const msg = "helper determine p unicode supplements"

	// must use doublequotes ("") for the UTF-8 byte representations of •, ─, █
	if bytes.Contains(p, []byte("•")) ||
		bytes.Contains(p, []byte("─")) ||
		bytes.Contains(p, []byte("█")) {
		if sl != nil {
			sl.Info(msg, slog.Bool("found match", true))
		}
		return uni.UTF8
	}

	return nil
}

// DetermineChar returns the encoding based on the presence of common CP-437 or ISO-8859-1 characters.
// A nil encoding is returned if no encoding is determined.
func DetermineChar(sl *slog.Logger, p []byte) encoding.Encoding {
	const msg = "helper determine p chars"
	const bullet, interpunct = 0xf9, 0xfa
	tick := time.Now()

	for i, char := range p {
		switch {
		case char == escape:
			// escape control character commonly used in ANSI escaped sequences
			continue

		case char == kcfAltEsc, char == bell:
			// oddball control characters that are sometimes found in Amiga ANSI files
			continue

		case char == formFeed, char == newline, char == carriageReturn, char == tab, char == verticalTab:
			// common whitespace control characters
			continue

		case char >= undefinedStart && char <= undefinedEnd:
			// unused ASCII, which we can probably assume to be CP-437
			logChar(sl, tick, i, char)
			return charmap.CodePage437

		case char >= controlStart && char <= controlEnd:
			// ASCII control characters, which we can probably assume to be CP-437 glyphs
			logChar(sl, tick, i, char)
			return charmap.CodePage437

		case char == interpunct, char == bullet:
			logChar(sl, tick, i, char)
			return charmap.CodePage437
		}
	}

	if sl != nil {
		sl.Info(msg, slog.Duration("time", time.Since(tick)))
	}
	return nil
}

func logChar(sl *slog.Logger, tick time.Time, i int, char byte) {
	const msg = "helper determine p chars"
	if sl == nil {
		return
	}

	r := rune(char)
	name := runenames.Name(r)
	if name == "" {
		name = "UNKNOWN"
	}

	// printable character representation (%q prevents raw binary pollution in logs)
	charInfo := fmt.Sprintf("%d> %q (0x%02X / %d) %s", i, char, char, char, name)

	sl.Info(msg,
		slog.String("character", charInfo),
		slog.Duration("time", time.Since(tick)),
	)
}

type cp437Pattern struct {
	sequence []byte
	name     string
}

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
	pair           = 2
	four           = 4
)

// Pre-allocated static sequences to avoid heap allocations.
var cp437Sequences = []cp437Pattern{ //nolint:gochecknoglobals
	{bytes.Repeat([]byte{shadeLight}, pair), "shade light pair"},
	{bytes.Repeat([]byte{shadeMedium}, pair), "shade medium pair"},
	{bytes.Repeat([]byte{shadeDark}, pair), "shade dark pair"},
	{bytes.Repeat([]byte{singleHorizBar}, pair), "single horizontal bar pair"},
	{bytes.Repeat([]byte{doubleHorizBar}, four), "double horizontal bar quad"},
	{bytes.Repeat([]byte{fullBlock}, four), "full block quad"},
	{bytes.Repeat([]byte{lowerHalfBlock}, four), "lower half block quad"},
	{bytes.Repeat([]byte{upperHalfBlock}, pair), "upper half block pair"},
	{bytes.Repeat([]byte{interpunct}, pair), "interpunct pair"},
	{bytes.Repeat([]byte{bulletpoint}, pair), "bulletpoint pair"},
	{[]byte{0xae, 0xaf}, "guillemets pair"},
}

// DetermineSequences returns the encoding based on the presence of common CP-437 or ISO-8859-1 character sequences.
func DetermineSequences(sl *slog.Logger, p []byte) encoding.Encoding {
	const msg = "helper determine p seqs"
	tick := time.Now()

	for _, pat := range cp437Sequences {
		if bytes.Contains(p, pat.sequence) {
			if sl != nil {
				sl.Info(msg,
					slog.String("matched sequence", pat.name),
					slog.Duration("time", time.Since(tick)),
				)
			}
			return charmap.CodePage437
		}
	}

	if sl != nil {
		sl.Info(msg, slog.Duration("time", time.Since(tick)))
	}
	return nil
}
