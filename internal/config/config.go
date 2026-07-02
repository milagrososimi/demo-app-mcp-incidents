// Package config reads a service's settings from the environment.
//
// Settings come from the environment because that is what the manifests under
// deploy/ set. Changing how a service behaves in production is a change to its
// manifest, not to its code — which is why those files are worth reading when
// a service starts behaving differently.
package config

import (
	"log/slog"
	"os"
	"strconv"
	"time"
)

// Int returns the value of key, or def when it is unset or unparseable.
//
// An unparseable value warns rather than exits: a service that refuses to
// start cannot report what was wrong with its configuration.
func Int(key string, def int) int {
	raw, ok := lookup(key)
	if !ok {
		return def
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		slog.Warn("ignoring unparseable setting", "key", key, "value", raw, "using", def)
		return def
	}
	return value
}

// Duration returns the value of key as a duration (e.g. "5s"), or def.
func Duration(key string, def time.Duration) time.Duration {
	raw, ok := lookup(key)
	if !ok {
		return def
	}
	value, err := time.ParseDuration(raw)
	if err != nil {
		slog.Warn("ignoring unparseable setting", "key", key, "value", raw, "using", def)
		return def
	}
	return value
}

// String returns the value of key, or def when it is unset or blank.
func String(key, def string) string {
	if raw, ok := lookup(key); ok {
		return raw
	}
	return def
}

func lookup(key string) (string, bool) {
	raw, ok := os.LookupEnv(key)
	if !ok || raw == "" {
		return "", false
	}
	return raw, true
}
