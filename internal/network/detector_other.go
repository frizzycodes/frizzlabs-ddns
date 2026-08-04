//go:build !linux

package network

import "fmt"

// DefaultRouteInterface returns an error on non-Linux platforms where procfs is unavailable.
func DefaultRouteInterface() (string, error) {
	return "", fmt.Errorf("default IPv6 route interface auto-matching is only supported natively on Linux")
}
