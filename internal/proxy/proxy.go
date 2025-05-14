package proxy

import (
	"bytes"
	_ "embed"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/dimmerz92/eavesdrop/internal/config"
)

//go:embed refresh.js
var SSE_SCRIPT string

type Proxy struct {
	AppPort     int
	ProxyPort   int
	Client      *http.Client
	Server      *http.Server
	SubMu       *sync.Mutex
	Subscribers map[chan struct{}]struct{}
}

// NewProxy returns a newly instantiated Proxy server.
func NewProxy(cfg *config.Config) *Proxy {
	proxy := &Proxy{
		AppPort:     cfg.AppPort,
		ProxyPort:   cfg.ProxyPort,
		Client:      &http.Client{},
		SubMu:       &sync.Mutex{},
		Subscribers: make(map[chan struct{}]struct{}),
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", proxy.proxyRequest)
	mux.HandleFunc("/eavesdrop_sse", proxy.refresher)

	proxy.Server = &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.ProxyPort),
		Handler: mux,
	}

	return proxy
}

// proxyRequest forwards the request to the target app server, injects a server
// sent event script for automatic browser refreshing.
func (p *Proxy) proxyRequest(w http.ResponseWriter, r *http.Request) {
	// create a forwarding request
	req, err := http.NewRequest(
		r.Method,
		fmt.Sprintf("http://localhost:%d%s", p.AppPort, r.URL.Path),
		r.Body,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// copy headers and set proxy headers
	req.Header = r.Header.Clone()
	req.Header.Set("X-Forwarded-For", r.RemoteAddr)
	req.Header.Set("Via", fmt.Sprintf("%s %s", r.Proto, r.Host))

	// forward the request
	var resp *http.Response
	for i := 0; i < 10; i++ {
		resp, err = p.Client.Do(req)
		if err == nil {
			break
		}

		time.Sleep(100 * time.Millisecond)
	}
	if err != nil {
		http.Error(w, "proxy error: app unresponsive", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	for k, v := range resp.Header {
		for _, vv := range v {
			if k == "Content-Length" {
				continue
			}
			w.Header().Add(k, vv)
		}
	}
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Via", fmt.Sprintf("%s %s", r.Proto, r.Host))
	w.WriteHeader(resp.StatusCode)

	if !strings.Contains(resp.Header.Get("Content-Type"), "text/html") {
		w.Header().Set("Content-Length", resp.Header.Get("Content-Length"))
		io.Copy(w, resp.Body)
	} else {
		body, err := p.injectRefresher(resp)
		if err != nil {
			http.Error(
				w,
				fmt.Sprintf("proxy error: %v", err),
				http.StatusInternalServerError,
			)
			return
		}
		w.Header().Set("Content-Length", strconv.Itoa(len([]byte(body))))
		_, err = io.WriteString(w, body)
		if err != nil {
			http.Error(
				w,
				fmt.Sprintf("proxy error: %v", err),
				http.StatusInternalServerError,
			)
			return
		}
	}
}

// injectRefresher injects the sse script if the body exists.
func (p *Proxy) injectRefresher(resp *http.Response) (string, error) {
	buf := new(bytes.Buffer)
	_, err := buf.ReadFrom(resp.Body)
	if err != nil {
		return "", fmt.Errorf("proxy error: failed to read body")
	}
	page := buf.String()

	// get the index of the body closing tag
	body := strings.LastIndex(page, "</body>")
	if body == -1 {
		return page, nil
	}

	script := page[:body] + "<script>" + SSE_SCRIPT + "</script>" + page[body:]

	return script, nil
}

// refresher handles the sse refresh events to the browser.
func (p *Proxy) refresher(w http.ResponseWriter, r *http.Request) {
	// set headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	// create the response flusher
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(
			w,
			"proxy error: streaming unsupported",
			http.StatusInternalServerError,
		)
		return
	}

	// create a subscriber channel for refresh signals
	ch := make(chan struct{}, 1)
	p.SubMu.Lock()
	p.Subscribers[ch] = struct{}{}
	p.SubMu.Unlock()

	// cleanup on return
	defer func() {
		p.SubMu.Lock()
		delete(p.Subscribers, ch)
		p.SubMu.Unlock()
	}()

	for {
		select {
		// subscriber disconnect
		case <-r.Context().Done():
			return

		// send refresh event
		case <-ch:
			fmt.Fprint(w, "data: refresh\n\n")
			flusher.Flush()
		}
	}
}

// Refresh broadcasts a signal to refresh to all proxy subscribers.
func (p *Proxy) Refresh() {
	p.SubMu.Lock()
	defer p.SubMu.Unlock()

	for sub := range p.Subscribers {
		select {
		// push refresh signal
		case sub <- struct{}{}:

		// skip if subscriber channel is full
		default:
		}
	}
}
