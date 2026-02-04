package helper_test

import (
	"bytes"
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
