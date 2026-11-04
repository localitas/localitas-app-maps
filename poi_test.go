package maps

import (
	"testing"
)

func TestJoinNonEmpty(t *testing.T) {
	tests := []struct {
		parts []string
		sep   string
		want  string
	}{
		{[]string{"Apple Park", "Cupertino"}, ", ", "Apple Park, Cupertino"},
		{[]string{"", "test", ""}, "-", "test"},
		{[]string{}, ", ", ""},
	}
	for _, tt := range tests {
		got := joinNonEmpty(tt.parts, tt.sep)
		if got != tt.want {
			t.Errorf("joinNonEmpty(%v, %q) = %q, want %q", tt.parts, tt.sep, got, tt.want)
		}
	}
}
