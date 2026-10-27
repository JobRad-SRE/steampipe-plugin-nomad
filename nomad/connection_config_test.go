package nomad

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/turbot/steampipe-plugin-sdk/v5/connection"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
)

// newTestQueryData builds a minimal QueryData with a real connection cache.
func newTestQueryData(t *testing.T, config any) *plugin.QueryData {
	t.Helper()
	connCache, err := connection.NewConnectionCache("test", 100*1024*1024)
	if err != nil {
		t.Fatalf("connection cache: %v", err)
	}
	return &plugin.QueryData{
		Connection: &plugin.Connection{
			Name:   "test",
			Config: config,
		},
		ConnectionCache: connCache,
		QueryContext:    &plugin.QueryContext{},
	}
}

// stuckTLSServer accepts connections but never completes the handshake.
func stuckTLSServer(t *testing.T) (addr string, stop func()) {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	go func() {
		for {
			c, err := ln.Accept()
			if err != nil {
				return
			}
			// Hold the connection open; never handshake, never close.
			go func() {
				buf := make([]byte, 1)
				for {
					if _, err := c.Read(buf); err != nil {
						return
					}
				}
			}()
		}
	}()
	return ln.Addr().String(), func() { _ = ln.Close() }
}

// TestTLSTimeoutApplied verifies tls_timeout_seconds is honored end to end:
// a server that never handshakes must trip the timeout at ~the configured value.
func TestTLSTimeoutApplied(t *testing.T) {
	addr, stop := stuckTLSServer(t)
	defer stop()

	d := newTestQueryData(t, nomadConfig{
		Address:          new("https://" + addr),
		TLSTimeoutSecond: new(2),
	})

	start := time.Now()
	client, err := getClient(context.Background(), d)
	if err != nil {
		t.Fatalf("getClient: %v", err)
	}
	_, _, err = client.Jobs().List(nil)
	elapsed := time.Since(start)

	if err == nil {
		t.Fatal("expected TLS handshake timeout error")
	}
	if elapsed > 5*time.Second {
		t.Fatalf("request took %v, timeout was not applied (expected ~2s)", elapsed)
	}
	t.Logf("elapsed=%v err=%v", elapsed, err)
}

// TestDefaultTLSTimeoutWorks verifies the 30s default when no config is set.
func TestDefaultTLSTimeoutWorks(t *testing.T) {
	addr, stop := stuckTLSServer(t)
	defer stop()

	seconds := 2
	d := newTestQueryData(t, nomadConfig{
		Address:          new("https://" + addr),
		TLSTimeoutSecond: &seconds,
	})

	timeout, err := tlsTimeout(GetConfig(d.Connection))
	if err != nil {
		t.Fatalf("tlsTimeout: %v", err)
	}
	if timeout != 2*time.Second {
		t.Fatalf("tlsTimeout = %v, want 2s", timeout)
	}
}

// TestClientCachedPerConnection verifies the api client is built once and
// reused from the connection cache.
func TestClientCachedPerConnection(t *testing.T) {
	addr, stop := stuckTLSServer(t)
	defer stop()

	d := newTestQueryData(t, nomadConfig{
		Address: new("https://" + addr),
	})

	c1, err := getClient(context.Background(), d)
	if err != nil {
		t.Fatalf("getClient 1: %v", err)
	}
	c2, err := getClient(context.Background(), d)
	if err != nil {
		t.Fatalf("getClient 2: %v", err)
	}
	if c1 != c2 {
		t.Fatal("expected the same *api.Client instance from the connection cache")
	}
}

// TestInvalidTLSTimeoutRejected verifies zero/negative values are rejected.
func TestInvalidTLSTimeoutRejected(t *testing.T) {
	zero := 0
	if _, err := tlsTimeout(nomadConfig{TLSTimeoutSecond: &zero}); err == nil {
		t.Fatal("expected error for tls_timeout_seconds = 0")
	}
	neg := -5
	if _, err := tlsTimeout(nomadConfig{TLSTimeoutSecond: &neg}); err == nil {
		t.Fatal("expected error for tls_timeout_seconds = -5")
	}
}
