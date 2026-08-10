package helper_test

import (
	"bytes"
	"fmt"
	"math/rand"
	"strings"
	"testing"
	"time"

	"github.com/Defacto2/helper"
	"github.com/google/uuid"
	"github.com/nalgeon/be"
)

const (
	unid = "00000000-0000-0000-0000-000000000000" // common universal unique identifier example
	cfid = "00000000-0000-0000-0000000000000000"  // coldfusion uuid example
)

func ExampleSplitAsSpaces() {
	fmt.Println(helper.SplitAsSpaces("TheQuickBrownFox"))
	// Output:
	// The Quick Brown Fox
}

func ExampleSearchTerm() {
	fmt.Println(helper.SearchTerm("quick,brown,fox"))
	// Output:
	// [quick brown fox]
}

func ExampleByteCount() {
	fmt.Println(helper.ByteCount(0))
	fmt.Println(helper.ByteCount(1024))
	fmt.Println(helper.ByteCount(1024 * 1024))
	// Output:
	// 0B
	// 1k
	// 1M
}

func ExampleByteCountFloat() {
	fmt.Println(helper.ByteCountFloat(1024 * 1024 * 1024))
	// Output:
	// 1.1 GB
}

func ExampleCapitalize() {
	fmt.Println(helper.Capitalize("hello world"))
	// Output:
	// Hello world
}

func ExampleCfUUID() {
	newid, _ := helper.CfUUID("00000000-0000-0000-0000000000000000")
	fmt.Println(newid)
	// Output:
	// 00000000-0000-0000-0000-000000000000
}

func ExampleDeleteDupe() {
	fmt.Println(helper.DeleteDupe("b", "a", "a"))
	// Output:
	// [a b]
}

func ExampleFinds() {
	fmt.Println(helper.Finds("bravo", "alfa", "bravo", "charlie", "delta"))
	fmt.Println(helper.Finds("bravo", "alfa", "charlie", "delta"))
	// Output:
	// true
	// false
}

func ExampleFmtSlice() {
	fmt.Println(helper.FmtSlice("alfa,bravo,charlie"))
	// Output:
	// Alfa, Bravo, Charlie
}

func ExampleIntegrityBytes() {
	fmt.Println(helper.IntegrityBytes([]byte("hello")))
	// Output:
	// sha384-WeF0h3dEjGnea4ANejO7+5/xtGPkQ1TDVTvNucZm+pASWjx5+QOXvfX2oT3oKGhP
}

func ExampleMaxLineLength() {
	s := strings.Repeat("a", 100) + "\n" + strings.Repeat("b", 50)
	fmt.Println(helper.MaxLineLength(s))
	// Output:
	// 100
}

func ExampleObfuscate() {
	fmt.Println(helper.Obfuscate("1"))
	fmt.Println(helper.Obfuscate("abc"))
	// Output:
	// 9b1c6
	// abc
}

func ExamplePageCount() {
	fmt.Println(helper.PageCount(1000, 100))
	// Output:
	// 10
}

func ExampleReleased() {
	year, month, day := helper.Released("2024-07-15")
	fmt.Println(year, month, day)
	// Output:
	// 2024 7 15
}

func ExampleReverserInt() {
	i := helper.ReverserInt(123456)
	fmt.Println(i)
	// Output:
	// 654321
}

func ExampleShortMonth() {
	fmt.Println(helper.ShortMonth(1))
	// Output:
	// Jan
}

func ExampleSlug() {
	fmt.Println(helper.Slug("Hello World Homepage!"))
	// Output:
	// hello-world-homepage
}

func ExampleTimeDistance() {
	oneMinuteAgo := time.Now().Add(-15 * time.Second)
	fmt.Println(helper.TimeDistance(oneMinuteAgo, time.Now(), true))
	fmt.Println(helper.TimeDistance(oneMinuteAgo, time.Now(), false))

	oneHourAgo := time.Now().Add(-time.Hour)
	fmt.Println(helper.TimeDistance(oneHourAgo, time.Now(), true))

	oneHourAhead := time.Now().Add(time.Hour)
	fmt.Println(helper.TimeDistance(oneHourAgo, oneHourAhead, true))
	// Output:
	// less than 20 seconds
	// less than a minute
	// about 1 hour
	// about 2 hours
}

func ExampleAdd1() {
	num := helper.Add1(2)
	fmt.Println(num)
	// Output:
	// 3
}

func ExampleTitleize() {
	fmt.Println(helper.Titleize("hello world"))
	// Output:
	// Hello World
}

func ExampleTrimPunct() {
	fmt.Println(helper.TrimPunct("OMG?!?"))
	// Output:
	// OMG
}

func ExampleTrimRoundBracket() {
	fmt.Println(helper.TrimRoundBracket("Hello (world)"))
	// Output:
	// Hello
}

func ExampleYears() {
	fmt.Println(helper.Years(1990, 2000))
	fmt.Println(helper.Years(1990, 1991))
	fmt.Println(helper.Years(1990, 1990))
	// Output:
	// the years 1990 - 2000
	// the years 1990 and 1991
	// the year 1990
}

func ExampleDeObfuscate() {
	fmt.Println(helper.DeObfuscate("9b1c6"))
	// Output:
	// 1
}

func ExampleChrLast() {
	fmt.Println(helper.ChrLast("hello"))
	fmt.Println(helper.ChrLast("abc\n"))
	// Output:
	// o
	// c
}

func ExampleDetermine() {
	a := strings.NewReader("hello")
	fmt.Println(helper.Determine(a))
	// Output:
	// ISO 8859-1
}

func TestCfUUID(t *testing.T) {
	t.Parallel()
	err := uuid.Validate(unid)
	be.Err(t, err, nil)
	newid, err := helper.CfUUID(unid)
	be.Err(t, err, nil)
	err = uuid.Validate(newid)
	be.Err(t, err, nil)
	newid, err = helper.CfUUID(cfid)
	be.Err(t, err, nil)
	err = uuid.Validate(newid)
	be.Err(t, err, nil)
}

func TestByteCount(t *testing.T) {
	t.Parallel()
	s := helper.ByteCount(0)
	be.Equal(t, s, "0B")
	s = helper.ByteCount(1023)
	be.Equal(t, s, "1023B")
	s = helper.ByteCount(1024)
	be.Equal(t, s, "1k")
	s = helper.ByteCount(-1026)
	be.Equal(t, s, "-1026B")
	s = helper.ByteCount(1024*1024*1024 - 1)
	be.Equal(t, s, "1024M")
}

func TestByteCountFloat(t *testing.T) {
	t.Parallel()
	s := helper.ByteCountFloat(0)
	be.Equal(t, "0 bytes", s)
	s = helper.ByteCountFloat(1023)
	be.Equal(t, "1 kB", s)
	s = helper.ByteCountFloat(1024)
	be.Equal(t, "1 kB", s)
	s = helper.ByteCountFloat(-1026)
	be.Equal(t, "-1026 bytes", s)
	s = helper.ByteCountFloat(1024*1024*1024 - 1)
	be.Equal(t, "1.1 GB", s)
	s = helper.ByteCountFloat(1024*1024*1024*1024 - 1)
	be.Equal(t, "1.1 TB", s)
	s = helper.ByteCountFloat(1024*1024*1024*1024*1024 - 1)
	be.Equal(t, "1.1 PB", s)
}

func TestCapitalize(t *testing.T) {
	t.Parallel()
	s := helper.Capitalize("")
	be.Equal(t, s, "")
	s = helper.Capitalize("hello")
	be.Equal(t, "Hello", s)
	s = helper.Capitalize("hello world")
	be.Equal(t, "Hello world", s)
	s = helper.Capitalize(strings.ToUpper("hello world!"))
	be.Equal(t, "Hello WORLD!", s)
}

func TestDeleteDupe(t *testing.T) {
	t.Parallel()
	s := helper.DeleteDupe(nil...)
	be.Equal(t, []string{}, s)
	s = helper.DeleteDupe([]string{"a"}...)
	be.Equal(t, []string{"a"}, s)
	s = helper.DeleteDupe([]string{"a", "b", "abcde"}...)
	be.Equal(t, []string{"a", "abcde", "b"}, s) // sorted
	s = helper.DeleteDupe([]string{"a", "b", "a"}...)
	be.Equal(t, []string{"a", "b"}, s)
}

func TestFmtSlice(t *testing.T) {
	t.Parallel()
	s := helper.FmtSlice("")
	be.Equal(t, s, "")
	s = helper.FmtSlice("a")
	be.Equal(t, "A", s)
	s = helper.FmtSlice("a,b, abcde")
	be.Equal(t, "A, B, Abcde", s)
	s = helper.FmtSlice("a , b , abcde")
	be.Equal(t, "A, B, Abcde", s)
}

func TestChrLast(t *testing.T) {
	t.Parallel()
	tests := []struct {
		s    string
		want string
	}{
		{"", ""},
		{"abc", "c"},
		{"012", "2"},
		{"abc ", "c"},
		{"😃💁 People · 🐻🌻 Animals · 🎷", "🎷"},
	}
	for _, tt := range tests {
		t.Run(tt.s, func(t *testing.T) {
			t.Parallel()
			if got := helper.ChrLast(tt.s); got != tt.want {
				t.Errorf("ChrLast() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMaxLineLength(t *testing.T) {
	t.Parallel()
	i := helper.MaxLineLength("")
	be.Equal(t, 0, i)
	i = helper.MaxLineLength("a")
	be.Equal(t, 1, i)
	i = helper.MaxLineLength("a\nb")
	be.Equal(t, 1, i)
	i = helper.MaxLineLength("a\nabcdefghijklmnopqrstuvwxyz\nabcde.")
	be.Equal(t, 26, i)
}

func TestShortMonth(t *testing.T) {
	t.Parallel()
	s := helper.ShortMonth(0)
	be.Equal(t, s, "")
	s = helper.ShortMonth(1)
	be.Equal(t, "Jan", s)
	s = helper.ShortMonth(12)
	be.Equal(t, "Dec", s)
	s = helper.ShortMonth(13)
	be.Equal(t, s, "")
}

func TestSplitAsSpace(t *testing.T) {
	t.Parallel()
	s := helper.SplitAsSpaces("")
	be.Equal(t, s, "")
	s = helper.SplitAsSpaces("a")
	be.Equal(t, "a", s)
	s = helper.SplitAsSpaces("Hello world!")
	be.Equal(t, "Hello world!", s)
	s = helper.SplitAsSpaces("HTTP Dir")
	be.Equal(t, "HTTP Directory", s)
	s = helper.SplitAsSpaces("DirectPath")
	be.Equal(t, "Direct Path", s)
}

func TestTruncFilename(t *testing.T) {
	t.Parallel()
	const fn = "one_two-three.file"
	type args struct {
		w    int
		name string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"empty", args{-1, ""}, ""},
		{"zero", args{0, fn}, ""},
		{"ext", args{5, fn}, ".file"},
		{"too short", args{3, fn}, ".file"},
		{"short", args{14, fn}, "one_two-..file"},
		{"too short 2", args{6, "file"}, "file"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := helper.TruncFilename(tt.args.w, tt.args.name); got != tt.want {
				t.Errorf("TruncFilename() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTrimRoundBracket(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		s    string
		want string
	}{
		{"empty", "", ""},
		{"hi", "Hello world", "Hello world"},
		{"oops", "Hello (world", "Hello"},
		{"count", "Hello 1) world", "Hello 1) world"},
		{"okay", "Hello world (Hi!)", "Hello world"},
		{"search", "Razor 1911 (RZR, Razor)", "Razor 1911"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			be.Equal(t, helper.TrimRoundBracket(tt.s), tt.want)
		})
	}
}

func TestTrimPunct(t *testing.T) {
	t.Parallel()
	tests := []struct {
		s    string
		want string
	}{
		{"", ""},
		{"abc", "abc"},
		{"abc.", "abc"},
		{"abc?", "abc"},
		{"📙", "📙"},
		{"📙!?!", "📙"},
		{"📙 (a book)", "📙 (a book)"},
	}
	for _, tt := range tests {
		t.Run(tt.s, func(t *testing.T) {
			t.Parallel()
			if got := helper.TrimPunct(tt.s); got != tt.want {
				t.Errorf("TrimPunct() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestYears(t *testing.T) {
	t.Parallel()
	s := helper.Years(0, 0)
	be.Equal(t, "the year 0", s)
	s = helper.Years(1990, 1991)
	be.Equal(t, "the years 1990 and 1991", s)
	s = helper.Years(1990, 2000)
	be.Equal(t, "the years 1990 - 2000", s)
}

// https://defacto2.net/f/ab27b2e

func TestDeobfuscateURL(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		rawURL string
		want   int
	}{
		{"record", "https://defacto2.net/f/ab27b2e", 13526},
		{"download", "https://defacto2.net/d/ab27b2e", 13526},
		{"query", "https://defacto2.net/f/ab27b2e?blahblahblah", 13526},
		{"typo", "https://defacto2.net/f/ab27b2", 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			be.Equal(t, tt.want, helper.DeobfuscateURL(tt.rawURL))
		})
	}
}

func TestSlug(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		expect string
	}{
		{"the-group", "the_group"},
		{"group1, group2", "group1*group2"},
		{"group1 & group2", "group1-ampersand-group2"},
		{"group 1, group 2", "group-1*group-2"},
		{"GROUP 👾", "group"},
		{"Mooñpeople", "moonpeople"},
	}
	for _, tt := range tests {
		t.Run(tt.expect, func(t *testing.T) {
			t.Parallel()
			be.Equal(t, helper.Slug(tt.name), tt.expect)
		})
	}
}

func TestPageCount(t *testing.T) {
	t.Parallel()
	type args struct {
		sum   int
		limit int
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{"-1", args{-1, -1}, 0},
		{"0", args{0, 500}, 0},
		{"1", args{1, 500}, 1},
		{"500", args{500, 750}, 1},
		{"750", args{750, 500}, 2},
		{"1k", args{1000, 500}, 2},
		{"1001", args{1001, 500}, 3},
		{"want 10", args{1000, 100}, 10},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			be.Equal(t, tt.want, helper.PageCount(tt.args.sum, tt.args.limit))
		})
	}
}

func TestObfuscates(t *testing.T) {
	t.Parallel()
	keys := []int{1, 1000, 1236346, -123, 0}
	for _, key := range keys {
		s := helper.ObfuscateID(int64(key))
		be.Equal(t, key, helper.DeobfuscateID(s))
	}
}

func TestSearchTerm(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{"empty", "", []string{}},
		{"spaces", "   ", []string{}},
		{"one", "one", []string{"one"}},
		{"two", "one two", []string{"one two"}},
		{"three", "one two three", []string{"one two three"}},
		{"quotes", `"one two" three`, []string{"\"one two\" three"}},
		{"onetwo", "one,two", []string{"one", "two"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			be.Equal(t, tt.want, helper.SearchTerm(tt.input))
		})
	}
}

func TestTitleize(t *testing.T) {
	t.Parallel()
	s := helper.Titleize("")
	be.Equal(t, s, "")
	s = helper.Titleize("hello")
	be.Equal(t, "Hello", s)
	s = helper.Titleize("hello world, how are you?")
	be.Equal(t, "Hello World, How Are You?", s)
	s = helper.Titleize("hello i am from the UK and we use GBP!")
	be.Equal(t, s, "Hello I Am From The UK And We Use GBP!")
}

func TestMask(t *testing.T) {
	m1 := []byte("1234-5678-ABCD-EFGH-IJKL-MNOP")
	m2 := []byte("12345-67890-ABCDE-FGHIJ-LMNOP")
	m3 := []byte("12345-67890-abcde-fghij-lmnop")
	m4 := []byte("1234-567890A-BCDEFGH-IJKL")
	d1 := []byte("1234 1240000 1234000 0000")
	x1 := []byte("1234-5678-ABCD-EFGH-IJKLZMNOP")
	x2 := []byte("  A -5678-ABCD-EFGH-IJKL-MNO")
	x3 := []byte("1234 I240000 1234000 0000")
	want29 := []byte(strings.Repeat("0", helper.Chrs29))
	want25 := []byte(strings.Repeat("0", helper.Chrs25))
	t.Parallel()
	// too short
	p := []byte("this string is too short")
	x := helper.Mask(p...)
	be.Equal(t, p, x)
	// no key in string
	p = random(1000)
	x = helper.Mask(p...)
	be.Equal(t, p, x)
	// 6 multiples of 4 chars match
	s := [][]byte{p, m1, p}
	x = bytes.Join(s, []byte(" "))
	be.True(t, bytes.Contains(x, m1))
	x = helper.Mask(x...)
	be.True(t, !bytes.Contains(x, m1))
	be.True(t, bytes.Contains(x, want29))
	// 5 multiples of 5 chars match
	s = [][]byte{p, m2, p}
	x = bytes.Join(s, []byte(" "))
	be.True(t, bytes.Contains(x, m2))
	x = helper.Mask(x...)
	be.True(t, !bytes.Contains(x, m2))
	be.True(t, bytes.Contains(x, want29))
	// 5 multiples of 5 lowercase chars match
	s = [][]byte{p, m3, p}
	x = bytes.Join(s, []byte(" "))
	be.True(t, bytes.Contains(x, m3))
	x = helper.Mask(x...)
	be.True(t, !bytes.Contains(x, m3))
	be.True(t, bytes.Contains(x, want29))
	// 4x7x7x4 chars match
	s = [][]byte{p, m4, p}
	x = bytes.Join(s, []byte(" "))
	be.True(t, bytes.Contains(x, m4))
	x = helper.Mask(x...)
	be.True(t, !bytes.Contains(x, m4))
	be.True(t, bytes.Contains(x, want25))
	// 4x7x7x4 chars match
	s = [][]byte{p, d1, p}
	x = bytes.Join(s, []byte(" "))
	be.True(t, bytes.Contains(x, d1))
	x = helper.Mask(x...)
	be.True(t, !bytes.Contains(x, d1))
	be.True(t, bytes.Contains(x, want25))

	// non-matches
	s = [][]byte{p, x1, p}
	x = bytes.Join(s, []byte(" "))
	be.True(t, bytes.Contains(x, x1))
	x = helper.Mask(x...)
	be.True(t, !bytes.Contains(x, x1)) // after an unpdate, this gets masked by Phone
	s = [][]byte{p, x2, p}
	x = bytes.Join(s, []byte(" "))
	be.True(t, bytes.Contains(x, x2))
	x = helper.Mask(x...)
	be.True(t, bytes.Contains(x, x2))
	s = [][]byte{p, x3, p}
	x = bytes.Join(s, []byte(" "))
	be.True(t, bytes.Contains(x, x3))
	x = helper.Mask(x...)
	be.True(t, bytes.Contains(x, x3))
}

const chars = " abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func random(n int) []byte {
	b := make([]byte, n)
	for i := range b {
		b[i] = chars[rand.Intn(len(chars))] //nolint:gosec
	}
	return b
}

func TestAlpha09(t *testing.T) {
	t.Parallel()
	bad := []string{"123-1234", "000_0000", "$1000.00", "Hello world!"}
	for _, number := range bad {
		l := len(number)
		ok := helper.Digits([]byte(number), 0, l)
		be.True(t, !ok)
	}
	good := []string{"1231234", "Abcdefghi", "0A1b2C3d", "09x876"}
	for _, s := range good {
		l := len(s)
		ok := helper.Alpha09([]byte(s), 0, l)
		be.True(t, ok)
	}
	idx := []byte("Abc123.")
	ok := helper.Alpha09(idx, 0, 6) // Abc123
	be.True(t, ok)
	ok = helper.Alpha09(idx, 6, 2) // 3.
	be.True(t, !ok)
}

func TestDigits(t *testing.T) {
	t.Parallel()
	bad := []string{"123-1234", "000-0000", "000 0000", "000A000"}
	for _, number := range bad {
		l := len(number)
		ok := helper.Digits([]byte(number), 0, l)
		be.True(t, !ok)
	}
	good := []string{"1231234", "0000000", "0000000", "09876"}
	for _, number := range good {
		l := len(number)
		ok := helper.Digits([]byte(number), 0, l)
		be.True(t, ok)
	}
	idx := []byte("ABC123")
	ok := helper.Digits(idx, 3, 3) // 123
	be.True(t, ok)
	ok = helper.Digits(idx, 4, 1) // 2
	be.True(t, ok)
	ok = helper.Digits(idx, 0, 3) // ABC
	be.True(t, !ok)
	ok = helper.Digits(idx, 1, 4) // BC12
	be.True(t, !ok)
}

func TestPhone(t *testing.T) {
	t.Parallel()
	numbers := []string{"123-1234", "000-0000", "999-9999"}
	for _, numb := range numbers {
		ok := helper.Phone(0, []byte(numb))
		be.True(t, ok)
	}
	notnumb := []string{"123 1234", "123A1234", "123_1234", "2005-12-12", "9999-555", "999-555X"}
	for _, s := range notnumb {
		ok := helper.Phone(0, []byte(s))
		be.True(t, !ok)
	}
}

func TestMaskTerm(t *testing.T) {
	t.Parallel()
	p := []byte("Use the following CD key: 123456")
	g := string(helper.MaskTerm(p...))
	be.True(t, g == "Use the following Cxxxxx: 123456")
	p = []byte("Use the following number: 123456")
	g = string(helper.MaskTerm(p...))
	be.True(t, g == "Use the following number: 123456")
}

func TestNANP(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		in   string
		exp  bool
	}{
		{"valid NANP", "305-555-1234", true},
		{"valid high area code", "999-555-1234", true},
		{"invalid low area code", "199-555-1234", false},
		{"invalid area code 100", "100-555-1234", false},
		{"missing hyphens", "3055551234", false},
		{"too short", "305-555-123", false},
		{"too long", "305-555-123456", true}, // First 12 bytes form valid NANP
		{"wrong format", "305/555/1234", false},
		{"empty", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := helper.NANP(0, []byte(tt.in))
			if got != tt.exp {
				t.Errorf("NANP() = %v, want %v for input %s", got, tt.exp, tt.in)
			}
		})
	}
}

func TestReleased(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		in    string
		year  int16
		month int16
		day   int16
	}{
		{"full date", "2024-07-15", 2024, 7, 15},
		{"year-month", "2024-07", 2024, 7, 0},
		{"year only", "2024", 2024, 0, 0},
		{"min year", "0001-01-01", 1, 1, 1},
		{"max year", "9999-12-31", 9999, 12, 31},
		{"invalid date", "2024-13-32", 2024, 0, 0}, // Invalid month/day
		{"partial invalid", "2024-13", 2024, 0, 0}, // Invalid month
		{"empty", "", 0, 0, 0},
		{"malformed", "not-a-date", 0, 0, 0},
		{"extra dashes", "2024-07-15-extra", 2024, 7, 15}, // Should parse first 3
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			year, month, day := helper.Released(tt.in)
			if year != tt.year || month != tt.month || day != tt.day {
				t.Errorf("Released() = (%v, %v, %v), want (%v, %v, %v) for input %s",
					year, month, day, tt.year, tt.month, tt.day, tt.in)
			}
		})
	}
}
