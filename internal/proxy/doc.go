// Package proxy routes accepted SOCKS5 targets according to explicit rules.
//
// Supported methods are:
// - http: forward-proxy HTTP request/response exchange via upstream proxy
// - connect: CONNECT and full-duplex byte tunneling via upstream proxy
// - direct: direct target TCP connection and full-duplex byte tunneling
// - reject: explicit request denial
//
// Maintainer: Ruslan Ovsyannikov <ovsyannikov@helmholtz-berlin.de>
package proxy
