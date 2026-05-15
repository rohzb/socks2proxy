package main

import (
	"fmt"
	"runtime"
)

var (
	project = "socks2proxy"
	version = "dev"
	commit  = "unknown"
	date    = "unknown"
	author  = "Ruslan Ovsyannikov"
	license = "MIT"
)

func buildVersionString() string {
	return fmt.Sprintf(
		"%s %s\n\nA lightweight SOCKS5 bridge with rule-based routing.\n\nAuthor:   %s\nLicense:  %s\nCommit:   %s\nBuilt:    %s\nGo:       %s\nPlatform: %s/%s",
		project,
		version,
		author,
		license,
		commit,
		date,
		runtime.Version(),
		runtime.GOOS,
		runtime.GOARCH,
	)
}

func buildHelpHeader() string {
	return "socks2proxy - SOCKS5 bridge with rule-based routing"
}

func buildStartupBanner() string {
	return fmt.Sprintf("%s %s | commit=%s | built=%s", project, version, commit, date)
}
