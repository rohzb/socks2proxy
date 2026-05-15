// Package socks5 implements protocol negotiation, request parsing, and server
// connection handling for SOCKS5 clients.
//
// It enforces ACL checks before delegating accepted traffic to the upstream
// routing layer.
//
// Maintainer: Ruslan Ovsyannikov <ovsyannikov@helmholtz-berlin.de>
package socks5
