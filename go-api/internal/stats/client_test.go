package stats

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/jinfanzon/reto-tecnico/go-api/internal/matrix"
)

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
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "invalid matrix", http.StatusUnprocessableEntity)
	}))
	defer server.Close()

	_, err := NewClient(server.URL).Fetch(matrix.Matrix{{1}}, matrix.Matrix{{1}}, "Bearer xyz")
	if err == nil || !strings.Contains(err.Error(), "422") {
		t.Errorf("want an error mentioning 422, got %v", err)
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
