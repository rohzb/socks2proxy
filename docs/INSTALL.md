# Installation and Deployment

This guide covers common ways to install and run `socks2proxy`.

## 1. Build from source

```bash
cd /workspaces/ansible-stack-ist/local/gosr/socks2http
make build
```

This produces:

- `./socks2proxy`

## 2. Prepare configuration

```bash
cp ./examples/config.example.yaml ./config.yaml
./socks2proxy --check --config ./config.yaml
```

Configuration details are documented in [CONFIG.md](./CONFIG.md).

## 3. Run manually

```bash
./socks2proxy --config ./config.yaml
```

Notes:

- `--config` is required
- `--log-level` can temporarily override `logging.level`

## 4. Build release artifacts

```bash
make build-all
```

Artifacts are written to `dist/`.

## 5. Install as a system service

Platform assets are organized under `platform/`:

- Linux: `platform/linux/`
- macOS: `platform/darwin/`
- Windows: `platform/windows/`

Linux deployment is fully documented below. Darwin and Windows currently
provide service/script templates that should be adapted to local policy.

### Standard service example

```bash
sudo install -m 0755 ./socks2proxy /usr/local/bin/socks2proxy
sudo install -d -m 0750 /usr/local/etc/socks2proxy
sudo install -m 0640 ./config.yaml /usr/local/etc/socks2proxy/config.yaml
sudo install -m 0644 ./platform/linux/systemd/socks2proxy.service /etc/systemd/system/socks2proxy.service
sudo systemctl daemon-reload
sudo systemctl enable --now socks2proxy.service
```

### Hardened service + AppArmor

Use the hardened profile and helper scripts:

```bash
sudo ./platform/linux/scripts/install-hardening.sh
```

Details and rationale:

- [platform/linux/HARDENING.md](./platform/linux/HARDENING.md)

### Darwin and Windows templates

- macOS launchd template and scripts:
  - [platform/darwin/README.md](./platform/darwin/README.md)
- Windows service template and scripts:
  - [platform/windows/README.md](./platform/windows/README.md)

## 6. Verify runtime

```bash
./socks2proxy --version
./socks2proxy --check --config ./config.yaml
systemctl status socks2proxy.service --no-pager
```

## 7. Uninstall hardened deployment

```bash
sudo ./platform/linux/scripts/uninstall-hardening.sh
```
