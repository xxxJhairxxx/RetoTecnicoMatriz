package stats

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/jinfanzon/reto-tecnico/go-api/internal/matrix"
)

func TestMain(m *testing.M) {
	retryDelay = 0
	os.Exit(m.Run())
}

func TestFetchSendsMatricesAndReturnsStatistics(t *testing.T) {
	var received map[string]matrix.Matrix
	var authHeader string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader = r.Header.Get("Authorization")
		if r.URL.Path != "/api/v1/statistics" {
			t.Errorf("unexpected path %q", r.URL.Path)
		}
		json.NewDecoder(r.Body).Decode(&received)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"max":4,"min":1,"sum":10,"average":2.5,"count":4,"diagonal":{"q":false,"r":true},"anyDiagonal":true}`))
	}))
	defer server.Close()

	out, err := NewClient(server.URL).Fetch(matrix.Matrix{{1, 2}}, matrix.Matrix{{3, 4}}, "Bearer xyz")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if authHeader != "Bearer xyz" {
		t.Errorf("Authorization = %q, want %q", authHeader, "Bearer xyz")
	}
	if len(received["q"]) != 1 || received["q"][0][1] != 2 || received["r"][0][0] != 3 {
		t.Errorf("q and r were not sent correctly: %v", received)
	}
	if out.Sum != 10 || !out.AnyDiagonal {
		t.Errorf("statistics decoded incorrectly: %+v", out)
	}
}

func TestFetchPropagatesStatusError(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		http.Error(w, "invalid matrix", http.StatusUnprocessableEntity)
	}))
	defer server.Close()

	_, err := NewClient(server.URL).Fetch(matrix.Matrix{{1}}, matrix.Matrix{{1}}, "Bearer xyz")
	if err == nil || !strings.Contains(err.Error(), "422") {
		t.Errorf("want an error mentioning 422, got %v", err)
	}
	if attempts != 1 {
		t.Errorf("a client error must not be retried, got %d attempts", attempts)
	}
}

func TestFetchRetriesWhileTheServiceWakesUp(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 3 {
			w.WriteHeader(http.StatusBadGateway)
			w.Write([]byte("<html>502</html>"))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"max":4,"min":1,"sum":10,"average":2.5,"count":4,"diagonal":{"q":false,"r":true},"anyDiagonal":true}`))
	}))
	defer server.Close()

	out, err := NewClient(server.URL).Fetch(matrix.Matrix{{1}}, matrix.Matrix{{1}}, "Bearer xyz")
	if err != nil {
		t.Fatalf("unexpected error after retrying: %v", err)
	}
	if out.Sum != 10 {
		t.Errorf("statistics decoded incorrectly: %+v", out)
	}
	if attempts != 3 {
		t.Errorf("want 3 attempts, got %d", attempts)
	}
}

func TestFetchGivesUpAfterMaxAttempts(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		w.WriteHeader(http.StatusBadGateway)
	}))
	defer server.Close()

	_, err := NewClient(server.URL).Fetch(matrix.Matrix{{1}}, matrix.Matrix{{1}}, "Bearer xyz")
	if err == nil || !strings.Contains(err.Error(), "502") {
		t.Errorf("want an error mentioning 502, got %v", err)
	}
	if attempts != maxAttempts {
		t.Errorf("want %d attempts, got %d", maxAttempts, attempts)
	}
}

func TestFetchOmitsNonJSONErrorBodies(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusUnprocessableEntity)
		w.Write([]byte("<!DOCTYPE html><html><head><title>422</title></head></html>"))
	}))
	defer server.Close()

	_, err := NewClient(server.URL).Fetch(matrix.Matrix{{1}}, matrix.Matrix{{1}}, "Bearer xyz")
	if err == nil || strings.Contains(err.Error(), "DOCTYPE") {
		t.Errorf("an HTML body must not leak into the error message, got %v", err)
	}
}

func TestNewClientTrimsTrailingSlash(t *testing.T) {
	var path string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"max":1,"min":1,"sum":1,"average":1,"count":1,"diagonal":{},"anyDiagonal":false}`))
	}))
	defer server.Close()

	NewClient(server.URL+"/").Fetch(matrix.Matrix{{1}}, matrix.Matrix{{1}}, "Bearer xyz")
	if path != "/api/v1/statistics" {
		t.Errorf("path = %q, want %q", path, "/api/v1/statistics")
	}
}

func TestFetchPropagatesNetworkError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	server.Close()

	_, err := NewClient(server.URL).Fetch(matrix.Matrix{{1}}, matrix.Matrix{{1}}, "Bearer xyz")
	if err == nil || !strings.Contains(err.Error(), "no responde") {
		t.Errorf("want a connection error, got %v", err)
	}
}

func TestFetchRejectsUnreadableResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("this is not json"))
	}))
	defer server.Close()

	_, err := NewClient(server.URL).Fetch(matrix.Matrix{{1}}, matrix.Matrix{{1}}, "Bearer xyz")
	if err == nil || !strings.Contains(err.Error(), "ilegible") {
		t.Errorf("want a decoding error, got %v", err)
	}
}
