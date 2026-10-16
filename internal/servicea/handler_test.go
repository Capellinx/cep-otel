package servicea_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/capella/cep-otel/internal/servicea"
)

func post(h http.Handler, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func TestServiceARejectsInvalidZipcode(t *testing.T) {
	called := false
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	}))
	defer backend.Close()

	h := servicea.New(backend.URL)

	for _, body := range []string{`{"cep":"123"}`, `{"cep":"abcdefgh"}`, `{"cep":12345678}`, `{"cep":"123456789"}`} {
		rec := post(h, body)
		if rec.Code != http.StatusUnprocessableEntity {
			t.Fatalf("payload %q: esperava 422, obteve %d", body, rec.Code)
		}
		if strings.TrimSpace(rec.Body.String()) != "invalid zipcode" {
			t.Fatalf("payload %q: corpo inesperado %q", body, rec.Body.String())
		}
	}
	if called {
		t.Fatal("Serviço B não deveria ser chamado para CEP inválido")
	}
}

func TestServiceAForwardsValidZipcode(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		if !strings.Contains(string(body), "29902555") {
			t.Errorf("Serviço B recebeu corpo inesperado: %s", body)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"city":"São Paulo","temp_C":28.5,"temp_F":83.3,"temp_K":301.5}`))
	}))
	defer backend.Close()

	h := servicea.New(backend.URL)
	rec := post(h, `{"cep":"29902555"}`)

	if rec.Code != http.StatusOK {
		t.Fatalf("esperava 200, obteve %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "São Paulo") {
		t.Fatalf("esperava repassar a resposta do Serviço B, obteve %q", rec.Body.String())
	}
}
