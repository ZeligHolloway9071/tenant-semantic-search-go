package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSearchHandlerUsesTenantFilter(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var b map[string]any
		json.NewDecoder(r.Body).Decode(&b)
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/v1/embeddings":
			json.NewEncoder(w).Encode(map[string]any{"ok": true, "data": []any{map[string]any{"embedding": []float64{1, 2}}}})
		case "/v1/vector/query":
			if b["top_k"] != float64(5) {
				t.Fatalf("top_k=%v", b["top_k"])
			}
			f := b["filter"].(map[string]any)
			if f["tenant"] != "acme" {
				t.Fatalf("filter=%v", f)
			}
			json.NewEncoder(w).Encode(map[string]any{"ok": true, "data": map[string]any{"results": []any{}}})
		default:
			json.NewEncoder(w).Encode(map[string]any{"ok": true, "data": map[string]any{}})
		}
	}))
	defer s.Close()
	c := &InfraiClient{BaseURL: s.URL, Key: "test", HTTP: s.Client()}
	req := httptest.NewRequest(http.MethodPost, "/search", strings.NewReader(`{"tenant":"acme","query":"admin"}`))
	rr := httptest.NewRecorder()
	searchHandler(c).ServeHTTP(rr, req)
	if rr.Code != 200 {
		t.Fatalf("status %d", rr.Code)
	}
}
