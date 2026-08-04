# Troubleshooting Guide - frizzlabs-ddns

This document provides diagnostic steps for common issues encountered when running `frizzlabs-ddns`.

---

## Exit Code Taxonomy

| Exit Code | Classification | Root Cause & Resolution |
| :---: | :--- | :--- |
| `0` | **Success** | Execution succeeded or IPv6 address was unchanged. |
| `1` | **Configuration Error** | Missing domain, invalid token, unsupported provider, or unparseable JSON file in `/etc/frizzlabs-ddns/config.json`. |
| `2` | **Network Error** | No global IPv6 address configured on the interface, interface offline, or route discovery failed. |
| `3` | **State Error** | Permission denied writing to `/var/lib/frizzlabs-ddns/state.json` or corrupted state JSON file. |
| `4` | **Provider Error** | DNS Provider HTTP request failed after retries (e.g. invalid DuckDNS token, network timeout, HTTP 500/KO). |

---

## Common Issues & Diagnoses

### Issue 1: Exit Code 2 - `no valid global IPv6 address found`
- **Cause**: Network interface only has Link-Local IPv6 (`fe80::/10`) or ULA (`fc00::/7`), or privacy extensions are delaying global address assignment.
- **Diagnostic**: Run `ip -6 addr show dev <interface>` and verify a global unicast address (e.g. `2001:...` or `2600:...`) is present and `scope global` is active.

### Issue 2: Exit Code 3 - `permission denied saving state`
- **Cause**: The process lacks write permission to `/var/lib/frizzlabs-ddns`.
- **Diagnostic**: Ensure directory exists with proper permissions:
  ```bash
  sudo mkdir -p /var/lib/frizzlabs-ddns
  sudo chmod 0755 /var/lib/frizzlabs-ddns
  ```

### Issue 3: Exit Code 4 - `duckdns: update failed with server response: "KO"`
- **Cause**: DuckDNS rejected the update request.
- **Resolution**: Verify that `domain` (sub-domain without `.duckdns.org`) and `token` in `config.json` match your DuckDNS dashboard.
