//go:build linux

package network

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// DefaultRouteInterface inspects Linux /proc/net/ipv6_route to find the interface associated with the default IPv6 route.
func DefaultRouteInterface() (string, error) {
	file, err := os.Open("/proc/net/ipv6_route")
	if err != nil {
		return "", fmt.Errorf("opening /proc/net/ipv6_route: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) < 10 {
			continue
		}

		destPrefix := fields[0]
		prefixLen := fields[1]
		ifaceName := fields[9]

		// Default IPv6 route has destination 00000000000000000000000000000000/00
		if destPrefix == "00000000000000000000000000000000" && prefixLen == "00" && ifaceName != "lo" {
			return ifaceName, nil
		}
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("reading /proc/net/ipv6_route: %w", err)
	}

	return "", fmt.Errorf("default IPv6 route interface not found in /proc/net/ipv6_route")
}
