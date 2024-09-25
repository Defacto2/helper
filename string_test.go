package helper_test

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/Defacto2/helper"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	unid = "00000000-0000-0000-0000-000000000000" // common universal unique identifier example
	cfid = "00000000-0000-0000-0000000000000000"  // coldfusion uuid example
)

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
	fmt.Println(helper.DeleteDupe("a", "b", "a"))
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

func ExampleReverseInt() {
	i, _ := helper.ReverseInt(123456)
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
	oneHourAgo := time.Now().Add(-time.Hour)
	oneHourAhead := time.Now().Add(time.Hour)
	fmt.Println(helper.TimeDistance(oneHourAgo, time.Now(), true))
	fmt.Println(helper.TimeDistance(oneHourAgo, oneHourAhead, true))
	// Output:
	// about 1 hour
	// about 2 hours
}

func ExampleAdd1() {
	num := 2
	fmt.Println(helper.Add1(num))
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

func ExampleTrimRoundBraket() {
	fmt.Println(helper.TrimRoundBraket("Hello (world)"))
	// Output:
	// Hello
}

func ExampleYears() {
	fmt.Println(helper.Years(1990, 2000))
	// Output:
	// the years 1990 - 2000
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
	require.NoError(t, err)

	newid, err := helper.CfUUID(unid)
	require.NoError(t, err)
	err = uuid.Validate(newid)
	require.NoError(t, err)

	newid, err = helper.CfUUID(cfid)
	require.NoError(t, err)
	err = uuid.Validate(newid)
	require.NoError(t, err)
}

func TestByteCount(t *testing.T) {
	s := helper.ByteCount(0)
	assert.Equal(t, "0B", s)
	s = helper.ByteCount(1023)
	assert.Equal(t, "1023B", s)
	s = helper.ByteCount(1024)
	assert.Equal(t, "1k", s)
	s = helper.ByteCount(-1026)
	assert.Equal(t, "-1026B", s)
	s = helper.ByteCount(1024*1024*1024 - 1)
	assert.Equal(t, "1024M", s)
}

func TestByteCountFloat(t *testing.T) {
	s := helper.ByteCountFloat(0)
	assert.Equal(t, "0 bytes", s)
	s = helper.ByteCountFloat(1023)
	assert.Equal(t, "1 kB", s)
	s = helper.ByteCountFloat(1024)
	assert.Equal(t, "1 kB", s)
	s = helper.ByteCountFloat(-1026)
	assert.Equal(t, "-1026 bytes", s)
	s = helper.ByteCountFloat(1024*1024*1024 - 1)
	assert.Equal(t, "1.1 GB", s)
	s = helper.ByteCountFloat(1024*1024*1024*1024 - 1)
	assert.Equal(t, "1.1 TB", s)
	s = helper.ByteCountFloat(1024*1024*1024*1024*1024 - 1)
	assert.Equal(t, "1.1 PB", s)
}

func TestCapitalize(t *testing.T) {
	s := helper.Capitalize("")
	assert.Equal(t, "", s)
	s = helper.Capitalize("hello")
	assert.Equal(t, "Hello", s)
	s = helper.Capitalize("hello world")
	assert.Equal(t, "Hello world", s)
	s = helper.Capitalize(strings.ToUpper("hello world!"))
	assert.Equal(t, "Hello WORLD!", s)
}

func TestDeleteDupe(t *testing.T) {
	s := helper.DeleteDupe(nil...)
	assert.EqualValues(t, []string{}, s)
	s = helper.DeleteDupe([]string{"a"}...)
	assert.EqualValues(t, []string{"a"}, s)
	s = helper.DeleteDupe([]string{"a", "b", "abcde"}...)
	assert.EqualValues(t, []string{"a", "abcde", "b"}, s) // sorted
	s = helper.DeleteDupe([]string{"a", "b", "a"}...)
	assert.EqualValues(t, []string{"a", "b"}, s)
}

func TestFmtSlice(t *testing.T) {
	s := helper.FmtSlice("")
	assert.Equal(t, "", s)
	s = helper.FmtSlice("a")
	assert.Equal(t, "A", s)
	s = helper.FmtSlice("a,b, abcde")
	assert.Equal(t, "A, B, Abcde", s)
	s = helper.FmtSlice("a , b , abcde")
	assert.Equal(t, "A, B, Abcde", s)
}

func TestChrLast(t *testing.T) {
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
			if got := helper.ChrLast(tt.s); got != tt.want {
				t.Errorf("ChrLast() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMaxLineLength(t *testing.T) {
	i := helper.MaxLineLength("")
	assert.Equal(t, 0, i)
	i = helper.MaxLineLength("a")
	assert.Equal(t, 1, i)
	i = helper.MaxLineLength("a\nb")
	assert.Equal(t, 1, i)
	i = helper.MaxLineLength("a\nabcdefghijklmnopqrstuvwxyz\nabcde.")
	assert.Equal(t, 26, i)
}

func TestShortMonth(t *testing.T) {
	s := helper.ShortMonth(0)
	assert.Equal(t, "", s)
	s = helper.ShortMonth(1)
	assert.Equal(t, "Jan", s)
	s = helper.ShortMonth(12)
	assert.Equal(t, "Dec", s)
	s = helper.ShortMonth(13)
	assert.Equal(t, "", s)
}

func TestSplitAsSpace(t *testing.T) {
	s := helper.SplitAsSpaces("")
	assert.Equal(t, "", s)
	s = helper.SplitAsSpaces("a")
	assert.Equal(t, "a", s)
	s = helper.SplitAsSpaces("Hello world!")
	assert.Equal(t, "Hello world!", s)
	s = helper.SplitAsSpaces("HTTP Dir")
	assert.Equal(t, "HTTP Directory", s)
}

func TestTruncFilename(t *testing.T) {
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
		{"empty", args{}, ""},
		{"zero", args{0, fn}, ""},
		{"ext", args{5, fn}, ".file"},
		{"too short", args{4, fn}, ".file"},
		{"short", args{14, fn}, "one_two-..file"},
		{"too short 2", args{6, "file"}, "file"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := helper.TruncFilename(tt.args.w, tt.args.name); got != tt.want {
				t.Errorf("TruncFilename() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTrimRoundBraket(t *testing.T) {
	tests := []struct {
		name string
		s    string
		want string
	}{
		{"empty", "", ""},
		{"hi", "Hello world", "Hello world"},
		{"okay", "Hello world (Hi!)", "Hello world"},
		{"search", "Razor 1911 (RZR, Razor)", "Razor 1911"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, helper.TrimRoundBraket(tt.s))
		})
	}
}

func TestTrimPunct(t *testing.T) {
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
			if got := helper.TrimPunct(tt.s); got != tt.want {
				t.Errorf("TrimPunct() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestYears(t *testing.T) {
	s := helper.Years(0, 0)
	assert.Equal(t, "the year 0", s)
	s = helper.Years(1990, 1991)
	assert.Equal(t, "the years 1990 and 1991", s)
	s = helper.Years(1990, 2000)
	assert.Equal(t, "the years 1990 - 2000", s)
}

// https://defacto2.net/f/ab27b2e

func TestDeobfuscateURL(t *testing.T) {
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
			assert.Equal(t, tt.want, helper.DeobfuscateURL(tt.rawURL))
		})
	}
}

func TestSlug(t *testing.T) {
	tests := []struct {
		name      string
		expect    string
		assertion assert.ComparisonAssertionFunc
	}{
		{"the-group", "the_group", assert.Equal},
		{"group1, group2", "group1*group2", assert.Equal},
		{"group1 & group2", "group1-ampersand-group2", assert.Equal},
		{"group 1, group 2", "group-1*group-2", assert.Equal},
		{"GROUP 👾", "group", assert.Equal},
		{"Mooñpeople", "moonpeople", assert.Equal},
	}
	for _, tt := range tests {
		t.Run(tt.expect, func(t *testing.T) {
			tt.assertion(t, tt.expect, helper.Slug(tt.name))
		})
	}
}

func TestPageCount(t *testing.T) {
	type args struct {
		sum   int
		limit int
	}
	tests := []struct {
		name string
		args args
		want uint
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
			assert.Equal(t, tt.want, helper.PageCount(tt.args.sum, tt.args.limit))
		})
	}
}

func TestObfuscates(t *testing.T) {
	keys := []int{1, 1000, 1236346, -123, 0}
	for _, key := range keys {
		s := helper.ObfuscateID(int64(key))
		assert.Equal(t, key, helper.DeobfuscateID(s))
	}
}

func TestSearchTerm(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{"empty", "", []string{}},
		{"spaces", "   ", []string{""}},
		{"one", "one", []string{"one"}},
		{"two", "one two", []string{"one two"}},
		{"three", "one two three", []string{"one two three"}},
		{"quotes", `"one two" three`, []string{"\"one two\" three"}},
		{"two", "one,two", []string{"one", "two"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, helper.SearchTerm(tt.input))
		})
	}
}

func TestTitleize(t *testing.T) {
	s := helper.Titleize("")
	assert.Empty(t, s)

	s = helper.Titleize("hello")
	assert.Equal(t, "Hello", s)

	s = helper.Titleize("hello world, how are you?")
	assert.Equal(t, "Hello World, How Are You?", s)
}

// func TestReleased(t *testing.T) {
// 	t.Parallel()
// 	tests := []struct {
// 		name          string
// 		releaseDate   string
// 		expectedYear  int16
// 		expectedMonth int16
// 		expectedDay   int16
// 	}{
// 		{
// 			name:          "Valid release date",
// 			releaseDate:   "2024-07-15",
// 			expectedYear:  2024,
// 			expectedMonth: 7,
// 			expectedDay:   15,
// 		},
// 		{
// 			name:          "Valid release date",
// 			releaseDate:   "2024-07",
// 			expectedYear:  2024,
// 			expectedMonth: 7,
// 			expectedDay:   0,
// 		},
// 		{
// 			name:          "Invalid release date",
// 			releaseDate:   "2024-07-15-01",
// 			expectedYear:  0,
// 			expectedMonth: 0,
// 			expectedDay:   0,
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			t.Parallel()
// 			p := pouet.Production{
// 				ReleaseDate: tt.releaseDate,
// 			}
// 			year, month, day := p.Released()
// 			assert.Equal(t, tt.expectedYear, year)
// 			assert.Equal(t, tt.expectedMonth, month)
// 			assert.Equal(t, tt.expectedDay, day)
// 		})
// 	}
// }
