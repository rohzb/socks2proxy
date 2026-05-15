// File copy.go implements bidirectional stream copy helpers.
package proxy

import (
	"io"
	"net"
	"sync"
)

type closeWriter interface {
	CloseWrite() error
}

// BidirectionalCopy streams data in both directions until both sides finish.
func BidirectionalCopy(a net.Conn, b net.Conn) {
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		_, _ = io.Copy(a, b)
		if cw, ok := a.(closeWriter); ok {
			_ = cw.CloseWrite()
		}
	}()

	go func() {
		defer wg.Done()
		_, _ = io.Copy(b, a)
		if cw, ok := b.(closeWriter); ok {
			_ = cw.CloseWrite()
		}
	}()

	wg.Wait()
}
