package socks5

import (
	"net"
	"testing"
)

func TestNegotiateNoAuthSuccess(t *testing.T) {
	c1, c2 := net.Pipe()
	defer c1.Close()
	defer c2.Close()

	errCh := make(chan error, 1)
	go func() { errCh <- NegotiateNoAuth(c1) }()

	_, _ = c2.Write([]byte{version5, 1, authNoAuth})
	resp := make([]byte, 2)
	_, _ = c2.Read(resp)
	if resp[0] != version5 || resp[1] != authNoAuth {
		t.Fatalf("unexpected negotiate response: %v", resp)
	}
	if err := <-errCh; err != nil {
		t.Fatalf("unexpected negotiate error: %v", err)
	}
}

func TestNegotiateNoAuthUnsupported(t *testing.T) {
	c1, c2 := net.Pipe()
	defer c1.Close()
	defer c2.Close()

	errCh := make(chan error, 1)
	go func() { errCh <- NegotiateNoAuth(c1) }()

	_, _ = c2.Write([]byte{version5, 1, 0x02})
	resp := make([]byte, 2)
	_, _ = c2.Read(resp)
	if resp[1] != authNoAcceptable {
		t.Fatalf("expected no acceptable auth response")
	}
	if err := <-errCh; err == nil {
		t.Fatalf("expected negotiate error for unsupported auth")
	}
}

func TestReadRequestIPv4(t *testing.T) {
	c1, c2 := net.Pipe()
	defer c1.Close()
	defer c2.Close()

	errCh := make(chan error, 1)
	var req *Request
	go func() {
		var err error
		req, err = ReadRequest(c1)
		errCh <- err
	}()

	packet := []byte{version5, cmdConnect, 0x00, atypIPv4, 1, 2, 3, 4, 0x01, 0xbb}
	_, _ = c2.Write(packet)

	if err := <-errCh; err != nil {
		t.Fatalf("unexpected read request error: %v", err)
	}
	if req.Host != "1.2.3.4" || req.Port != 443 || req.Command != cmdConnect {
		t.Fatalf("unexpected parsed request: %+v", req)
	}
}

func TestWriteReply(t *testing.T) {
	c1, c2 := net.Pipe()
	defer c1.Close()
	defer c2.Close()

	errCh := make(chan error, 1)
	go func() { errCh <- WriteReply(c1, repSuccess) }()

	buf := make([]byte, 10)
	_, _ = c2.Read(buf)
	if buf[0] != version5 || buf[1] != repSuccess {
		t.Fatalf("unexpected reply payload: %v", buf)
	}
	if err := <-errCh; err != nil {
		t.Fatalf("unexpected write reply error: %v", err)
	}
}

func TestReplyForCommand(t *testing.T) {
	if !CommandSupported(cmdConnect) {
		t.Fatalf("connect command must be supported")
	}
	if CommandSupported(cmdBind) {
		t.Fatalf("bind command must not be supported")
	}
	if ReplyForCommand(cmdBind) != repCmdUnsupported {
		t.Fatalf("bind should map to cmd unsupported")
	}
	if ReplyForCommand(0x7f) != repGeneralFailure {
		t.Fatalf("unknown should map to general failure")
	}
	if AddrTypeUnsupportedReply() != repAddrUnsupported {
		t.Fatalf("unexpected addr unsupported reply")
	}
	if AllowedDeniedReply() != repNotAllowed {
		t.Fatalf("unexpected not-allowed reply")
	}
	if SuccessReply() != repSuccess {
		t.Fatalf("unexpected success reply")
	}
}
