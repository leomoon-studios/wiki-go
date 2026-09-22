package handlers

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"wiki-go/internal/auth"
	"wiki-go/internal/config"
)

func TestEditorCanFetchMetadataFromAllowedPublicDestination(t *testing.T) {
	testConfig := installUserSessionTestConfig(t, nil)
	editorCookie := sessionCookieForUser(t, testConfig, "metadata-editor", config.RoleEditor)
	resolver := staticMetadataResolver{addresses: map[string][]net.IPAddr{
		"public.example": {{IP: net.ParseIP("192.0.2.70")}},
	}}
	var requests atomic.Int32
	dialer := metadataDialFunc(func(_ context.Context, _, address string) (net.Conn, error) {
		if address != "192.0.2.70:8080" {
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
			requests.Add(1)
			body := `<html><head><title>Controlled public page</title><meta name="description" content="Allowed metadata response"></head></html>`
			fmt.Fprintf(serverConnection, "HTTP/1.1 200 OK\r\nContent-Type: text/html; charset=utf-8\r\nContent-Length: %d\r\nConnection: close\r\n\r\n%s", len(body), body)
		}()
		return clientConnection, nil
	})

	request := metadataHandlerRequest("http://public.example:8080/article")
	request.AddCookie(editorCookie)
	response := httptest.NewRecorder()
	handler := requireEditorForMetadataTest(func(w http.ResponseWriter, r *http.Request) {
		fetchMetadataHandler(w, r, resolver, newMetadataHTTPClient(resolver, dialer))
	})
	handler(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body: %s", response.Code, response.Body.String())
	}
	var result MetadataResponse
	if err := json.Unmarshal(response.Body.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if !result.Success || result.Title != "Controlled public page" || result.Description != "Allowed metadata response" {
		t.Fatalf("unexpected metadata response: %+v", result)
	}
	if requests.Load() != 1 {
		t.Fatalf("controlled server requests = %d, want 1", requests.Load())
	}
}

func TestMetadataHandlerRejectsProhibitedDestinationsWithoutDialing(t *testing.T) {
	tests := []struct {
		name      string
		targetURL string
	}{
		{name: "loopback", targetURL: "http://127.0.0.1/admin"},
		{name: "private network", targetURL: "http://10.20.30.40/admin"},
		{name: "metadata service", targetURL: "http://169.254.169.254/latest/meta-data/"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var dialCalls atomic.Int32
			dialer := metadataDialFunc(func(context.Context, string, string) (net.Conn, error) {
				dialCalls.Add(1)
				return nil, fmt.Errorf("prohibited destination was dialed")
			})
			resolver := staticMetadataResolver{}
			response := httptest.NewRecorder()
			fetchMetadataHandler(response, metadataHandlerRequest(test.targetURL), resolver, newMetadataHTTPClient(resolver, dialer))

			if response.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want 400; body: %s", response.Code, response.Body.String())
			}
			if dialCalls.Load() != 0 {
				t.Fatalf("underlying dialer was called %d times", dialCalls.Load())
			}
			if !strings.Contains(response.Body.String(), "URL destination is not allowed") {
				t.Fatalf("response did not contain the generic validation failure: %s", response.Body.String())
			}
		})
	}
}

func TestMetadataHandlerDoesNotDiscloseConnectionFailure(t *testing.T) {
	logOutput := captureHandlerLogs(t)
	resolver := staticMetadataResolver{addresses: map[string][]net.IPAddr{
		"internal-name.example": {{IP: net.ParseIP("192.0.2.80")}},
	}}
	dialer := metadataDialFunc(func(context.Context, string, string) (net.Conn, error) {
		return nil, fmt.Errorf("dial tcp 192.0.2.80:8443: connection refused")
	})
	response := httptest.NewRecorder()
	fetchMetadataHandler(
		response,
		metadataHandlerRequest("http://internal-name.example:8443/private-service?token=QUERY-SECRET"),
		resolver,
		newMetadataHTTPClient(resolver, dialer),
	)

	if response.Code != http.StatusBadGateway {
		t.Fatalf("status = %d, want 502; body: %s", response.Code, response.Body.String())
	}
	body := response.Body.String()
	if !strings.Contains(body, "Failed to fetch metadata") {
		t.Fatalf("response did not contain the generic fetch failure: %s", body)
	}
	for _, secret := range []string{"internal-name.example", "192.0.2.80", "8443", "private-service", "dial tcp", "connection refused"} {
		if strings.Contains(body, secret) {
			t.Errorf("response disclosed %q: %s", secret, body)
		}
	}
	if strings.Contains(logOutput.String(), "QUERY-SECRET") || strings.Contains(logOutput.String(), "token=") {
		t.Fatalf("log output disclosed URL query credentials: %q", logOutput.String())
	}
}

func metadataHandlerRequest(targetURL string) *http.Request {
	body := fmt.Sprintf(`{"url":%q}`, targetURL)
	request := httptest.NewRequest(http.MethodPost, "/api/links/fetch-metadata", strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	return request
}

func requireEditorForMetadataTest(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !auth.RequireRole(r, config.RoleEditor) {
			http.Error(w, "editor access required", http.StatusForbidden)
			return
		}
		next(w, r)
	}
}
