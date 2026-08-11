package helper

// Package string contains helper functions for string operations.

import (
	"bytes"
	"fmt"
	"math/bits"
	"net/url"
	"path"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/google/uuid"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

const (
	hexadecimal  = 16
	obfuscateXOR = 461
	obfuscateSum = 154
)

// ByteCount formats b as in a compact, human-readable unit of measure.
func ByteCount(b int64) string {
	const unit = 1024
	const base = 10
	if b < unit {
		return strconv.FormatInt(b, base) + "B"
	}

	u := uint64(b)
	exp := (bits.Len64(u)-1)/base - 1
	if exp >= len(byteUnits) {
		exp = len(byteUnits) - 1
	}
	div := uint64(1) << ((exp + 1) * base)
	val := float64(u) / float64(div)
	return strconv.FormatFloat(val, 'f', 0, 64) + string(byteUnits[exp])
}

// ByteCountFloat formats b in a human-readable unit of measure.
// Units measured in gigabytes or larger are returned with 1 decimal place.
func ByteCountFloat(b int64) string {
	const unit = 1000
	if b < unit {
		return strconv.FormatInt(b, 10) + " bytes"
	}

	u := uint64(b)
	exp := 0
	div := uint64(unit)

	for n := u / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	if exp >= len(byteUnits) {
		exp = len(byteUnits) - 1
	}

	val := float64(u) / float64(div)
	prec := 0
	const giga = 2
	if exp >= giga {
		prec = 1
	}

	return strconv.FormatFloat(val, 'f', prec, 64) + " " + string(byteUnits[exp]) + "B"
}

// Capitalize returns a string with the first letter of the first word capitalized.
// If the first word is an acronym, it is capitalized as a word.
func Capitalize(s string) string {
	if s == "" {
		return ""
	}

	caser := cases.Title(language.English)

	i := strings.IndexByte(s, ' ')
	if i == -1 {
		return caser.String(s)
	}

	return caser.String(s[:i]) + s[i:]
}

// ChrLast returns the last character or rune of the string.
func ChrLast(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	_, size := utf8.DecodeLastRuneInString(s)
	return s[len(s)-size:]
}

// CfUUID formats a 35 character, Coldfusion Universally Unique Identifier.
// to a standard, 36 character, Universally Unique Identifier.
func CfUUID(cfid string) (string, error) {
	const format = "cfuuid validate: %w"
	if err := uuid.Validate(cfid); err == nil {
		return cfid, nil
	}

	old := strings.TrimSpace(cfid)

	// ColdFusion UUID format: xxxxxxxx-xxxx-xxxx-xxxxxxxxxxxxxxxx (35 chars)
	// Standard UUID format:   xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx (36 chars)
	const cfLen = 35
	const pos = 23
	if len(old) != cfLen {
		return "", fmt.Errorf(format, uuid.Validate(old))
	}

	var b [36]byte
	copy(b[:pos], old[:pos])
	b[pos] = '-'
	copy(b[pos+1:], old[pos:])

	newid := string(b[:])
	if err := uuid.Validate(newid); err != nil {
		return "", fmt.Errorf(format, err)
	}
	return newid, nil
}

// DeleteDupe removes duplicate strings from a slice.
// The returned slice is sorted and compacted.
func DeleteDupe(s ...string) []string {
	if len(s) == 0 {
		return []string{}
	}
	x := slices.Clone(s)
	slices.Sort(x)
	return slices.Compact(x)
}

// DeObfuscate deobfuscates the obfuscated string, or returns the original string.
//
// This function is a port of the [deobfuscateParam] function programmed in ColdFusion (CFML).
//
// [deobfuscateParam]: https://github.com/cfwheels/cfwheels/blob/main/wheels/global/misc.cfm
func DeObfuscate(s string) string {
	const checksum = 2
	if len(s) < checksum {
		return s
	}

	if _, err := strconv.Atoi(s); err == nil {
		return s
	}

	num, err := strconv.ParseInt(s[checksum:], hexadecimal, 64)
	if err != nil {
		return s
	}

	// format XOR'd number to base-10 digits byte-buffer
	num ^= obfuscateXOR
	baseNum := strconv.FormatInt(num, 10)
	length := len(baseNum)
	if length <= 1 {
		return s
	}

	// reverse the digits (excluding original index 0)
	// and calculate checksum sum in a single pass without string allocations
	values := make([]byte, length-1)
	chksumTest := 0

	for i := range length - 1 {
		b := baseNum[length-1-i]
		if b < '0' || b > '9' {
			return s
		}
		values[i] = b
		chksumTest += int(b - '0')
	}

	headerChecksum, err := strconv.ParseInt(s[:checksum], hexadecimal, 64)
	if err != nil {
		return s
	}
	// validate header checksum against sum of reversed digits + obfuscateSum offset
	if headerChecksum != int64(chksumTest+obfuscateSum) {
		return s
	}

	return string(values)
}

// DeobfuscateID deobfuscates an obfuscated ID to return the primary key of the record.
// Returns a 0 if the id is not valid.
func DeobfuscateID(id string) int {
	key, err := strconv.Atoi(DeObfuscate(id))
	if err != nil {
		return 0
	}
	return key
}

// DeobfuscateURL deobfuscates an obfuscated record URL to return a record's primary key.
// A URL can point to a Defacto2 record download or detail page.
// Returns a 0 if the URL is not valid.
func DeobfuscateURL(rawURL string) int {
	u, err := url.Parse(rawURL)
	if err != nil {
		return 0
	}
	p := strings.TrimRight(u.Path, "/")
	if p == "" {
		return 0
	}
	return DeobfuscateID(path.Base(p))
}

// FmtSlice formats a comma separated string.
func FmtSlice(s string) string {
	var b strings.Builder
	const sep = ","

	for part := range strings.SplitSeq(s, sep) {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		if b.Len() > 0 {
			b.WriteString(", ")
		}
		b.WriteString(Capitalize(part))
	}

	return b.String()
}

// MaxLineLength counts the character/rune length of the longest line in a string,
// that is split strictly by newline (\n).
func MaxLineLength(s string) int {
	maxLen := 0
	for line := range strings.SplitSeq(s, "\n") {
		if n := utf8.RuneCountInString(line); n > maxLen {
			maxLen = n
		}
	}
	return maxLen
}

const (
	Chrs24 = 24
	Chrs25 = 25
	Chrs29 = 29
)

var (
	// Pre-computed mask strings for Mask() - computed once, reused many times.
	//nolint:gochecknoglobals
	maskChrs29 = bytes.Repeat([]byte{'0'}, Chrs29)
	//nolint:gochecknoglobals
	maskChrs25 = bytes.Repeat([]byte{'0'}, Chrs25)
	//nolint:gochecknoglobals
	maskChrs24 = bytes.Repeat([]byte{'0'}, Chrs24)
)

// Mask runs a performant scan of the bytes and replaces any matching.
// serials or key sequences with a sequence of 0 characters of the same length.
//
// Currently the following patterns are matched
//
//   - 12345-67890-ABCDE-FGHIJ-LMNOP
//   - 1234-5678-ABCD-EFGH-IJKL-MNOP
//   - 1234-567890A-BCDEFGH-IJKL
//   - 1234 1240000 1234000 0000
//   - 1234-5678-ABCD-EFGH-IJKL
//
// In addition telephone numbers matching seven digit (or more) phone numbers
// will have two digits replaced by $$.
//
// And finally, there's are collection words and combos that will be masked with xxx.
// For example, any occurrence of the word "password" will be masked as "pxxxxxx".
//
// For example:
// "Hello world B014-56789A-BCDEFGH-G45X example".
//
// Would be masked with:
// "Hello world 000000000000000000000000 example".
func Mask(p ...byte) []byte {
	out := bytes.NewBuffer(make([]byte, 0, len(p)))
	i := 0
	for i < len(p) {
		switch {
		case serial5x5(i, p), serial6x4(i, p):
			out.Write(maskChrs29)
			i += Chrs29
			continue
		case serial4774(i, p), digit4774(i, p):
			out.Write(maskChrs25)
			i += Chrs25
			continue
		case serial5x4(i, p):
			out.Write(maskChrs24)
			i += Chrs24
			continue
		case Phone(i, p), PhoneEuro(i, p), PhoneDE(i, p):
			// 123-5678
			out.Write(p[i : i+5])
			out.WriteString("$$")
			out.WriteByte(p[i+7])
			i += 8
			continue
		default:
			if x := IndexTerm(i, p); x > 0 {
				out.WriteByte(p[i])
				for range x - 1 {
					out.WriteByte('x')
				}
				i += x
				continue
			}
		}
		if i < len(p) {
			out.WriteByte(p[i])
			i++
		}
	}
	return out.Bytes()
}

// MaskTerm replaces a predefined list of words and combinations with a series of x characters.
// For example, any occurrence of the word "password" will be masked as "pxxxxxx".
//
// MaskTerm should not be used with [Mask] as it duplicates functionality.
func MaskTerm(p ...byte) []byte {
	out := bytes.NewBuffer(make([]byte, 0, len(p)))
	i := 0

	for i < len(p) {
		if x := IndexTerm(i, p); x > 0 {
			out.WriteByte(p[i])
			for range x - 1 {
				out.WriteByte('x')
			}
			i += x
			continue
		}

		out.WriteByte(p[i])
		i++
	}

	return out.Bytes()
}

// IndexTerm searches the p byte array for a predefined list of words and combinations.
// Any matches will return the location index of the match.
// If no matches are found a 0 is returned.
//
// The predefined list are items that can trigger online bots.
func IndexTerm(i int, p []byte) int {
	sub := p[i:]
	terms := terms()
	for _, term := range terms {
		l := len(term)
		if len(sub) >= l && bytes.EqualFold(sub[:l], term) {
			return l
		}
	}
	return 0
}

// terms are intentionally fragmented and are kept as a var for caching.
//
//nolint:gochecknoglobals
var terms = sync.OnceValue(func() [][]byte {
	return [][]byte{
		// generic
		[]byte("cd-k" + "ey"), []byte("cd" + " key"), []byte("c" + "racke" + "d"),
		[]byte("key " + "code"), []byte("k" + "ey file"), []byte("k" + "eyfile"),
		[]byte("k" + "ey gen"), []byte("k" + "eygen"), []byte("k" + "eymaker"),
		[]byte("l" + "ice" + "nse " + "code"), []byte("pas" + "sword"), []byte("s" + "er" + "ial"),
		// brands
		[]byte("m" + "icro" + "soft"), []byte("c" + "orel"),
		[]byte("s" + "pacial audio " + "solution"), []byte("p" + "arallels " + "inc"),
		[]byte("p" + "aint" + "shop"), []byte("a" + "dobe"), []byte("a" + "cronis"),
		[]byte("s" + "am " + "b" + "road" + "caster"), []byte("n" + "intend" + "o"), []byte("s" + "ony"),
	}
})

// Alpha09 returns true if the slice of bytes exclusively contains.
// alphanumeric characters. Everything else including punctuation returns false.
// i is the index position and n is the number of bytes to match.
func Alpha09(b []byte, i, n int) bool {
	if i < 0 || n <= 0 || i+n > len(b) {
		return false
	}
	for _, c := range b[i : i+n] {
		if !isAlphaNum(c) {
			return false
		}
	}
	return true
}

func isAlphaNum(c byte) bool {
	return (c >= '0' && c <= '9') ||
		(c >= 'a' && c <= 'z') ||
		(c >= 'A' && c <= 'Z')
}

// Digits returns true if the slice of bytes is a sequence of digits.
// i is the index position and n is the number of bytes to match.
func Digits(b []byte, i, n int) bool {
	if i < 0 || n <= 0 || i+n > len(b) {
		return false
	}
	for _, c := range b[i : i+n] {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

// Phone matches a 3-4, 7 digit telephone number, i.e., "555-1234".
func Phone(i int, p []byte) bool {
	const length = 8 // 3 digits + '-' + 4 digits
	if i < 0 || i+length > len(p) {
		return false
	}
	const prefix, suffix = 3, 4
	return p[i+3] == '-' && Digits(p, i, prefix) && Digits(p, i+suffix, suffix)
}

// PhoneDE matches a 3-3-3, 9 digit telephone number, i.e., "555-123-456".
func PhoneDE(i int, p []byte) bool {
	const length = 11 // 3 + '-' + 3 + '-' + 3
	if i < 0 || i+length > len(p) {
		return false
	}
	const n = 3
	const midoff = 4
	const endoff = 8
	return p[i+3] == '-' &&
		p[i+7] == '-' &&
		Digits(p, i, n) &&
		Digits(p, i+midoff, n) &&
		Digits(p, i+endoff, n)
}

// PhoneEuro matches a 2-6, 8 digit telephone number, i.e., "55-123456".
func PhoneEuro(i int, p []byte) bool {
	const length = 9 // 2 + '-' + 6
	if i < 0 || i+length > len(p) {
		return false
	}
	const prefix = 2
	const suffix = 3
	const digits = 6
	return p[i+2] == '-' &&
		Digits(p, i, prefix) &&
		Digits(p, i+suffix, digits)
}

// NANP matches an areacode and a 7 digit number, i.e., 305-555-1234.
// However, area codes below 200 are not matched, ie 199-555-1234.
func NANP(i int, p []byte) bool {
	const length = 12 // 3 + '-' + 3 + '-' + 4
	if i < 0 || i+length > len(p) {
		return false
	}
	const n = 3
	const midoff = 4
	const endoff = 8
	return p[i+3] == '-' &&
		p[i+7] == '-' &&
		p[i] >= '2' &&
		Digits(p, i, n) &&
		Digits(p, i+midoff, n) &&
		Digits(p, i+endoff, n+1)
}

// serial5x5 matches 12345-67890-ABCDE-FGHIJ-LMNOP.
func serial5x5(i int, p []byte) bool { //nolint:cyclop
	const length = 29 // (5 * 5) + 4 hyphens = 29
	if i < 0 || i+length > len(p) {
		return false
	}
	const (
		n    = 5
		off1 = 6
		off2 = 12
		off3 = 18
		off4 = 24
	)
	// check hyphen delimiters first for fast short-circuiting
	return p[i+5] == '-' &&
		p[i+11] == '-' &&
		p[i+17] == '-' &&
		p[i+23] == '-' &&
		Alpha09(p, i, n) &&
		Alpha09(p, i+off1, n) &&
		Alpha09(p, i+off2, n) &&
		Alpha09(p, i+off3, n) &&
		Alpha09(p, i+off4, n)
}

// serial5x4 matches 1234-5678-ABCD-EFGH-IJKL.
func serial5x4(i int, p []byte) bool { //nolint:cyclop
	const length = 25 // (5 * 4) + 4 hyphens + 1 trailing delimiter = 25
	if i < 0 || i+length > len(p) {
		return false
	}
	const (
		n    = 4
		off1 = 5
		off2 = 10
		off3 = 15
		off4 = 20
	)
	// check delimiters and trailing boundary first for fast short-circuiting
	return (p[i+24] == ' ' || p[i+24] == '\n') &&
		p[i+4] == '-' &&
		p[i+9] == '-' &&
		p[i+14] == '-' &&
		p[i+19] == '-' &&
		Alpha09(p, i, n) &&
		Alpha09(p, i+off1, n) &&
		Alpha09(p, i+off2, n) &&
		Alpha09(p, i+off3, n) &&
		Alpha09(p, i+off4, n)
}

// serial6x4 matches 1234-5678-ABCD-EFGH-IJKL-MNOP.
func serial6x4(i int, p []byte) bool { //nolint:cyclop
	const length = 29 // (6 * 4) + 5 hyphens = 29
	if i < 0 || i+length > len(p) {
		return false
	}
	const (
		n    = 4
		off1 = 5
		off2 = 10
		off3 = 15
		off4 = 20
		off5 = 25
	)
	return p[i+4] == '-' &&
		p[i+9] == '-' &&
		p[i+14] == '-' &&
		p[i+19] == '-' &&
		p[i+24] == '-' &&
		Alpha09(p, i, n) &&
		Alpha09(p, i+off1, n) &&
		Alpha09(p, i+off2, n) &&
		Alpha09(p, i+off3, n) &&
		Alpha09(p, i+off4, n) &&
		Alpha09(p, i+off5, n)
}

// serial4774 matches 1234-567890A-BCDEFGH-IJKL.
func serial4774(i int, p []byte) bool {
	const length = 25 // 4 + '-' + 7 + '-' + 7 + '-' + 4 = 25
	if i < 0 || i+length > len(p) {
		return false
	}
	const (
		n4   = 4
		n7   = 7
		off1 = 5
		off2 = 13
		off3 = 21
	)
	return p[i+4] == '-' &&
		p[i+12] == '-' &&
		p[i+20] == '-' &&
		Alpha09(p, i, n4) &&
		Alpha09(p, i+off1, n7) &&
		Alpha09(p, i+off2, n7) &&
		Alpha09(p, i+off3, n4)
}

// digits4774 matches 1234 1234567 1234567 1234.
func digit4774(i int, p []byte) bool {
	const length = 25 // 4 + ' ' + 7 + ' ' + 7 + ' ' + 4 = 25
	if i < 0 || i+length > len(p) {
		return false
	}
	const (
		n4   = 4
		n7   = 7
		off1 = 5
		off2 = 13
		off3 = 21
	)
	return p[i+4] == ' ' &&
		p[i+12] == ' ' &&
		p[i+20] == ' ' &&
		Digits(p, i, n4) &&
		Digits(p, i+off1, n7) &&
		Digits(p, i+off2, n7) &&
		Digits(p, i+off3, n4)
}

// ObfuscateID obfuscates the primary key of a record as a string that is used as a URL param or path.
func ObfuscateID(key int64) string {
	return Obfuscate(strconv.Itoa(int(key)))
}

// Obfuscate obfuscates a numeric string to insecurely hide database primary key values when passed along a URL.
//
// This function is a port of the [obfuscateParam] function programmed in ColdFusion (CFML).
//
// [obfuscateParam]: https://github.com/cfwheels/cfwheels/blob/main/wheels/global/misc.cfm
func Obfuscate(s string) string {
	l := len(s)
	if l == 0 || s[0] == '0' {
		return s
	}

	var (
		val      int
		digitSum int
		pow10    = 1
	)

	const base = 10
	for i := range l {
		c := s[i]
		if c < '0' || c > '9' {
			return s // return un-obfuscated if string contains non-digits
		}
		digit := int(c - '0')
		val = val*base + digit
		digitSum += digit
		pow10 *= 10
	}

	reverse := ReverserInt(val)
	a := (pow10 + reverse) ^ obfuscateXOR
	b := digitSum + obfuscateSum
	hexB := strconv.FormatInt(int64(b), hexadecimal)
	hexA := strconv.FormatInt(int64(a), hexadecimal)

	return hexB + hexA
}

// ReverserInt reverses the digits of n.
func ReverserInt(n int) int {
	const base = 10
	rev := 0
	for n > 0 {
		rev = rev*base + (n % base)
		n /= 10
	}
	return rev
}

// Deprecated: use [ReverserInt] instead.
//
// The error always returns nil.
func ReverseInt(i int) (int, error) {
	return ReverserInt(i), nil
}

// PageCount returns the maximum pages possible for the sum of records with a record limit per-page.
func PageCount(sum, limit int) int {
	if sum <= 0 || limit <= 0 {
		return 0
	}
	// integer arithmetic is more performant that using the floating point math package.
	return (sum + limit - 1) / limit
}

// Released returns a string release date as year, month, day int16 values.
// The string is expected to be in the format "2024-07-15" or "2024-07" or "2024".
func Released(s string) (year, month, day int16) { //nolint:cyclop,nonamedreturns
	const minimum = 4
	if len(s) < minimum {
		return 0, 0, 0
	}

	y := parse4Digits(s[0:4])
	if y <= 0 {
		return 0, 0, 0
	}
	year = y

	if len(s) < 7 || s[4] != '-' {
		return year, 0, 0
	}

	m := parse2Digits(s[5:7])
	if m < 1 || m > 12 {
		return year, 0, 0
	}
	month = m

	if len(s) < 10 || s[7] != '-' {
		return year, month, 0
	}

	d := parse2Digits(s[8:10])
	if d < 1 || d > 31 {
		return year, month, 0
	}
	day = d

	return year, month, day
}

// a fast four-digit ASCII byte parser.
func parse4Digits(b string) int16 {
	if b[0] < '0' || b[0] > '9' ||
		b[1] < '0' || b[1] > '9' ||
		b[2] < '0' || b[2] > '9' ||
		b[3] < '0' || b[3] > '9' {
		return -1
	}
	return int16(b[0]-'0')*1000 + int16(b[1]-'0')*100 + int16(b[2]-'0')*10 + int16(b[3]-'0')
}

// a fast two-digit ASCII byte parser.
func parse2Digits(b string) int16 {
	if b[0] < '0' || b[0] > '9' || b[1] < '0' || b[1] > '9' {
		return -1
	}
	return int16(b[0]-'0')*10 + int16(b[1]-'0')
}

// SearchTerm returns a list of search terms from the input string.
// The input string is split by commas.
func SearchTerm(input string) []string {
	if input == "" {
		return []string{}
	}

	const sep = ","
	terms := strings.Split(input, sep)
	s := make([]string, 0, len(terms))

	for _, term := range terms {
		trimmed := strings.TrimSpace(term)
		if trimmed != "" {
			s = append(s, trimmed)
		}
	}

	return s
}

// ShortMonth takes a month integer and abbreviates it to a three letter English month.
func ShortMonth(month int) string {
	if month < 1 || month > 12 {
		return ""
	}
	const length = 3
	return time.Month(month).String()[:length]
}

// Slug returns a URL friendly string of the named group.
func Slug(name string) string {
	if name == "" {
		return ""
	}

	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	s, _, err := transform.String(t, name)
	if err != nil {
		s = name // fallback on transform error
	}
	// strings replace all is more performant than regex
	s = strings.ReplaceAll(s, "-", "_")             // hyphen to underscores
	s = strings.ReplaceAll(s, ", ", "*")            // multiple groups get separated with asterisks
	s = strings.ReplaceAll(s, " & ", " ampersand ") // remove & characters from URI usage

	var b strings.Builder
	b.Grow(len(s))

	runesList := []rune(s)
	n := len(runesList)

	for i := range n {
		r := runesList[i]
		// numbers receive leading hyphens
		if r == ' ' && i+1 < n && unicode.IsDigit(runesList[i+1]) {
			b.WriteRune('-')
			continue
		}
		// remove all unexpected characters
		if isPermitted(r) {
			b.WriteRune(r)
		}
	}
	res := strings.TrimSpace(strings.ToLower(b.String()))
	return strings.ReplaceAll(res, " ", "-")
}

func isPermitted(r rune) bool {
	if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
		return true
	}
	switch r {
	case ' ', '-', '+', '.', '_', '*':
		return true
	}
	return false
}

// SplitAsSpaces splits a string at each capital letter.
func SplitAsSpaces(s string) string { //nolint:cyclop
	if s == "" {
		return ""
	}

	var result strings.Builder
	const heuristic = 8
	result.Grow(len(s) + heuristic)

	runes := []rune(s)
	n := len(runes)

	for i := 0; i < n; {
		r := runes[i]

		if i > 0 && unicode.IsUpper(r) {
			prev := runes[i-1]
			// case 1: from lowercase/digit to uppercase ("getHTTP" -> "get HTTP")
			// case 2: from acronym to regular word ("HTTPResponse" -> "HTTP Response")
			if !unicode.IsUpper(prev) || (i+1 < n && unicode.IsLower(runes[i+1])) {
				if prev != ' ' {
					result.WriteRune(' ')
				}
			}
		}

		if r == 'D' && i+2 < n && runes[i+1] == 'i' && runes[i+2] == 'r' {
			isStandalone := (i+3 == n) || unicode.IsUpper(runes[i+3]) || !unicode.IsLetter(runes[i+3])
			if isStandalone {
				result.WriteString("Directory")
				i += 3 // Advance past 'D', 'i', and 'r'
				continue
			}
		}

		result.WriteRune(r)
		i++
	}

	return result.String()
}

var englishCaser = sync.OnceValue(func() cases.Caser { //nolint:gochecknoglobals
	return cases.Title(language.English, cases.NoLower)
})

// Titleize returns a string with the first letter of each word capitalized.
// If a word is an acronym, it is capitalized as a word.
func Titleize(s string) string {
	if s == "" {
		return ""
	}
	return englishCaser().String(s)
}

// TruncFilename reduces a filename to the length of w characters.
// The file extension is always preserved with the truncation.
func TruncFilename(w int, name string) string {
	if w <= 0 {
		return ""
	}

	count := utf8.RuneCountInString(name)
	if w >= count {
		return name
	}

	// identify hidden files (".gitignore", ".env") that have no main stem
	// treat the dot-prefix name as a stem with no extension
	ext := filepath.Ext(name)
	if strings.HasPrefix(name, ".") && ext == name {
		ext = ""
	}

	const trunc = "."
	truncLen := utf8.RuneCountInString(trunc)
	extLen := utf8.RuneCountInString(ext)
	stemLen := w - extLen - truncLen
	if stemLen <= 0 {
		if ext != "" {
			return ext
		}
		// if purely a hidden file stem with no ext, truncate it directly
		runes := []rune(name)
		return string(runes[:w])
	}

	stem := name[:len(name)-len(ext)]
	runes := []rune(stem)

	return string(runes[:stemLen]) + trunc + ext
}

// TrimRoundBracket removes the trailing round brackets and any whitespace.
func TrimRoundBracket(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}

	// find the start of the bracket section.
	before, _, ok := strings.Cut(s, "(")
	if ok {
		return strings.TrimSpace(before)
	}

	return s
}

// TrimRoundBraket is the deprecated name for TrimRoundBracket.
//
// Deprecated: Use TrimRoundBracket instead.
func TrimRoundBraket(s string) string {
	return TrimRoundBracket(s)
}

// TrimPunct removes any trailing, common punctuation characters from the string.
func TrimPunct(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}

	// Strip trailing 'Po' category characters
	s = strings.TrimRightFunc(s, func(r rune) bool {
		return unicode.Is(unicode.Po, r)
	})

	// Trim any whitespace that was sitting behind the removed punctuation
	return strings.TrimSpace(s)
}

// Years returns a formatted string representing a single year, consecutive years,
// a range of years, or if they are the same, it returns a singular year.
func Years(a, b int16) string {
	if a == b {
		return "the year " + strconv.Itoa(int(a))
	}

	if a > b {
		a, b = b, a
	}

	if b-a == 1 {
		return "the years " + strconv.Itoa(int(a)) + " and " + strconv.Itoa(int(b))
	}

	return "the years " + strconv.Itoa(int(a)) + " - " + strconv.Itoa(int(b))
}
