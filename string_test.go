package helper_test

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/Defacto2/helper"
	"github.com/google/uuid"
	"github.com/nalgeon/be"
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

	be.Err(t, uuid.Validate(testUNID), nil)

	s, err := helper.CfUUID(testUNID)
	be.Err(t, err, nil)
	be.Err(t, uuid.Validate(s), nil)

	s, err = helper.CfUUID(testCUID)
	be.Err(t, err, nil)
	be.Err(t, uuid.Validate(s), nil)
}

func TestByteCount(t *testing.T) {
	t.Parallel()

	const mb = 1024*1024*1024 - 1
	be.Equal(t, helper.ByteCount(0), "0B")
	be.Equal(t, helper.ByteCount(1023), "1023B")
	be.Equal(t, helper.ByteCount(1024), "1k")
	be.Equal(t, helper.ByteCount(-1026), "-1026B")
	be.Equal(t, helper.ByteCount(mb), "1024M")
	be.Equal(t, helper.ByteCount(mb+2), "1G")
}

func TestByteCountFloat(t *testing.T) {
	t.Parallel()

	be.Equal(t, helper.ByteCountFloat(0), "0 bytes")
	be.Equal(t, helper.ByteCountFloat(1023), "1 kB")
	be.Equal(t, helper.ByteCountFloat(1024), "1 kB")
	be.Equal(t, helper.ByteCountFloat(-1026), "-1026 bytes")

	const (
		mb = 1024*1024*1024 - 1
		gb = mb * 1024
		tb = gb * 1024
	)
	be.Equal(t, helper.ByteCountFloat(mb), "1.1 GB")
	be.Equal(t, helper.ByteCountFloat(gb), "1.1 TB")
	be.Equal(t, helper.ByteCountFloat(tb), "1.1 PB")
}

func TestCapitalize(t *testing.T) {
	t.Parallel()

	be.Equal(t, helper.Capitalize(""), "")
	be.Equal(t, helper.Capitalize("hello"), "Hello")
	be.Equal(t, helper.Capitalize("the uk"), "The uk")
	be.Equal(t, helper.Capitalize("the UK"), "The UK")
	be.Equal(t, helper.Capitalize("hello world"), "Hello world")
	be.Equal(t, helper.Capitalize("HELLO WORLD!"), "Hello WORLD!")
}

func TestDeleteDupe(t *testing.T) {
	t.Parallel()

	got := helper.DeleteDupe(nil...)
	be.Equal(t, got, []string{})

	got = helper.DeleteDupe([]string{"a"}...)
	be.Equal(t, got, []string{"a"})

	got = helper.DeleteDupe([]string{"a", "b", "abcde"}...)
	be.Equal(t, got, []string{"a", "abcde", "b"}) // sorted

	got = helper.DeleteDupe([]string{"a", "b", "a"}...)
	be.Equal(t, got, []string{"a", "b"})
}

func TestFmtSlice(t *testing.T) {
	t.Parallel()

	got := helper.FmtSlice("")
	be.Equal(t, got, "")

	got = helper.FmtSlice("a")
	be.Equal(t, got, "A")

	got = helper.FmtSlice("a,b, abcde")
	be.Equal(t, got, "A, B, Abcde")

	got = helper.FmtSlice("a , b , abcde")
	be.Equal(t, got, "A, B, Abcde")
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
				t.Errorf("ChrLast(%s) = %v, want %v", tt.s, got, tt.want)
			}
		})
	}
}

func TestMaxLineLength(t *testing.T) {
	t.Parallel()

	be.Equal(t, helper.MaxLineLength(""), 0)
	be.Equal(t, helper.MaxLineLength("a"), 1)
	be.Equal(t, helper.MaxLineLength("a\nb"), 1)
	const s = "a\n" +
		"abcdefghijklmnopqrstuvwxyz\n" +
		"abcde.\n"
	be.Equal(t, helper.MaxLineLength(s), 26)
}

func TestShortMonth(t *testing.T) {
	t.Parallel()

	be.Equal(t, helper.ShortMonth(-1), "")
	be.Equal(t, helper.ShortMonth(0), "")
	be.Equal(t, helper.ShortMonth(1), "Jan")
	be.Equal(t, helper.ShortMonth(12), "Dec")
	be.Equal(t, helper.ShortMonth(13), "")
}

func TestSplitAsSpace(t *testing.T) {
	t.Parallel()

	be.Equal(t, helper.SplitAsSpaces(""), "")
	be.Equal(t, helper.SplitAsSpaces("a"), "a")
	be.Equal(t, helper.SplitAsSpaces("Hello world!"), "Hello world!")
	be.Equal(t, helper.SplitAsSpaces("HTTP Dir"), "HTTP Directory")
	be.Equal(t, helper.SplitAsSpaces("DirectPath"), "Direct Path")
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
				t.Errorf("TruncFilename() %s = %v, want %v", tt.name, got, tt.want)
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
			got := helper.TrimRoundBracket(tt.s)
			be.Equal(t, got, tt.want)
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
				t.Errorf("TrimPunct(%s) = %v, want %v", tt.s, got, tt.want)
			}
		})
	}
}

func TestYears(t *testing.T) {
	t.Parallel()

	be.Equal(t, helper.Years(0, 0), "the year 0")
	be.Equal(t, helper.Years(1990, 1990), "the year 1990")
	be.Equal(t, helper.Years(1990, 1991), "the years 1990 and 1991")
	be.Equal(t, helper.Years(1990, 2000), "the years 1990 - 2000")
	be.Equal(t, helper.Years(2000, 1990), "the years 1990 - 2000")
	be.Equal(t, helper.Years(-100, 0), "the years -100 - 0")
}

func TestDeobfuscateURL(t *testing.T) {
	t.Parallel()

	const d = `https://defacto2.net/d/ab27b2e`
	const f = `https://defacto2.net/f/ab27b2e`
	tests := []struct {
		name   string
		rawURL string
		want   int
	}{
		{"record", f, 13526},
		{"record", f + "/", 13526},
		{"download", d, 13526},
		{"query", f + "?blahblahblah", 13526},
		{"query extra", f + "?blahblahblah%more", 13526},
		{"invalid index", f + "/index.htm", 0},
		{"trim 1 byte (typo)", d[:len(d)-1], 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := helper.DeobfuscateURL(tt.rawURL)
			be.Equal(t, got, tt.want)
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
			got := helper.Slug(tt.name)
			be.Equal(t, got, tt.expect)
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
			got := helper.PageCount(tt.args.sum, tt.args.limit)
			be.Equal(t, got, tt.want)
		})
	}
}

func TestObfuscates(t *testing.T) {
	t.Parallel()

	keys := []int{1, 1000, 1236346, -123, 0}
	for _, key := range keys {
		id := helper.ObfuscateID(int64(key))
		got := helper.DeobfuscateID(id)
		be.Equal(t, got, key)
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
			got := helper.SearchTerm(tt.input)
			be.Equal(t, got, tt.want)
		})
	}
}

func TestTitleize(t *testing.T) {
	t.Parallel()

	got := helper.Titleize("")
	be.Equal(t, got, "")
	got = helper.Titleize("hello")
	be.Equal(t, got, "Hello")
	got = helper.Titleize("hello world, how are you?")
	be.Equal(t, got, "Hello World, How Are You?")
	got = helper.Titleize("hello i am from the UK and we use GBP!")
	be.Equal(t, got, "Hello I Am From The UK And We Use GBP!")
}

func TestMask_errors(t *testing.T) {
	t.Parallel()

	p := []byte("this string is too short")
	be.Equal(t, helper.Mask(p...), p)

	p = random(t, 100_000)
	be.Equal(t, helper.Mask(p...), p)
}

func TestMask(t *testing.T) {
	t.Parallel()

	p := []byte("this string is too short")
	m1 := []byte("1234-5678-ABCD-EFGH-IJKL-MNOP")
	m2 := []byte("12345-67890-ABCDE-FGHIJ-LMNOP")
	m3 := []byte("12345-67890-abcde-fghij-lmnop")
	m4 := []byte("1234-567890A-BCDEFGH-IJKL")
	d1 := []byte("1234 1240000 1234000 0000")
	want29 := []byte(strings.Repeat("0", helper.Chrs29))
	want25 := []byte(strings.Repeat("0", helper.Chrs25))

	// 6 multiples of 4 chars match
	s := [][]byte{p, m1, p}
	got := bytes.Join(s, []byte(" "))
	be.True(t, bytes.Contains(got, m1))
	got = helper.Mask(got...)
	be.True(t, !bytes.Contains(got, m1))
	be.True(t, bytes.Contains(got, want29))
	// 5 multiples of 5 chars match
	s = [][]byte{p, m2, p}
	got = bytes.Join(s, []byte(" "))
	be.True(t, bytes.Contains(got, m2))
	got = helper.Mask(got...)
	be.True(t, !bytes.Contains(got, m2))
	be.True(t, bytes.Contains(got, want29))
	// 5 multiples of 5 lowercase chars match
	s = [][]byte{p, m3, p}
	got = bytes.Join(s, []byte(" "))
	be.True(t, bytes.Contains(got, m3))
	got = helper.Mask(got...)
	be.True(t, !bytes.Contains(got, m3))
	be.True(t, bytes.Contains(got, want29))
	// 4x7x7x4 chars match
	s = [][]byte{p, m4, p}
	got = bytes.Join(s, []byte(" "))
	be.True(t, bytes.Contains(got, m4))
	got = helper.Mask(got...)
	be.True(t, !bytes.Contains(got, m4))
	be.True(t, bytes.Contains(got, want25))
	// 4x7x7x4 chars match
	s = [][]byte{p, d1, p}
	got = bytes.Join(s, []byte(" "))
	be.True(t, bytes.Contains(got, d1))
	got = helper.Mask(got...)
	be.True(t, !bytes.Contains(got, d1))
	be.True(t, bytes.Contains(got, want25))
}

func TestMask_nonmatches(t *testing.T) {
	t.Parallel()

	p := []byte("this string is too short")
	x1 := []byte("1234-5678-ABCD-EFGH-IJKLZMNOP")
	x2 := []byte("  A -5678-ABCD-EFGH-IJKL-MNO")
	x3 := []byte("1234 I240000 1234000 0000")

	// non-matches
	s := [][]byte{p, x1, p}
	got := bytes.Join(s, []byte(" "))
	be.True(t, bytes.Contains(got, x1))

	got = helper.Mask(got...)
	be.True(t, !bytes.Contains(got, x1)) // gets masked by Phone

	s = [][]byte{p, x2, p}
	got = bytes.Join(s, []byte(" "))

	be.True(t, bytes.Contains(got, x2))
	got = helper.Mask(got...)

	be.True(t, bytes.Contains(got, x2))
	s = [][]byte{p, x3, p}
	got = bytes.Join(s, []byte(" "))
	be.True(t, bytes.Contains(got, x3))

	got = helper.Mask(got...)
	be.True(t, bytes.Contains(got, x3))
}

func TestAlpha09(t *testing.T) {
	t.Parallel()

	bads := []string{"123-1234", "000_0000", "$1000.00", "Hello world!"}
	for _, number := range bads {
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

	bads := []string{"123-1234", "000-0000", "000 0000", "000A000"}
	for _, number := range bads {
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
	be.True(t, helper.Digits(idx, 3, 3))  // 123
	be.True(t, helper.Digits(idx, 4, 1))  // 2
	be.True(t, !helper.Digits(idx, 0, 3)) // ABC
	be.True(t, !helper.Digits(idx, 1, 4)) // BC12
}

func TestPhone(t *testing.T) {
	t.Parallel()

	numbers := []string{"123-1234", "000-0000", "999-9999"}
	for _, numb := range numbers {
		ok := helper.Phone(0, []byte(numb))
		be.True(t, ok)
	}

	nans := []string{"123 1234", "123A1234", "123_1234", "2005-12-12", "9999-555", "999-555X"}
	for _, s := range nans {
		ok := helper.Phone(0, []byte(s))
		be.True(t, !ok)
	}
}

func TestMaskTerm(t *testing.T) {
	t.Parallel()

	p := []byte("Use the following CD key: 123456")

	got := string(helper.MaskTerm(p...))
	be.True(t, got == "Use the following Cxxxxx: 123456")
	p = []byte("Use the following number: 123456")

	got = string(helper.MaskTerm(p...))
	be.True(t, got == "Use the following number: 123456")
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
		{"too long", "305-555-123456", true}, // first 12 bytes form valid NANP
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
