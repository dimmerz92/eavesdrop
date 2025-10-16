package eavesdrop

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/fatih/color"
)

const (
	MIN_PORT      = 2000
	MAX_PORT      = 65535
	BACKOFF_RETRY = 100
	CLOSE_DELAY   = 5
)

//go:embed proxy.js
var SSE_SCRIPT string

type ProxyConfig struct {
	Enabled   bool `json:"enabled" toml:"enabled" yaml:"enabled"`
	AppPort   int  `json:"app_port" toml:"app_port" yaml:"app_port"`
	ProxyPort int  `json:"proxy_port" toml:"proxy_port" yaml:"proxy_port"`
}

// Validate checks to make sure the ProxyConfig fields are valid.
func (p *ProxyConfig) Validate() error {
	if p.Enabled {
		if p.AppPort < MIN_PORT || p.AppPort > MAX_PORT {
			return fmt.Errorf("app_port must be between %d and %d", MIN_PORT, MAX_PORT)
		}

		if p.ProxyPort < MIN_PORT || p.ProxyPort > MAX_PORT {
			return fmt.Errorf("proxy_port must be between %d and %d", MIN_PORT, MAX_PORT)
		}

		if p.AppPort == p.ProxyPort {
			return fmt.Errorf("app_port and proxy_port must be different")
		}
	}

	return nil
}

// ToProxy returns a newly instantiated Proxy server for browser reloading.
func (p *ProxyConfig) ToProxy() *Proxy {
	if !p.Enabled {
		return nil
	}

	proxy := &Proxy{
		AppPort:   fmt.Sprintf(":%d", p.AppPort),
		ProxyPort: fmt.Sprintf(":%d", p.ProxyPort),
		Client: &http.Client{CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}},
		Subscribers:   make(map[chan struct{}]struct{}),
		SubscribersMu: &sync.Mutex{},
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", proxy.HandleRequest)
	mux.HandleFunc("/eavesdrop_sse", proxy.ClientEvent)

	proxy.Server = &http.Server{
		ReadHeaderTimeout: 0,
		Addr:              proxy.ProxyPort,
		Handler:           mux,
	}

	return proxy
}

type Proxy struct {
	AppPort       string
	ProxyPort     string
	Client        *http.Client
	Server        *http.Server
	Subscribers   map[chan struct{}]struct{}
	SubscribersMu *sync.Mutex
}

// HandleRequest sends the incoming request to the target app server and returns the response.
// If the response contains valid HTML, the SSE script will be injected prior to returning the response to the browser.
func (p *Proxy) HandleRequest(w http.ResponseWriter, r *http.Request) {
	// reconstruct the path with query params if found.
	path := r.URL.Path
	params := r.URL.RawQuery
	if params != "" {
		path = fmt.Sprintf("%s?%s", path, params)
	}

	// create a forwarding request and clone headers.
	req, err := http.NewRequest(r.Method, fmt.Sprintf("127.0.0.1%s%s", p.AppPort, path), r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		color.Red("proxy error: %v", err)
		return
	}
	req.Header = r.Header.Clone()
	req.Header.Set("X-Forwarded-For", r.RemoteAddr)
	req.Header.Set("Via", fmt.Sprintf("%s %s", r.Proto, r.Host))

	// forward the request to the app.
	var resp *http.Response
	for range 10 {
		resp, err = p.Client.Do(req)
		if err == nil {
			break
		}
		time.Sleep(BACKOFF_RETRY * time.Millisecond) // constant time backoff retry.
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		color.Red("proxy error: %v", err)
		return
	}

	defer resp.Body.Close()

	// add and write response headers to the response writer.
	for key, values := range resp.Header {
		for _, value := range values {
			if key == "Content-Length" {
				continue
			}
			w.Header().Add(key, value)
		}
	}
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Via", fmt.Sprintf("%s %s", r.Proto, r.Host))
	w.WriteHeader(resp.StatusCode)

	// write the response body.
	if !strings.Contains(resp.Header.Get("Content-Type"), "text/html") {
		w.Header().Set("Content-Length", resp.Header.Get("Content-Length"))
		_, err := io.Copy(w, req.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			color.Red("proxy error: %v", err)
			return
		}
	} else {
		body, err := p.InjectSSE(resp)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			color.Red("proxy error: %v", err)
			return
		}

		w.Header().Set("Content-Length", strconv.Itoa(len(body)))
		_, err = w.Write(body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			color.Red("proxy error: %v", err)
			return
		}
	}
}

// InjectSSE prepends the SSE script before the last </body> tag if it exists.
func (p *Proxy) InjectSSE(resp *http.Response) ([]byte, error) {
	buf := new(bytes.Buffer)

	_, err := buf.ReadFrom(resp.Body)
	if err != nil {
		return nil, err
	}

	page := buf.String()

	bodyIdx := strings.LastIndex(page, "</body>")
	if bodyIdx == -1 {
		return buf.Bytes(), nil
	}

	return fmt.Appendf([]byte{}, "%s<script>%s</script>%s", page[:bodyIdx], SSE_SCRIPT, page[bodyIdx:]), nil
}

// ClientEvent handles the SSE refresh events to and from the browser.
func (p *Proxy) ClientEvent(w http.ResponseWriter, r *http.Request) {
	// set SSE stream headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	// add a subscriber channel for refresh signals.
	sub := make(chan struct{}, 1)
	p.SubscribersMu.Lock()
	p.Subscribers[sub] = struct{}{}
	p.SubscribersMu.Unlock()

	// cleanup on return.
	defer func() {
		p.SubscribersMu.Lock()
		delete(p.Subscribers, sub)
		p.SubscribersMu.Unlock()
	}()

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "proxy error: streaming unsupported", http.StatusInternalServerError)
		color.Red("proxy error: streaming unsupported")
		return
	}

	for {
		select {
		case <-r.Context().Done():
			return

		case <-sub:
			_, err := fmt.Fprint(w, "data: refresh\n\n")
			if err != nil {
				color.Red("proxy error: %v", err)
			}
			flusher.Flush()
		}
	}
}

// Refresh broadcasts a signal to refresh to all proxy subscribers.
func (p *Proxy) Refresh() {
	p.SubscribersMu.Lock()
	defer p.SubscribersMu.Unlock()

	for sub := range p.Subscribers {
		select {
		case sub <- struct{}{}: // push a refresh signal.

		default: // skip if the subscriber channel is full.
		}
	}
}

func (p *Proxy) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), CLOSE_DELAY*time.Second)
	defer cancel()

	err := p.Server.Shutdown(ctx)
	if err != nil {
		err = p.Server.Close()
	}

	return err
}
