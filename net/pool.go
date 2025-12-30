// Package net implements HTTP networking for the browser.
// This file contains HTTP connection pooling for Keep-Alive support.
package net

import (
	"go-web-browser/logger"
	"net"
	"sync"
)

// MaxConnectionsPerHost is the maximum number of idle Keep-Alive connections
// per host, as recommended by HTTP/1.1 (RFC 2616).
const MaxConnectionsPerHost = 6

// ConnectionPool manages persistent HTTP connections for Keep-Alive.
//
// It maintains a pool of idle connections per server address, allowing
// connection reuse across multiple HTTP requests to the same host.
// This significantly reduces latency by avoiding repeated TCP handshakes.
//
// The pool is thread-safe and can be used concurrently from multiple goroutines.
type ConnectionPool struct {
	connections map[string][]net.Conn // "host:port" → []net.Conn
	mu          sync.Mutex            // protects connections map
	maxPerHost  int                   // maximum idle connections per host
}

// NewConnectionPool creates a new ConnectionPool with default settings.
//
// The pool will maintain up to MaxConnectionsPerHost idle connections
// per server address. Connections exceeding this limit are closed immediately.
func NewConnectionPool() *ConnectionPool {
	return &ConnectionPool{
		connections: make(map[string][]net.Conn),
		maxPerHost:  MaxConnectionsPerHost,
	}
}

// Get retrieves an idle connection from the pool for the given address.
//
// It returns (conn, true) if an idle connection is available, or (nil, false)
// if the pool is empty for this address. The retrieved connection is removed
// from the pool (check-out pattern) and should be returned with Put after use.
//
// Get is safe for concurrent use.
func (pool *ConnectionPool) Get(address string) (net.Conn, bool) {
	pool.mu.Lock()
	defer pool.mu.Unlock()

	conns := pool.connections[address]
	if len(conns) == 0 {
		return nil, false
	}

	// Pop last connection (LIFO - most recently used)
	lastIdx := len(conns) - 1
	conn := conns[lastIdx]
	pool.connections[address] = conns[:lastIdx]

	logger.Logger.Printf("Reusing connection to %s (remaining: %d)", address, len(conns)-1)
	return conn, true
}

// Put returns a connection to the pool for future reuse.
//
// If the pool already contains maxPerHost connections for this address,
// the connection is closed immediately to prevent resource leaks.
// Otherwise, the connection is stored for reuse by future requests.
//
// Put is safe for concurrent use.
func (pool *ConnectionPool) Put(address string, conn net.Conn) {
	pool.mu.Lock()
	defer pool.mu.Unlock()

	conns := pool.connections[address]

	if len(conns) < pool.maxPerHost {
		pool.connections[address] = append(conns, conn)
		logger.Logger.Printf("Stored connection to %s (total: %d/%d)", address, len(conns)+1, pool.maxPerHost)
	} else {
		conn.Close()
		logger.Logger.Printf("Pool full, closed connection to %s (%d/%d)", address, pool.maxPerHost, pool.maxPerHost)
	}
}

// Close closes all idle connections for the given address and removes them from the pool.
//
// This is useful when you want to force new connections on the next request,
// or when shutting down.
//
// Close is safe for concurrent use.
func (pool *ConnectionPool) Close(address string) {
	pool.mu.Lock()
	defer pool.mu.Unlock()

	conns := pool.connections[address]
	for _, conn := range conns {
		conn.Close()
	}
	delete(pool.connections, address)
	logger.Logger.Printf("Closed all connections to %s (%d connections)", address, len(conns))
}

// GlobalConnectionPool is the global ConnectionPool instance used by the HTTP fetcher
var GlobalConnectionPool = NewConnectionPool()
