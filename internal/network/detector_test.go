package network_test

import (
	"context"
	"net/netip"
	"testing"
	"time"

	"github.com/frizzlabs/frizzlabs-ddns/internal/network"
)

func TestIsGlobalIPv6(t *testing.T) {
	tests := []struct {
		name     string
		addrStr  string
		isGlobal bool
	}{
		{
			name:     "Global Unicast 2001:db8::1",
			addrStr:  "2001:db8::1",
			isGlobal: true,
		},
		{
			name:     "Global Unicast 2607:f8b0:4005:805::200e",
			addrStr:  "2607:f8b0:4005:805::200e",
			isGlobal: true,
		},
		{
			name:     "Loopback ::1",
			addrStr:  "::1",
			isGlobal: false,
		},
		{
			name:     "Link-Local fe80::1",
			addrStr:  "fe80::1",
			isGlobal: false,
		},
		{
			name:     "ULA fc00::1",
			addrStr:  "fc00::1",
			isGlobal: false,
		},
		{
			name:     "ULA fd00::1234",
			addrStr:  "fd00::1234",
			isGlobal: false,
		},
		{
			name:     "Multicast ff02::1",
			addrStr:  "ff02::1",
			isGlobal: false,
		},
		{
			name:     "Unspecified ::",
			addrStr:  "::",
			isGlobal: false,
		},
		{
			name:     "IPv4 192.168.1.1",
			addrStr:  "192.168.1.1",
			isGlobal: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			addr := netip.MustParseAddr(tt.addrStr)
			got := network.IsGlobalIPv6(addr)
			if got != tt.isGlobal {
				t.Errorf("IsGlobalIPv6(%s) = %v; want %v", tt.addrStr, got, tt.isGlobal)
			}
		})
	}
}

func TestDetectorContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	det := network.NewDetector()
	_, err := det.DetectIPv6(ctx, "", false)
	if err == nil {
		t.Errorf("expected error when context is cancelled, got nil")
	}
}

func BenchmarkDetectorIsGlobalIPv6(b *testing.B) {
	addr := netip.MustParseAddr("2001:db8::1")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = network.IsGlobalIPv6(addr)
	}
}

func BenchmarkDetectorExecution(b *testing.B) {
	det := network.NewDetector()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = det.DetectIPv6(ctx, "", false)
	}
}
