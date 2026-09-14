package wss

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

// ConfigureAPIProxy configures the HTTP client used by Mend API requests.
// An empty proxy keeps Go's standard HTTP_PROXY, HTTPS_PROXY, and NO_PROXY
// environment-variable behavior.
func ConfigureAPIProxy(proxyAddress string) error {
	client, err := newAPIHTTPClient(proxyAddress)
	if err != nil {
		return err
	}
	apiHTTPClient.CloseIdleConnections()
	apiHTTPClient = client
	return nil
}

func newAPIHTTPClient(proxyAddress string) (*http.Client, error) {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	proxyAddress = strings.TrimSpace(proxyAddress)
	if proxyAddress != "" {
		proxyURL, err := parseProxyURL(proxyAddress)
		if err != nil {
			return nil, err
		}
		transport.Proxy = http.ProxyURL(proxyURL)
	}
	return &http.Client{Transport: transport, Timeout: httpTimeout}, nil
}

func parseProxyURL(proxyAddress string) (*url.URL, error) {
	proxyURL, err := url.Parse(proxyAddress)
	if err != nil || proxyURL.Host == "" {
		return nil, fmt.Errorf("proxy must be a valid URL with a host")
	}
	switch strings.ToLower(proxyURL.Scheme) {
	case "http", "https", "socks5", "socks5h":
	default:
		return nil, fmt.Errorf("proxy scheme must be http, https, socks5, or socks5h")
	}
	return proxyURL, nil
}
