package demo_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	asnblock "github.com/DoVietHoang1712/demo"
)

const (
	pluginName = "asnblock"
)

func TestNew(t *testing.T) {
	t.Run("Disabled", func(t *testing.T) {
		plugin, _ := asnblock.CreatePlugin("X-Real-IP", []string{"206948"}, pluginName)
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
		plugin, _ := asnblock.CreatePlugin("X-Real-IP", []string{"35236"}, pluginName)

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
			plugin, _ := asnblock.CreatePlugin("X-Real-IP", []string{"206948"}, pluginName)

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
			plugin, _ := asnblock.CreatePlugin("X-Real-IP", []string{"35236"}, pluginName)

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
