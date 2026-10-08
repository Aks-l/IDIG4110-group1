// Package proxy forwards requests to a backend service and returns
// JSON errors instead of the reverse proxy's blank default responses.
package proxy

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"time"

	"IDIG4110/shared/httperror"
)

func New(target string, responseHeaderTimeout time.Duration) (http.Handler, error) {
	u, err := url.Parse(target)
	if err != nil {
		return nil, fmt.Errorf("parsing upstream url %q: %w", target, err)
	}

	rp := &httputil.ReverseProxy{
		Rewrite: func(pr *httputil.ProxyRequest) {
			pr.SetURL(u)
			pr.SetXForwarded()
		},
		Transport: &http.Transport{
			ResponseHeaderTimeout: responseHeaderTimeout,
		},
		ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
			if os.IsTimeout(err) {
				httperror.HandleError(w, http.StatusGatewayTimeout, err, httperror.ErrGatewayTimeout)
				return
			}
			httperror.HandleError(w, http.StatusBadGateway, err, httperror.ErrBadGateway)
		},
	}

	return rp, nil
}
