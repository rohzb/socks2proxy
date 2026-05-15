package socks5

import (
	"net"
	"testing"
)

func TestReadRequestIPv6AndDomainAndErrors(t *testing.T) {
	t.Run("ipv6", func(t *testing.T) {
		c1, c2 := net.Pipe()
		defer c1.Close()
		defer c2.Close()
		errCh := make(chan error, 1)
		var req *Request
		go func() { req, _ = ReadRequest(c1); errCh <- nil }()
		pkt := append([]byte{0x05, 0x01, 0x00, 0x04}, net.ParseIP("2001:db8::1").To16()...)
		pkt = append(pkt, 0x00, 0x50)
		_, _ = c2.Write(pkt)
		<-errCh
		if req == nil || req.Port != 80 {
			t.Fatalf("unexpected req: %+v", req)
		}
	})

	t.Run("domain", func(t *testing.T) {
		c1, c2 := net.Pipe()
		defer c1.Close()
		defer c2.Close()
		ch := make(chan *Request, 1)
		go func() {
			r, _ := ReadRequest(c1)
			ch <- r
		}()
		d := []byte("example.com")
		pkt := []byte{0x05, 0x01, 0x00, 0x03, byte(len(d))}
		pkt = append(pkt, d...)
		pkt = append(pkt, 0x01, 0xbb)
		_, _ = c2.Write(pkt)
		r := <-ch
		if r == nil || r.Host != "example.com" || r.Port != 443 {
			t.Fatalf("unexpected domain request: %+v", r)
		}
	})

	t.Run("unsupported atyp", func(t *testing.T) {
		c1, c2 := net.Pipe()
		defer c1.Close()
		defer c2.Close()
		errCh := make(chan error, 1)
		go func() { _, err := ReadRequest(c1); errCh <- err }()
		_, _ = c2.Write([]byte{0x05, 0x01, 0x00, 0x09})
		if err := <-errCh; err == nil {
			t.Fatalf("expected unsupported atyp error")
		}
	})

	t.Run("bad version", func(t *testing.T) {
		c1, c2 := net.Pipe()
		defer c1.Close()
		defer c2.Close()
		errCh := make(chan error, 1)
		go func() { _, err := ReadRequest(c1); errCh <- err }()
		_, _ = c2.Write([]byte{0x04, 0x01, 0x00, 0x01})
		if err := <-errCh; err == nil {
			t.Fatalf("expected unsupported version error")
		}
	})
}
