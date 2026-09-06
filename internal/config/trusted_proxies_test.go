package config

import (
	"bytes"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestSaveConfigPreservesTrustedProxies(t *testing.T) {
	cfg := &Config{}
	cfg.Server.TrustedProxies = []string{"127.0.0.1", "10.0.0.0/8", "2001:db8::/32"}

	var output bytes.Buffer
	if err := SaveConfig(cfg, &output); err != nil {
		t.Fatal(err)
	}

	var decoded Config
	if err := yaml.Unmarshal(output.Bytes(), &decoded); err != nil {
		t.Fatalf("generated config is invalid YAML: %v\n%s", err, output.String())
	}
	if len(decoded.Server.TrustedProxies) != len(cfg.Server.TrustedProxies) {
		t.Fatalf("trusted proxies = %#v, want %#v", decoded.Server.TrustedProxies, cfg.Server.TrustedProxies)
	}
	for i, proxy := range cfg.Server.TrustedProxies {
		if decoded.Server.TrustedProxies[i] != proxy {
			t.Fatalf("trusted proxy %d = %q, want %q", i, decoded.Server.TrustedProxies[i], proxy)
		}
	}
}
