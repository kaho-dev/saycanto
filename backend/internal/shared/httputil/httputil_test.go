package httputil

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWriteJSON(t *testing.T) {
	rec := httptest.NewRecorder()
	WriteJSON(rec, http.StatusOK, map[string]string{"hello": "world"})

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json; charset=utf-8" {
		t.Fatalf("unexpected Content-Type: %s", ct)
	}
	var body map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode JSON: %v", err)
	}
	if body["hello"] != "world" {
		t.Fatalf("expected world, got %s", body["hello"])
	}
}

func TestWriteError(t *testing.T) {
	rec := httptest.NewRecorder()
	WriteError(rec, http.StatusBadRequest, "bad input")

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
	var body ErrorResponse
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode JSON: %v", err)
	}
	if body.Error != "bad input" {
		t.Fatalf("expected 'bad input', got %q", body.Error)
	}
}

func TestParseIntQuery(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/?limit=15&bad=abc", nil)
	if got := ParseIntQuery(req, "limit", 20); got != 15 {
		t.Fatalf("expected 15, got %d", got)
	}
	if got := ParseIntQuery(req, "bad", 20); got != 20 {
		t.Fatalf("expected default 20, got %d", got)
	}
	if got := ParseIntQuery(req, "missing", 10); got != 10 {
		t.Fatalf("expected default 10, got %d", got)
	}
}

func TestCORSMiddleware(t *testing.T) {
	handler := CORSMiddleware("")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Preflight OPTIONS
	reqOpt := httptest.NewRequest(http.MethodOptions, "/api/search", nil)
	recOpt := httptest.NewRecorder()
	handler.ServeHTTP(recOpt, reqOpt)
	if recOpt.Code != http.StatusNoContent {
		t.Fatalf("expected 204 for OPTIONS, got %d", recOpt.Code)
	}
	if recOpt.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Fatalf("expected * origin")
	}

	// Normal GET with explicit origin
	handlerCustom := CORSMiddleware("http://localhost:3000")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	reqGet := httptest.NewRequest(http.MethodGet, "/api/search", nil)
	recGet := httptest.NewRecorder()
	handlerCustom.ServeHTTP(recGet, reqGet)
	if recGet.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recGet.Code)
	}
	if recGet.Header().Get("Access-Control-Allow-Origin") != "http://localhost:3000" {
		t.Fatalf("expected custom origin")
	}
}
