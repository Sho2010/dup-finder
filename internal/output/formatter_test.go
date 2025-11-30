package output

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"dup-finder/internal/models"
)

func TestSimpleFormatter_FormatPairComparison_NoMatches(t *testing.T) {
	formatter := NewSimpleFormatter(false)
	comparison := models.PairComparison{
		Dir1:    "/path/to/dir1",
		Dir2:    "/path/to/dir2",
		Matches: []models.FileMatch{},
	}

	result := formatter.FormatPairComparison(comparison)

	assert.Contains(t, result, "=== /path/to/dir1 ↔ /path/to/dir2 ===")
	assert.Contains(t, result, "(No duplicates)")
}

func TestSimpleFormatter_FormatPairComparison_WithMatches(t *testing.T) {
	formatter := NewSimpleFormatter(false)
	comparison := models.PairComparison{
		Dir1: "/path/to/dir1",
		Dir2: "/path/to/dir2",
		Matches: []models.FileMatch{
			{
				Filename:    "file1.txt",
				HashChecked: false,
			},
			{
				Filename:    "file2.txt",
				HashChecked: false,
			},
		},
	}

	result := formatter.FormatPairComparison(comparison)

	assert.Contains(t, result, "=== /path/to/dir1 ↔ /path/to/dir2 ===")
	assert.Contains(t, result, "file1.txt:")
	assert.Contains(t, result, "file2.txt:")
	assert.Contains(t, result, "✓")
}

func TestSimpleFormatter_FormatPairComparison_WithHashIdentical(t *testing.T) {
	formatter := NewSimpleFormatter(true)
	comparison := models.PairComparison{
		Dir1: "/path/to/dir1",
		Dir2: "/path/to/dir2",
		Matches: []models.FileMatch{
			{
				Filename:    "file1.txt",
				HashChecked: true,
				HashMatch:   true,
			},
		},
	}

	result := formatter.FormatPairComparison(comparison)

	assert.Contains(t, result, "file1.txt:")
	assert.Contains(t, result, "[Hash: ✓ Identical]")
}

func TestSimpleFormatter_FormatPairComparison_WithHashDifferent(t *testing.T) {
	formatter := NewSimpleFormatter(true)
	comparison := models.PairComparison{
		Dir1: "/path/to/dir1",
		Dir2: "/path/to/dir2",
		Matches: []models.FileMatch{
			{
				Filename:    "file1.txt",
				HashChecked: true,
				HashMatch:   false,
			},
		},
	}

	result := formatter.FormatPairComparison(comparison)

	assert.Contains(t, result, "file1.txt:")
	assert.Contains(t, result, "[Hash: ✗ Different]")
}

func TestFormatAllComparisons(t *testing.T) {
	comparisons := []models.PairComparison{
		{
			Dir1: "/dir1",
			Dir2: "/dir2",
			Matches: []models.FileMatch{
				{Filename: "file1.txt", HashChecked: true, HashMatch: true},
			},
		},
		{
			Dir1: "/dir1",
			Dir2: "/dir3",
			Matches: []models.FileMatch{
				{Filename: "file2.txt", HashChecked: true, HashMatch: false},
			},
		},
	}

	result := FormatAllComparisons(comparisons, true)

	// Should contain both comparisons
	assert.Contains(t, result, "/dir1 ↔ /dir2")
	assert.Contains(t, result, "/dir1 ↔ /dir3")
	assert.Contains(t, result, "file1.txt:")
	assert.Contains(t, result, "file2.txt:")

	// Should have blank line between comparisons
	lines := strings.Split(result, "\n")
	assert.Greater(t, len(lines), 4, "Should have multiple lines with separator")
}

func TestFormatAllComparisons_WithoutHash(t *testing.T) {
	comparisons := []models.PairComparison{
		{
			Dir1: "/dir1",
			Dir2: "/dir2",
			Matches: []models.FileMatch{
				{Filename: "file1.txt"},
			},
		},
	}

	result := FormatAllComparisons(comparisons, false)

	assert.Contains(t, result, "file1.txt:")
	assert.Contains(t, result, "✓")
	assert.NotContains(t, result, "Hash:")
}
