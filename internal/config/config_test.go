package config

import (
	"testing"
	"time"
)

func TestInt(t *testing.T) {
	t.Setenv("POOL", "12")
	t.Setenv("BLANK", "")
	t.Setenv("JUNK", "twelve")

	for name, tc := range map[string]struct {
		key  string
		def  int
		want int
	}{
		"set":         {"POOL", 5, 12},
		"unset":       {"MISSING", 5, 5},
		"blank":       {"BLANK", 5, 5},
		"unparseable": {"JUNK", 5, 5},
	} {
		t.Run(name, func(t *testing.T) {
			if got := Int(tc.key, tc.def); got != tc.want {
				t.Fatalf("Int(%q, %d) = %d, want %d", tc.key, tc.def, got, tc.want)
			}
		})
	}
}

func TestDuration(t *testing.T) {
	t.Setenv("TIMEOUT", "250ms")
	t.Setenv("JUNK", "soon")

	for name, tc := range map[string]struct {
		key  string
		def  time.Duration
		want time.Duration
	}{
		"set":         {"TIMEOUT", time.Second, 250 * time.Millisecond},
		"unset":       {"MISSING", time.Second, time.Second},
		"unparseable": {"JUNK", time.Second, time.Second},
	} {
		t.Run(name, func(t *testing.T) {
			if got := Duration(tc.key, tc.def); got != tc.want {
				t.Fatalf("Duration(%q) = %v, want %v", tc.key, got, tc.want)
			}
		})
	}
}

func TestString(t *testing.T) {
	t.Setenv("ADDR", ":9090")
	t.Setenv("BLANK", "")

	if got := String("ADDR", ":8080"); got != ":9090" {
		t.Fatalf("String = %q, want :9090", got)
	}
	if got := String("BLANK", ":8080"); got != ":8080" {
		t.Fatalf("String on a blank value = %q, want the default", got)
	}
	if got := String("MISSING", ":8080"); got != ":8080" {
		t.Fatalf("String on an unset key = %q, want the default", got)
	}
}
