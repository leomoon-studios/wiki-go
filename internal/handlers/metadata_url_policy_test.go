package handlers

import (
	"context"
	"errors"
	"net"
	"testing"
)

type staticMetadataResolver struct {
	addresses map[string][]net.IPAddr
	errors    map[string]error
}

func (resolver staticMetadataResolver) LookupIPAddr(_ context.Context, hostname string) ([]net.IPAddr, error) {
	if err := resolver.errors[hostname]; err != nil {
		return nil, err
	}
	return resolver.addresses[hostname], nil
}

func TestValidateOutboundMetadataURLAllowsPublicDestinations(t *testing.T) {
	resolver := staticMetadataResolver{addresses: map[string][]net.IPAddr{
		"public.example": {
			{IP: net.ParseIP("192.0.2.10")},
			{IP: net.ParseIP("2001:db8::10")},
		},
	}}
	tests := []string{
		"https://public.example/article?id=10",
		"http://public.example:8080/",
		"http://198.51.100.20/resource",
		"https://[2001:db8::20]/resource",
		"https://[64:ff9b::808:808]/resource",
	}
	for _, targetURL := range tests {
		t.Run(targetURL, func(t *testing.T) {
			parsedURL, err := validateOutboundMetadataURL(context.Background(), targetURL, resolver)
			if err != nil {
				t.Fatalf("validateOutboundMetadataURL(%q): %v", targetURL, err)
			}
			if parsedURL.String() != targetURL {
				t.Fatalf("validated URL = %q, want %q", parsedURL.String(), targetURL)
			}
		})
	}
}

func TestValidateOutboundMetadataURLRejectsProhibitedDestinations(t *testing.T) {
	resolver := staticMetadataResolver{
		addresses: map[string][]net.IPAddr{
			"localhost.example": {{IP: net.ParseIP("127.0.0.1")}},
			"private.example":   {{IP: net.ParseIP("10.20.30.40")}},
			"mixed.example": {
				{IP: net.ParseIP("192.0.2.30")},
				{IP: net.ParseIP("192.168.1.10")},
			},
			"nat64-metadata.example": {{IP: net.ParseIP("64:ff9b::a9fe:a9fe")}},
			"nat64-mixed.example": {
				{IP: net.ParseIP("192.0.2.31")},
				{IP: net.ParseIP("64:ff9b::a9fe:a9fe")},
			},
			"empty.example": {},
		},
		errors: map[string]error{
			"failure.example": errors.New("lookup failed"),
		},
	}
	tests := []struct {
		name      string
		targetURL string
	}{
		{name: "relative URL", targetURL: "/internal"},
		{name: "unsupported scheme", targetURL: "ftp://public.example/file"},
		{name: "credentials", targetURL: "https://user:password@public.example/"},
		{name: "empty hostname", targetURL: "https:///path"},
		{name: "malformed hostname", targetURL: "https://bad_host.example/"},
		{name: "empty port", targetURL: "https://public.example:/"},
		{name: "out of range port", targetURL: "https://public.example:70000/"},
		{name: "IPv4 loopback", targetURL: "http://127.0.0.1/"},
		{name: "IPv4 unspecified", targetURL: "http://0.0.0.0/"},
		{name: "IPv4 multicast", targetURL: "http://224.0.0.1/"},
		{name: "IPv4 broadcast", targetURL: "http://255.255.255.255/"},
		{name: "IPv4 private 10", targetURL: "http://10.0.0.1/"},
		{name: "IPv4 private 172", targetURL: "http://172.16.0.1/"},
		{name: "IPv4 private 192", targetURL: "http://192.168.1.1/"},
		{name: "IPv4 link local", targetURL: "http://169.254.1.1/"},
		{name: "cloud metadata IPv4", targetURL: "http://169.254.169.254/latest/meta-data/"},
		{name: "shared metadata range", targetURL: "http://100.100.100.200/latest/meta-data/"},
		{name: "Azure platform address", targetURL: "http://168.63.129.16/metadata/"},
		{name: "IPv6 loopback", targetURL: "http://[::1]/"},
		{name: "IPv6 unspecified", targetURL: "http://[::]/"},
		{name: "IPv6 multicast", targetURL: "http://[ff02::1]/"},
		{name: "IPv6 link local", targetURL: "http://[fe80::1]/"},
		{name: "IPv6 unique local", targetURL: "http://[fd12:3456::1]/"},
		{name: "AWS metadata IPv6", targetURL: "http://[fd00:ec2::254]/latest/meta-data/"},
		{name: "NAT64 loopback", targetURL: "http://[64:ff9b::7f00:1]/"},
		{name: "NAT64 private", targetURL: "http://[64:ff9b::a00:1]/"},
		{name: "NAT64 cloud metadata", targetURL: "http://[64:ff9b::a9fe:a9fe]/latest/meta-data/"},
		{name: "local-use translation", targetURL: "http://[64:ff9b:1::808:808]/"},
		{name: "DNS loopback", targetURL: "http://localhost.example/"},
		{name: "DNS private", targetURL: "http://private.example/"},
		{name: "mixed DNS answers", targetURL: "http://mixed.example/"},
		{name: "DNS NAT64 metadata", targetURL: "http://nat64-metadata.example/"},
		{name: "mixed DNS NAT64 answers", targetURL: "http://nat64-mixed.example/"},
		{name: "empty DNS answers", targetURL: "http://empty.example/"},
		{name: "DNS failure", targetURL: "http://failure.example/"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if parsedURL, err := validateOutboundMetadataURL(context.Background(), test.targetURL, resolver); err == nil {
				t.Fatalf("validateOutboundMetadataURL(%q) = %q, want error", test.targetURL, parsedURL)
			}
		})
	}
}

func TestEmbeddedIPv4FromNAT64WellKnownAddress(t *testing.T) {
	tests := []struct {
		name    string
		address string
		want    string
	}{
		{name: "NAT64 loopback", address: "64:ff9b::7f00:1", want: "127.0.0.1"},
		{name: "NAT64 metadata", address: "64:ff9b::a9fe:a9fe", want: "169.254.169.254"},
		{name: "native IPv6", address: "2001:db8::1"},
		{name: "IPv4", address: "192.0.2.10"},
		{name: "invalid address"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := embeddedIPv4FromNAT64WellKnownAddress(net.ParseIP(test.address))
			if test.want == "" {
				if got != nil {
					t.Fatalf("embeddedIPv4FromNAT64WellKnownAddress(%q) = %s, want nil", test.address, got)
				}
				return
			}
			if got == nil || got.String() != test.want {
				t.Fatalf("embeddedIPv4FromNAT64WellKnownAddress(%q) = %v, want %s", test.address, got, test.want)
			}
		})
	}
}

func TestIsProhibitedMetadataAddressHandlesIPv4IPv6Translation(t *testing.T) {
	tests := []struct {
		name       string
		address    string
		prohibited bool
	}{
		{name: "NAT64 loopback", address: "64:ff9b::7f00:1", prohibited: true},
		{name: "NAT64 private", address: "64:ff9b::a00:1", prohibited: true},
		{name: "NAT64 cloud metadata", address: "64:ff9b::a9fe:a9fe", prohibited: true},
		{name: "NAT64 public", address: "64:ff9b::808:808", prohibited: false},
		{name: "local-use translation", address: "64:ff9b:1::808:808", prohibited: true},
		{name: "IPv4-mapped metadata", address: "::ffff:169.254.169.254", prohibited: true},
		{name: "native public IPv6", address: "2001:db8::1", prohibited: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := isProhibitedMetadataAddress(net.ParseIP(test.address)); got != test.prohibited {
				t.Fatalf("isProhibitedMetadataAddress(%q) = %t, want %t", test.address, got, test.prohibited)
			}
		})
	}
}
