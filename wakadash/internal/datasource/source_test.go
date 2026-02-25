package datasource

import (
	"testing"
	"time"
)

// TestIsRecent verifies date-based routing logic.
func TestIsRecent(t *testing.T) {
	tests := []struct {
		name     string
		date     string
		expected bool
	}{
		{
			name:     "today is recent",
			date:     time.Now().Format("2006-01-02"),
			expected: true,
		},
		{
			name:     "7 days ago is recent (boundary)",
			date:     time.Now().AddDate(0, 0, -7).Format("2006-01-02"),
			expected: true,
		},
		{
			name:     "8 days ago is not recent (boundary)",
			date:     time.Now().AddDate(0, 0, -8).Format("2006-01-02"),
			expected: false,
		},
		{
			name:     "30 days ago is not recent",
			date:     time.Now().AddDate(0, 0, -30).Format("2006-01-02"),
			expected: false,
		},
		{
			name:     "1 day ago is recent",
			date:     time.Now().AddDate(0, 0, -1).Format("2006-01-02"),
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ds := &DataSource{}
			result := ds.IsRecent(tt.date)
			if result != tt.expected {
				t.Errorf("IsRecent(%s) = %v, want %v", tt.date, result, tt.expected)
			}
		})
	}
}
