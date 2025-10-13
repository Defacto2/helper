package helper

// Package file string.go contains the helper functions for string operations.

import (
	"bytes"
	"fmt"
	"math"
	"net/url"
	"path"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
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
//
// source, [yourbasic]
//
// [yourbasic]: https://yourbasic.org/golang/formatting-byte-size-to-human-readable-format/
func ByteCount(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d%s", b, strings.Repeat("B", 1))
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.0f%c",
		float64(b)/float64(div), byteUnits[exp])
}

// ByteCountFloat formats b as in a human-readable unit of measure.
// Units measured in gigabytes or larger are returned with 1 decimal place.
func ByteCountFloat(b int64) string {
	const unit = 1000
	if b < unit {
		return fmt.Sprintf("%d bytes", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	const gigabyte = 2
	if exp < gigabyte {
		return fmt.Sprintf("%.0f %cB",
			float64(b)/float64(div), byteUnits[exp])
	}
	return fmt.Sprintf("%.1f %cB",
		float64(b)/float64(div), byteUnits[exp])
}

// Capitalize returns a string with the first letter of the first word capitalized.
// If the first word is an acronym, it is capitalized as a word.
func Capitalize(s string) string {
	if s == "" {
		return ""
	}
	const sep = " "
	caser := cases.Title(language.English)
	x := strings.Split(s, sep)
	const req = 2
	if len(x) < req {
		return caser.String(s)
	}
	return caser.String(x[0]) + sep + strings.Join(x[1:], sep)
}

// ChrLast returns the last character or rune of the string.
func ChrLast(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	r, _ := utf8.DecodeLastRuneInString(s)
	return string(r)
}

// CfUUID formats a 35 character, Coldfusion Universally Unique Identifier
// to a standard, 36 character, Universally Unique Identifier.
func CfUUID(cfid string) (string, error) {
	if err := uuid.Validate(cfid); err == nil {
		return cfid, nil
	}
	const pos = 23
	const hyphen = '-'
	old := strings.TrimSpace(cfid)
	r := []rune(old)
	r = append(r[:pos], append([]rune{hyphen}, r[pos:]...)...)
	newid := string(r)
	err := uuid.Validate(newid)
	if err != nil {
		return "", fmt.Errorf("cfuuid validate %w", err)
	}
	return newid, nil
}

// DeleteDupe removes duplicate strings from a slice.
// The returned slice is sorted and compacted.
func DeleteDupe(s ...string) []string {
	x := make([]string, 0, len(s))
	for val := range slices.Values(s) {
		if !slices.Contains(x, val) {
			x = append(x, val)
		}
	}
	slices.Sort(x)
	return slices.Compact(x)
}

// DeObfuscate the obfuscated string, or return the original string.
//
// This function is a port of the [deobfuscateParam] function programmed in ColdFusion (CFML).
//
// [deobfuscateParam]: https://github.com/cfwheels/cfwheels/blob/main/wheels/global/misc.cfm
func DeObfuscate(s string) string {
	const checksum, decimal = 2, 10
	if len(s) < checksum {
		return s
	}
	if i, _ := strconv.Atoi(s); i > 0 {
		return s
	}
	// deobfuscate string
	num, err := strconv.ParseInt(s[checksum:], hexadecimal, 0)
	if err != nil {
		return s
	}
	num ^= obfuscateXOR
	baseNum := strconv.Itoa(int(num))
	l := len(baseNum) - 1
	value := ""
	for i := range l {
		f := baseNum[l-i:][:1]
		value += f
	}
	// create checks
	l = len(value)
	chksumTest := 0
	for i := range l {
		chr := value[i : i+1]
		n, err1 := strconv.Atoi(chr)
		if err1 != nil {
			return s
		}
		chksumTest += n
	}
	// run checks
	chksum, err := strconv.ParseInt(s[:2], hexadecimal, 0)
	if err != nil {
		return s
	}
	chksumX := strconv.FormatInt(chksum, decimal)
	chksumY := strconv.FormatInt(int64(chksumTest+obfuscateSum), decimal)
	if err := chksumX != chksumY; err {
		return s
	}

	return value
}

// DeobfuscateID an obfuscated ID to return the primary key of the record.
// Returns a 0 if the id is not valid.
func DeobfuscateID(id string) int {
	key, _ := strconv.Atoi(DeObfuscate(id))
	return key
}

// DeobfuscateURL deobfuscate an obfuscated record URL to return a record's primary key.
// A URL can point to a Defacto2 record download or detail page.
// Returns a 0 if the URL is not valid.
func DeobfuscateURL(rawURL string) int {
	u, err := url.Parse(rawURL)
	if err != nil {
		return 0
	}
	return DeobfuscateID(path.Base(u.Path))
}

// FmtSlice formats a comma separated string.
func FmtSlice(s string) string {
	x := []string{}
	for part := range strings.SplitSeq(s, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		x = append(x, Capitalize(part))
	}
	return strings.Join(x, ", ")
}

// MaxLineLength counts the character length of the longest line in a string.
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
	Chrs25 = 25
	Chrs29 = 29
)

// Mask runs a performant scan of the bytes and replaces any matching
// serials or key sequences with a sequence of 0 characters of the same length.
//
// Currently the following patterns are matched
//
//   - 12345-67890-ABCDE-FGHIJ-LMNOP
//   - 1234-5678-ABCD-EFGH-IJKL-MNOP
//   - 1234-567890A-BCDEFGH-IJKL
//
// For example:
// "Hello world B014-56789A-BCDEFGH-G45X example".
//
// Would be masked with:
// "Hello world 000000000000000000000000 example".
func Mask(p ...byte) []byte {
	out := bytes.NewBuffer(nil)
	i := 0
	for i < len(p) {
		switch {
		case serial5x5(i, p), serial6x4(i, p):
			mask := strings.Repeat("0", Chrs29)
			out.WriteString(mask)
			i += Chrs29
			continue
		case serial4774(i, p):
			mask := strings.Repeat("0", Chrs25)
			out.WriteString(mask)
			i += Chrs25
		}
		out.WriteByte(p[i])
		i++
	}
	return out.Bytes()
}

func matcher(b []byte, i, n int) bool {
	if i < 0 || n <= 0 || i+n > len(b) {
		return false
	}
	for k := range n {
		c := b[i+k]
		switch {
		case c >= '0' && c <= '9':
			return true
		case c >= 'A' && c <= 'Z':
			return true
		case c >= 'a' && c <= 'z':
			return true
		default:
			return false
		}
	}
	return false
}

// serial5x5 matches 12345-67890-ABCDE-FGHIJ-LMNOP.
//
//nolint:mnd,cyclop
func serial5x5(i int, p []byte) bool {
	if i+29 <= len(p) &&
		matcher(p, i, 5) &&
		p[i+5] == '-' &&
		matcher(p, i+6, 5) &&
		p[i+11] == '-' &&
		matcher(p, i+12, 5) &&
		p[i+17] == '-' &&
		matcher(p, i+18, 5) &&
		p[i+23] == '-' &&
		matcher(p, i+24, 5) {
		return true
	}
	return false
}

// serial6x4 matches 1234-5678-ABCD-EFGH-IJKL-MNOP.
//
//nolint:mnd,cyclop
func serial6x4(i int, p []byte) bool {
	if i+29 <= len(p) &&
		matcher(p, i, 4) &&
		p[i+4] == '-' &&
		matcher(p, i+5, 4) &&
		p[i+9] == '-' &&
		matcher(p, i+10, 4) &&
		p[i+14] == '-' &&
		matcher(p, i+15, 4) &&
		p[i+19] == '-' &&
		matcher(p, i+20, 4) &&
		p[i+24] == '-' &&
		matcher(p, i+25, 4) {
		return true
	}
	return false
}

// serial4774 matches 1234-567890A-BCDEFGH-IJKL.
//
//nolint:mnd
func serial4774(i int, p []byte) bool {
	if i+25 <= len(p) &&
		matcher(p, i, 4) &&
		p[i+4] == '-' &&
		matcher(p, i+5, 7) &&
		p[i+12] == '-' &&
		matcher(p, i+13, 7) &&
		p[i+20] == '-' &&
		matcher(p, i+21, 4) {
		return true
	}
	return false
}

// ObfuscateID the primary key of a record as a string that is used as a URL param or path.
func ObfuscateID(key int64) string {
	return Obfuscate(strconv.Itoa(int(key)))
}

// Obfuscate a numeric string to insecurely hide database primary key values when passed along a URL.
//
// This function is a port of the [obfuscateParam] function programmed in ColdFusion (CFML).
//
// [obfuscateParam]: https://github.com/cfwheels/cfwheels/blob/main/wheels/global/misc.cfm
func Obfuscate(s string) string {
	i, err := strconv.Atoi(s)
	if err != nil {
		return s
	}
	// confirm the first digit of i isn't a zero
	if s[0] == '0' {
		return s
	}
	reverse, err := ReverseInt(i)
	if err != nil {
		return s
	}
	l := len(s)
	a := int(math.Pow10(l) + float64(reverse))
	b := 0
	for i := 1; i <= l; i++ {
		// slice and sum the individual digits
		digit, err := strconv.Atoi(string(s[l-i]))
		if err != nil {
			return s
		}
		b += digit
	}
	// base64 conversion
	a ^= obfuscateXOR
	b += obfuscateSum

	return fmt.Sprintf("%s%s",
		strconv.FormatInt(int64(b), hexadecimal),
		strconv.FormatInt(int64(a), hexadecimal),
	)
}

// PageCount returns the maximum pages possible for the sum of records with a record limit per-page.
func PageCount(sum, limit int) int {
	if sum <= 0 || limit <= 0 {
		return 0
	}
	x := math.Ceil(float64(sum) / float64(limit))
	return int(math.Abs(x))
}

// Released returns a string release date as year, month, day int16 values.
// The string is expected to be in the format "2024-07-15" or "2024-07" or "2024".
func Released(s string) (int16, int16, int16) {
	dates := strings.Split(s, "-") // "2024-07-15"
	const (
		y = 0
		m = 1
		d = 2
	)
	if len(dates) < 1 {
		return 0, 0, 0
	}
	var year, month, day int16
	yv, _ := strconv.ParseInt(dates[y], 10, 16)
	if yv > 0 && yv <= math.MaxInt16 {
		year = int16(yv)
	}
	if len(dates) < m+1 {
		return year, 0, 0
	}
	if mv, _ := strconv.ParseInt(dates[m], 10, 16); mv > 0 && mv <= 12 {
		month = int16(mv)
	}
	if len(dates) < d+1 {
		return year, month, 0
	}
	if dv, _ := strconv.ParseInt(dates[d], 10, 16); dv > 0 && dv <= 31 {
		day = int16(dv)
	}
	return year, month, day
}

// ReverseInt reverses an integer.
//
// credit, [Wade73]
//
// [Wade73]: http://stackoverflow.com/questions/35972561/reverse-int-golang
func ReverseInt(i int) (int, error) {
	itoa, str := strconv.Itoa(i), ""
	for x := len(itoa); x > 0; x-- {
		str += string(itoa[x-1])
	}

	reverse, err := strconv.Atoi(str)
	if err != nil {
		return 0, fmt.Errorf("reverse integer %d: %w", i, err)
	}

	return reverse, nil
}

// SearchTerm returns a list of search terms from the input string.
// The input string is split by commas.
func SearchTerm(input string) []string {
	if input == "" {
		return []string{}
	}
	terms := strings.Split(input, ",")
	// join the two slices
	s := make([]string, 0, len(terms))
	for term := range slices.Values(terms) {
		s = append(s, strings.TrimSpace(term))
	}
	return s
}

// ShortMonth takes a month integer and abbreviates it to a three letter English month.
func ShortMonth(month int) string {
	if month < 1 || month > 12 {
		return ""
	}
	const abbreviated = 3
	s := fmt.Sprint(time.Month(month))
	if len(s) >= abbreviated {
		return s[0:abbreviated]
	}
	return ""
}

// Slug returns a URL friendly string of the named group.
func Slug(name string) string {
	s := name
	// remove diacritics
	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	s, _, _ = transform.String(t, s)
	// hyphen to underscore
	re := regexp.MustCompile(`\-`)
	s = re.ReplaceAllString(s, "_")
	// multiple groups get separated with asterisk
	re = regexp.MustCompile(`\, `)
	s = re.ReplaceAllString(s, "*")
	// any & characters need replacement due to HTML escaping
	re = regexp.MustCompile(` \& `)
	s = re.ReplaceAllString(s, " ampersand ")
	// numbers receive a leading hyphen
	re = regexp.MustCompile(` ([0-9])`)
	s = re.ReplaceAllString(s, "-$1")
	// delete all other characters
	const deleteAllExcept = `[^A-Za-z0-9 \-\+\.\_\*]`
	re = regexp.MustCompile(deleteAllExcept)
	s = re.ReplaceAllString(s, "")
	// trim whitespace and replace any space separators with hyphens
	s = strings.TrimSpace(strings.ToLower(s))
	re = regexp.MustCompile(` `)
	s = re.ReplaceAllString(s, "-")
	return s
}

// SplitAsSpaces splits a string at each capital letter.
func SplitAsSpaces(s string) string {
	var result strings.Builder
	for i, r := range s {
		if unicode.IsUpper(r) && i != 0 {
			result.WriteRune(' ')
		}
		result.WriteRune(r)
	}
	x := result.String()
	x = strings.ReplaceAll(x, "Dir", "Directory")
	x = strings.ReplaceAll(x, "H T T P", "HTTP") //nolint:dupword
	x = strings.ReplaceAll(x, "T L S", "TLS")
	x = strings.ReplaceAll(x, "P S ", "PS ")
	x = strings.ReplaceAll(x, "I D", "ID")
	x = strings.ReplaceAll(x, "  ", " ")
	return x
}

// Titleize returns a string with the first letter each word capitalized.
// If a word is an acronym, it is capitalized as a word.
func Titleize(s string) string {
	if s == "" {
		return ""
	}
	const sep = " "
	caser := cases.Title(language.English)
	words := strings.Split(s, sep)
	if len(words) == 1 {
		return caser.String(s)
	}
	for i, word := range words {
		if word == "" {
			continue
		}
		words[i] = caser.String(word)
	}
	return strings.Join(words, sep)
}

// TruncFilename reduces a filename to the length of w characters.
// The file extension is always preserved with the truncation.
func TruncFilename(w int, name string) string {
	const trunc = "."
	if w == 0 {
		return ""
	}
	l := len(name)
	if w >= l {
		return name
	}
	ext := filepath.Ext(name)
	if w <= len(ext) {
		return ext
	}
	s := name[0 : w-len(ext)-len(trunc)]
	return fmt.Sprintf("%s%s%s", s, trunc, ext)
}

// TrimRoundBraket removes the tailing round brakets and any whitespace.
func TrimRoundBraket(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	l, r := strings.Index(s, "("), strings.Index(s, ")")
	if l < r {
		return strings.TrimSpace(s[:l-1])
	}
	return s
}

// TrimPunct removes any trailing, common punctuation characters from the string.
func TrimPunct(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	rs := []rune(s)
	for i := len(rs) - 1; i >= 0; i-- {
		r := rs[i]
		// https://www.compart.com/en/unicode/category/Po
		if !unicode.Is(unicode.Po, r) {
			punctless := string(rs[0 : i+1])
			return strings.TrimSpace(punctless)
		}
	}
	return s
}

// Years returns a string of the years if they are different.
// If they are the same, it returns a singular year.
func Years(a, b int16) string {
	if a == b {
		return fmt.Sprintf("the year %d", a)
	}
	if b-a == 1 {
		return fmt.Sprintf("the years %d and %d", a, b)
	}
	return fmt.Sprintf("the years %d - %d", a, b)
}
