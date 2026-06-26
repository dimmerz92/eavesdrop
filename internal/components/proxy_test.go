package components_test

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/dimmerz92/eavesdrop/v2/internal/components"
)

func freePort(t *testing.T) uint16 {
	t.Helper()
	ln, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatalf("freePort: %v", err)
	}
	port := uint16(ln.Addr().(*net.TCPAddr).Port)
	ln.Close()
	return port
}

func appPort(s *httptest.Server) uint16 {
	_, p, _ := net.SplitHostPort(s.Listener.Addr().String())
	port, _ := strconv.ParseUint(p, 10, 16)
	return uint16(port)
}

func TestNewProxy(t *testing.T) {
	tests := []struct {
		name        string
		appPort     uint16
		proxyPort   uint16
		expectedErr bool
	}{
		{"zero app port", 0, 8001, true},
		{"zero proxy port", 8000, 0, true},
		{"equal ports", 8000, 8000, true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()

			_, err := components.NewProxy(ctx, test.appPort, test.proxyPort)
			if test.expectedErr {
				if err == nil {
					t.Fatal("expected panic")
				}
				return
			}
			if err != nil {
				t.Fatalf("failed to return new proxy: %v", err)
			}
		})
	}

	t.Run("valid ports", func(t *testing.T) {
		ctx, cancel := context.WithCancel(t.Context())
		defer cancel()
		if p, _ := components.NewProxy(ctx, freePort(t), freePort(t)); p == nil {
			t.Error("NewProxy() returned nil")
		}
	})
}

func TestProxy_Forwarding(t *testing.T) {
	tests := []struct {
		name                string
		appBody             string
		appContentType      string
		appStatus           int
		expectedStatus      int
		expectedContains    string
		expectedNotContains string
	}{
		{
			name:                "non-html passes through unchanged",
			appBody:             `{"key":"value"}`,
			appContentType:      "application/json",
			appStatus:           http.StatusOK,
			expectedStatus:      http.StatusOK,
			expectedContains:    `{"key":"value"}`,
			expectedNotContains: "EventSource",
		},
		{
			name:             "html gets SSE script injected before closing body tag",
			appBody:          "<html><body>hello</body></html>",
			appContentType:   "text/html",
			appStatus:        http.StatusOK,
			expectedStatus:   http.StatusOK,
			expectedContains: "EventSource",
		},
		{
			name:                "html without body tag is not modified",
			appBody:             "<html>no body tag</html>",
			appContentType:      "text/html",
			appStatus:           http.StatusOK,
			expectedStatus:      http.StatusOK,
			expectedNotContains: "EventSource",
		},
		{
			name:             "non-200 status is forwarded",
			appBody:          "not found",
			appContentType:   "text/plain",
			appStatus:        http.StatusNotFound,
			expectedStatus:   http.StatusNotFound,
			expectedContains: "not found",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			app := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", test.appContentType)
				w.WriteHeader(test.appStatus)
				fmt.Fprint(w, test.appBody)
			}))
			defer app.Close()

			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()

			proxyPort := freePort(t)
			_, _ = components.NewProxy(ctx, appPort(app), proxyPort)

			resp, err := http.Get(fmt.Sprintf("http://localhost:%d/", proxyPort))
			if err != nil {
				t.Fatalf("proxy GET: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != test.expectedStatus {
				t.Errorf("status = %d, expected %d", resp.StatusCode, test.expectedStatus)
			}

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatalf("read body: %v", err)
			}
			bodyStr := string(body)

			if test.expectedContains != "" && !strings.Contains(bodyStr, test.expectedContains) {
				t.Errorf("body does not contain %q\nbody: %s", test.expectedContains, bodyStr)
			}

			if test.expectedNotContains != "" && strings.Contains(bodyStr, test.expectedNotContains) {
				t.Errorf("body unexpectedly contains %q\nbody: %s", test.expectedNotContains, bodyStr)
			}
		})
	}
}

func TestProxy_SSE(t *testing.T) {
	tests := []struct {
		name      string
		doRefresh bool
	}{
		{"connected message received on connect", false},
		{"refresh message received after RefreshBrowser", true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			app := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
			defer app.Close()

			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()

			proxyPort := freePort(t)
			p, _ := components.NewProxy(ctx, appPort(app), proxyPort)

			resp, err := http.Get(fmt.Sprintf("http://localhost:%d/eavesdrop_sse", proxyPort))
			if err != nil {
				t.Fatalf("SSE connect: %v", err)
			}
			defer resp.Body.Close()

			if ct := resp.Header.Get("Content-Type"); ct != "text/event-stream" {
				t.Errorf("Content-Type = %q, expected %q", ct, "text/event-stream")
			}

			msgs := make(chan string, 10)
			go func() {
				scanner := bufio.NewScanner(resp.Body)
				for scanner.Scan() {
					if line := scanner.Text(); strings.HasPrefix(line, "data:") {
						msgs <- line
					}
				}
			}()

			select {
			case msg := <-msgs:
				if msg != "data: connected" {
					t.Errorf("first message = %q, expected %q", msg, "data: connected")
				}
			case <-time.After(time.Second):
				t.Fatal("timeout waiting for connected message")
			}

			if test.doRefresh {
				p.RefreshBrowser()
				select {
				case msg := <-msgs:
					if msg != "data: refresh" {
						t.Errorf("refresh message = %q, expected %q", msg, "data: refresh")
					}
				case <-time.After(time.Second):
					t.Fatal("timeout waiting for refresh message")
				}
			}
		})
	}
}
