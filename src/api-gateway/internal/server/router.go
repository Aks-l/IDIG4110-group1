package server

import (
	"fmt"
	"net/http"

	"IDIG4110/api-gateway/internal/config"
	"IDIG4110/api-gateway/internal/handlers"
	"IDIG4110/api-gateway/internal/middleware"
	"IDIG4110/api-gateway/internal/proxy"
)

func NewRouter(cfg *config.Config) (http.Handler, error) {
	mux := http.NewServeMux()

	healthUpstreams := make([]handlers.Upstream, 0, len(cfg.Upstreams))
	for _, u := range cfg.Upstreams {
		p, err := proxy.New(u.Url)
		if err != nil {
			return nil, fmt.Errorf("upstream %s: %w", u.Name, err)
		}
		mux.Handle(u.Prefix+"/", p)
		healthUpstreams = append(healthUpstreams, handlers.Upstream{Name: u.Name, Url: u.Url})
	}

	mux.HandleFunc("GET "+HEALTHZ, handlers.GetHealth(healthUpstreams))

	return middleware.Chain(mux, middleware.Recovery(), middleware.CORS(), middleware.Logging()), nil
}
