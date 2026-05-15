package proxy

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestHandleHTTPRoundTrip(t *testing.T) {
	clientServer, clientPeer := net.Pipe()
	upstreamServer, upstreamPeer := net.Pipe()
	defer clientPeer.Close()
	defer upstreamPeer.Close()
	_ = clientPeer.SetDeadline(time.Now().Add(3 * time.Second))
	_ = upstreamPeer.SetDeadline(time.Now().Add(3 * time.Second))

	hErr := make(chan error, 1)
	go func() {
		hErr <- HandleHTTP(clientServer, upstreamServer, "example.com", 8192)
	}()

	upErr := make(chan error, 1)
	go func() {
		br := bufio.NewReader(upstreamPeer)
		req, err := http.ReadRequest(br)
		if err != nil {
			upErr <- err
			return
		}
		if req.Method != "GET" {
			upErr <- fmt.Errorf("unexpected method %s", req.Method)
			return
		}
		resp := "HTTP/1.1 200 OK\r\nContent-Length: 2\r\nConnection: close\r\n\r\nOK"
		_, err = io.WriteString(upstreamPeer, resp)
		upErr <- err
	}()

	_, _ = io.WriteString(clientPeer, "GET / HTTP/1.1\r\nHost: example.com\r\nConnection: close\r\n\r\n")
	resp, err := http.ReadResponse(bufio.NewReader(clientPeer), nil)
	if err != nil {
		t.Fatalf("failed reading client response: %v", err)
	}
	bodyBytes, _ := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if resp.StatusCode != 200 || string(bodyBytes) != "OK" {
		t.Fatalf("unexpected client response status=%d body=%q", resp.StatusCode, string(bodyBytes))
	}

	if err := <-upErr; err != nil {
		t.Fatalf("upstream side error: %v", err)
	}
	if err := <-hErr; err != nil {
		t.Fatalf("handleHTTP error: %v", err)
	}
}

func TestHandleConnectFailureStatus(t *testing.T) {
	clientServer, clientPeer := net.Pipe()
	upstreamServer, upstreamPeer := net.Pipe()
	defer clientServer.Close()
	defer clientPeer.Close()
	defer upstreamServer.Close()
	defer upstreamPeer.Close()

	errCh := make(chan error, 1)
	go func() {
		errCh <- HandleConnect(clientServer, upstreamServer, "example.com", 443)
	}()

	br := bufio.NewReader(upstreamPeer)
	req, err := http.ReadRequest(br)
	if err != nil {
		t.Fatalf("failed reading connect request: %v", err)
	}
	if req.Method != http.MethodConnect {
		t.Fatalf("expected CONNECT method, got %s", req.Method)
	}
	_, _ = io.WriteString(upstreamPeer, "HTTP/1.1 403 Forbidden\r\nContent-Length: 0\r\n\r\n")

	err = <-errCh
	if err == nil || !strings.Contains(err.Error(), "CONNECT failed") {
		t.Fatalf("expected connect failed error, got: %v", err)
	}
}
