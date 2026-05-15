// File connect.go handles CONNECT tunneling via an upstream proxy.
package proxy

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
)

// HandleConnect establishes an upstream CONNECT tunnel and relays bytes bidirectionally.
func HandleConnect(client net.Conn, upstream net.Conn, targetHost string, targetPort int) error {
	hostPort := net.JoinHostPort(targetHost, fmt.Sprintf("%d", targetPort))
	req := &http.Request{
		Method: "CONNECT",
		Host:   hostPort,
		URL:    &url.URL{Opaque: hostPort},
		Header: make(http.Header),
	}
	req.Header.Set("Host", hostPort)
	if err := req.Write(upstream); err != nil {
		return fmt.Errorf("write CONNECT to upstream: %w", err)
	}

	br := newBufferedReader(upstream)
	resp, err := http.ReadResponse(br, req)
	if err != nil {
		return fmt.Errorf("read CONNECT response: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		_, _ = io.Copy(io.Discard, resp.Body)
		return fmt.Errorf("upstream CONNECT failed: %s", resp.Status)
	}

	BidirectionalCopy(client, upstream)
	return nil
}
