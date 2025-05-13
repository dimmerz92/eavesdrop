package proxy_test

import (
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/dimmerz92/eavesdrop/internal/config"
	"github.com/dimmerz92/eavesdrop/internal/proxy"
)

const (
	html = "<!DOCTYPE html><html><head><title>test</title><body></body></html>"
)

func makeServer() *http.Server {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte(html))
	})

	mux.HandleFunc("POST /", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("posted"))
	})

	return &http.Server{
		Addr:    ":8000",
		Handler: mux,
	}
}

func TestProxy_ProxyRequest(t *testing.T) {
	origin := makeServer()
	originErrChan := make(chan error, 1)

	go func() {
		if err := origin.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			originErrChan <- err
		}
	}()
	defer origin.Close()

	time.Sleep(100 * time.Millisecond)

	cfg := config.DefaultConfig("")
	cfg.Proxy = true

	p := proxy.NewProxy(cfg)
	proxyErrChan := make(chan error, 1)

	go func() {
		if err := p.Server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			proxyErrChan <- err
		}
	}()
	defer p.Server.Close()

	time.Sleep(100 * time.Millisecond)

	select {
	case err := <-originErrChan:
		t.Fatalf("origin server failed: %v", err)
	case err := <-proxyErrChan:
		t.Fatalf("proxy server failed: %v", err)
	default:
	}

	t.Run("test successful GET through proxy", func(t *testing.T) {
		resp, err := http.Get("http://localhost:8001")
		if err != nil {
			t.Fatalf("proxy request failed: %v", err)
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("failed to read response body: %v", err)
		}

		if string(body) != html {
			t.Errorf("expected: %s\ngot: %s", html, string(body))
		}
	})

	t.Run("test successful POST through proxy", func(t *testing.T) {
		resp, err := http.Post("http://localhost:8001", "text/plain", nil)
		if err != nil {
			t.Fatalf("proxy request failed: %v", err)
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("failed to read response body: %v", err)
		}

		if string(body) != "posted" {
			t.Errorf("expected: %s\ngot: %s", "posted", string(body))
		}
	})

	t.Run("test PUT not allowed through proxy", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodPut, "http://localhost:8001", nil)
		if err != nil {
			t.Fatalf("creating request failed: %v", err)
		}

		client := http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("proxy request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusMethodNotAllowed {
			t.Errorf("expected status: %d got: %d", http.StatusMethodNotAllowed, resp.StatusCode)
		}
	})
}
