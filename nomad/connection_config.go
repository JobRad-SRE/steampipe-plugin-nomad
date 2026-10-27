package nomad

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	cleanhttp "github.com/hashicorp/go-cleanhttp"
	"github.com/hashicorp/nomad/api"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
)

type nomadConfig struct {
	Address          *string `hcl:"address"`
	Namespace        *string `hcl:"namespace"`
	SecretID         *string `hcl:"secret_id"`
	TLSTimeoutSecond *int    `hcl:"tls_timeout_seconds"`
	MaxConnections   *int    `hcl:"max_connections"`
}

func ConfigInstance() interface{} {
	return &nomadConfig{}
}

func GetConfig(connection *plugin.Connection) nomadConfig {
	if connection == nil || connection.Config == nil {
		return nomadConfig{}
	}
	config, _ := connection.Config.(nomadConfig)
	return config
}

// defaultTLSTimeout is applied to the TLS handshake when tls_timeout_seconds
// is not set. The Nomad SDK default is 10s, which is too tight for slow or
// remote links.
const defaultTLSTimeout = 3 * time.Minute

func tlsTimeout(config nomadConfig) (time.Duration, error) {
	if config.TLSTimeoutSecond == nil {
		return defaultTLSTimeout, nil
	}
	t := time.Duration(*config.TLSTimeoutSecond) * time.Second
	if t <= 0 {
		return 0, fmt.Errorf("invalid tls_timeout_seconds %d: must be a positive number of seconds", *config.TLSTimeoutSecond)
	}
	return t, nil
}

// defaultMaxConnections caps concurrent TCP connections to the Nomad HTTP API.
// Nomad agents close connections from a single IP beyond
// limits.http_max_conns_per_client (default 100), which surfaces as
// "connection reset by peer". Steampipe runs up to 500 rows concurrently, and
// over HTTP/1.1 every in-flight request needs its own connection, so without a
// cap a single query can exceed the server limit.
const defaultMaxConnections = 6

func maxConnections(config nomadConfig) (int, error) {
	if config.MaxConnections == nil {
		return defaultMaxConnections, nil
	}
	if *config.MaxConnections <= 0 {
		return 0, fmt.Errorf("invalid max_connections %d: must be a positive number", *config.MaxConnections)
	}
	return *config.MaxConnections, nil
}

// getClient returns a Nomad API client for this connection, creating and
// caching it on first use. The client is reused for the lifetime of the
// connection so TLS connections are kept alive instead of being re-negotiated
// on every hydrate call.
func getClient(ctx context.Context, d *plugin.QueryData) (*api.Client, error) {
	nomadConfig := GetConfig(d.Connection)

	cacheKey := "nomad_api_client"
	if cached, ok := d.ConnectionCache.Get(ctx, cacheKey); ok {
		return cached.(*api.Client), nil
	}

	timeout, err := tlsTimeout(nomadConfig)
	if err != nil {
		return nil, err
	}
	maxConns, err := maxConnections(nomadConfig)
	if err != nil {
		return nil, err
	}

	address := os.Getenv("NOMAD_ADDR")
	namespace := os.Getenv("NOMAD_NAMESPACE")
	secretId := os.Getenv("NOMAD_TOKEN")

	if nomadConfig.Address != nil {
		address = *nomadConfig.Address
	}
	if nomadConfig.SecretID != nil {
		secretId = *nomadConfig.SecretID
	}
	if nomadConfig.Namespace != nil {
		namespace = *nomadConfig.Namespace
	}

	if address == "" {
		return nil, errors.New("'address' or ('address' and 'secret_id') must be set in the connection configuration. Edit your connection configuration file and then restart Steampipe.")
	}

	// Build our own HTTP client: the SDK pins TLSHandshakeTimeout to 10s and
	// offers no override, so we construct the transport ourselves.
	httpClient := cleanhttp.DefaultPooledClient()
	transport := httpClient.Transport.(*http.Transport)
	// ConfigureTLS writes to TLSClientConfig; cleanhttp leaves it nil.
	transport.TLSClientConfig = &tls.Config{
		MinVersion: tls.VersionTLS12,
	}
	transport.TLSHandshakeTimeout = timeout
	// Requests beyond maxConns wait for a free connection instead of opening
	// new ones the Nomad agent would reset. Keeping the same number idle lets
	// the next burst reuse them rather than re-handshaking (cleanhttp keeps only
	// GOMAXPROCS+1 idle per host).
	transport.MaxConnsPerHost = maxConns
	transport.MaxIdleConnsPerHost = maxConns

	con := api.DefaultConfig()
	con.Address = address
	con.SecretID = secretId
	con.Namespace = namespace
	con.HttpClient = httpClient

	// NewClient skips ConfigureTLS when an HttpClient is supplied, but it
	// still applies TLSConfig-derived env vars (NOMAD_CACERT etc.), so run it
	// here to preserve that behavior.
	if err := api.ConfigureTLS(httpClient, con.TLSConfig); err != nil {
		return nil, fmt.Errorf("failed to configure TLS: %w", err)
	}

	client, err := api.NewClient(con)
	if err != nil {
		return nil, err
	}

	if err := d.ConnectionCache.Set(ctx, cacheKey, client); err != nil {
		return nil, fmt.Errorf("failed to cache Nomad client: %w", err)
	}

	return client, nil
}
