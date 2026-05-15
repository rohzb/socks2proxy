package main

import (
	"strings"
	"testing"
)

func TestBuildHelpHeader(t *testing.T) {
	h := buildHelpHeader()
	if !strings.Contains(h, "socks2proxy") {
		t.Fatalf("help header must contain project name, got: %q", h)
	}
	if !strings.Contains(h, "SOCKS5") {
		t.Fatalf("help header must mention SOCKS5, got: %q", h)
	}
}

func TestBuildVersionStringContainsCoreFields(t *testing.T) {
	origProject, origVersion, origCommit, origDate, origAuthor, origLicense := project, version, commit, date, author, license
	project = "socks2proxy"
	version = "1.2.3"
	commit = "abc1234"
	date = "2026-05-15T00:00:00Z"
	author = "Ruslan Ovsyannikov"
	license = "MIT"
	t.Cleanup(func() {
		project, version, commit, date, author, license = origProject, origVersion, origCommit, origDate, origAuthor, origLicense
	})

	out := buildVersionString()
	mustContain := []string{
		"socks2proxy 1.2.3",
		"Author:",
		"License:",
		"Commit:",
		"Built:",
		"Go:",
		"Platform:",
	}
	for _, s := range mustContain {
		if !strings.Contains(out, s) {
			t.Fatalf("version output missing %q; got: %q", s, out)
		}
	}
}

func TestBuildStartupBanner(t *testing.T) {
	origProject, origVersion, origCommit, origDate := project, version, commit, date
	project = "socks2proxy"
	version = "9.9.9"
	commit = "deadbee"
	date = "2026-05-15T00:00:00Z"
	t.Cleanup(func() {
		project, version, commit, date = origProject, origVersion, origCommit, origDate
	})

	b := buildStartupBanner()
	if !strings.Contains(b, "socks2proxy 9.9.9") {
		t.Fatalf("startup banner missing name/version: %q", b)
	}
	if !strings.Contains(b, "commit=deadbee") || !strings.Contains(b, "built=2026-05-15T00:00:00Z") {
		t.Fatalf("startup banner missing commit/date: %q", b)
	}
}
