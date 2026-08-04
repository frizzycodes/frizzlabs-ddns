// Package network provides network interface inspection and IPv6 address detection using net/netip.
package network

import (
	"context"
	"fmt"
	"net"
	"net/netip"
	"strings"
)

// Detector defines the interface for discovering the system's global IPv6 address.
type Detector interface {
	DetectIPv6(ctx context.Context, interfaceName string, matchDefaultRoute bool) (netip.Addr, error)
}

// DefaultDetector implements Detector using Go's net package and net/netip.
type DefaultDetector struct{}

// NewDetector creates a new instance of DefaultDetector.
func NewDetector() *DefaultDetector {
	return &DefaultDetector{}
}

// IsGlobalIPv6 checks whether a netip.Addr is a routable global IPv6 address.
// It filters out Loopback (::1), Link-Local (fe80::/10), Multicast (ff00::/8),
// Unspecified (::), and Unique Local Addresses (ULA, fc00::/7).
func IsGlobalIPv6(addr netip.Addr) bool {
	if !addr.IsValid() || !addr.Is6() || addr.Is4In6() {
		return false
	}

	if addr.IsLoopback() || addr.IsLinkLocalUnicast() || addr.IsMulticast() || addr.IsUnspecified() {
		return false
	}

	// ULA check: fc00::/7 (Prefix range fc00:: to fdff::)
	b := addr.As16()
	if b[0]&0xfe == 0xfc {
		return false
	}

	return true
}

// DetectIPv6 discovers the primary global IPv6 address based on interface name or default route.
func (d *DefaultDetector) DetectIPv6(ctx context.Context, interfaceName string, matchDefaultRoute bool) (netip.Addr, error) {
	select {
	case <-ctx.Done():
		return netip.Addr{}, ctx.Err()
	default:
	}

	targetInterface := strings.TrimSpace(interfaceName)

	if matchDefaultRoute && targetInterface == "" {
		defIface, err := DefaultRouteInterface()
		if err == nil && defIface != "" {
			targetInterface = defIface
		}
	}

	if targetInterface != "" {
		addr, err := detectIPv6OnInterface(targetInterface)
		if err == nil {
			return addr, nil
		}
		// If specific interface detection failed and matchDefaultRoute was set, fall back to scanning interfaces.
		if !matchDefaultRoute {
			return netip.Addr{}, fmt.Errorf("detecting IPv6 on interface %q: %w", targetInterface, err)
		}
	}

	// Scan all interfaces if targetInterface was empty or fallback needed
	ifaces, err := net.Interfaces()
	if err != nil {
		return netip.Addr{}, fmt.Errorf("enumerating network interfaces: %w", err)
	}

	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		addr, err := detectIPv6OnInterface(iface.Name)
		if err == nil {
			return addr, nil
		}
	}

	return netip.Addr{}, fmt.Errorf("no valid global IPv6 address found on any active network interface")
}

// detectIPv6OnInterface finds the first valid global IPv6 address on a named interface.
func detectIPv6OnInterface(ifaceName string) (netip.Addr, error) {
	iface, err := net.InterfaceByName(ifaceName)
	if err != nil {
		return netip.Addr{}, fmt.Errorf("interface %q not found: %w", ifaceName, err)
	}

	addrs, err := iface.Addrs()
	if err != nil {
		return netip.Addr{}, fmt.Errorf("reading addresses for interface %q: %w", ifaceName, err)
	}

	for _, address := range addrs {
		var ip net.IP
		switch v := address.(type) {
		case *net.IPNet:
			ip = v.IP
		case *net.IPAddr:
			ip = v.IP
		}

		if ip == nil {
			continue
		}

		netAddr, ok := netip.AddrFromSlice(ip)
		if !ok {
			continue
		}

		// Ensure zone identifier is stripped for standard comparisons
		netAddr = netAddr.Unmap()

		if IsGlobalIPv6(netAddr) {
			return netAddr, nil
		}
	}

	return netip.Addr{}, fmt.Errorf("no global IPv6 address configured on interface %q", ifaceName)
}
