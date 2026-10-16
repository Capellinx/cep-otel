package serviceb_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/capella/cep-otel/internal/serviceb"
	"github.com/capella/cep-otel/internal/viacep"
)

type fakeLocations struct {
	city string
	err  error
}

func (f fakeLocations) FindCity(ctx context.Context, cep string) (string, error) {
	return f.city, f.err
}

type fakeTemps struct {
	celsius float64
	err     error
}

func (f fakeTemps) CurrentCelsius(ctx context.Context, city string) (float64, error) {
	return f.celsius, f.err
}

func do(h http.Handler, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func TestServiceBInvalidZipcode(t *testing.T) {
	h := serviceb.New(fakeLocations{}, fakeTemps{})
	rec := do(h, `{"cep":"123"}`)
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("esperava 422, obteve %d", rec.Code)
	}
	if strings.TrimSpace(rec.Body.String()) != "invalid zipcode" {
		t.Fatalf("corpo inesperado: %q", rec.Body.String())
	}
}

func TestServiceBZipcodeNotFound(t *testing.T) {
	h := serviceb.New(fakeLocations{err: viacep.ErrNotFound}, fakeTemps{})
	rec := do(h, `{"cep":"99999999"}`)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("esperava 404, obteve %d", rec.Code)
	}
	if strings.TrimSpace(rec.Body.String()) != "can not find zipcode" {
		t.Fatalf("corpo inesperado: %q", rec.Body.String())
	}
}

func TestServiceBSuccess(t *testing.T) {
	h := serviceb.New(fakeLocations{city: "São Paulo"}, fakeTemps{celsius: 28.5})
	rec := do(h, `{"cep":"29902555"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("esperava 200, obteve %d", rec.Code)
	}
	var out struct {
		City  string  `json:"city"`
		TempC float64 `json:"temp_C"`
		TempF float64 `json:"temp_F"`
		TempK float64 `json:"temp_K"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("erro ao decodificar: %v", err)
	}
	if out.City != "São Paulo" || out.TempC != 28.5 || out.TempF != 83.3 || out.TempK != 301.5 {
		t.Fatalf("resposta inesperada: %+v", out)
	}
}
