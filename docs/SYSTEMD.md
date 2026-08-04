# Systemd Setup Guide - frizzlabs-ddns

`frizzlabs-ddns` follows the Unix oneshot daemon model. Instead of running an infinite sleeping loop in memory, systemd manages execution timing efficiently using a systemd timer unit.

---

## 1. Systemd Service File (`/etc/systemd/system/frizzlabs-ddns.service`)

```ini
[Unit]
Description=FrizzLabs Dynamic DNS IPv6 Updater Daemon
Documentation=https://github.com/frizzlabs/frizzlabs-ddns
After=network-online.target
Wants=network-online.target

[Service]
Type=oneshot
User=root
ExecStart=/usr/local/bin/frizzlabs-ddns -config /etc/frizzlabs-ddns/config.json -state /var/lib/frizzlabs-ddns/state.json

# Hardening security sandboxing
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/var/lib/frizzlabs-ddns
PrivateTmp=true
NoNewPrivileges=true

[Install]
WantedBy=multi-user.target
```

---

## 2. Systemd Timer File (`/etc/systemd/system/frizzlabs-ddns.timer`)

```ini
[Unit]
Description=Run FrizzLabs Dynamic DNS Updater every minute
Documentation=https://github.com/frizzlabs/frizzlabs-ddns

[Timer]
OnCalendar=*:0/1
Persistent=true
RandomizedDelaySec=5

[Install]
WantedBy=timers.target
```

---

## 3. Installation & Activation Commands

```bash
# Copy systemd units
sudo cp systemd/frizzlabs-ddns.service /etc/systemd/system/
sudo cp systemd/frizzlabs-ddns.timer /etc/systemd/system/

# Reload systemd manager
sudo systemctl daemon-reload

# Enable and start the timer
sudo systemctl enable --now frizzlabs-ddns.timer

# Check timer status
sudo systemctl status frizzlabs-ddns.timer

# View journald logs
sudo journalctl -u frizzlabs-ddns.service -f
```
