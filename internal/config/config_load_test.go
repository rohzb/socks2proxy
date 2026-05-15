package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestPortSpecsUnmarshalAndMarshalDuration(t *testing.T) {
	type doc struct {
		Ports PortSpecs `yaml:"ports"`
		Dur   Duration  `yaml:"dur"`
	}
	var d doc
	if err := yaml.Unmarshal([]byte("ports: \"80,443,1000-1002\"\ndur: \"5s\"\n"), &d); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if len(d.Ports) != 5 {
		t.Fatalf("unexpected ports len: %d", len(d.Ports))
	}
	v, err := d.Dur.MarshalYAML()
	if err != nil {
		t.Fatalf("marshal duration failed: %v", err)
	}
	if s, ok := v.(string); !ok || s != "5s" {
		t.Fatalf("unexpected marshal duration value: %v", v)
	}
}

func TestPortSpecsUnmarshalRejectsInvalid(t *testing.T) {
	type doc struct {
		Ports PortSpecs `yaml:"ports"`
	}
	cases := []string{
		"ports: \"1-\"\n",
		"ports: \"abc\"\n",
		"ports: \"10-1\"\n",
	}
	for _, c := range cases {
		var d doc
		if err := yaml.Unmarshal([]byte(c), &d); err == nil {
			t.Fatalf("expected invalid ports to fail: %q", c)
		}
	}
}

func TestLoadFile(t *testing.T) {
	d := t.TempDir()
	p := filepath.Join(d, "cfg.yaml")
	good := `
listen: ":41080"
allowed_client_addresses: ["127.0.0.1/32"]
routing:
  rules:
    - dst_ports: [1]
      dst_addresses: ["0.0.0.0/0"]
      method: "reject"
  default:
    method: "reject"
timeouts:
  connect: "10s"
  idle: "10s"
http:
  max_header_bytes: 1024
logging:
  level: "info"
`
	if err := os.WriteFile(p, []byte(good), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	cfg, err := LoadFile(p)
	if err != nil {
		t.Fatalf("load file failed: %v", err)
	}
	if cfg.Listen != ":41080" {
		t.Fatalf("unexpected listen: %s", cfg.Listen)
	}

	if _, err := LoadFile(filepath.Join(d, "missing.yaml")); err == nil {
		t.Fatalf("expected missing file load error")
	}

	badPath := filepath.Join(d, "bad.yaml")
	_ = os.WriteFile(badPath, []byte("::bad::yaml"), 0o644)
	if _, err := LoadFile(badPath); err == nil || !strings.Contains(err.Error(), "failed to parse config") {
		t.Fatalf("expected parse error, got: %v", err)
	}
}
