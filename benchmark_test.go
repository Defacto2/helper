package helper_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/Defacto2/helper"
)

// Benchmarks for performance-critical functions

// BenchmarkDeleteDupe - Deduplication (O(n²) before fix).
func BenchmarkDeleteDupe(b *testing.B) {
	input := []string{
		"group1", "group2", "group1", "group3", "group2",
		"group4", "group5", "group3", "group1", "group6",
	}
	b.ResetTimer()
	for b.Loop() {
		helper.DeleteDupe(input...)
	}
}

// BenchmarkDeleteDupeLarge - Deduplication with larger input.
func BenchmarkDeleteDupeLarge(b *testing.B) {
	input := make([]string, 100)
	for i := range 100 {
		input[i] = "group" + string(rune(i%20))
	}
	b.ResetTimer()
	for b.Loop() {
		helper.DeleteDupe(input...)
	}
}

// BenchmarkSlug - String slug generation (regex recompilation before fix).
func BenchmarkSlug(b *testing.B) {
	tests := []string{
		"The-Group",
		"group1, group2",
		"GROUP 👾",
		"Mooñpeople",
		"test_slug-name",
	}
	b.ResetTimer()
	for b.Loop() {
		for _, s := range tests {
			helper.Slug(s)
		}
	}
}

// BenchmarkMask - String masking with patterns.
func BenchmarkMask(b *testing.B) {
	input := []byte("1234-5678-ABCD-EFGH-IJKL-MNOP serial key test")
	b.ResetTimer()
	for b.Loop() {
		helper.Mask(input...)
	}
}

// BenchmarkObfuscateID - ID obfuscation.
func BenchmarkObfuscateID(b *testing.B) {
	ids := []int64{1, 100, 1000, 10000, 999999}
	b.ResetTimer()
	for b.Loop() {
		for _, id := range ids {
			helper.ObfuscateID(id)
		}
	}
}

// BenchmarkObfuscate - String obfuscation.
func BenchmarkObfuscate(b *testing.B) {
	inputs := []string{"1", "100", "1000", "12345", "999999"}
	b.ResetTimer()
	for b.Loop() {
		for _, s := range inputs {
			helper.Obfuscate(s)
		}
	}
}

// BenchmarkSplitAsSpaces - Space normalization.
func BenchmarkSplitAsSpaces(b *testing.B) {
	input := "test\nwith\r\nmultiple\ttypes\rof\fwhitespace"
	b.ResetTimer()
	for b.Loop() {
		helper.SplitAsSpaces(input)
	}
}

// BenchmarkIntegrity - File comparison (uses fileMatch internally).
func BenchmarkIntegrity(b *testing.B) {
	data1 := bytes.Repeat([]byte("x"), 10000)
	data2 := bytes.Repeat([]byte("y"), 10000)
	b.ResetTimer()
	for b.Loop() {
		helper.IntegrityBytes(data1)
		_ = data2
	}
}

// BenchmarkAdd1 - Numeric increment (uses reflection before fix).
func BenchmarkAdd1(b *testing.B) {
	tests := []any{"test", int32(100), int64(1000), "another"}
	b.ResetTimer()
	for b.Loop() {
		for _, val := range tests {
			helper.Add1(val)
		}
	}
}

// BenchmarkDetermineUTF8 - UTF-8 text detection with emojis.
func BenchmarkDetermineUTF8(b *testing.B) {
	utf8Text := "Hello 👾 😀 🎮 This is UTF-8 text with emojis!"
	b.ResetTimer()
	for b.Loop() {
		helper.Determine(strings.NewReader(utf8Text))
	}
}

// BenchmarkDetermineUTF8WithBOM - UTF-8 with BOM (fast path).
func BenchmarkDetermineUTF8WithBOM(b *testing.B) {
	utf8WithBOM := "\xEF\xBB\xBFHello World - UTF-8 with BOM"
	b.ResetTimer()
	for b.Loop() {
		helper.Determine(strings.NewReader(utf8WithBOM))
	}
}

// BenchmarkDetermineCP437 - CP-437 text detection with ANSI art.
func BenchmarkDetermineCP437(b *testing.B) {
	cp437Text := "┌─────────────────────────────┐\n│ CP-437 ANSI Art Example │\n└─────────────────────────────┘"
	b.ResetTimer()
	for b.Loop() {
		helper.Determine(strings.NewReader(cp437Text))
	}
}

// BenchmarkDetermineLatin1 - ISO-8859-1 text detection.
func BenchmarkDetermineLatin1(b *testing.B) {
	latin1Text := "Café résumé naïve façade - Latin-1 text with accents"
	b.ResetTimer()
	for b.Loop() {
		helper.Determine(strings.NewReader(latin1Text))
	}
}

// BenchmarkDetermineASCII - Plain ASCII text (default case).
func BenchmarkDetermineASCII(b *testing.B) {
	asciiText := "This is plain ASCII text without any special characters"
	b.ResetTimer()
	for b.Loop() {
		helper.Determine(strings.NewReader(asciiText))
	}
}

// BenchmarkDetermineLargeUTF8 - Large UTF-8 file (64KB).
func BenchmarkDetermineLargeUTF8(b *testing.B) {
	largeUTF8 := strings.Repeat("Hello 👾 World! ", 1000)
	b.ResetTimer()
	for b.Loop() {
		helper.Determine(strings.NewReader(largeUTF8))
	}
}

// BenchmarkDetermineLargeCP437 - Large CP-437 file (64KB).
func BenchmarkDetermineLargeCP437(b *testing.B) {
	largeCP437 := strings.Repeat("┌────┐\n│Test│\n└────┘", 100)
	b.ResetTimer()
	for b.Loop() {
		helper.Determine(strings.NewReader(largeCP437))
	}
}
