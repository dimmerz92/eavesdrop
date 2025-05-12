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

func NewProxy(cfg config.Config) *Proxy {
	mux := http.NewServeMux()

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.ProxyPort),
		Handler: mux,
	}

	return &Proxy{
		AppPort:   cfg.AppPort,
		ProxyPort: cfg.ProxyPort,
		Client:    &http.Client{},
		Server:    server,
	}
}
