package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// 回归：桌面 App 模式下 /icons/* 必须可访问，否则页头 logo 会显示成破图（问号）。
func TestIconsRouteServed(t *testing.T) {
	mux := newMux()
	for _, p := range []string{"/icons/icon-192.png", "/icons/icon-512.png", "/icons/apple-touch-icon.png"} {
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, p, nil))
		if rec.Code != http.StatusOK {
			t.Fatalf("%s: status=%d, want 200", p, rec.Code)
		}
		if ct := rec.Header().Get("Content-Type"); ct != "image/png" {
			t.Fatalf("%s: content-type=%q, want image/png", p, ct)
		}
	}
}
