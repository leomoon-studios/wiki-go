package handlers

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"
)

type metadataHostnameResolver interface {
	LookupIPAddr(context.Context, string) ([]net.IPAddr, error)
}

var metadataProhibitedNetworks = []*net.IPNet{
	mustParseMetadataNetwork("169.254.0.0/16"),
	mustParseMetadataNetwork("100.64.0.0/10"),
	mustParseMetadataNetwork("168.63.129.16/32"),
	mustParseMetadataNetwork("fd00:ec2::254/128"),
}

// validateOutboundMetadataURL validates the initial destination before a
// metadata request is created. Connection-time and redirect validation reuse
// this policy in the metadata HTTP client.
func validateOutboundMetadataURL(ctx context.Context, rawURL string, resolver metadataHostnameResolver) (*url.URL, error) {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("parse metadata URL: %w", err)
	}
	if !parsedURL.IsAbs() || (parsedURL.Scheme != "http" && parsedURL.Scheme != "https") {
		return nil, fmt.Errorf("metadata URL must use HTTP or HTTPS")
	}
	if parsedURL.User != nil {
		return nil, fmt.Errorf("metadata URL must not contain credentials")
	}

	hostname := parsedURL.Hostname()
	if hostname == "" {
		return nil, fmt.Errorf("metadata URL hostname is empty")
	}
	if port := parsedURL.Port(); port != "" {
		portNumber, err := strconv.Atoi(port)
		if err != nil || portNumber < 1 || portNumber > 65535 {
			return nil, fmt.Errorf("metadata URL port is invalid")
		}
	} else if strings.HasSuffix(parsedURL.Host, ":") {
		return nil, fmt.Errorf("metadata URL port is empty")
	}

	if _, err := resolveAllowedMetadataAddresses(ctx, hostname, resolver); err != nil {
		return nil, err
	}

	return parsedURL, nil
}

func resolveAllowedMetadataAddresses(ctx context.Context, hostname string, resolver metadataHostnameResolver) ([]net.IPAddr, error) {
	if literalIP := net.ParseIP(hostname); literalIP != nil {
		if isProhibitedMetadataAddress(literalIP) {
			return nil, fmt.Errorf("metadata URL resolves to a prohibited address")
		}
		return []net.IPAddr{{IP: literalIP}}, nil
	}
	if !isValidMetadataHostname(hostname) {
		return nil, fmt.Errorf("metadata URL hostname is malformed")
	}
	if resolver == nil {
		return nil, fmt.Errorf("metadata hostname resolver is unavailable")
	}

	addresses, err := resolver.LookupIPAddr(ctx, hostname)
	if err != nil {
		return nil, fmt.Errorf("resolve metadata URL hostname: %w", err)
	}
	if len(addresses) == 0 {
		return nil, fmt.Errorf("metadata URL hostname has no addresses")
	}
	for _, address := range addresses {
		if address.Zone != "" || isProhibitedMetadataAddress(address.IP) {
			return nil, fmt.Errorf("metadata URL resolves to a prohibited address")
		}
	}
	return addresses, nil
}

func isProhibitedMetadataAddress(ip net.IP) bool {
	if ip == nil || !ip.IsGlobalUnicast() || ip.IsUnspecified() || ip.IsLoopback() || ip.IsMulticast() || ip.IsInterfaceLocalMulticast() || ip.IsLinkLocalMulticast() || ip.IsLinkLocalUnicast() || ip.IsPrivate() {
		return true
	}
	for _, prohibitedNetwork := range metadataProhibitedNetworks {
		if prohibitedNetwork.Contains(ip) {
			return true
		}
	}
	return false
}

func isValidMetadataHostname(hostname string) bool {
	hostname = strings.TrimSuffix(hostname, ".")
	if hostname == "" || len(hostname) > 253 {
		return false
	}
	for _, label := range strings.Split(hostname, ".") {
		if label == "" || len(label) > 63 || label[0] == '-' || label[len(label)-1] == '-' {
			return false
		}
		for _, character := range label {
			if character >= 'a' && character <= 'z' || character >= 'A' && character <= 'Z' || character >= '0' && character <= '9' || character == '-' {
				continue
			}
			return false
		}
	}
	return true
}

func mustParseMetadataNetwork(value string) *net.IPNet {
	_, network, err := net.ParseCIDR(value)
	if err != nil {
		panic(err)
	}
	return network
}
