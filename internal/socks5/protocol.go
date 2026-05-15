// File protocol.go implements SOCKS5 negotiation, request parsing, and replies.
package socks5

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
)

const (
	version5           = 0x05
	cmdConnect         = 0x01
	cmdBind            = 0x02
	cmdUDPAssociate    = 0x03
	atypIPv4           = 0x01
	atypDomain         = 0x03
	atypIPv6           = 0x04
	authNoAuth         = 0x00
	authNoAcceptable   = 0xFF
	repSuccess         = 0x00
	repGeneralFailure  = 0x01
	repNotAllowed      = 0x02
	repCmdUnsupported  = 0x07
	repAddrUnsupported = 0x08
)

// Request is a parsed SOCKS5 client request.
type Request struct {
	Command byte
	ATYP    byte
	Host    string
	Port    int
}

// NegotiateNoAuth performs SOCKS5 greeting and selects no-auth.
func NegotiateNoAuth(conn net.Conn) error {
	head := make([]byte, 2)
	if _, err := io.ReadFull(conn, head); err != nil {
		return fmt.Errorf("read greeting header: %w", err)
	}
	if head[0] != version5 {
		return fmt.Errorf("unsupported socks version: %d", head[0])
	}
	nMethods := int(head[1])
	methods := make([]byte, nMethods)
	if _, err := io.ReadFull(conn, methods); err != nil {
		return fmt.Errorf("read auth methods: %w", err)
	}

	supported := false
	for _, m := range methods {
		if m == authNoAuth {
			supported = true
			break
		}
	}
	if !supported {
		_, _ = conn.Write([]byte{version5, authNoAcceptable})
		return fmt.Errorf("no supported auth methods")
	}

	_, err := conn.Write([]byte{version5, authNoAuth})
	return err
}

// ReadRequest reads and decodes a SOCKS5 request.
func ReadRequest(conn net.Conn) (*Request, error) {
	head := make([]byte, 4)
	if _, err := io.ReadFull(conn, head); err != nil {
		return nil, fmt.Errorf("read request header: %w", err)
	}
	if head[0] != version5 {
		return nil, fmt.Errorf("unsupported socks version: %d", head[0])
	}

	req := &Request{Command: head[1], ATYP: head[3]}

	switch req.ATYP {
	case atypIPv4:
		addr := make([]byte, 4)
		if _, err := io.ReadFull(conn, addr); err != nil {
			return nil, fmt.Errorf("read ipv4 addr: %w", err)
		}
		req.Host = net.IP(addr).String()
	case atypIPv6:
		addr := make([]byte, 16)
		if _, err := io.ReadFull(conn, addr); err != nil {
			return nil, fmt.Errorf("read ipv6 addr: %w", err)
		}
		req.Host = net.IP(addr).String()
	case atypDomain:
		lenBuf := make([]byte, 1)
		if _, err := io.ReadFull(conn, lenBuf); err != nil {
			return nil, fmt.Errorf("read domain length: %w", err)
		}
		domain := make([]byte, int(lenBuf[0]))
		if _, err := io.ReadFull(conn, domain); err != nil {
			return nil, fmt.Errorf("read domain: %w", err)
		}
		req.Host = string(domain)
	default:
		return nil, fmt.Errorf("unsupported address type: %d", req.ATYP)
	}

	portBuf := make([]byte, 2)
	if _, err := io.ReadFull(conn, portBuf); err != nil {
		return nil, fmt.Errorf("read dst port: %w", err)
	}
	req.Port = int(binary.BigEndian.Uint16(portBuf))
	return req, nil
}

// WriteReply writes a SOCKS5 reply byte with an empty bind address.
func WriteReply(conn net.Conn, rep byte) error {
	// BND.ADDR/BND.PORT set to 0.0.0.0:0 for simplicity.
	resp := []byte{version5, rep, 0x00, atypIPv4, 0, 0, 0, 0, 0, 0}
	_, err := conn.Write(resp)
	return err
}

// CommandSupported reports whether a SOCKS command is handled.
func CommandSupported(cmd byte) bool {
	return cmd == cmdConnect
}

// ReplyForCommand maps unsupported commands to protocol reply codes.
func ReplyForCommand(cmd byte) byte {
	switch cmd {
	case cmdBind, cmdUDPAssociate:
		return repCmdUnsupported
	default:
		return repGeneralFailure
	}
}

// AddrTypeUnsupportedReply returns the address type not supported reply code.
func AddrTypeUnsupportedReply() byte {
	return repAddrUnsupported
}

// AllowedDeniedReply returns the rule-denied reply code.
func AllowedDeniedReply() byte {
	return repNotAllowed
}

// SuccessReply returns the success reply code.
func SuccessReply() byte {
	return repSuccess
}
