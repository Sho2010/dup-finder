package interactive

import (
	"testing"
)

func TestFormatSize(t *testing.T) {
	tests := []struct {
		name     string
		bytes    int64
		expected string
	}{
		{"zero bytes", 0, "0 B"},
		{"bytes", 100, "100 B"},
		{"kilobytes", 1024, "1.0 KB"},
		{"kilobytes with decimal", 1536, "1.5 KB"},
		{"megabytes", 1048576, "1.0 MB"},
		{"megabytes with decimal", 1572864, "1.5 MB"},
		{"gigabytes", 1073741824, "1.0 GB"},
		{"gigabytes with decimal", 1610612736, "1.5 GB"},
		{"terabytes", 1099511627776, "1.0 TB"},
		{"large file", 5368709120, "5.0 GB"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatSize(tt.bytes)
			if result != tt.expected {
				t.Errorf("formatSize(%d) = %s; want %s", tt.bytes, result, tt.expected)
			}
		})
	}
}

func TestFormatSizeEdgeCases(t *testing.T) {
	t.Run("1023 bytes", func(t *testing.T) {
		result := formatSize(1023)
		if result != "1023 B" {
			t.Errorf("formatSize(1023) = %s; want 1023 B", result)
		}
	})

	t.Run("exactly 1 KB", func(t *testing.T) {
		result := formatSize(1024)
		if result != "1.0 KB" {
			t.Errorf("formatSize(1024) = %s; want 1.0 KB", result)
		}
	})

	t.Run("exactly 1 MB", func(t *testing.T) {
		result := formatSize(1048576)
		if result != "1.0 MB" {
			t.Errorf("formatSize(1048576) = %s; want 1.0 MB", result)
		}
	})
}
