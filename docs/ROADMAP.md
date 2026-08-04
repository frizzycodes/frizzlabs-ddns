# Project Roadmap - frizzlabs-ddns

This document outlines the planned evolutionary stages for `frizzlabs-ddns`.

---

## Version 1.0 (Current Release)
- [x] Global IPv6 address auto-detection (`net/netip`).
- [x] Linux default route interface auto-matching (`/proc/net/ipv6_route`).
- [x] Filtering out Loopback, Link-Local, ULA (`fc00::/7`), and Multicast IPv6 addresses.
- [x] DuckDNS provider implementation with HTTP retries and exponential backoff (250ms, 500ms, 1000ms).
- [x] Cloudflare provider implementation (API v4).
- [x] NoOp provider implementation for testing and dry-runs.
- [x] Atomic state persistence via JSON (`state.json`).
- [x] Structured logging with `log/slog` (Text, JSON, journald support).
- [x] Oneshot Systemd service and 1-minute timer units.
- [x] Zero third-party dependencies.

---

## Version 2.0 (Planned)
- [ ] Additional DNS Providers: Porkbun, Namecheap, Route53.
- [ ] Netlink event listener for instant IP change triggers (avoiding timer polling).
- [ ] Multi-interface support.
- [ ] Webhook notifications on IP address change (Discord, Slack, Matrix).

---

## Version 3.0 (Future Vision)
- [ ] Native Prometheus metrics exporter endpoint (`/metrics`).
- [ ] OpenTelemetry tracing.
- [ ] Healthcheck HTTP endpoint for container orchestrators.
