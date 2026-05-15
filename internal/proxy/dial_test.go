package proxy

import (
	"context"
	"crypto/tls"
	"net"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"socks2proxy/internal/config"
)

func TestDialUpstreamHTTP(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer ln.Close()
	go func() {
		c, e := ln.Accept()
		if e == nil {
			_ = c.Close()
		}
	}()

	d := net.Dialer{Timeout: time.Second}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	conn, err := dialUpstream(ctx, d, config.UpstreamEndpoint{Scheme: "http", Address: ln.Addr().String()}, nil)
	if err != nil {
		t.Fatalf("dialUpstream http failed: %v", err)
	}
	_ = conn.Close()
}

func TestDialUpstreamHTTPS(t *testing.T) {
	ts := httptest.NewTLSServer(nil)
	defer ts.Close()

	hostPort := ts.Listener.Addr().String()
	host, _, _ := net.SplitHostPort(hostPort)

	d := net.Dialer{Timeout: time.Second}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	conn, err := dialUpstream(ctx, d, config.UpstreamEndpoint{Scheme: "https", Address: hostPort, ServerName: host}, &config.UpstreamTLS{InsecureSkipVerify: true, MinVersion: "1.2"})
	if err != nil {
		t.Fatalf("dialUpstream https failed: %v", err)
	}
	_ = conn.Close()
}

func TestDialUpstreamHTTPSWithCACertAndUnsupported(t *testing.T) {
	ts := httptest.NewTLSServer(nil)
	defer ts.Close()

	hostPort := ts.Listener.Addr().String()
	host, _, _ := net.SplitHostPort(hostPort)

	dir := t.TempDir()
	caPath := filepath.Join(dir, "ca.pem")
	certDER := ts.TLS.Certificates[0].Certificate[0]
	pem := tls.Certificate{Certificate: [][]byte{certDER}}
	_ = pem
	// write leaf certificate as PEM for test trust roots
	pemBytes := []byte("-----BEGIN CERTIFICATE-----\n")
	pemBytes = append(pemBytes, []byte(encodeBase64Lines(certDER))...)
	pemBytes = append(pemBytes, []byte("\n-----END CERTIFICATE-----\n")...)
	if err := os.WriteFile(caPath, pemBytes, 0o644); err != nil {
		t.Fatalf("write ca file: %v", err)
	}

	dialer := net.Dialer{Timeout: time.Second}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	conn, err := dialUpstream(ctx, dialer, config.UpstreamEndpoint{Scheme: "https", Address: hostPort, ServerName: host}, &config.UpstreamTLS{CACertFile: caPath, MinVersion: "1.2"})
	if err != nil {
		t.Fatalf("dialUpstream https with ca failed: %v", err)
	}
	_ = conn.Close()

	if _, err := dialUpstream(ctx, dialer, config.UpstreamEndpoint{Scheme: "socks5", Address: "127.0.0.1:1"}, nil); err == nil {
		t.Fatalf("expected unsupported scheme to fail")
	}
}

func encodeBase64Lines(in []byte) string {
	const table = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"
	out := make([]byte, 0, ((len(in)+2)/3)*4)
	for i := 0; i < len(in); i += 3 {
		var b0, b1, b2 byte
		b0 = in[i]
		have1 := i+1 < len(in)
		have2 := i+2 < len(in)
		if have1 {
			b1 = in[i+1]
		}
		if have2 {
			b2 = in[i+2]
		}
		out = append(out,
			table[b0>>2],
			table[((b0&0x03)<<4)|(b1>>4)],
		)
		if have1 {
			out = append(out, table[((b1&0x0f)<<2)|(b2>>6)])
		} else {
			out = append(out, '=')
		}
		if have2 {
			out = append(out, table[b2&0x3f])
		} else {
			out = append(out, '=')
		}
	}
	return string(out)
}
