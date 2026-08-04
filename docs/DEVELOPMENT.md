# Development Guide - frizzlabs-ddns

Welcome to the development documentation for `frizzlabs-ddns`.

## Prerequisites

- Go 1.21 or higher installed on your host system.
- Git.
- `make` (optional, for convenient targets).

---

## Building & Testing

### Build Binary
```bash
make build
# Output compiled binary: bin/frizzlabs-ddns
```

### Run Unit Tests & Race Detector
```bash
make test
```

### Run Benchmarks
```bash
make bench
```

### Code Formatting & Static Analysis
```bash
make fmt
make vet
make lint
```

---

## Adding a New DNS Provider

To add a new DNS provider (e.g. `porkbun`):

1. Create directory `internal/provider/porkbun/`.
2. Implement `porkbun.go` adhering to the `provider.Provider` interface:
   ```go
   type Provider interface {
       Name() string
       Update(ctx context.Context, ipv6 netip.Addr) error
   }
   ```
3. Register the new provider in `internal/provider/provider.go` inside the `New(...)` factory function.
4. Add unit tests in `internal/provider/porkbun/porkbun_test.go` using `net/http/httptest`.
