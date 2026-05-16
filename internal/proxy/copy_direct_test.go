package proxy

import (
	"io"
	"net"
	"testing"
	"time"
)

func TestBidirectionalCopy(t *testing.T) {
	a, b := net.Pipe()
	defer a.Close()
	defer b.Close()
	_ = a.SetDeadline(time.Now().Add(2 * time.Second))
	_ = b.SetDeadline(time.Now().Add(2 * time.Second))

	done := make(chan struct{})
	go func() {
		BidirectionalCopy(a, b)
		close(done)
	}()

	msg := []byte("ping")
	go func() {
		_, _ = b.Write(msg)
		_ = b.Close()
	}()
	buf := make([]byte, 4)
	_, err := io.ReadFull(a, buf)
	if err != nil {
		t.Fatalf("readfull failed: %v", err)
	}
	if string(buf) != "ping" {
		t.Fatalf("unexpected copied data: %q", string(buf))
	}
	<-done
}

func TestHandleDirectDialFailure(t *testing.T) {
	client, peer := net.Pipe()
	defer client.Close()
	defer peer.Close()
	err := HandleDirect(client, "127.0.0.1", 1, "", 20*time.Millisecond, 20*time.Millisecond)
	if err == nil {
		t.Fatalf("expected direct dial failure")
	}
}

func TestHandleDirectSuccess(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer ln.Close()
	go func() {
		conn, e := ln.Accept()
		if e != nil {
			return
		}
		defer conn.Close()
		_, _ = io.Copy(conn, conn)
	}()

	clientServer, clientPeer := net.Pipe()
	defer clientPeer.Close()
	defer clientServer.Close()
	_ = clientPeer.SetDeadline(time.Now().Add(2 * time.Second))

	errCh := make(chan error, 1)
	go func() {
		host, portStr, _ := net.SplitHostPort(ln.Addr().String())
		p, _ := ParsePort(portStr)
		errCh <- HandleDirect(clientServer, host, p, "", time.Second, time.Second)
	}()

	_, _ = clientPeer.Write([]byte("ok"))
	buf := make([]byte, 2)
	_, err = io.ReadFull(clientPeer, buf)
	if err != nil {
		t.Fatalf("read echoed bytes failed: %v", err)
	}
	if string(buf) != "ok" {
		t.Fatalf("unexpected echoed bytes: %q", string(buf))
	}
	_ = clientPeer.Close()
	if err := <-errCh; err != nil {
		t.Fatalf("handle direct should succeed, got: %v", err)
	}
}

func TestHandleDirectInvalidSourceIP(t *testing.T) {
	client, peer := net.Pipe()
	defer client.Close()
	defer peer.Close()
	err := HandleDirect(client, "127.0.0.1", 80, "not-an-ip", 20*time.Millisecond, 20*time.Millisecond)
	if err == nil {
		t.Fatalf("expected invalid source_ip error")
	}
}
