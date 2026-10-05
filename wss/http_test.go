package wss

import (
	"fmt"
	"net/http"
	"testing"
)

func TestNewAPIHTTPClientUsesConfiguredProxy(t *testing.T) {
	const proxyAddress = "http://proxy.example.test:8080"
	client, err := newAPIHTTPClient(proxyAddress)
	if err != nil {
		t.Fatalf("newAPIHTTPClient() error = %v", err)
	}
	request, err := http.NewRequest(http.MethodPost, "https://mend.example.test/api", nil)
	if err != nil {
		t.Fatal(err)
	}
	transport, ok := client.Transport.(*http.Transport)
	if !ok {
		t.Fatalf("transport type = %T, want *http.Transport", client.Transport)
	}
	proxyURL, err := transport.Proxy(request)
	if err != nil {
		t.Fatalf("transport.Proxy() error = %v", err)
	}
	if got := proxyURL.String(); got != proxyAddress {
		t.Fatalf("proxy URL = %q, want %q", got, proxyAddress)
	}
}

func TestNewAPIHTTPClientRejectsInvalidProxy(t *testing.T) {
	tests := []string{
		"proxy.example.test:8080",
		"ftp://proxy.example.test:21",
		"http://",
		fmt.Sprintf("http://%c", 0x7f),
	}
	for _, proxyAddress := range tests {
		if _, err := newAPIHTTPClient(proxyAddress); err == nil {
			t.Errorf("newAPIHTTPClient(%q) accepted invalid proxy", proxyAddress)
		}
	}
}
