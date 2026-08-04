# Architecture Overview - frizzlabs-ddns

`frizzlabs-ddns` is designed following **Clean Architecture**, **SOLID principles**, and **idiomatic Go patterns**.

It is constructed as a lightweight, single-execution daemon (oneshot model) triggered periodically by systemd timers.

---

## High-Level Architecture Diagram

```mermaid
graph TD
    A[Systemd Timer / CLI] -->|Execute| B[cmd/frizzlabs-ddns]
    B -->|1. Load & Validate| C[internal/config]
    B -->|2. Detect IPv6| D[internal/network]
    B -->|3. Read Cache| E[internal/state]
    B -->|4. Reconcile & Compare| F[internal/updater]
    F -->|5. If Changed -> Update| G[internal/provider]
    G -->|DuckDNS Provider| H[DuckDNS REST API]
    G -->|Cloudflare Provider| I[Cloudflare API v4]
    F -->|6. Atomic Save| E
```

---

## Key Modules & Contracts

### 1. Provider Interface (`internal/provider`)

The `Provider` interface abstracts DNS record modification. New DNS providers (e.g. Porkbun, Namecheap) can be added by implementing this contract without modifying any reconciliation logic:

```go
type Provider interface {
    Name() string
    Update(ctx context.Context, ipv6 netip.Addr) error
}
```

### 2. Network Address Detector (`internal/network`)

Uses `net/netip` to inspect interface addresses and parse Linux IPv6 routing tables (`/proc/net/ipv6_route`).

#### IPv6 Address Filtering Rules
- Must be a valid IPv6 address (`addr.Is6()`).
- Excludes IPv4-mapped IPv6 addresses (`!addr.Is4In6()`).
- Excludes Loopback (`::1`), Link-Local (`fe80::/10`), Multicast (`ff00::/8`), and Unspecified (`::`).
- Excludes Unique Local Addresses (ULA, `fc00::/7`).

### 3. State Management (`internal/state`)

State is cached in JSON format containing:
- `version`: Schema versioning.
- `provider`: Provider name.
- `lastIPv6`: Last set global IPv6 address string.
- `lastUpdated`: RFC3339 timestamp.

**Atomic File Writes**: State modifications are written to a temporary file (`state-*.tmp`), flushed to disk via `fsync`, and atomically renamed to the destination state file path.

### 4. Reconciliation Engine (`internal/updater`)

- Detects current IPv6 address.
- Loads cached state.
- Skips DNS API calls if the address is unchanged.
- Executes provider updates with **250ms, 500ms, and 1000ms backoff retries** upon network failure.
