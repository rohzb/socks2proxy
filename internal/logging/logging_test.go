package logging

import (
	"bytes"
	"log"
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestParseLevel(t *testing.T) {
	if parseLevel("debug") != LevelDebug {
		t.Fatalf("debug level parse failed")
	}
	if parseLevel("warn") != LevelWarn {
		t.Fatalf("warn level parse failed")
	}
	if parseLevel("error") != LevelError {
		t.Fatalf("error level parse failed")
	}
	if parseLevel("whatever") != LevelInfo {
		t.Fatalf("unknown level should default to info")
	}
}

func TestLoggerLevelFiltering(t *testing.T) {
	buf := &bytes.Buffer{}
	l := New("warn")
	l.base = log.New(buf, "", 0)

	l.Debugf("debug")
	l.Infof("info")
	l.Warnf("warn")
	l.Errorf("error")

	out := buf.String()
	if strings.Contains(out, "DEBUG") || strings.Contains(out, "INFO") {
		t.Fatalf("debug/info logs should be filtered, got: %s", out)
	}
	if !strings.Contains(out, "WARN") || !strings.Contains(out, "ERROR") {
		t.Fatalf("warn/error logs should be present, got: %s", out)
	}
}

func TestLoggerPrintf(t *testing.T) {
	buf := &bytes.Buffer{}
	l := New("info")
	l.base = log.New(buf, "", 0)
	l.Printf("hello %s", "world")
	out := buf.String()
	if !strings.Contains(out, "INFO") || !strings.Contains(out, "hello world") {
		t.Fatalf("unexpected printf output: %q", out)
	}
}

func TestLoggerFatalfHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_LOG_FATAL_HELPER") == "1" {
		l := New("info")
		l.Fatalf("boom")
		return
	}
	cmd := exec.Command(os.Args[0], "-test.run=TestLoggerFatalfHelperProcess")
	cmd.Env = append(os.Environ(), "GO_WANT_LOG_FATAL_HELPER=1")
	err := cmd.Run()
	if err == nil {
		t.Fatalf("expected helper process to exit non-zero")
	}
}
