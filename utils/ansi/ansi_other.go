//go:build !windows

package ansi

// EnableANSI is a no-op on non-Windows; ANSI escape sequences are supported by default.
func EnableANSI() {
}
