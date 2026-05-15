# socks2proxy configuration reference

This document describes all supported configuration keys for `socks2proxy`.
It reflects current runtime and `--check` validation behavior.

## Full example

```yaml
listen: ":41080"

allowed_client_addresses:
  - "172.31.11.0/24"
  - "127.0.0.1/32"

tls:
  insecure_skip_verify: false
  min_version: "1.2"

routing:
  default:
    method: "reject"

  rules:
    - dst_ports: [80]
      dst_addresses: ["0.0.0.0/0", "::/0"]
      method: "http"
      upstream: "http://proxy.example.com:3128"

    - dst_ports: [443]
      dst_addresses: ["0.0.0.0/0", "::/0"]
      method: "connect"
      upstream: "http://proxy.example.com:3128"

    - dst_ports: [22]
      dst_addresses: ["0.0.0.0/0", "::/0"]
      method: "direct"

    - dst_ports: [25]
      dst_addresses: ["0.0.0.0/0", "::/0"]
      method: "reject"

timeouts:
  connect: "10s"
  idle: "300s"

http:
  max_header_bytes: 65536

logging:
  level: "info"
```

## Top-level keys

| Key | Type | Required | Description |
|---|---|---|---|
| `listen` | string | yes | TCP listen address for SOCKS5 server. |
| `allowed_client_addresses` | list[string] | yes | Client source address allowlist (single IP, CIDR, or IP range). |
| `tls` | object | no | Global TLS defaults for HTTPS upstream proxy connections. |
| `routing` | object | yes | Routing rules by destination ports/addresses plus optional default rule. |
| `timeouts` | object | yes | Connect and idle timeout values. |
| `http` | object | yes | HTTP parser/server limits. |
| `logging` | object | yes | Logging settings. |

## `listen`

- Format: Go TCP address format, for example:
  - `":41080"` (all interfaces, port 41080)
  - `"127.0.0.1:41080"`
  - `"[::1]:41080"`
- Validation:
  - must resolve as TCP listen address
  - port must be in range `1..65535`

## `allowed_client_addresses`

List of source client address selectors allowed to connect to the SOCKS5 listener.

Accepted forms (scalar or list, CSV supported):
- single IP (for example `10.0.0.10`)
- CIDR (for example `10.0.0.0/24`)
- IP range (for example `10.0.0.10-10.0.0.50`)

Notes:
- Each entry must be a valid address selector.
- If a client IP does not match any selector, connection is denied.

## `routing`

Routing has two parts:

- `routing.rules`: first-match rules by `dst_port(s)` and required `dst_address(es)`
- `routing.default`: fallback when no rule matches

### `routing.default`

`routing.default` is optional.

If `routing.default` is not specified, an implicit fallback is used:

```yaml
routing:
  default:
    method: "reject"
```

That means unmatched ports are rejected unless you set another default.

### `routing.rules[]` fields

| Field | Type | Required | Description |
|---|---|---|---|
| `dst_ports` / `dst_port` | list/scalar | yes | Destination TCP port selector(s); accepts list or comma-separated/range scalar. |
| `dst_addresses` / `dst_address` | list/scalar | yes | Destination address selector(s); accepts list or comma-separated scalar. |
| `method` | string | yes | One of `http`, `connect`, `direct`, `reject`. |
| `upstream` | string | conditional | Required for `http` and `connect`; forbidden for `direct` and `reject`. |
| `tls` | object | optional | Optional HTTPS upstream TLS settings. |

### `routing.default` fields

| Field | Type | Required | Description |
|---|---|---|---|
| `method` | string | yes | One of `http`, `connect`, `direct`, `reject`. |
| `upstream` | string | conditional | Required for `http` and `connect`; forbidden for `direct` and `reject`. |
| `tls` | object | optional | Optional HTTPS upstream TLS settings. |

### Methods

- `http`
  - Reads HTTP/1.x request from client.
  - Sends absolute-form request to upstream proxy.

- `connect`
  - Sends upstream `CONNECT host:port`.
  - On success, tunnels bytes bidirectionally.

- `direct`
  - Opens direct TCP connection to target host:port.
  - Does not use upstream proxy.

- `reject`
  - Rejects request immediately.

### `upstream` format

- Preferred format: URL with scheme:
  - `http://proxy.example.com:3128`
  - `https://proxy.example.com:8443`
- Validation:
  - URL scheme must be `http` or `https`
  - host must be non-empty
  - port must be numeric and in `1..65535`

### `tls` options (optional)

`tls` may be defined only when `upstream` uses `https://`.

| Key | Type | Default | Description |
|---|---|---|---|
| `insecure_skip_verify` | bool | `false` | Disables certificate chain and hostname verification (not recommended). |
| `ca_cert_file` | string | unset | Path to PEM certificate bundle used as root CA store for this upstream. |
| `min_version` | string | `1.2` | Minimum TLS version for HTTPS upstream handshake (`1.2` or `1.3`). |

Behavior:
- Default (no `tls` block): verify certificate and hostname using system CA trust store.
- With `ca_cert_file`: use provided PEM bundle as root CAs for this upstream.
- `min_version` controls minimum TLS version for HTTPS upstream (`1.2` or `1.3`).
- `tls` options with `http://` upstream are invalid.

### Top-level `tls` defaults

You can define a top-level `tls` block to set defaults for HTTPS upstream proxy
connections:

```yaml
tls:
  insecure_skip_verify: false
  min_version: "1.2"
```

Resolution order:
1. Per-rule/default `tls` block, if present.
2. Top-level `tls` block, if present.
3. Built-in defaults (`insecure_skip_verify: false`, system trust store, `min_version: "1.2"`).

Important:
- Per-rule/default `tls` fully overrides top-level `tls` (it is not field-wise merged).
- Top-level `tls` is ignored for non-HTTPS upstream URLs.

### Rule matching order

1. First `routing.rules` match in list order (port + address match)
2. `routing.default` fallback
3. If no `routing.default` configured, implicit `reject`

### Rule validation

- each expanded port value from `dst_port(s)` must be in `1..65535`.
- each `dst_addresses` entry must be one of:
  - single IP (for example `10.0.0.10`)
  - CIDR (for example `10.0.0.0/24`)
  - IP range (for example `10.0.0.10-10.0.0.50`)
- `method` must be one of `http`, `connect`, `direct`, `reject`.

## `timeouts`

Duration values use Go duration syntax (`10s`, `300ms`, `2m`, etc.).

| Key | Type | Required | Description |
|---|---|---|---|
| `connect` | duration string | yes | Dial timeout for upstream/direct target connection. |
| `idle` | duration string | yes | Connection idle timeout for client/upstream/target sockets. |

Validation:
- all timeout values must be greater than `0`.

## `http`

| Key | Type | Required | Description |
|---|---|---|---|
| `max_header_bytes` | int | yes | Max header read buffer size for client HTTP request parsing. |

Validation:
- must be greater than `0`.

## `logging`

| Key | Type | Required | Allowed values | Description |
|---|---|---|---|---|
| `level` | string | yes | `debug`, `info`, `warn`, `error` | Log verbosity threshold. |

Validation:
- unknown values cause startup/check failure.

## Effective behavior and failure mode

- Validation runs when loading config for normal runtime.
- Validation also runs for `--check` mode.
- If CLI overrides are used (for example `--log-level`), config is re-validated.
- Any invalid value causes immediate exit with an error.

## Minimal examples

### Default reject (implicit)

```yaml
routing:
  rules:
    - dst_ports: [443]
      dst_addresses: ["0.0.0.0/0", "::/0"]
      method: "connect"
      upstream: "http://proxy.example.com:3128"
```

Unmatched ports are rejected because `routing.default` is omitted.

### Default direct

```yaml
routing:
  default:
    method: "direct"
  rules:
    - dst_ports: [443]
      dst_addresses: ["0.0.0.0/0", "::/0"]
      method: "connect"
      upstream: "http://proxy.example.com:3128"
```

### Default connect via upstream

```yaml
routing:
  default:
    method: "connect"
    upstream: "https://proxy.example.com:8443"
  rules:
    - dst_ports: [80]
      dst_addresses: ["0.0.0.0/0", "::/0"]
      method: "http"
      upstream: "https://proxy.example.com:8443"
```

## Common mistakes (and why they fail)

- `method: "direct"` with `upstream` set -> invalid (upstream forbidden).
- `method: "reject"` with `upstream` set -> invalid (upstream forbidden).
- `method: "http"` or `"connect"` without `upstream` -> invalid (upstream required).
- missing `dst_addresses`/`dst_address` in a rule -> invalid (destination selector required).
- `logging.level: "verbose"` -> invalid (not in allowed set).
- `listen: ":0"` -> invalid (port out of allowed range).
