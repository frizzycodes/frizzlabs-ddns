# DuckDNS Setup Guide - frizzlabs-ddns

[DuckDNS](https://www.duckdns.org) is a free Dynamic DNS service.

---

## 1. Obtain DuckDNS Token & Domain

1. Log in to [DuckDNS](https://www.duckdns.org) using Persona, GitHub, Google, or Reddit.
2. Note your **Account Token** listed on the DuckDNS header (UUID format, e.g. `a1b2c3d4-e5f6-7890-abcd-ef0123456789`).
3. Create a domain under **domains** (e.g. `frizzlabs`). Your full domain will be `frizzlabs.duckdns.org`.

---

## 2. Configure `frizzlabs-ddns`

Create or edit `/etc/frizzlabs-ddns/config.json`:

```json
{
  "version": 1,
  "provider": "duckdns",
  "interface": "enp4s0",
  "matchDefaultRoute": true,
  "domain": "frizzlabs",
  "token": "a1b2c3d4-e5f6-7890-abcd-ef0123456789"
}
```

---

## 3. Verify Setup with Dry-Run Mode

```bash
frizzlabs-ddns -config /etc/frizzlabs-ddns/config.json -dry-run -verbose
```

If successful, run without `-dry-run` to execute the first update:

```bash
frizzlabs-ddns -config /etc/frizzlabs-ddns/config.json -verbose
```
