# Configuration Reference - frizzlabs-ddns

`frizzlabs-ddns` uses a JSON configuration file. By default, it looks for `/etc/frizzlabs-ddns/config.json` or `config.json` in the current working directory.

---

## Configuration Example

```json
{
  "version": 1,
  "provider": "duckdns",
  "interface": "enp4s0",
  "matchDefaultRoute": true,
  "domain": "my-homelab",
  "token": "00000000-0000-0000-0000-000000000000",
  "timeoutSec": 10
}
```

---

## Field Specifications

| Field | Type | Required | Description |
| :--- | :--- | :--- | :--- |
| `version` | Integer | **Yes** | Config schema version. Must be `>= 1`. |
| `provider` | String | **Yes** | DNS provider identifier (`"duckdns"`, `"cloudflare"`, or `"noop"`). Defaults to `"duckdns"`. |
| `interface` | String | Conditional | Target network interface name (e.g. `"enp4s0"`, `"eth0"`). Required if `matchDefaultRoute` is `false`. |
| `matchDefaultRoute` | Boolean | Conditional | Automatically discovers interface owning the default IPv6 route on Linux when `true`. |
| `domain` | String | **Yes** | Sub-domain or domain name registered with the DNS provider. |
| `token` | String | **Yes** | Provider authentication token or API token. |
| `timeoutSec` | Integer | No | HTTP request timeout in seconds. Defaults to `10` seconds. |
| `stateFile` | String | No | Custom override path for `state.json`. |

---

## Environment & Path Precedence

1. CLI Flag `-config <path>` / `-c <path>`
2. `/etc/frizzlabs-ddns/config.json` (if present)
3. `./config.json`
