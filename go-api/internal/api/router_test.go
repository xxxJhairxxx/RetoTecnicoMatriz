package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/jinfanzon/reto-tecnico/go-api/internal/matrix"
)

var testConfig = Config{JWTSecret: "test-secret", APIUser: "demo", APIPassword: "demo123"}

type fakeStats struct {
	gotQ, gotR matrix.Matrix
	gotAuth    string
	result     *matrix.Statistics
	err        error
}

func (f *fakeStats) Fetch(q, r matrix.Matrix, authHeader string) (*matrix.Statistics, error) {
	f.gotQ, f.gotR, f.gotAuth = q, r, authHeader
	if f.err != nil {
		return nil, f.err
	}
	return f.result, nil
}

func newFakeStats() *fakeStats {
	return &fakeStats{result: &matrix.Statistics{
		Max: 2, Min: 0, Sum: 3, Average: 0.75, Count: 4,
		Diagonal: map[string]bool{"q": false, "r": true}, AnyDiagonal: true,
	}}
}

func post(t *testing.T, app *fiber.App, path, token string, body any) (*http.Response, map[string]any) {
	t.Helper()
	raw, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, path, bytes.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	res, err := app.Test(req, 5000)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	var out map[string]any
	json.NewDecoder(res.Body).Decode(&out)
	return res, out
}

func TestQREndpointEndToEnd(t *testing.T) {
	stats := newFakeStats()
	app := NewRouter(testConfig, stats)

	_, login := post(t, app, "/api/v1/auth/token", "", map[string]string{"username": "demo", "password": "demo123"})
	token, ok := login["token"].(string)
	if !ok || token == "" {
		t.Fatalf("no token returned: %v", login)
	}

	res, out := post(t, app, "/api/v1/qr", token, map[string]matrix.Matrix{"matrix": {{1, 2}, {3, 4}}})
	if res.StatusCode != http.StatusOK {
		t.Fatalf("status %d, want 200: %v", res.StatusCode, out)
	}
	for _, key := range []string{"q", "r", "statistics", "rows", "columns"} {
		if _, present := out[key]; !present {
			t.Errorf("missing key %q in response: %v", key, out)
		}
	}
	if s := out["statistics"].(map[string]any); s["sum"] != 3.0 {
		t.Errorf("statistics were not propagated: %v", s)
	}
	if stats.gotAuth != "Bearer "+token {
		t.Errorf("Authorization was not forwarded to the provider: %q", stats.gotAuth)
	}
	if len(stats.gotQ) != 2 || len(stats.gotR) != 2 {
		t.Errorf("want Q and R with 2 rows, got Q=%v R=%v", stats.gotQ, stats.gotR)
	}
}

func TestQREndpointRequiresToken(t *testing.T) {
	app := NewRouter(testConfig, newFakeStats())
	res, _ := post(t, app, "/api/v1/qr", "", map[string]matrix.Matrix{"matrix": {{1, 2}, {3, 4}}})
	if res.StatusCode != http.StatusUnauthorized {
		t.Errorf("status %d, want 401", res.StatusCode)
	}
}

func TestQREndpointRejectsForeignToken(t *testing.T) {
	app := NewRouter(testConfig, newFakeStats())
	foreignToken, _, _ := issueToken("another-secret", "intruder")
	res, _ := post(t, app, "/api/v1/qr", foreignToken, map[string]matrix.Matrix{"matrix": {{1, 2}, {3, 4}}})
	if res.StatusCode != http.StatusUnauthorized {
		t.Errorf("status %d, want 401", res.StatusCode)
	}
}

func TestLoginRejectsBadCredentials(t *testing.T) {
	app := NewRouter(testConfig, newFakeStats())
	res, _ := post(t, app, "/api/v1/auth/token", "", map[string]string{"username": "demo", "password": "wrong"})
	if res.StatusCode != http.StatusUnauthorized {
		t.Errorf("status %d, want 401", res.StatusCode)
	}
}

func TestQREndpointRejectsInvalidMatrix(t *testing.T) {
	app := NewRouter(testConfig, newFakeStats())
	token, _, _ := issueToken(testConfig.JWTSecret, "demo")

	res, out := post(t, app, "/api/v1/qr", token, map[string]matrix.Matrix{"matrix": {{1, 2}, {2, 4}}})
	if res.StatusCode != http.StatusUnprocessableEntity {
		t.Errorf("status %d, want 422: %v", res.StatusCode, out)
	}
}

func TestQREndpointReportsProviderFailure(t *testing.T) {
	stats := newFakeStats()
	stats.err = errors.New("la API de estadisticas no responde")
	app := NewRouter(testConfig, stats)
	token, _, _ := issueToken(testConfig.JWTSecret, "demo")

	res, _ := post(t, app, "/api/v1/qr", token, map[string]matrix.Matrix{"matrix": {{1, 2}, {3, 4}}})
	if res.StatusCode != http.StatusBadGateway {
		t.Errorf("status %d, want 502", res.StatusCode)
	}
}
