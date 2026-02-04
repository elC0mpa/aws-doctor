//go:build !windows

package utils //nolint:revive

// EnableANSI is a no-op on non-Windows; ANSI escape sequences are supported by default.
func EnableANSI() {
}
