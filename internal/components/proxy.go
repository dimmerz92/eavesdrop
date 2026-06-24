package components

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/fatih/color"
	"github.com/hashicorp/go-retryablehttp"
)

const SSE_SCRIPT = `eventSource = new EventSource("/eavesdrop_sse");

eventSource.onmessage = (event) => {
	if (event.data === "refresh") window.location.reload();
}

eventSource.onerror = (error) => console.error("eavesdrop sse error:", error);`

type Proxy struct {
	ctx         context.Context
	appPort     uint16
	proxyPort   uint16
	mu          sync.Mutex
	subscribers map[chan struct{}]struct{}
}

func NewProxy(ctx context.Context, appPort, proxyPort uint16) (*Proxy, error) {
	if appPort == 0 || proxyPort == 0 || appPort == proxyPort {
		return nil, fmt.Errorf("app and proxy port must be non-zero and different")
	}

	p := &Proxy{
		ctx:         ctx,
		appPort:     appPort,
		proxyPort:   proxyPort,
		subscribers: make(map[chan struct{}]struct{}),
	}

	retryClient := retryablehttp.NewClient()
	retryClient.Logger = nil

	target, _ := url.Parse(fmt.Sprintf("http://127.0.0.1:%d", p.appPort))
	rp := &httputil.ReverseProxy{
		Rewrite: func(r *httputil.ProxyRequest) {
			r.SetURL(target)
			r.SetXForwarded()
			r.Out.RequestURI = "" // retryablehttp uses http.Client.Do which rejects non-empty RequestURI
		},
		Transport:      &retryablehttp.RoundTripper{Client: retryClient},
		ModifyResponse: p.injectSSE,
		ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
			color.Red("proxy error: %v", err)
			http.Error(w, err.Error(), http.StatusBadGateway)
		},
	}

	mux := http.NewServeMux()
	mux.Handle("/", rp)
	mux.HandleFunc("/eavesdrop_sse", p.handleSSE)

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", p.proxyPort),
		Handler: mux,
	}

	ln, err := net.Listen("tcp", server.Addr)
	if err != nil {
		panic(err)
	}

	go func() {
		if err := server.Serve(ln); err != nil && !errors.Is(err, http.ErrServerClosed) {
			panic(err)
		}
	}()

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			color.Red("proxy shutdown error: %v", err)
			server.Close()
		}
	}()

	return p, nil
}

func (p *Proxy) RefreshBrowser() {
	p.mu.Lock()
	defer p.mu.Unlock()
	for subscriber := range p.subscribers {
		select {
		case subscriber <- struct{}{}:
		default:
		}
	}
}

func (p *Proxy) handleSSE(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusNotAcceptable)
		return
	}

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
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
			fmt.Fprint(w, "data: refresh\n\n")
			flusher.Flush()
		}
	}
}

func (p *Proxy) injectSSE(resp *http.Response) error {
	if !strings.Contains(resp.Header.Get("Content-Type"), "text/html") {
		return nil
	}

	body, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return err
	}

	if idx := strings.LastIndex(string(body), "</body>"); idx != -1 {
		body = fmt.Appendf(nil, "%s<script>%s</script>%s", body[:idx], SSE_SCRIPT, body[idx:])
	}

	resp.Body = io.NopCloser(bytes.NewReader(body))
	resp.ContentLength = int64(len(body))
	resp.Header.Set("Content-Length", fmt.Sprintf("%d", len(body)))

	return nil
}
