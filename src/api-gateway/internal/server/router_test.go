package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"IDIG4110/api-gateway/internal/config"
)

// A bare prefix must proxy unchanged: upstreams register exact patterns
// such as "GET /api/v1/rules", so a mux-generated redirect to the
// trailing-slash subtree would 404 there. This test fails with 307 when
// only the subtree pattern is registered.
func TestRouterProxiesBarePrefix(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/rules" {
			t.Errorf("forwarded path = %q, want /api/v1/rules", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{}`))
	}))
	defer backend.Close()

	cfg := &config.Config{
		Proxy:     config.ProxyConfig{ResponseHeaderTimeout: 5},
		Upstreams: []config.UpstreamConfig{{Name: "rules", Prefix: "/api/v1/rules", Url: backend.URL}},
	}
	handler, err := NewRouter(cfg)
	if err != nil {
		t.Fatal(err)
	}
	ts := httptest.NewServer(handler)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/api/v1/rules")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200: a bare prefix must proxy, not redirect to the subtree", resp.StatusCode)
	}
}

// The subtree pattern still proxies deeper paths.
func TestRouterProxiesSubtreePaths(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/rules/9d2c8f9a-1111-4222-8333-444455556666" {
			t.Errorf("forwarded path = %q", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer backend.Close()

	cfg := &config.Config{
		Proxy:     config.ProxyConfig{ResponseHeaderTimeout: 5},
		Upstreams: []config.UpstreamConfig{{Name: "rules", Prefix: "/api/v1/rules", Url: backend.URL}},
	}
	handler, err := NewRouter(cfg)
	if err != nil {
		t.Fatal(err)
	}
	ts := httptest.NewServer(handler)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/api/v1/rules/9d2c8f9a-1111-4222-8333-444455556666")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}
