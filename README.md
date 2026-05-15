# socks2proxy

`socks2proxy` is a lightweight SOCKS5 egress router with explicit rule-based
routing.

[![Checks](https://github.com/rohzb/socks2http/actions/workflows/checks.yml/badge.svg)](https://github.com/rohzb/socks2http/actions/workflows/checks.yml)
[![Coverage](https://github.com/rohzb/socks2http/actions/workflows/coverage.yml/badge.svg)](https://github.com/rohzb/socks2http/actions/workflows/coverage.yml)

## Background

In isolated environments, direct web access is often blocked completely and all
outbound connectivity is only allowed through a central HTTP proxy.

That setup works reasonably well for web browsers, but quickly becomes painful
for everything else:

- many applications do not support HTTP proxies at all
- every tool has different proxy configuration behavior
- command-line tools, containers, package managers, and legacy software all
  behave differently and often require separate proxy configuration files,
  environment variables, or application-specific settings
- transparent traffic redirection becomes difficult
- managing proxy settings across systems quickly turns into operational chaos

At first approximation, replacing one proxy configuration mechanism with another
does not sound like much of an improvement. SOCKS can initially look like just
another layer of proxy configuration complexity.

In practice, though, SOCKS is often significantly easier to operationalize.

Instead of configuring every individual application differently, traffic can be
redirected centrally at the system, container, firewall, or router level. That
makes it possible to handle entire groups of applications transparently,
including software that has no native HTTP proxy support at all.

SOCKS-based routing is often a much cleaner solution. Traffic can be redirected
through SOCKS using wrappers, firewall/NAT rules, or router-level policies
without modifying the applications themselves.

The problem is that many restricted environments only provide an HTTP upstream
proxy and no SOCKS infrastructure.

Existing SOCKS-to-HTTP bridge tools came close, but none fit reliably for the
use cases this project was built for. Some lacked routing flexibility, some
were difficult to operate, and some simply did not behave reliably enough in
production environments.

`socks2proxy` was created as a small and predictable building block for exactly
those situations.

## Features

- SOCKS5 server support
- Explicit rule-based routing by destination address and port
- Multiple routing methods:
  - `http`
  - `connect`
  - `direct`
  - `reject`
- HTTP and HTTPS upstream proxy support
- Global and per-rule TLS configuration for HTTPS upstreams
- Simple YAML configuration
- Lightweight single-binary deployment

## Typical Use Cases

- Restricted enterprise or scientific networks
- Environments with HTTP-only outbound access
- Transparent proxy routing with firewall/NAT rules
- MikroTik or Linux router-based traffic redirection
- Routing selected traffic through different upstream proxies
- Providing SOCKS access to applications without native HTTP proxy support

## Quick Start

```bash
make build

cp examples/config.example.yaml config.yaml

./socks2proxy --check --config ./config.yaml
./socks2proxy --config ./config.yaml
```

## Command Line Options

```text
-c, --config string      Path to YAML config file (required)
-l, --log-level string   Override logging.level from config
-t, --check              Parse and validate config, then exit
    --check-config       Alias for --check
-V, --version            Show build version information and exit
-h, --help               Show help and exit
```

## Documentation

- Configuration reference: [docs/CONFIG.md](./docs/CONFIG.md)
- Installation and deployment: [docs/INSTALL.md](./docs/INSTALL.md)
- Release specification: [docs/RELEASE.md](./docs/RELEASE.md)
- Hardening (systemd + AppArmor): [platform/linux/HARDENING.md](./platform/linux/HARDENING.md)
- Example configuration: [examples/config.example.yaml](./examples/config.example.yaml)
- Test matrix examples: [examples/config.test-combinations.yaml](./examples/config.test-combinations.yaml)

## Development

```bash
make fmt
make test
make build
```

## License

Released under the MIT License.

See [LICENSE](./LICENSE) for details.
