package weather

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

type Client struct {
	BaseURL string
	APIKey  string
	HTTP    *http.Client
}

func NewClient(apiKey string) *Client {
	return &Client{
		BaseURL: "https://api.weatherapi.com",
		APIKey:  apiKey,
		HTTP: &http.Client{
			Timeout:   5 * time.Second,
			Transport: otelhttp.NewTransport(http.DefaultTransport),
		},
	}
}

type response struct {
	Current *struct {
		TempC float64 `json:"temp_c"`
	} `json:"current"`
}

func (c *Client) CurrentCelsius(ctx context.Context, city string) (float64, error) {
	q := url.Values{"key": {c.APIKey}, "q": {city}, "aqi": {"no"}}
	endpoint := c.BaseURL + "/v1/current.json?" + q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return 0, errors.New("weather: build request")
	}
	resp, err := c.HTTP.Do(req)
	if err != nil {
		var ue *url.Error
		if errors.As(err, &ue) {
			err = ue.Err
		}
		return 0, fmt.Errorf("weather: request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("weather: unexpected status %d", resp.StatusCode)
	}
	var r response
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return 0, fmt.Errorf("weather: decode: %w", err)
	}
	if r.Current == nil {
		return 0, errors.New("weather: response without current data")
	}
	return r.Current.TempC, nil
}
