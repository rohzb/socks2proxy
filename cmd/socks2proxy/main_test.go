package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestMainHelpVersionAndCheck(t *testing.T) {
	t.Run("help", func(t *testing.T) {
		out, err := runMainSubprocess(t, []string{"--help"}, "")
		if err != nil {
			t.Fatalf("help subprocess failed: %v; out=%s", err, out)
		}
		if !strings.Contains(out, "Usage:") || !strings.Contains(out, "socks2proxy") {
			t.Fatalf("unexpected help output: %s", out)
		}
	})

	t.Run("version", func(t *testing.T) {
		out, err := runMainSubprocess(t, []string{"--version"}, "")
		if err != nil {
			t.Fatalf("version subprocess failed: %v; out=%s", err, out)
		}
		if !strings.Contains(out, "socks2proxy") || !strings.Contains(out, "License:") {
			t.Fatalf("unexpected version output: %s", out)
		}
	})

	t.Run("check", func(t *testing.T) {
		d := t.TempDir()
		cfgPath := filepath.Join(d, "config.yaml")
		cfg := `
listen: ":41080"
allowed_client_addresses:
  - "127.0.0.1/32"
routing:
  rules:
    - dst_ports: [1]
      dst_addresses: ["0.0.0.0/0"]
      method: "reject"
  default:
    method: "reject"
timeouts:
  connect: "10s"
  idle: "30s"
http:
  max_header_bytes: 65536
logging:
  level: "info"
`
		if err := os.WriteFile(cfgPath, []byte(cfg), 0o644); err != nil {
			t.Fatalf("write config: %v", err)
		}
		out, err := runMainSubprocess(t, []string{"--check", "--config", cfgPath}, "")
		if err != nil {
			t.Fatalf("check subprocess failed: %v; out=%s", err, out)
		}
		if !strings.Contains(out, "config is valid") {
			t.Fatalf("unexpected check output: %s", out)
		}
	})

	t.Run("missing-config-fails", func(t *testing.T) {
		out, err := runMainSubprocess(t, []string{}, "")
		if err == nil {
			t.Fatalf("expected missing-config subprocess to fail")
		}
		if !strings.Contains(out, "--config is required") {
			t.Fatalf("unexpected missing-config output: %s", out)
		}
	})

	t.Run("bad-log-level-override-fails", func(t *testing.T) {
		d := t.TempDir()
		cfgPath := filepath.Join(d, "config.yaml")
		cfg := `
listen: ":41080"
allowed_client_addresses:
  - "127.0.0.1/32"
routing:
  rules:
    - dst_ports: [1]
      dst_addresses: ["0.0.0.0/0"]
      method: "reject"
  default:
    method: "reject"
timeouts:
  connect: "10s"
  idle: "30s"
http:
  max_header_bytes: 65536
logging:
  level: "info"
`
		if err := os.WriteFile(cfgPath, []byte(cfg), 0o644); err != nil {
			t.Fatalf("write config: %v", err)
		}
		out, err := runMainSubprocess(t, []string{"--check", "--config", cfgPath, "--log-level", "verbose"}, "")
		if err == nil {
			t.Fatalf("expected bad log-level override to fail")
		}
		if !strings.Contains(out, "unsupported logging.level") {
			t.Fatalf("unexpected bad log-level output: %s", out)
		}
	})
}

func runMainSubprocess(t *testing.T, args []string, extraEnv string) (string, error) {
	t.Helper()
	cmd := exec.Command(os.Args[0], append([]string{"-test.run=TestMainHelperProcess", "--"}, args...)...)
	cmd.Env = append(os.Environ(), "GO_WANT_MAIN_HELPER_PROCESS=1")
	if extraEnv != "" {
		cmd.Env = append(cmd.Env, extraEnv)
	}
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	err := cmd.Run()
	return out.String(), err
}

func TestMainHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_MAIN_HELPER_PROCESS") != "1" {
		return
	}
	idx := 0
	for i, a := range os.Args {
		if a == "--" {
			idx = i + 1
			break
		}
	}
	os.Args = append([]string{os.Args[0]}, os.Args[idx:]...)
	main()
	os.Exit(0)
}
