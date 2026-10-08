package server

import (
	"fmt"
	"net/http"
	"time"

	"IDIG4110/api-gateway/internal/config"
	"IDIG4110/api-gateway/internal/handlers"
	"IDIG4110/api-gateway/internal/proxy"
	"IDIG4110/shared/middleware"
)

func NewRouter(cfg *config.Config) (http.Handler, error) {
	mux := http.NewServeMux()

	healthUpstreams := make([]handlers.Upstream, 0, len(cfg.Upstreams))
	for _, u := range cfg.Upstreams {
		p, err := proxy.New(u.Url, time.Duration(cfg.Proxy.ResponseHeaderTimeout)*time.Second)
		if err != nil {
			return nil, fmt.Errorf("upstream %s: %w", u.Name, err)
		}
		mux.Handle(u.Prefix, p)
		mux.Handle(u.Prefix+"/", p)
		healthUpstreams = append(healthUpstreams, handlers.Upstream{Name: u.Name, Url: u.Url})
	}

	mux.HandleFunc("GET "+HEALTHZ, handlers.GetHealth(healthUpstreams, time.Duration(cfg.Health.Timeout)*time.Second))

	return middleware.Chain(mux, middleware.Recovery(), middleware.CORS(), middleware.Logging()), nil
}
