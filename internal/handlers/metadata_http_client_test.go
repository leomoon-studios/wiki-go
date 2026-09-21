package handlers

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"net/http"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type metadataDialFunc func(context.Context, string, string) (net.Conn, error)

func (dial metadataDialFunc) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	return dial(ctx, network, address)
}

type sequenceMetadataResolver struct {
	mu        sync.Mutex
	responses map[string][][]net.IPAddr
	calls     map[string]int
}

func (resolver *sequenceMetadataResolver) LookupIPAddr(_ context.Context, hostname string) ([]net.IPAddr, error) {
	resolver.mu.Lock()
	defer resolver.mu.Unlock()
	index := resolver.calls[hostname]
	resolver.calls[hostname] = index + 1
	responses := resolver.responses[hostname]
	if len(responses) == 0 {
		return nil, fmt.Errorf("no response for %s", hostname)
	}
	if index >= len(responses) {
		index = len(responses) - 1
	}
	return responses[index], nil
}

func TestMetadataHTTPClientRejectsDNSRebindingBeforeDial(t *testing.T) {
	resolver := &sequenceMetadataResolver{
		responses: map[string][][]net.IPAddr{
			"rebind.example": {
				{{IP: net.ParseIP("192.0.2.40")}},
				{{IP: net.ParseIP("127.0.0.1")}},
			},
		},
		calls: make(map[string]int),
	}
	targetURL := "http://rebind.example/metadata"
	if _, err := validateOutboundMetadataURL(context.Background(), targetURL, resolver); err != nil {
		t.Fatalf("initial validation failed: %v", err)
	}

	var dialCalls atomic.Int32
	dialer := metadataDialFunc(func(context.Context, string, string) (net.Conn, error) {
		dialCalls.Add(1)
		return nil, fmt.Errorf("unexpected dial")
	})
	_, err := fetchURLMetadataWithClient(targetURL, newMetadataHTTPClient(resolver, dialer))
	if err == nil {
		t.Fatal("metadata fetch succeeded after DNS changed to loopback")
	}
	if dialCalls.Load() != 0 {
		t.Fatalf("underlying dialer was called %d times", dialCalls.Load())
	}
}

func TestMetadataHTTPClientRejectsRedirectToProhibitedDestination(t *testing.T) {
	const allowedPort = "8080"
	const blockedPort = "9090"
	var blockedRequests atomic.Int32
	var allowedRequests atomic.Int32
	resolver := staticMetadataResolver{addresses: map[string][]net.IPAddr{
		"allowed.example": {{IP: net.ParseIP("192.0.2.50")}},
		"blocked.example": {{IP: net.ParseIP("127.0.0.1")}},
	}}
	var dialedAddressesMu sync.Mutex
	var dialedAddresses []string
	dialer := metadataDialFunc(func(_ context.Context, _, address string) (net.Conn, error) {
		dialedAddressesMu.Lock()
		dialedAddresses = append(dialedAddresses, address)
		dialedAddressesMu.Unlock()
		if address == net.JoinHostPort("127.0.0.1", blockedPort) {
			blockedRequests.Add(1)
			return nil, fmt.Errorf("prohibited destination was dialed")
		}
		if address != net.JoinHostPort("192.0.2.50", allowedPort) {
			return nil, fmt.Errorf("unexpected dial address %s", address)
		}

		clientConnection, serverConnection := net.Pipe()
		go func() {
			defer serverConnection.Close()
			request, err := http.ReadRequest(bufio.NewReader(serverConnection))
			if err != nil {
				return
			}
			request.Body.Close()
			allowedRequests.Add(1)
			fmt.Fprintf(serverConnection, "HTTP/1.1 302 Found\r\nLocation: http://blocked.example:%s/metadata\r\nContent-Length: 0\r\nConnection: close\r\n\r\n", blockedPort)
		}()
		return clientConnection, nil
	})

	targetURL := "http://allowed.example:" + allowedPort + "/start"
	_, err := fetchURLMetadataWithClient(targetURL, newMetadataHTTPClient(resolver, dialer))
	if err == nil {
		t.Fatal("metadata fetch followed a redirect to a prohibited destination")
	}
	if allowedRequests.Load() != 1 {
		t.Fatalf("allowed server requests = %d, want 1", allowedRequests.Load())
	}
	if blockedRequests.Load() != 0 {
		t.Fatalf("blocked server received %d requests", blockedRequests.Load())
	}
	dialedAddressesMu.Lock()
	defer dialedAddressesMu.Unlock()
	if len(dialedAddresses) != 1 || dialedAddresses[0] != net.JoinHostPort("192.0.2.50", allowedPort) {
		t.Fatalf("dialed addresses = %v, want only the validated allowed address", dialedAddresses)
	}
}

func TestMetadataHTTPClientPreservesTimeoutAndDirectTransport(t *testing.T) {
	client := newMetadataHTTPClient(staticMetadataResolver{}, &net.Dialer{})
	if client.Timeout != 2*time.Second {
		t.Fatalf("timeout = %s, want 2s", client.Timeout)
	}
	transport, ok := client.Transport.(*http.Transport)
	if !ok {
		t.Fatalf("transport type = %T, want *http.Transport", client.Transport)
	}
	if transport.Proxy != nil {
		t.Fatal("metadata transport unexpectedly permits proxy routing")
	}
}
