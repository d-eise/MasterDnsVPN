package tunnel

import (
	"fmt"
	"log"
	"net"
	"sync"
	"time"
)

// TunnelConfig holds the configuration for a tunnel connection
type TunnelConfig struct {
	ServerAddr    string
	ServerPort    int
	LocalPort     int
	Protocol      string // "tcp" or "udp"
	ReconnectDelay time.Duration
	MaxRetries    int
}

// Tunnel represents an active tunnel connection
type Tunnel struct {
	config     TunnelConfig
	conn       net.Conn
	mu         sync.Mutex
	active     bool
	stopChan   chan struct{}
	doneChan   chan struct{}
}

// NewTunnel creates a new Tunnel instance with the given configuration
func NewTunnel(cfg TunnelConfig) *Tunnel {
	if cfg.ReconnectDelay == 0 {
		cfg.ReconnectDelay = 5 * time.Second
	}
	if cfg.MaxRetries == 0 {
		cfg.MaxRetries = 10
	}
	return &Tunnel{
		config:   cfg,
		stopChan: make(chan struct{}),
		doneChan: make(chan struct{}),
	}
}

// Start initiates the tunnel connection and begins forwarding traffic
func (t *Tunnel) Start() error {
	t.mu.Lock()
	if t.active {
		t.mu.Unlock()
		return fmt.Errorf("tunnel is already active")
	}
	t.active = true
	t.mu.Unlock()

	go t.run()
	return nil
}

// Stop gracefully shuts down the tunnel
func (t *Tunnel) Stop() {
	t.mu.Lock()
	if !t.active {
		t.mu.Unlock()
		return
	}
	t.mu.Unlock()

	close(t.stopChan)
	<-t.doneChan

	t.mu.Lock()
	t.active = false
	t.mu.Unlock()
	log.Println("[tunnel] Tunnel stopped")
}

// IsActive returns whether the tunnel is currently running
func (t *Tunnel) IsActive() bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.active
}

// run manages the tunnel lifecycle including reconnection logic
func (t *Tunnel) run() {
	defer close(t.doneChan)

	retries := 0
	for {
		select {
		case <-t.stopChan:
			if t.conn != nil {
				t.conn.Close()
			}
			return
		default:
		}

		addr := fmt.Sprintf("%s:%d", t.config.ServerAddr, t.config.ServerPort)
		log.Printf("[tunnel] Connecting to %s via %s", addr, t.config.Protocol)

		conn, err := net.DialTimeout(t.config.Protocol, addr, 10*time.Second)
		if err != nil {
			retries++
			log.Printf("[tunnel] Connection failed (attempt %d/%d): %v", retries, t.config.MaxRetries, err)
			if t.config.MaxRetries > 0 && retries >= t.config.MaxRetries {
				log.Println("[tunnel] Max retries reached, stopping tunnel")
				return
			}
			select {
			case <-t.stopChan:
				return
			case <-time.After(t.config.ReconnectDelay):
				continue
			}
		}

		retries = 0
		t.mu.Lock()
		t.conn = conn
		t.mu.Unlock()

		log.Printf("[tunnel] Connected to %s", addr)
		t.handleConnection(conn)

		log.Println("[tunnel] Connection lost, attempting to reconnect...")
	}
}

// handleConnection manages data flow for an established connection
func (t *Tunnel) handleConnection(conn net.Conn) {
	defer conn.Close()

	done := make(chan struct{})
	go func() {
		defer close(done)
		buf := make([]byte, 4096)
		for {
			conn.SetReadDeadline(time.Now().Add(30 * time.Second))
			_, err := conn.Read(buf)
			if err != nil {
				log.Printf("[tunnel] Read error: %v", err)
				return
			}
		}
	}()

	select {
	case <-t.stopChan:
		return
	case <-done:
		return
	}
}
