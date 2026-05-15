// File http.go handles HTTP forwarding via an upstream proxy.
package proxy

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
)

// HandleHTTP proxies HTTP/1.x traffic via forward-proxy mode.
func HandleHTTP(client net.Conn, upstream net.Conn, targetHost string, maxHeaderBytes int) error {
	reader := bufio.NewReaderSize(client, maxHeaderBytes)

	for {
		req, err := http.ReadRequest(reader)
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return fmt.Errorf("read client HTTP request: %w", err)
		}

		if req.URL == nil {
			return fmt.Errorf("invalid HTTP request: missing URL")
		}

		host := req.Host
		if host == "" {
			host = targetHost
		}
		if !strings.Contains(host, ":") {
			host = net.JoinHostPort(host, "80")
		}

		if req.URL.Scheme == "" {
			req.URL.Scheme = "http"
		}
		if req.URL.Host == "" {
			req.URL.Host = host
		}

		if err := req.WriteProxy(upstream); err != nil {
			return fmt.Errorf("write proxy request upstream: %w", err)
		}

		resp, err := http.ReadResponse(newBufferedReader(upstream), req)
		if err != nil {
			return fmt.Errorf("read upstream response: %w", err)
		}
		if err := resp.Write(client); err != nil {
			resp.Body.Close()
			return fmt.Errorf("write response to client: %w", err)
		}
		resp.Body.Close()

		if req.Close || resp.Close {
			return nil
		}
	}
}
