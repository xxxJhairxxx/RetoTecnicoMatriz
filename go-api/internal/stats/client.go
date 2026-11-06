package stats

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/jinfanzon/reto-tecnico/go-api/internal/matrix"
)

type Client struct {
	baseURL string
	http    *http.Client
}

func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		http:    &http.Client{Timeout: 10 * time.Second},
	}
}

func (c *Client) Fetch(q, r matrix.Matrix, authHeader string) (*matrix.Statistics, error) {
	body, err := json.Marshal(map[string]matrix.Matrix{"q": q, "r": r})
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(http.MethodPost, c.baseURL+"/api/v1/statistics", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", authHeader)

	res, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("la API de estadisticas no responde: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		snippet, _ := io.ReadAll(io.LimitReader(res.Body, 512))
		return nil, fmt.Errorf("la API de estadisticas respondio %d: %s", res.StatusCode, snippet)
	}
	var out matrix.Statistics
	if err := json.NewDecoder(res.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("respuesta ilegible de la API de estadisticas: %w", err)
	}
	return &out, nil
}
