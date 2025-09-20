package provider

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"os"

	retryablehttp "github.com/hashicorp/go-retryablehttp"
	pihole "github.com/ryanwholey/go-pihole"
)

// Config defines the configuration options for the Pi-hole client
type Config struct {
	// The Pi-hole URL
	URL string

	// The Pi-hole admin password
	Password string

	// UserAgent for requests
	UserAgent string

	// Custom CA file
	CAFile string

	// SessionID can be passed to reduce the number of requests against the /api/auth endpoint
	SessionID string
}

func (c Config) Client(ctx context.Context) (*pihole.Client, error) {
	retryClient := retryablehttp.NewClient()

	if c.CAFile != "" {
		ca, err := os.ReadFile(c.CAFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read CA file %q: %w", c.CAFile, err)
		}

		rootCAs := x509.NewCertPool()
		if ok := rootCAs.AppendCertsFromPEM(ca); !ok {
			return nil, fmt.Errorf("failed to parse CA file %q: no certificates found", c.CAFile)
		}

		baseTransport := retryClient.HTTPClient.Transport
		if baseTransport == nil {
			baseTransport = http.DefaultTransport
		}

		transport, ok := baseTransport.(*http.Transport)
		if !ok {
			return nil, fmt.Errorf("unexpected transport type %T", baseTransport)
		}

		clonedTransport := transport.Clone()

		var tlsConfig *tls.Config
		if transport.TLSClientConfig != nil {
			tlsConfig = transport.TLSClientConfig.Clone()
		} else {
			tlsConfig = &tls.Config{}
		}

		tlsConfig.RootCAs = rootCAs
		clonedTransport.TLSClientConfig = tlsConfig

		retryClient.HTTPClient.Transport = clonedTransport
	}

	httpClient := retryClient.StandardClient()

	headers := http.Header{}
	headers.Add("User-Agent", c.UserAgent)

	config := pihole.Config{
		BaseURL:    c.URL,
		Password:   c.Password,
		Headers:    headers,
		HttpClient: httpClient,
		SessionID:  c.SessionID,
	}

	return pihole.New(config)
}
