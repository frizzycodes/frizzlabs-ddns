# frizzlabs-ddns

[![CI](https://github.com/frizzlabs/frizzlabs-ddns/actions/workflows/ci.yml/badge.svg)](https://github.com/frizzlabs/frizzlabs-ddns/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/frizzlabs/frizzlabs-ddns)](https://goreportcard.com/report/github.com/frizzlabs/frizzlabs-ddns)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/Go-1.21%2B-00ADD8.svg)](go.mod)

A lightweight, zero-dependency Dynamic DNS (DDNS) daemon written in Go for Linux homelab servers.

`frizzlabs-ddns` monitors global IPv6 addresses on network interfaces, detects default routes automatically, and updates DuckDNS (and Cloudflare) whenever your address changes.

---

## Key Features

- **IPv6 First**: Specialized for modern homelab architectures with static IPv4 edge gateways and dynamic IPv6 host addresses.
- **Zero Third-Party Dependencies**: Built exclusively using standard Go packages (`net/netip`, `net/http`, `log/slog`, `encoding/json`).
- **Modern `net/netip`**: High-performance, allocation-free IP parsing and filtering.
- **Dual Interface Modes**: Select a specific interface (e.g. `enp4s0`) or enable automatic default route discovery (`matchDefaultRoute = true`).
- **Strict IPv6 Address Filtering**: Ignores Loopback (`::1`), Link-Local (`fe80::/10`), ULA (`fc00::/7`), Multicast (`ff00::/8`), and Unspecified (`::`).
- **Pluggable Architecture**: Easily extensible provider design (`DuckDNS`, `Cloudflare`, `NoOp`).
- **Atomic State Persistence**: Prevents state file corruption via atomic temporary writes and `fsync`.
- **Systemd Oneshot Execution**: Designed for periodic systemd timer invocation (no background sleeping loops).
- **HTTP Backoff Retries**: Automatic retry mechanism with 250ms, 500ms, and 1000ms delays.
- **Journald Logging**: Clean structured output powered by `log/slog`.

---

## Quickstart & Installation

### Option 1: Automated Shell Installer (Linux)

```bash
curl -fsSL https://raw.githubusercontent.com/frizzlabs/frizzlabs-ddns/main/scripts/install.sh | sudo bash
```

### Option 2: Build & Install via Makefile

```bash
git clone https://github.com/frizzlabs/frizzlabs-ddns.git
cd frizzlabs-ddns
make
sudo make install
```

---

## Configuration

Configuration is stored in `/etc/frizzlabs-ddns/config.json`:

```json
{
  "version": 1,
  "provider": "duckdns",
  "interface": "enp4s0",
  "matchDefaultRoute": true,
  "domain": "frizzlabs",
  "token": "your-duckdns-token-here",
  "timeoutSec": 10
}
```

### Test Configuration (Dry-Run Mode)

Run `frizzlabs-ddns` in dry-run mode to verify network detection and provider setup without making external API changes:

```bash
frizzlabs-ddns -config /etc/frizzlabs-ddns/config.json -dry-run -verbose
```

For complete field explanations, see [docs/CONFIGURATION.md](docs/CONFIGURATION.md).

---

## Systemd Integration

`frizzlabs-ddns` is executed every minute via a systemd timer.

```bash
# Enable and start systemd timer
sudo systemctl enable --now frizzlabs-ddns.timer

# View execution logs
sudo journalctl -u frizzlabs-ddns.service -f
```

For detailed systemd service setup, see [docs/SYSTEMD.md](docs/SYSTEMD.md).

---

## Command Line Usage

```
frizzlabs-ddns - Dynamic DNS IPv6 Updater Daemon

Usage:
  frizzlabs-ddns [options]

Options:
  -c, -config string  Path to JSON configuration file
  -s, -state string   Path to state file (overrides config setting)
  -n, -dry-run        Simulate execution without modifying DNS or state file
  -v, -verbose        Enable verbose debug logging
  -json               Output log messages in JSON format
  -V, -version        Display version and build metadata
```

---

## Documentation

- [Architecture Overview](docs/ARCHITECTURE.md)
- [Configuration Reference](docs/CONFIGURATION.md)
- [DuckDNS Setup Guide](docs/DUCKDNS.md)
- [Systemd Integration](docs/SYSTEMD.md)
- [Development Guide](docs/DEVELOPMENT.md)
- [Roadmap](docs/ROADMAP.md)
- [Troubleshooting](docs/TROUBLESHOOTING.md)

---

## License

`frizzlabs-ddns` is released under the [MIT License](LICENSE).
