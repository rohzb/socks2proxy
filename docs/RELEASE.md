# Release Guide

This document describes how `socks2proxy` is versioned, built, packaged, and
published.

It applies to both local releases (`make`-driven) and CI/CD releases. The
expectation is that both paths produce equivalent artifacts.

## Versioning

`socks2proxy` uses Semantic Versioning with git tags in the form
`vMAJOR.MINOR.PATCH`.

Version meaning:

- `MAJOR` is for backward-incompatible changes (config semantics, CLI behavior,
  deployment defaults, or other breaking runtime behavior).
- `MINOR` is for backward-compatible feature work and improvements.
- `PATCH` is for backward-compatible bugfixes, security fixes, and maintenance
  updates.

Examples:

- `v1.0.0`: first stable release
- `v1.1.0`: feature release without breaking changes
- `v1.1.1`: patch release for `v1.1.x`

## Release Types

Two release types are used:

- snapshot builds for branch/testing workflows (non-tag builds)
- stable releases from semver tags

## Tagging and Release Channels

Stable releases are cut from immutable semver tags:

- `vMAJOR.MINOR.PATCH` (for example `v1.4.2`)

In addition to immutable tags, moving channel tags can be maintained:

- `latest` -> newest stable release
- `vMAJOR` -> newest release in a major line (for example latest `v1.x.y`)
- `vMAJOR.MINOR` -> newest patch in a minor line (for example latest `v1.4.z`)

Consumers should use exact tags when they need reproducibility. Channel tags are
for controlled update streams.

In this repository, channel tags are updated automatically by the release
workflow on every stable release tag (`v*.*.*`).

When channel tags are enabled, assets are refreshed for:

- the exact release tag (always)
- moving channel tags (`latest`, `vMAJOR`, `vMAJOR.MINOR`)

Operational behavior:

- the semver tag release (`vMAJOR.MINOR.PATCH`) is created/updated first
- moving tags are force-updated to the same commit
- release assets are uploaded to the moving channel releases as well

This gives users two consumption modes:

- pin exact immutable versions (`v1.4.2`)
- track a channel (`latest`, `v1`, `v1.4`)

## Platform Matrix

Default build matrix:

- `linux/amd64`
- `linux/arm64`
- `darwin/amd64`
- `darwin/arm64`
- `windows/amd64`
- `windows/arm64`

## Archive Formats

Use standard archive formats by platform:

- Linux/macOS: `.tar.gz`
- Windows: `.zip`

## Artifact Naming

Archive names:

- `socks2proxy_<version>_<os>_<arch>.tar.gz` (Linux/macOS)
- `socks2proxy_<version>_windows_<arch>.zip` (Windows)

Binary names inside archives:

- `socks2proxy` (Linux/macOS)
- `socks2proxy.exe` (Windows)

## Package Contents

Each release archive includes:

- binary (`socks2proxy` or `socks2proxy.exe`)
- `README.md`
- `LICENSE`
- `docs/CONFIG.md`
- `docs/INSTALL.md`
- `examples/config.example.yaml`
- platform-specific assets from `platform/<os>/`

Platform asset mapping:

- Linux archives include `platform/linux/**`
- Darwin archives include `platform/darwin/**`
- Windows archives include `platform/windows/**`

## Build Metadata

`./socks2proxy --version` should report at least:

- project name
- version
- commit
- build date (UTC)
- Go version
- platform

## Verification Artifacts

Stable releases publish:

- `SHA256SUMS`
- optional signature file (if signing is enabled)

## Quality Gates

A release candidate must pass all checks below.

1. Formatting

- `make fmt`
- `make go-fmt`

2. Static analysis

- `make vet`

3. Tests

- `make test`
- `go test ./... -cover`

4. Build

- `make build`
- `make build-all`

5. Runtime sanity

- `./socks2proxy --version`
- `./socks2proxy --help`
- `./socks2proxy --check --config ./examples/config.example.yaml`

## Local Release Flow

Recommended local release sequence:

1. run `make check`
2. run `make build-all`
3. package artifacts using the OS archive rules in this document
4. generate checksums
5. verify checksums and smoke test at least one artifact per OS family

## CI/CD Release Flow

PR pipeline:

- run `check`
- validate cross-platform builds
- do not publish artifacts

Tag pipeline (`v*.*.*`):

- run the same gates as PR
- build matrix artifacts
- package (`tar.gz` for Linux/macOS, `zip` for Windows)
- generate and publish checksums
- publish release notes and assets
- force-update moving channel tags (`latest`, `vMAJOR`, `vMAJOR.MINOR`)
- upload matching assets to channel releases

## Compatibility Policy

Compatibility-sensitive surfaces:

- config schema and semantics
- CLI flags and behavior
- default deployment paths

Current Linux deployment defaults:

- binary: `/usr/local/bin/socks2proxy`
- config: `/usr/local/etc/socks2proxy/config.yaml`

If a release introduces incompatibility, call it out explicitly in release
notes.

## Release Notes Checklist

Each stable release note should include:

- version and date
- highlights
- fixes
- breaking changes (if any)
- migration notes (if any)
- checksum reference
