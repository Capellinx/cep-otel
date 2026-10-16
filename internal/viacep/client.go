package viacep

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

var ErrNotFound = errors.New("zipcode not found")

type Client struct {
	BaseURL string
	HTTP    *http.Client
}

func NewClient() *Client {
	return &Client{
		BaseURL: "https://viacep.com.br",
		HTTP: &http.Client{
			Timeout:   5 * time.Second,
			Transport: otelhttp.NewTransport(http.DefaultTransport),
		},
	}
}

type response struct {
	Localidade string `json:"localidade"`
	Erro       any    `json:"erro"`
}

func (c *Client) FindCity(ctx context.Context, cep string) (string, error) {
	endpoint := c.BaseURL + "/ws/" + url.PathEscape(cep) + "/json/"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return "", fmt.Errorf("viacep: build request: %w", err)
	}
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return "", fmt.Errorf("viacep: request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("viacep: unexpected status %d", resp.StatusCode)
	}
	var r response
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return "", fmt.Errorf("viacep: decode: %w", err)
	}
	if r.Erro == true || r.Erro == "true" {
		return "", ErrNotFound
	}
	if r.Localidade == "" {
		return "", errors.New("viacep: response without city")
	}
	return r.Localidade, nil
}
