package socks5

import (
	"io"
	"net"
	"testing"
	"time"

	"socks2proxy/internal/acl"
	"socks2proxy/internal/logging"
	"socks2proxy/internal/proxy"
)

func baseServer(t *testing.T) *Server {
	t.Helper()
	clientACL, err := acl.NewAddressAllowlist([]string{"127.0.0.1/32"})
	if err != nil {
		t.Fatalf("client ACL: %v", err)
	}
	return &Server{
		ListenAddr:  "127.0.0.1:0",
		ClientACL:   clientACL,
		PortACL:     acl.NewAllowAllPortAllowlist(),
		Router:      &proxy.Router{DefaultRule: proxy.Rule{Method: proxy.MethodReject}},
		Logger:      logging.New("error"),
		IdleTimeout: 2 * time.Second,
	}
}

func withTCPPair(t *testing.T, fn func(serverConn net.Conn, clientConn net.Conn)) {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer ln.Close()

	accepted := make(chan net.Conn, 1)
	errCh := make(chan error, 1)
	go func() {
		c, e := ln.Accept()
		if e != nil {
			errCh <- e
			return
		}
		accepted <- c
	}()

	clientConn, err := net.Dial("tcp", ln.Addr().String())
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer clientConn.Close()

	select {
	case e := <-errCh:
		t.Fatalf("accept error: %v", e)
	case serverConn := <-accepted:
		defer serverConn.Close()
		fn(serverConn, clientConn)
	}
}

func TestHandleConnNegotiationFailure(t *testing.T) {
	s := baseServer(t)
	withTCPPair(t, func(serverConn net.Conn, clientConn net.Conn) {
		go s.handleConn(serverConn)
		_, _ = clientConn.Write([]byte{0x04, 0x01, 0x00})
		_ = clientConn.SetReadDeadline(time.Now().Add(200 * time.Millisecond))
		_, _ = io.ReadAll(clientConn)
	})
}

func TestHandleConnHappyPathReturnsSuccessReply(t *testing.T) {
	s := baseServer(t)
	withTCPPair(t, func(serverConn net.Conn, clientConn net.Conn) {
		go s.handleConn(serverConn)

		_, _ = clientConn.Write([]byte{0x05, 0x01, 0x00})
		neg := make([]byte, 2)
		if _, err := io.ReadFull(clientConn, neg); err != nil {
			t.Fatalf("failed reading negotiation response: %v", err)
		}
		if neg[1] != 0x00 {
			t.Fatalf("expected no-auth selected")
		}

		req := []byte{0x05, 0x01, 0x00, 0x01, 1, 2, 3, 4, 0x01, 0xbb}
		_, _ = clientConn.Write(req)
		rep := make([]byte, 10)
		if _, err := io.ReadFull(clientConn, rep); err != nil {
			t.Fatalf("failed reading reply: %v", err)
		}
		if rep[1] != SuccessReply() {
			t.Fatalf("expected success reply, got %d", rep[1])
		}
	})
}

func TestHandleConnUnsupportedCommand(t *testing.T) {
	s := baseServer(t)
	withTCPPair(t, func(serverConn net.Conn, clientConn net.Conn) {
		go s.handleConn(serverConn)
		_, _ = clientConn.Write([]byte{0x05, 0x01, 0x00})
		neg := make([]byte, 2)
		_, _ = io.ReadFull(clientConn, neg)

		req := []byte{0x05, 0x02, 0x00, 0x01, 1, 2, 3, 4, 0x01, 0xbb}
		_, _ = clientConn.Write(req)
		rep := make([]byte, 10)
		_, _ = io.ReadFull(clientConn, rep)
		if rep[1] != ReplyForCommand(0x02) {
			t.Fatalf("expected unsupported-command reply")
		}
	})
}

func TestHandleConnDeniedByPortACL(t *testing.T) {
	s := baseServer(t)
	s.PortACL = acl.NewPortAllowlist([]int{80})
	withTCPPair(t, func(serverConn net.Conn, clientConn net.Conn) {
		go s.handleConn(serverConn)
		_, _ = clientConn.Write([]byte{0x05, 0x01, 0x00})
		neg := make([]byte, 2)
		_, _ = io.ReadFull(clientConn, neg)

		req := []byte{0x05, 0x01, 0x00, 0x01, 1, 2, 3, 4, 0x01, 0xbb}
		_, _ = clientConn.Write(req)
		rep := make([]byte, 10)
		_, _ = io.ReadFull(clientConn, rep)
		if rep[1] != AllowedDeniedReply() {
			t.Fatalf("expected not-allowed reply")
		}
	})
}
