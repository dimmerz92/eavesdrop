package proxy

import (
	"fmt"
	"io"
	"net/http"

	"github.com/dimmerz92/eavesdrop/internal/config"
)

type Proxy struct {
	AppPort   int
	ProxyPort int
	Client    *http.Client
	Server    *http.Server
}

// NewProxy returns a newly instantiated Proxy server.
func NewProxy(cfg *config.Config) *Proxy {
	proxy := &Proxy{
		AppPort:   cfg.AppPort,
		ProxyPort: cfg.ProxyPort,
		Client:    &http.Client{},
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", proxy.proxyRequest)

	proxy.Server = &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.ProxyPort),
		Handler: mux,
	}

	return proxy
}

// proxyRequest forwards the request to the target app server, injects a server sent event script for automatic
// browser refreshing.
func (p *Proxy) proxyRequest(w http.ResponseWriter, r *http.Request) {
	req, err := http.NewRequest(r.Method, fmt.Sprintf("http://localhost:%d", p.AppPort), r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	req.Header = r.Header.Clone()
	resp, err := p.Client.Do(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	for k, v := range resp.Header {
		for _, vv := range v {
			w.Header().Add(k, vv)
		}
	}

	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}
