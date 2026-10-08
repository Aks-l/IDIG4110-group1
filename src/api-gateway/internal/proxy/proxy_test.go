package proxy

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestProxyForwardsRequest(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/ingest/sensors" {
			t.Errorf("unexpected forwarded path: %s", r.URL.Path)
		}
		if r.Header.Get("X-Forwarded-For") == "" {
			t.Error("X-Forwarded-For not set")
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`[{"id":"1"}]`))
	}))
	defer backend.Close()

	p, err := New(backend.URL, 30*time.Second)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	ts := httptest.NewServer(p)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/api/v1/ingest/sensors")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want 200", resp.StatusCode)
	}

	var body []map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if len(body) != 1 || body[0]["id"] != "1" {
		t.Errorf("unexpected body: %v", body)
	}
}

func TestProxyReturnsJSONErrorOnUnreachableUpstream(t *testing.T) {
	p, err := New("http://127.0.0.1:1", 30*time.Second)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	ts := httptest.NewServer(p)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/api/v1/rules")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadGateway {
		t.Errorf("status = %d, want 502", resp.StatusCode)
	}

	var body map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if body["code"] != float64(http.StatusBadGateway) {
		t.Errorf("unexpected error body: %v", body)
	}
}
