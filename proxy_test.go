package eavesdrop_test

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/dimmerz92/eavesdrop"
)

func TestProxy(t *testing.T) {
	t.Run("failed construction", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("expected panic")
			}
		}()
		eavesdrop.NewProxy(t.Context(),
			eavesdrop.WithAppPort(8000),
			eavesdrop.WithProxyPort(8000),
		)
	})

	t.Run("sse handling and script injection", func(t *testing.T) {
		proxy := eavesdrop.NewProxy(t.Context(),
			eavesdrop.WithAppPort(8000),
			eavesdrop.WithProxyPort(8001),
		)

		htmlWithBody := "<html><body>Test</body></html>"
		expectedWithBody := fmt.Sprintf("<html><body>Test<script>%s</script></body></html>", eavesdrop.SSE_SCRIPT)
		htmlWithBodyHandler := func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(htmlWithBody))
		}

		htmlNoBody := "<html>Test</html>"
		expectedNoBody := htmlNoBody
		htmlNoBodyHandler := func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(htmlNoBody))
		}

		http.HandleFunc("GET /with", htmlWithBodyHandler)
		http.HandleFunc("GET /without", htmlNoBodyHandler)

		errCh := make(chan error, 1)
		go func() {
			errCh <- http.ListenAndServe(":8000", nil)
		}()

		ctx, cancel := context.WithCancel(t.Context())
		defer cancel()

		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("unexpected panic: %v", r)
			}
		}()

		go func() {
			select {
			case err := <-errCh:
				if err != nil && !errors.Is(err, http.ErrServerClosed) {
					panic(err)
				}

			case <-ctx.Done():
				return
			}
		}()

		t.Run("script injected", func(t *testing.T) {
			resp, err := http.Get("http://127.0.0.1:8001/with")
			if err != nil {
				t.Fatalf("failed to send GET: %v", err)
			}
			defer resp.Body.Close()

			var buf bytes.Buffer
			if _, err = buf.ReadFrom(resp.Body); err != nil {
				t.Fatalf("failed to read from body: %v", err)
			}

			if got := buf.String(); got != expectedWithBody {
				t.Errorf("expected %s, got %s", expectedWithBody, got)
			}
		})

		t.Run("no script injected", func(t *testing.T) {
			resp, err := http.Get("http://127.0.0.1:8001/without")
			if err != nil {
				t.Fatalf("failed to send GET: %v", err)
			}
			defer resp.Body.Close()

			var buf bytes.Buffer
			if _, err = buf.ReadFrom(resp.Body); err != nil {
				t.Fatalf("failed to read from body: %v", err)
			}

			if got := buf.String(); got != expectedNoBody {
				t.Errorf("expected %s, got %s", expectedNoBody, got)
			}
		})

		t.Run("sse refresh handling", func(t *testing.T) {
			req, _ := http.NewRequest("GET", "http://127.0.0.1:8001/eavesdrop_sse", nil)
			req.Header.Set("Accept", "text/event-stream")

			client := http.Client{Timeout: 0}
			resp, err := client.Do(req)
			if err != nil {
				t.Fatalf("failed to resolve request: %v", err)
			}
			defer resp.Body.Close()

			out := make(chan string, 1)
			go func() {
				scanner := bufio.NewScanner(resp.Body)
				for scanner.Scan() {
					line := scanner.Text()
					if line == "data: refresh" {
						out <- line
					}
				}
			}()

			proxy.RefreshBrowser()

			ctx, cancel := context.WithTimeout(t.Context(), 200*time.Millisecond)
			defer cancel()

			expected := "data: refresh"
			select {
			case <-ctx.Done():
				t.Error("failed to receive refresh broadcast")

			case got := <-out:
				if got != expected {
					t.Errorf("expected %s, got %s", expected, got)
				}
			}
		})

	})
}
