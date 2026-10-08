package handlers

import (
	"context"
	"log/slog"
	"net/http"
	"strings"
	"time"

	jsonutils "IDIG4110/shared/json-utils"
)

type Upstream struct {
	Name string
	Url  string
}

// GetHealth pings each upstream's /healthz and reports an aggregate status.
func GetHealth(upstreams []Upstream, timeout time.Duration) http.HandlerFunc {
	type healthResponse struct {
		Status    string          `json:"status"`
		Upstreams map[string]bool `json:"upstreams"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), timeout)
		defer cancel()

		type result struct {
			name string
			ok   bool
		}
		results := make(chan result, len(upstreams))

		client := &http.Client{}
		for _, u := range upstreams {
			go func(u Upstream) {
				req, err := http.NewRequestWithContext(ctx, http.MethodGet, strings.TrimSuffix(u.Url, "/")+"/healthz", nil)
				if err != nil {
					results <- result{u.Name, false}
					return
				}

				resp, err := client.Do(req)
				if err != nil {
					results <- result{u.Name, false}
					return
				}
				defer resp.Body.Close()

				results <- result{u.Name, resp.StatusCode == http.StatusOK}
			}(u)
		}

		statuses := make(map[string]bool, len(upstreams))
		healthy := true
		for range upstreams {
			res := <-results
			statuses[res.name] = res.ok
			if !res.ok {
				healthy = false
			}
		}

		code := http.StatusOK
		status := "ok"
		if !healthy {
			code = http.StatusServiceUnavailable
			status = "unavailable"
		}

		if err := jsonutils.Encode(w, code, healthResponse{Status: status, Upstreams: statuses}); err != nil {
			slog.Error("encoding health response", "error", err)
		}
	}
}
