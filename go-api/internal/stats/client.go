package stats

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/jinfanzon/reto-tecnico/go-api/internal/matrix"
)

var (
	maxAttempts   = 5
	retryDelay    = 12 * time.Second
	clientTimeout = 30 * time.Second
)

type Client struct {
	baseURL string
	http    *http.Client
}

func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: strings.TrimSuffix(baseURL, "/"),
		http:    &http.Client{Timeout: clientTimeout},
	}
}

func (c *Client) Fetch(q, r matrix.Matrix, authHeader string) (*matrix.Statistics, error) {
	body, err := json.Marshal(map[string]matrix.Matrix{"q": q, "r": r})
	if err != nil {
		return nil, err
	}

	var lastErr error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		if attempt > 1 {
			time.Sleep(retryDelay)
		}
		out, retryable, err := c.post(body, authHeader)
		if err == nil {
			return out, nil
		}
		lastErr = err
		if !retryable {
			return nil, err
		}
	}
	return nil, lastErr
}

func (c *Client) post(body []byte, authHeader string) (*matrix.Statistics, bool, error) {
	req, err := http.NewRequest(http.MethodPost, c.baseURL+"/api/v1/statistics", bytes.NewReader(body))
	if err != nil {
		return nil, false, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", authHeader)

	res, err := c.http.Do(req)
	if err != nil {
		return nil, true, fmt.Errorf("la API de estadisticas no responde: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, res.StatusCode >= http.StatusInternalServerError,
			fmt.Errorf("la API de estadisticas respondio %d%s", res.StatusCode, jsonDetail(res))
	}

	var out matrix.Statistics
	if err := json.NewDecoder(res.Body).Decode(&out); err != nil {
		return nil, false, fmt.Errorf("respuesta ilegible de la API de estadisticas: %w", err)
	}
	return &out, false, nil
}

func jsonDetail(res *http.Response) string {
	if !strings.HasPrefix(res.Header.Get("Content-Type"), "application/json") {
		return ""
	}
	snippet, _ := io.ReadAll(io.LimitReader(res.Body, 512))
	return ": " + string(snippet)
}
