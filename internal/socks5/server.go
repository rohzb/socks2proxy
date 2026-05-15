// File server.go contains SOCKS5 server lifecycle and client handling.
package socks5

import (
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"socks2proxy/internal/acl"
	"socks2proxy/internal/logging"
	"socks2proxy/internal/proxy"
)

// Server accepts SOCKS5 clients and routes approved traffic.
type Server struct {
	ListenAddr  string
	ClientACL   *acl.AddressAllowlist
	PortACL     *acl.PortAllowlist
	Router      *proxy.Router
	Logger      *logging.Logger
	IdleTimeout time.Duration
}

var listenFunc = net.Listen

// Serve accepts incoming SOCKS5 clients and proxies supported requests.
func (s *Server) Serve() error {
	ln, err := listenFunc("tcp", s.ListenAddr)
	if err != nil {
		return fmt.Errorf("listen %s: %w", s.ListenAddr, err)
	}
	s.Logger.Infof("listening on %s", s.ListenAddr)
	for {
		conn, err := ln.Accept()
		if err != nil {
			var ne net.Error
			if errors.As(err, &ne) && ne.Temporary() {
				s.Logger.Warnf("accept temporary error: %v", err)
				continue
			}
			return fmt.Errorf("accept: %w", err)
		}
		s.Logger.Debugf("accepted client remote=%s", conn.RemoteAddr())
		go s.handleConn(conn)
	}
}

// handleConn processes a single accepted client connection.
func (s *Server) handleConn(conn net.Conn) {
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(s.IdleTimeout))

	remoteAddr := conn.RemoteAddr().String()
	clientIP, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		s.Logger.Warnf("invalid remote addr %q: %v", remoteAddr, err)
		return
	}

	parsedIP := net.ParseIP(clientIP)
	if parsedIP == nil || !s.ClientACL.Contains(parsedIP) {
		s.Logger.Warnf("deny client %s by client ACL", remoteAddr)
		return
	}

	if err := NegotiateNoAuth(conn); err != nil {
		s.Logger.Warnf("socks negotiation failed from %s: %v", remoteAddr, err)
		return
	}

	req, err := ReadRequest(conn)
	if err != nil {
		_ = WriteReply(conn, AddrTypeUnsupportedReply())
		s.Logger.Warnf("read socks request failed from %s: %v", remoteAddr, err)
		return
	}

	if !CommandSupported(req.Command) {
		_ = WriteReply(conn, ReplyForCommand(req.Command))
		s.Logger.Warnf("unsupported socks cmd=%d from %s", req.Command, remoteAddr)
		return
	}

	if !s.PortACL.Allows(req.Port) {
		_ = WriteReply(conn, AllowedDeniedReply())
		s.Logger.Warnf("deny target %s:%d from %s by port ACL", req.Host, req.Port, remoteAddr)
		return
	}

	if err := WriteReply(conn, SuccessReply()); err != nil {
		s.Logger.Warnf("write success reply failed: %v", err)
		return
	}

	host := normalizeTargetHost(req.Host, req.Port)
	s.Logger.Infof("proxy %s -> %s:%d", remoteAddr, host, req.Port)

	if err := s.Router.Route(conn, host, req.Port); err != nil {
		s.Logger.Errorf("proxy error for %s -> %s:%d: %v", remoteAddr, host, req.Port, err)
		return
	}
	s.Logger.Debugf("completed proxy flow %s -> %s:%d", remoteAddr, host, req.Port)
}

// normalizeTargetHost canonicalizes host formats before upstream routing.
func normalizeTargetHost(host string, port int) string {
	if ip := net.ParseIP(host); ip != nil {
		return host
	}
	if strings.HasPrefix(host, "[") && strings.Contains(host, "]") {
		return strings.TrimPrefix(strings.Split(host, "]")[0], "[")
	}
	if strings.Contains(host, ":") {
		if h, p, err := net.SplitHostPort(host); err == nil {
			if parsed, convErr := strconv.Atoi(p); convErr == nil && parsed == port {
				return h
			}
		}
	}
	return host
}
