// Package main provides the socks2proxy service executable.
//
// The binary bridges SOCKS5 client traffic using explicit per-port routing
// rules from YAML config, supporting upstream HTTP proxy forwarding and direct
// connections.
//
// This entrypoint wires configuration, ACLs, routing, and logging together
// into the runtime server process.
//
// Maintainer: Ruslan Ovsyannikov <ovsyannikov@helmholtz-berlin.de>
package main
