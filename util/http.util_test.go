package util

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHttpGetSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}))
	defer server.Close()

	resp := HttpGet(server.URL)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read body: %v", err)
	}

	if string(body) != "ok" {
		t.Fatalf("expected body ok, got %q", string(body))
	}
}

func TestHttpGetFailureReturns500(t *testing.T) {
	resp := HttpGet("://invalid-url")
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", resp.StatusCode)
	}
}
