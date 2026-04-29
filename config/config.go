// Package config provides configuration structures and loading utilities
// for the MasterDnsVPN client and server.
package config

import (
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
)

// ClientConfig holds the full client-side configuration.
type ClientConfig struct {
	General  GeneralConfig  `toml:"general"`
	DNS      DNSConfig      `toml:"dns"`
	Tunnel   TunnelConfig   `toml:"tunnel"`
	Proxy    ProxyConfig    `toml:"proxy"`
	Logging  LoggingConfig  `toml:"logging"`
}

// ServerConfig holds the full server-side configuration.
type ServerConfig struct {
	General  GeneralConfig  `toml:"general"`
	DNS      DNSConfig      `toml:"dns"`
	Tunnel   TunnelConfig   `toml:"tunnel"`
	Logging  LoggingConfig  `toml:"logging"`
}

// GeneralConfig contains general application settings.
type GeneralConfig struct {
	Mode       string `toml:"mode"`        // "client" or "server"
	ServerAddr string `toml:"server_addr"` // Remote server address (client mode)
	ServerPort int    `toml:"server_port"` // Remote server port
	ListenAddr string `toml:"listen_addr"` // Local listen address (server mode)
	ListenPort int    `toml:"listen_port"` // Local listen port
	Secret     string `toml:"secret"`      // Shared secret / pre-shared key
}

// DNSConfig contains DNS-over-tunnel settings.
type DNSConfig struct {
	Enabled       bool     `toml:"enabled"`
	LocalPort     int      `toml:"local_port"`      // Local DNS listener port
	UpstreamDNS   []string `toml:"upstream_dns"`    // Upstream DNS servers
	DomainSuffix  string   `toml:"domain_suffix"`   // Domain suffix used for DNS tunneling
	QueryTimeout  int      `toml:"query_timeout"`   // DNS query timeout in seconds
	MaxRetries    int      `toml:"max_retries"`     // Maximum DNS query retries
}

// TunnelConfig contains VPN tunnel settings.
type TunnelConfig struct {
	Interface  string `toml:"interface"`   // TUN interface name
	MTU        int    `toml:"mtu"`         // Maximum transmission unit
	IPAddress  string `toml:"ip_address"`  // Tunnel IP address (CIDR)
	RemoteIP   string `toml:"remote_ip"`   // Remote tunnel endpoint IP
	KeepAlive  int    `toml:"keep_alive"`  // Keep-alive interval in seconds
}

// ProxyConfig contains optional HTTP/SOCKS proxy settings for the client.
type ProxyConfig struct {
	Enabled  bool   `toml:"enabled"`
	Type     string `toml:"type"`     // "http" or "socks5"
	Address  string `toml:"address"`  // Proxy address
	Port     int    `toml:"port"`     // Proxy port
	Username string `toml:"username"` // Optional proxy username
	Password string `toml:"password"` // Optional proxy password
}

// LoggingConfig contains logging preferences.
type LoggingConfig struct {
	Level  string `toml:"level"`   // "debug", "info", "warn", "error"
	File   string `toml:"file"`    // Log file path; empty means stdout
	Format string `toml:"format"`  // "text" or "json"
}

// LoadClientConfig reads and parses a TOML configuration file into a ClientConfig.
// Returns an error if the file cannot be read or parsed.
func LoadClientConfig(path string) (*ClientConfig, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, fmt.Errorf("config file not found: %s", path)
	}

	var cfg ClientConfig
	if _, err := toml.DecodeFile(path, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file %s: %w", path, err)
	}

	if err := validateClientConfig(&cfg); err != nil {
		return nil, fmt.Errorf("invalid client config: %w", err)
	}

	return &cfg, nil
}

// LoadServerConfig reads and parses a TOML configuration file into a ServerConfig.
func LoadServerConfig(path string) (*ServerConfig, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, fmt.Errorf("config file not found: %s", path)
	}

	var cfg ServerConfig
	if _, err := toml.DecodeFile(path, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file %s: %w", path, err)
	}

	if err := validateServerConfig(&cfg); err != nil {
		return nil, fmt.Errorf("invalid server config: %w", err)
	}

	return &cfg, nil
}

// validateClientConfig performs basic sanity checks on a ClientConfig.
func validateClientConfig(cfg *ClientConfig) error {
	if cfg.General.ServerAddr == "" {
		return fmt.Errorf("general.server_addr must not be empty")
	}
	if cfg.General.ServerPort <= 0 || cfg.General.ServerPort > 65535 {
		return fmt.Errorf("general.server_port must be between 1 and 65535")
	}
	if cfg.General.Secret == "" {
		return fmt.Errorf("general.secret must not be empty")
	}
	if cfg.DNS.Enabled && len(cfg.DNS.UpstreamDNS) == 0 {
		return fmt.Errorf("dns.upstream_dns must contain at least one server when DNS is enabled")
	}
	return nil
}

// validateServerConfig performs basic sanity checks on a ServerConfig.
func validateServerConfig(cfg *ServerConfig) error {
	if cfg.General.ListenPort <= 0 || cfg.General.ListenPort > 65535 {
		return fmt.Errorf("general.listen_port must be between 1 and 65535")
	}
	if cfg.General.Secret == "" {
		return fmt.Errorf("general.secret must not be empty")
	}
	return nil
}
