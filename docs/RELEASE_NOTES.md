# frizzlabs-ddns v1.1.1 Release Notes

Welcome to **`frizzlabs-ddns` v1.1.1** — a lightweight, zero-dependency Dynamic DNS IPv6 reconciliation daemon written in Go.

---

## What's New in v1.1.0

- **3-Way Reconciliation Engine**: Reconciles local IPv6 addresses, cached state, and live DNS AAAA records.
- **DNS Drift Repair**: Automatically repairs out-of-sync or manually edited DNS records.
- **Authoritative NXDOMAIN Repair**: Detects cleared/deleted records on DNS servers and restores them.
- **Zero External Dependencies**: Built exclusively with Go standard library (`net/netip`, `net/http`, `log/slog`, `encoding/json`).
- **Cross-Platform Release Binaries**: Pre-compiled binaries for Linux, Windows, and macOS.

---

## Why Go? (Design Rationale)

`frizzlabs-ddns` was intentionally engineered in Go rather than Python or Bash scripts:
- **Single Native Binary**: No Python interpreter, no `pip` packages, and no `venv` virtual environments to manage, break, or update.
- **Ultra-Lightweight & Fast**: Execution starts in **< 5ms** using **< 5MB RAM** (compared to Python processes taking 30MB+ RAM and interpreter startup overhead).
- **Zero Third-Party Dependency Overhead**: Built 100% on Go's standard library to ensure long-term stability and zero supply-chain risk.

---

## Download Matrix (Which binary do I need?)

| Binary File Name | Target Operating System | Architecture / Hardware |
| :--- | :--- | :--- |
| **`frizzlabs-ddns-linux-amd64`** | Linux (Ubuntu, Debian, Fedora, Arch, CentOS, RHEL) | 64-bit x86_64 (Intel / AMD) |
| **`frizzlabs-ddns-linux-arm64`** | Linux | 64-bit ARM (Raspberry Pi 4/5 64-bit OS, AWS Graviton) |
| **`frizzlabs-ddns-linux-armv7`** | Linux | 32-bit ARM (Raspberry Pi 2/3 32-bit OS, ARMv7) |
| **`frizzlabs-ddns-windows-amd64.exe`** | Windows 10, 11, Windows Server | 64-bit x86_64 (Intel / AMD) |
| **`frizzlabs-ddns-windows-arm64.exe`** | Windows 10, 11 on ARM | 64-bit ARM (Surface Pro X, Windows ARM64) |
| **`frizzlabs-ddns-darwin-amd64`** | macOS | Intel Processors |
| **`frizzlabs-ddns-darwin-arm64`** | macOS | Apple Silicon (M1, M2, M3, M4 Macs) |
| **`checksums.txt`** | All | SHA-256 Checksums for binary verification |

---

## How to Install & Run

### 1. Linux Setup (Recommended)

#### Quick Automated Install Script:
```bash
curl -fsSL https://raw.githubusercontent.com/frizzycodes/frizzlabs-ddns/main/scripts/install.sh | sudo bash
```

#### Manual Binary Setup:
```bash
# Download binary matching your CPU architecture (e.g. amd64)
curl -LO https://github.com/frizzycodes/frizzlabs-ddns/releases/download/v1.1.1/frizzlabs-ddns-linux-amd64
chmod +x frizzlabs-ddns-linux-amd64
sudo mv frizzlabs-ddns-linux-amd64 /usr/local/bin/frizzlabs-ddns

# Create configuration directory and config file
sudo mkdir -p /etc/frizzlabs-ddns
sudo nano /etc/frizzlabs-ddns/config.json
```

**`/etc/frizzlabs-ddns/config.json` Example:**
```json
{
  "version": 1,
  "provider": "duckdns",
  "matchDefaultRoute": true,
  "domain": "your-subdomain-here",
  "token": "your-duckdns-token-here",
  "verifyDNS": true,
  "timeoutSec": 10
}
```

**Test with Dry-Run Mode:**
```bash
frizzlabs-ddns -config /etc/frizzlabs-ddns/config.json -dry-run -verbose
```

---

### 2. Windows Setup (PowerShell / Command Prompt)

1. Download **`frizzlabs-ddns-windows-amd64.exe`** from the release assets below.
2. Create **`config.json`** in the same folder:
   ```json
   {
     "version": 1,
     "provider": "duckdns",
     "interface": "Ethernet",
     "domain": "your-subdomain-here",
     "token": "your-duckdns-token-here",
     "verifyDNS": true,
     "timeoutSec": 10
   }
   ```
3. Test in PowerShell:
   ```powershell
   .\frizzlabs-ddns-windows-amd64.exe -config config.json -dry-run -verbose
   ```

---

### 3. macOS Setup (Apple Silicon & Intel)

```bash
# Download binary matching your CPU architecture (e.g. arm64 for M1/M2/M3/M4)
curl -LO https://github.com/frizzycodes/frizzlabs-ddns/releases/download/v1.1.0/frizzlabs-ddns-darwin-arm64
chmod +x frizzlabs-ddns-darwin-arm64

# Test execution
./frizzlabs-ddns-darwin-arm64 -config config.json -dry-run -verbose
```

---

## Verification & Security

Verify downloaded binaries using `checksums.txt`:
```bash
sha256sum -c checksums.txt --ignore-missing
```
