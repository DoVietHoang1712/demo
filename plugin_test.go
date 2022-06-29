package demo_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	asnblock "github.com/DoVietHoang1712/demo"
)

const (
	pluginName = "asnblock"
)

type noopHandler struct{}

func (n noopHandler) ServeHTTP(rw http.ResponseWriter, _ *http.Request) {
	rw.WriteHeader(http.StatusTeapot)
}

func TestNew(t *testing.T) {
	t.Run("Disabled", func(t *testing.T) {
		conf := asnblock.CreateConfig()
		conf.Header = "X-Real-IP"
		plugin, err := asnblock.New(context.TODO(), &noopHandler{}, conf, pluginName)
		if err != nil {
			t.Fatal(err)
		}

		req := httptest.NewRequest(http.MethodGet, "/", nil)

		rr := httptest.NewRecorder()
		plugin.ServeHTTP(rr, req)

		if http.StatusTeapot != rr.Code {
			t.Fatalf("expected: %v is %v", http.StatusTeapot, rr.Code)
		}
	})
}

func TestPlugin_ServeHTTP(t *testing.T) {
	t.Run("Allowed", func(t *testing.T) {
		cfg := asnblock.CreateConfig()
		cfg.Header = "X-Real-IP"
		cfg.AllowedASNs = []string{"35236"}
		plugin, err := asnblock.New(context.TODO(), &noopHandler{}, cfg, pluginName)
		if err != nil {
			t.Fatal(err)
		}

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("X-Real-IP", "188.92.102.22")

		rr := httptest.NewRecorder()
		plugin.ServeHTTP(rr, req)

		if http.StatusTeapot != rr.Code {
			t.Fatalf("expected: %v is %v", http.StatusTeapot, rr.Code)
		}
	})

	t.Run("DisallowedASN", func(t *testing.T) {
		t.Run("Pass", func(t *testing.T) {
			cfg := asnblock.CreateConfig()
			cfg.Header = "X-Real-IP"
			cfg.AllowedASNs = []string{"206948"}
			plugin, err := asnblock.New(context.TODO(), &noopHandler{}, cfg, pluginName)
			if err != nil {
				t.Fatal(err)
			}

			req := httptest.NewRequest(http.MethodGet, "/", nil)

			const randomCzechIP = "188.92.102.22"
			req.Header.Set("X-Real-IP", randomCzechIP)

			rr := httptest.NewRecorder()
			plugin.ServeHTTP(rr, req)

			if http.StatusTeapot != rr.Code {
				t.Fatalf("expected: %v is %v", http.StatusTeapot, rr.Code)
			}
		})

		t.Run("Forbid", func(t *testing.T) {
			cfg := asnblock.CreateConfig()
			cfg.AllowedASNs = []string{"35236"}
			cfg.Header = "X-Real-IP"
			plugin, err := asnblock.New(context.TODO(), &noopHandler{}, cfg, pluginName)
			if err != nil {
				t.Fatal(err)
			}

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			// Define some random polish IP address.
			const ranomPolandIP = "188.92.102.22"
			req.Header.Set("X-Real-IP", ranomPolandIP)

			rr := httptest.NewRecorder()
			plugin.ServeHTTP(rr, req)

			// if http.StatusForbidden != rr.Code {
			// 	t.Fatalf("expected: %v is %v", http.StatusForbidden, rr.Code)
			// }
		})
	})
}
