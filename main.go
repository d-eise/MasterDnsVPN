package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/BurntSushi/toml"
)

// Version information
const (
	AppName    = "MasterDnsVPN"
	AppVersion = "1.0.0"
)

// Config holds the top-level configuration loaded from the TOML file.
type Config struct {
	General  GeneralConfig  `toml:"general"`
	DNS      DNSConfig      `toml:"dns"`
	VPN      VPNConfig      `toml:"vpn"`
	Logging  LoggingConfig  `toml:"logging"`
}

// GeneralConfig contains general application settings.
type GeneralConfig struct {
	Mode       string `toml:"mode"`        // "client" or "server"
	ListenAddr string `toml:"listen_addr"` // e.g. "0.0.0.0:5300"
}

// DNSConfig holds DNS-related configuration.
type DNSConfig struct {
	Upstream    string   `toml:"upstream"`     // Upstream DNS server
	Fallback    string   `toml:"fallback"`     // Fallback DNS server
	Domains     []string `toml:"domains"`      // Domains to route through VPN
	CacheTTL    int      `toml:"cache_ttl"`    // DNS cache TTL in seconds
	EnableCache bool     `toml:"enable_cache"` // Whether to cache DNS responses
}

// VPNConfig holds VPN tunnel configuration.
type VPNConfig struct {
	ServerAddr string `toml:"server_addr"` // VPN server address
	ServerPort int    `toml:"server_port"` // VPN server port
	Protocol   string `toml:"protocol"`    // "tcp" or "udp"
	Secret     string `toml:"secret"`      // Shared secret / PSK
	MTU        int    `toml:"mtu"`         // MTU for the tunnel interface
}

// LoggingConfig controls log verbosity and output.
type LoggingConfig struct {
	Level  string `toml:"level"`   // "debug", "info", "warn", "error"
	Output string `toml:"output"`  // "stdout" or file path
}

func main() {
	// Parse command-line flags
	configPath := flag.String("config", "client_config.toml", "Path to the configuration file")
	showVersion := flag.Bool("version", false, "Print version information and exit")
	flag.Parse()

	if *showVersion {
		fmt.Printf("%s v%s\n", AppName, AppVersion)
		os.Exit(0)
	}

	// Load configuration
	cfg, err := loadConfig(*configPath)
	if err != nil {
		log.Fatalf("[ERROR] Failed to load configuration from %s: %v", *configPath, err)
	}

	log.Printf("[INFO] Starting %s v%s in '%s' mode", AppName, AppVersion, cfg.General.Mode)
	log.Printf("[INFO] Listening on %s", cfg.General.ListenAddr)

	// Initialize and start the application based on mode
	switch cfg.General.Mode {
	case "client":
		runClient(cfg)
	case "server":
		runServer(cfg)
	default:
		log.Fatalf("[ERROR] Unknown mode '%s'. Must be 'client' or 'server'.", cfg.General.Mode)
	}
}

// loadConfig reads and parses the TOML configuration file.
func loadConfig(path string) (*Config, error) {
	var cfg Config
	if _, err := toml.DecodeFile(path, &cfg); err != nil {
		return nil, fmt.Errorf("toml decode error: %w", err)
	}
	// Apply defaults
	if cfg.VPN.MTU == 0 {
		cfg.VPN.MTU = 1500
	}
	if cfg.VPN.Protocol == "" {
		cfg.VPN.Protocol = "udp"
	}
	return &cfg, nil
}

// runClient starts the DNS-VPN client.
func runClient(cfg *Config) {
	log.Printf("[INFO] Client mode: routing domains %v through VPN at %s:%d",
		cfg.DNS.Domains, cfg.VPN.ServerAddr, cfg.VPN.ServerPort)

	// TODO: initialize DNS listener, VPN tunnel, and routing logic
	waitForShutdown()
}

// runServer starts the DNS-VPN server.
func runServer(cfg *Config) {
	log.Printf("[INFO] Server mode: upstream DNS %s, fallback %s",
		cfg.DNS.Upstream, cfg.DNS.Fallback)

	// TODO: initialize DNS server, VPN endpoint, and forwarding logic
	waitForShutdown()
}

// waitForShutdown blocks until an OS interrupt or SIGTERM is received.
func waitForShutdown() {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	sig := <-sigCh
	log.Printf("[INFO] Received signal '%s', shutting down gracefully...", sig)
}
