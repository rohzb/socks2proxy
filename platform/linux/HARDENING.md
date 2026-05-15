# Hardening Guide (systemd + AppArmor)

This document provides example hardening assets for `socks2proxy`:

- `socks2proxy.hardened.service` (systemd sandboxing)
- `apparmor/usr.local.bin.socks2proxy` (AppArmor confinement)
- `install-hardening.sh` / `uninstall-hardening.sh` helper scripts

These examples are intentionally strict and may need local adjustments.

## Files

- Service unit: `platform/linux/systemd/socks2proxy.hardened.service`
- AppArmor profile: `platform/linux/apparmor/usr.local.bin.socks2proxy`
- Installer script: `platform/linux/scripts/install-hardening.sh`
- Uninstaller script: `platform/linux/scripts/uninstall-hardening.sh`

## Fast path (scripts)

From repository root:

```bash
sudo ./platform/linux/scripts/install-hardening.sh
```

Uninstall:

```bash
sudo ./platform/linux/scripts/uninstall-hardening.sh
```

Optional uninstall behavior:

- Keep config and user (default): `KEEP_CONFIG=1 KEEP_USER=1`
- Remove config too: `KEEP_CONFIG=0`
- Remove service user too: `KEEP_USER=0`

Example:

```bash
sudo KEEP_CONFIG=0 KEEP_USER=0 ./platform/linux/scripts/uninstall-hardening.sh
```

Installer script supports path overrides via env vars:

- `BINARY_SRC`, `CONFIG_SRC`
- `BINARY_DST`, `CONFIG_DIR`, `CONFIG_DST`
- `UNIT_SRC`, `UNIT_DST`
- `APPARMOR_SRC`, `APPARMOR_DST`

## Install steps

1. Create service user:

```bash
sudo useradd --system --home-dir /nonexistent --shell /usr/sbin/nologin socks2proxy
```

2. Install binary and config:

```bash
sudo install -m 0755 ./socks2proxy /usr/local/bin/socks2proxy
sudo install -d -m 0750 /usr/local/etc/socks2proxy
sudo install -m 0640 ./config.yaml /usr/local/etc/socks2proxy/config.yaml
sudo chown root:socks2proxy /usr/local/etc/socks2proxy/config.yaml
```

3. Install AppArmor profile:

```bash
sudo install -m 0644 ./platform/linux/apparmor/usr.local.bin.socks2proxy /etc/apparmor.d/usr.local.bin.socks2proxy
sudo apparmor_parser -r /etc/apparmor.d/usr.local.bin.socks2proxy
sudo aa-status | grep socks2proxy || true
```

4. Install hardened service unit:

```bash
sudo install -m 0644 ./platform/linux/systemd/socks2proxy.hardened.service /etc/systemd/system/socks2proxy.service
sudo systemctl daemon-reload
sudo systemctl enable --now socks2proxy.service
```

5. Validate service health:

```bash
systemctl status socks2proxy.service --no-pager
journalctl -u socks2proxy.service -n 100 --no-pager
```

## Why these settings

- `User` / `Group`: run without root privileges.
- `NoNewPrivileges=true`: prevents gaining extra privileges at runtime.
- `ProtectSystem=strict`, `ProtectHome=true`: make filesystem mostly read-only.
- `PrivateDevices=true`, `ProtectKernel*`, `ProtectControlGroups=true`: reduce kernel and device attack surface.
- `RestrictAddressFamilies=AF_INET AF_INET6 AF_UNIX`: allow only expected socket families.
- `SystemCallFilter=@system-service` + deny privileged/resource classes: narrow syscall surface.
- `MemoryDenyWriteExecute=true`: blocks writable+executable mappings.
- `UMask=0077`: strict default file permissions.
- AppArmor profile: explicit read-only paths and deny writes by default.

## Common adjustments

- If config path differs, update both unit and AppArmor profile.
- If `tls.ca_cert_file` points to a custom path, allow read access in:
  - systemd `ReadOnlyPaths=`
  - AppArmor profile path rules
- If DNS/NSS differs on your distro, add required read paths.

## Troubleshooting

- AppArmor denials:

```bash
sudo dmesg | grep -i apparmor
sudo journalctl -k -g apparmor --no-pager
```

- systemd sandbox denials:

```bash
journalctl -u socks2proxy.service -b --no-pager
```

Relax only the specific rule that blocks legitimate behavior.
