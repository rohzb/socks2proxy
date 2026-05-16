// File direct.go handles direct target TCP connections without upstream proxy.
package proxy

import (
	"context"
	"fmt"
	"net"
	"time"
)

// HandleDirect connects directly to the target and relays bytes bidirectionally.
func HandleDirect(client net.Conn, targetHost string, targetPort int, sourceIP string, connectTimeout time.Duration, idleTimeout time.Duration) error {
	targetAddr := net.JoinHostPort(targetHost, fmt.Sprintf("%d", targetPort))
	dialer := net.Dialer{Timeout: connectTimeout}
	if sourceIP != "" {
		localIP := net.ParseIP(sourceIP)
		if localIP == nil {
			return fmt.Errorf("invalid source_ip %q", sourceIP)
		}
		dialer.LocalAddr = &net.TCPAddr{IP: localIP}
	}
	ctx, cancel := context.WithTimeout(context.Background(), connectTimeout)
	defer cancel()

	targetConn, err := dialer.DialContext(ctx, "tcp", targetAddr)
	if err != nil {
		return fmt.Errorf("dial target directly: %w", err)
	}
	defer targetConn.Close()

	_ = targetConn.SetDeadline(time.Now().Add(idleTimeout))
	_ = client.SetDeadline(time.Now().Add(idleTimeout))

	BidirectionalCopy(client, targetConn)
	return nil
}
