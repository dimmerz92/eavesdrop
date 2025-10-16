package eavesdrop_test

import (
	"bufio"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/dimmerz92/eavesdrop"
)

func TestProxyConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  eavesdrop.ProxyConfig
		wantErr bool
	}{
		{"Valid config", eavesdrop.ProxyConfig{Enabled: true, AppPort: 3000, ProxyPort: 4000}, false},
		{"Same ports", eavesdrop.ProxyConfig{Enabled: true, AppPort: 3000, ProxyPort: 3000}, true},
		{"App port out of range", eavesdrop.ProxyConfig{Enabled: true, AppPort: 1024, ProxyPort: 4000}, true},
		{"Proxy port out of range", eavesdrop.ProxyConfig{Enabled: true, AppPort: 3000, ProxyPort: 70000}, true},
		{"Disabled config", eavesdrop.ProxyConfig{Enabled: false}, false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.config.Validate()
			if test.wantErr && err == nil {
				t.Fatal("expected error, got none")
			} else if !test.wantErr && err != nil {
				t.Fatalf("expected nil, got %v", err)
			}
		})
	}
}

func TestProxyConfig_ToProxy(t *testing.T) {
	cfg := eavesdrop.ProxyConfig{Enabled: true, AppPort: 3000, ProxyPort: 4000}
	proxy := cfg.ToProxy()

	if proxy == nil {
		t.Fatal("expected non-nil proxy")
	}

	if proxy.AppPort != ":3000" || proxy.ProxyPort != ":4000" {
		t.Fatalf("unexpected proxy ports: got %s / %s", proxy.AppPort, proxy.ProxyPort)
	}
}

func TestProxy_InjectSSE_WithBodyTag(t *testing.T) {
	resp := &http.Response{
		Body:   io.NopCloser(strings.NewReader(`<html><body>Hello world</body></html>`)),
		Header: http.Header{"Content-Type": []string{"text/html"}},
	}

	proxy := &eavesdrop.Proxy{}

	modified, err := proxy.InjectSSE(resp)
	if err != nil {
		t.Fatalf("expected injection, got %v", err)
	}

	result := string(modified)
	if !strings.Contains(result, "<script>") {
		t.Fatal("expected injected <script> tag")
	}

	if !strings.Contains(result, eavesdrop.SSE_SCRIPT) {
		t.Fatal("expected injected SSE_SCRIPT content")
	}

	if !strings.HasSuffix(result, "</body></html>") {
		t.Fatalf("expected content to end with </body></html>, got %s", result)
	}
}

func TestProxy_InjectSSE_NoBodyTag(t *testing.T) {
	resp := &http.Response{
		Body:   io.NopCloser(strings.NewReader(`<html><head></head><div>Content</div></html>`)),
		Header: http.Header{"Content-Type": []string{"text/html"}},
	}

	proxy := &eavesdrop.Proxy{}

	modified, err := proxy.InjectSSE(resp)
	if err != nil {
		t.Fatalf("InjectSSE error: %v", err)
	}

	result := string(modified)
	if strings.Contains(result, "<script>") {
		t.Fatalf("expected no <script> tag injection")
	}
}

func TestProxy_RefreshBroadcast(t *testing.T) {
	proxy := &eavesdrop.Proxy{
		Subscribers:   make(map[chan struct{}]struct{}),
		SubscribersMu: &sync.Mutex{},
	}

	ch1 := make(chan struct{}, 1)
	ch2 := make(chan struct{}, 1)

	proxy.SubscribersMu.Lock()
	proxy.Subscribers[ch1] = struct{}{}
	proxy.Subscribers[ch2] = struct{}{}
	proxy.SubscribersMu.Unlock()

	proxy.Refresh()

	select {
	case <-ch1:
	default:
		t.Fatalf("expected refresh signal on ch1")
	}

	select {
	case <-ch2:
	default:
		t.Fatalf("expected refresh signal on ch2")
	}
}

func TestClientEvent(t *testing.T) {
	proxy := &eavesdrop.Proxy{
		Subscribers:   make(map[chan struct{}]struct{}),
		SubscribersMu: &sync.Mutex{},
	}

	server := httptest.NewServer(http.HandlerFunc(proxy.ClientEvent))
	defer server.Close()

	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", server.URL, nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	respCh := make(chan *http.Response, 1)
	errCh := make(chan error, 1)
	go func() {
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			errCh <- err
			return
		}
		respCh <- resp
	}()

	var ch chan struct{}
	for start := time.Now(); time.Since(start) < 2*time.Second; {
		proxy.SubscribersMu.Lock()
		if len(proxy.Subscribers) == 1 {
			for c := range proxy.Subscribers {
				ch = c
			}
			proxy.SubscribersMu.Unlock()
			break
		}
		proxy.SubscribersMu.Unlock()
		time.Sleep(10 * time.Millisecond)
	}

	ch <- struct{}{}

	var resp *http.Response
	select {
	case resp = <-respCh:
	case err = <-errCh:
		t.Fatalf("failed to get a response: %v", err)
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for response headers")
	}

	defer resp.Body.Close()

	header := resp.Header.Get("Content-Type")
	if header != "text/event-stream" {
		t.Fatalf("expected 'text/event-stream', got %s", header)
	}

	reader := bufio.NewReader(resp.Body)
	line, err := reader.ReadString('\n')
	if err != nil {
		t.Fatalf("failed to read SSE data: %v", err)
	}

	if !strings.Contains(line, "data: refresh") {
		t.Fatalf("expected 'data: refresh', got %s", line)
	}

	cancel()
	time.Sleep(50 * time.Millisecond)

	proxy.SubscribersMu.Lock()
	defer proxy.SubscribersMu.Unlock()
	if len(proxy.Subscribers) != 0 {
		t.Fatal("expected subscriber cleanup")
	}
}
