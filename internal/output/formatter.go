package output

import (
	"fmt"
	"strings"

	"github.com/Sho2010/dup-finder/internal/models"
)

// Formatter defines the interface for formatting output
type Formatter interface {
	FormatPairComparison(comparison models.PairComparison) string
}

// SimpleFormatter provides a simple text-based output format
type SimpleFormatter struct {
	showHash bool
}

// NewSimpleFormatter creates a new simple formatter
func NewSimpleFormatter(showHash bool) *SimpleFormatter {
	return &SimpleFormatter{showHash: showHash}
}

// FormatPairComparison formats a pair comparison result
func (sf *SimpleFormatter) FormatPairComparison(comparison models.PairComparison) string {
	var builder strings.Builder

	// Header
	builder.WriteString(fmt.Sprintf("=== %s ↔ %s ===\n", comparison.Dir1, comparison.Dir2))

	if len(comparison.Matches) == 0 {
		builder.WriteString("(No duplicates)\n")
		return builder.String()
	}

	// List each matching file
	for _, match := range comparison.Matches {
		if sf.showHash && match.HashChecked {
			// Show hash comparison result
			hashStatus := "✓ Identical"
			if !match.HashMatch {
				hashStatus = "✗ Different"
			}
			builder.WriteString(fmt.Sprintf("%-20s ✓ [Hash: %s]\n", match.Filename+":", hashStatus))
		} else {
			// Just show the filename match
			builder.WriteString(fmt.Sprintf("%-20s ✓\n", match.Filename+":"))
		}
	}

	return builder.String()
}

// FormatAllComparisons formats all pair comparisons
func FormatAllComparisons(comparisons []models.PairComparison, showHash bool) string {
	formatter := NewSimpleFormatter(showHash)
	var builder strings.Builder

	for i, comparison := range comparisons {
		builder.WriteString(formatter.FormatPairComparison(comparison))
		// Add blank line between pairs (except after the last one)
		if i < len(comparisons)-1 {
			builder.WriteString("\n")
		}
	}

	return builder.String()
}
