package main

import (
	"context"
	"net"
	"sync"
	"time"

	"layeh.com/radius"
)

// RADIUSClientPool manages a pool of RADIUS clients for better performance
type RADIUSClientPool struct {
	mu      sync.RWMutex
	clients []*radius.Client
	maxSize int
	timeout time.Duration
}

// NewRADIUSClientPool creates a new RADIUS client pool
func NewRADIUSClientPool(maxSize int, timeout time.Duration) *RADIUSClientPool {
	return &RADIUSClientPool{
		clients: make([]*radius.Client, 0, maxSize),
		maxSize: maxSize,
		timeout: timeout,
	}
}

// Get retrieves a client from the pool or creates a new one
func (p *RADIUSClientPool) Get() *radius.Client {
	p.mu.Lock()
	defer p.mu.Unlock()

	if len(p.clients) > 0 {
		// Pop from the end for better performance
		client := p.clients[len(p.clients)-1]
		p.clients = p.clients[:len(p.clients)-1]
		return client
	}

	// Create new client if pool is empty
	client := &radius.Client{
		Retry:           3,
		MaxPacketErrors: 2,
	}

	return client
}

// Put returns a client to the pool
func (p *RADIUSClientPool) Put(client *radius.Client) {
	if client == nil {
		return
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	// Only add to pool if we haven't reached max size
	if len(p.clients) < p.maxSize {
		p.clients = append(p.clients, client)
	}
}

// NetworkInterface provides optimized network interface operations
type NetworkInterface struct {
	mu         sync.RWMutex
	interfaces map[string]*net.Interface
}

var networkCache = &NetworkInterface{
	interfaces: make(map[string]*net.Interface),
}

// GetInterface returns a cached network interface
func (ni *NetworkInterface) GetInterface(name string) (*net.Interface, error) {
	ni.mu.RLock()
	if intf, exists := ni.interfaces[name]; exists {
		ni.mu.RUnlock()
		return intf, nil
	}
	ni.mu.RUnlock()

	// Interface not in cache, look it up
	ni.mu.Lock()
	defer ni.mu.Unlock()

	// Double check after acquiring write lock
	if intf, exists := ni.interfaces[name]; exists {
		return intf, nil
	}

	intf, err := net.InterfaceByName(name)
	if err != nil {
		return nil, err
	}

	ni.interfaces[name] = intf
	return intf, nil
}

// ClearCache clears the interface cache
func (ni *NetworkInterface) ClearCache() {
	ni.mu.Lock()
	defer ni.mu.Unlock()
	ni.interfaces = make(map[string]*net.Interface)
}

// RateLimiter provides basic rate limiting functionality
type RateLimiter struct {
	tokens chan struct{}
	rate   time.Duration
	ctx    context.Context
	cancel context.CancelFunc
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(requestsPerSecond int) *RateLimiter {
	ctx, cancel := context.WithCancel(context.Background())

	rl := &RateLimiter{
		tokens: make(chan struct{}, requestsPerSecond),
		rate:   time.Second / time.Duration(requestsPerSecond),
		ctx:    ctx,
		cancel: cancel,
	}

	// Fill initial tokens
	for i := 0; i < requestsPerSecond; i++ {
		rl.tokens <- struct{}{}
	}

	// Start token replenishment
	go rl.refillTokens()

	return rl
}

// Wait waits for a token to become available
func (rl *RateLimiter) Wait() bool {
	select {
	case <-rl.tokens:
		return true
	case <-rl.ctx.Done():
		return false
	}
}

// TryWait attempts to get a token without blocking
func (rl *RateLimiter) TryWait() bool {
	select {
	case <-rl.tokens:
		return true
	default:
		return false
	}
}

// refillTokens replenishes tokens at the configured rate
func (rl *RateLimiter) refillTokens() {
	ticker := time.NewTicker(rl.rate)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			select {
			case rl.tokens <- struct{}{}:
			default:
				// Token bucket is full
			}
		case <-rl.ctx.Done():
			return
		}
	}
}

// Close stops the rate limiter
func (rl *RateLimiter) Close() {
	rl.cancel()
}
