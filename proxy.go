package eavesdrop

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/fatih/color"
	"github.com/hashicorp/go-retryablehttp"
)

const (
	DefaultAppPort   uint16 = 8000
	DefaultProxyPort uint16 = 8001
)

const SSE_SCRIPT = `eventSource = new EventSource("/eavesdrop_sse");

eventSource.onmessage = (event) => {
	if (event.data === "refresh") window.location.reload();
}

eventSource.onerror = (error) => console.error("eavesdrop sse error:", error);`

type Proxy interface {
	RefreshBrowser()
}

type proxy struct {
	ctx         context.Context
	appPort     uint16
	proxyPort   uint16
	client      *retryablehttp.Client
	mu          sync.Mutex
	subscribers map[chan struct{}]struct{}
}

type ProxyOption func(*proxy)

func WithAppPort(port uint16) ProxyOption {
	return func(p *proxy) {
		if port > 0 {
			if port == p.proxyPort {
				panic("app and proxy port must be different")
			}
			p.appPort = port
		}
	}
}

func WithProxyPort(port uint16) ProxyOption {
	return func(p *proxy) {
		if port > 0 {
			if port == p.appPort {
				panic("app and proxy port must be different")
			}
			p.proxyPort = port
		}
	}
}

// NewProxy returns an instance of the default Proxy implementation.
func NewProxy(ctx context.Context, opts ...ProxyOption) *proxy {
	proxy := &proxy{
		ctx:         ctx,
		appPort:     DefaultAppPort,
		proxyPort:   DefaultProxyPort,
		client:      retryablehttp.NewClient(),
		subscribers: make(map[chan struct{}]struct{}),
	}
	proxy.client.HTTPClient.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}
	proxy.client.Logger = nil // TODO: might support custom logger another time

	for _, opt := range opts {
		opt(proxy)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", proxy.handleClientRequest)
	mux.HandleFunc("/eavesdrop_sse", proxy.handleSSE)

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", proxy.proxyPort),
		Handler: mux,
	}

	go func() {
		err := server.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			panic(err)
		}
	}()

	go func() {
		<-ctx.Done()
		err := server.Shutdown(ctx)
		if err != nil {
			color.Red("proxy shutdown error: %v", err)
			server.Close()
		}
	}()

	return proxy
}

// RefreshBrowser broadcasts a refresh signal to all connected client browsers.
func (p *proxy) RefreshBrowser() {
	p.mu.Lock()
	defer p.mu.Unlock()

	for subscriber := range p.subscribers {
		subscriber <- struct{}{}
	}
}

// handleClientRequest proxies client requests to the application and returns the response back to the client.
func (p *proxy) handleClientRequest(w http.ResponseWriter, r *http.Request) {
	req, _ := retryablehttp.NewRequest(
		r.Method,
		fmt.Sprintf("http://127.0.0.1:%d%s", p.appPort, r.URL.RequestURI()),
		r.Body,
	)

	req.Header = r.Header.Clone()
	req.Header.Set("X-Forwarded-For", r.RemoteAddr)
	req.Header.Set("Via", fmt.Sprintf("%s %s", r.Proto, r.Host))

	resp, err := p.client.Do(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		color.Red("proxy error: %v", err)
		return
	}
	defer resp.Body.Close()

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Via", fmt.Sprintf("%s %s", r.Proto, r.Host))
	for key, values := range resp.Header {
		for _, value := range values {
			if key == "Content-Length" {
				continue
			}
			w.Header().Add(key, value)
		}
	}

	if !strings.Contains(resp.Header.Get("Content-Type"), "text/html") {
		w.WriteHeader(resp.StatusCode)
		w.Header().Set("Content-Length", resp.Header.Get("Content-Length"))
		_, err := io.Copy(w, resp.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			color.Red("proxy error: %v", err)
			return
		}
	}

	body, err := p.injectSSE(resp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		color.Red("proxy error: %v", err)
		return
	}

	w.Header().Set("Content-Length", strconv.Itoa(len(body)))
	// w.WriteHeader(resp.StatusCode)
	_, err = w.Write(body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		color.Red("proxy error: %v", err)
		return
	}
}

// handleSSE sets up and handles persistent SSE connections.
func (p *proxy) handleSSE(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusNotAcceptable)
		color.Red("proxy error: client does not support streaming")
		return
	}

	fmt.Fprint(w, "data: connected\n\n")
	flusher.Flush()

	subscriber := make(chan struct{}, 1)
	p.mu.Lock()
	p.subscribers[subscriber] = struct{}{}
	p.mu.Unlock()

	for {
		select {
		case <-r.Context().Done():
			p.mu.Lock()
			delete(p.subscribers, subscriber)
			close(subscriber)
			p.mu.Unlock()
			return

		case <-subscriber:
			_, err := fmt.Fprint(w, "data: refresh\n\n")
			if err != nil {
				color.Red("proxy error: %v", err)
			}
			flusher.Flush()
		}
	}
}

func (p *proxy) injectSSE(resp *http.Response) ([]byte, error) {
	var buf bytes.Buffer
	_, err := buf.ReadFrom(resp.Body)
	if err != nil {
		return nil, err
	}

	idx := strings.LastIndex(buf.String(), "</body>")
	if idx == -1 {
		return buf.Bytes(), nil
	}

	page := buf.String()
	return fmt.Appendf([]byte{}, "%s<script>%s</script>%s", page[:idx], SSE_SCRIPT, page[idx:]), nil
}
