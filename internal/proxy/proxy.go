package proxy

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/dimmerz92/eavesdrop/internal/config"
)

const SSE_SCRIPT = `<script>
	const eventSource = new EventSource("/eavesdrop_sse");
	eventSource.onmessage = (event) => event.data === "refresh" && window.location.reload();
	eventSource.onerror = (error) => console.error("eavesdrop sse error:", error);
</script>`

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
	// create a forwarding request
	req, err := http.NewRequest(r.Method, fmt.Sprintf("http://localhost:%d%s", p.AppPort, r.URL.Path), r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
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
		http.Error(w, "proxy error: application unresponsive", http.StatusBadGateway)
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
			http.Error(w, fmt.Sprintf("proxy error: %v", err), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Length", strconv.Itoa(len([]byte(body))))
		if _, err := io.WriteString(w, body); err != nil {
			http.Error(w, fmt.Sprintf("proxy error: %v", err), http.StatusInternalServerError)
			return
		}
	}
}

// injectRefresher injects the sse script if the body exists.
func (p *Proxy) injectRefresher(resp *http.Response) (string, error) {
	buf := new(bytes.Buffer)
	if _, err := buf.ReadFrom(resp.Body); err != nil {
		return "", fmt.Errorf("proxy error: failed to read body")
	}
	page := buf.String()

	// get the index of the body closing tag
	body := strings.LastIndex(page, "</body>")
	if body == -1 {
		return page, nil
	}

	return page[:body] + SSE_SCRIPT + page[body:], nil
}
