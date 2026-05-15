// File main.go wires the socks2proxy runtime components and starts the server.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"socks2proxy/internal/acl"
	"socks2proxy/internal/config"
	"socks2proxy/internal/logging"
	"socks2proxy/internal/proxy"
	"socks2proxy/internal/socks5"
)

// main loads configuration, constructs dependencies, and runs the server.
func main() {
	configPath := ""
	logLevelOverride := ""
	checkConfig := false
	showHelp := false
	showVersion := false

	flag.CommandLine.SetOutput(os.Stdout)
	flag.Usage = func() {
		_, _ = fmt.Fprintln(os.Stdout, buildHelpHeader())
		_, _ = fmt.Fprintln(os.Stdout, "")
		_, _ = fmt.Fprintln(os.Stdout, "Usage:")
		_, _ = fmt.Fprintln(os.Stdout, "  socks2proxy [options]")
		_, _ = fmt.Fprintln(os.Stdout, "")
		_, _ = fmt.Fprintln(os.Stdout, "Options:")
		_, _ = fmt.Fprintln(os.Stdout, "  -c, --config string      Path to YAML config file (required)")
		_, _ = fmt.Fprintln(os.Stdout, "  -l, --log-level string   Override logging.level from config")
		_, _ = fmt.Fprintln(os.Stdout, "  -t, --check              Parse and validate config, then exit")
		_, _ = fmt.Fprintln(os.Stdout, "      --check-config       Alias for --check")
		_, _ = fmt.Fprintln(os.Stdout, "  -V, --version            Show build version information and exit")
		_, _ = fmt.Fprintln(os.Stdout, "  -h, --help               Show this help and exit")
	}

	flag.StringVar(&configPath, "config", configPath, "Path to YAML config file")
	flag.StringVar(&configPath, "c", configPath, "Path to YAML config file")
	flag.StringVar(&logLevelOverride, "log-level", logLevelOverride, "Override logging.level from config")
	flag.StringVar(&logLevelOverride, "l", logLevelOverride, "Override logging.level from config")
	flag.BoolVar(&checkConfig, "check", checkConfig, "Parse and validate config, then exit")
	flag.BoolVar(&checkConfig, "t", checkConfig, "Parse and validate config, then exit")
	flag.BoolVar(&checkConfig, "check-config", checkConfig, "Parse and validate config, then exit")
	flag.BoolVar(&showVersion, "version", showVersion, "Show build version information and exit")
	flag.BoolVar(&showVersion, "V", showVersion, "Show build version information and exit")
	flag.BoolVar(&showHelp, "help", showHelp, "Show help and exit")
	flag.BoolVar(&showHelp, "h", showHelp, "Show help and exit")
	flag.Parse()
	if showHelp {
		flag.Usage()
		return
	}
	if showVersion {
		fmt.Println(buildVersionString())
		return
	}
	if configPath == "" {
		log.Fatalf("config error: --config is required")
	}

	cfg, err := config.LoadFile(configPath)
	if err != nil {
		log.Fatalf("config error: %v", err)
	}
	if logLevelOverride != "" {
		cfg.Logging.Level = logLevelOverride
		if err := cfg.Validate(); err != nil {
			log.Fatalf("config error after CLI overrides: %v", err)
		}
	}
	if checkConfig {
		fmt.Printf("config is valid: %s\n", configPath)
		return
	}

	logger := logging.New(cfg.Logging.Level)
	logger.Infof("%s", buildStartupBanner())
	logger.Debugf("loaded config path=%s listen=%s", configPath, cfg.Listen)

	clientACL, err := acl.NewAddressAllowlist([]string(cfg.AllowedClientAddresses))
	if err != nil {
		logger.Fatalf("client address ACL error: %v", err)
	}
	var portACL *acl.PortAllowlist
	if cfg.HasNonRejectDefault() {
		portACL = acl.NewAllowAllPortAllowlist()
	} else {
		portACL = acl.NewPortAllowlist(cfg.AllowedPorts())
	}

	router := &proxy.Router{
		ConnectTimeout: cfg.Timeouts.Connect.Duration,
		IdleTimeout:    cfg.Timeouts.Idle.Duration,
		MaxHeaderBytes: cfg.HTTP.MaxHeaderBytes,
		Rules:          proxy.RulesFromConfig(cfg.Routing, cfg.TLS),
		DefaultRule:    proxy.DefaultRuleFromConfig(cfg.Routing, cfg.TLS),
	}

	server := &socks5.Server{
		ListenAddr:  cfg.Listen,
		ClientACL:   clientACL,
		PortACL:     portACL,
		Router:      router,
		Logger:      logger,
		IdleTimeout: cfg.Timeouts.Idle.Duration,
	}

	logger.Infof("starting socks2proxy listen=%s rules=%d", cfg.Listen, len(cfg.Routing.Rules))
	if err := server.Serve(); err != nil {
		logger.Errorf("fatal server error: %v", err)
		os.Exit(1)
	}
}
