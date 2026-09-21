package handlers

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"
)

type metadataNetworkDialer interface {
	DialContext(context.Context, string, string) (net.Conn, error)
}

func newMetadataHTTPClient(resolver metadataHostnameResolver, dialer metadataNetworkDialer) *http.Client {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	// A proxy would receive the original target and bypass the validated direct
	// dial path, so metadata fetching always uses a direct connection.
	transport.Proxy = nil
	transport.DialContext = metadataDialContext(resolver, dialer)

	return &http.Client{
		Transport: transport,
		Timeout:   2 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 5 {
				return fmt.Errorf("too many redirects")
			}
			if _, err := validateOutboundMetadataURL(req.Context(), req.URL.String(), resolver); err != nil {
				return fmt.Errorf("redirect destination rejected: %w", err)
			}
			return nil
		},
	}
}

func metadataDialContext(resolver metadataHostnameResolver, dialer metadataNetworkDialer) func(context.Context, string, string) (net.Conn, error) {
	return func(ctx context.Context, network, address string) (net.Conn, error) {
		if dialer == nil {
			return nil, fmt.Errorf("metadata network dialer is unavailable")
		}
		hostname, port, err := net.SplitHostPort(address)
		if err != nil {
			return nil, fmt.Errorf("invalid metadata dial address: %w", err)
		}
		addresses, err := resolveAllowedMetadataAddresses(ctx, hostname, resolver)
		if err != nil {
			return nil, err
		}

		var lastErr error
		for _, resolvedAddress := range addresses {
			if network == "tcp4" && resolvedAddress.IP.To4() == nil {
				continue
			}
			if network == "tcp6" && resolvedAddress.IP.To4() != nil {
				continue
			}
			dialAddress := net.JoinHostPort(resolvedAddress.IP.String(), port)
			connection, err := dialer.DialContext(ctx, network, dialAddress)
			if err == nil {
				return connection, nil
			}
			lastErr = err
		}
		if lastErr != nil {
			return nil, fmt.Errorf("metadata connection failed: %w", lastErr)
		}
		return nil, fmt.Errorf("metadata hostname has no address for %s", network)
	}
}
