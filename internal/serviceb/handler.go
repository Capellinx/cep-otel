package serviceb

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"regexp"

	"go.opentelemetry.io/otel"

	"github.com/capella/cep-otel/internal/temperature"
	"github.com/capella/cep-otel/internal/viacep"
)

type LocationFinder interface {
	FindCity(ctx context.Context, cep string) (string, error)
}

type TemperatureFinder interface {
	CurrentCelsius(ctx context.Context, city string) (float64, error)
}

var cepPattern = regexp.MustCompile(`^\d{8}$`)

type request struct {
	Cep string `json:"cep"`
}

type successResponse struct {
	City  string  `json:"city"`
	TempC float64 `json:"temp_C"`
	TempF float64 `json:"temp_F"`
	TempK float64 `json:"temp_K"`
}

type Handler struct {
	locations LocationFinder
	temps     TemperatureFinder
}

func New(l LocationFinder, t TemperatureFinder) http.Handler {
	h := &Handler{locations: l, temps: t}
	mux := http.NewServeMux()
	mux.HandleFunc("POST /", h.handle)
	return mux
}

func (h *Handler) handle(w http.ResponseWriter, r *http.Request) {
	var req request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeText(w, http.StatusUnprocessableEntity, "invalid zipcode")
		return
	}
	if !cepPattern.MatchString(req.Cep) {
		writeText(w, http.StatusUnprocessableEntity, "invalid zipcode")
		return
	}

	tr := otel.Tracer("service-b")

	ctxCep, spanCep := tr.Start(r.Context(), "busca-cep-localizacao")
	city, err := h.locations.FindCity(ctxCep, req.Cep)
	spanCep.End()
	if errors.Is(err, viacep.ErrNotFound) {
		writeText(w, http.StatusNotFound, "can not find zipcode")
		return
	}
	if err != nil {
		upstream(w, err)
		return
	}

	ctxTemp, spanTemp := tr.Start(r.Context(), "busca-temperatura-clima")
	celsius, err := h.temps.CurrentCelsius(ctxTemp, city)
	spanTemp.End()
	if err != nil {
		upstream(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(successResponse{
		City:  city,
		TempC: celsius,
		TempF: temperature.CelsiusToFahrenheit(celsius),
		TempK: temperature.CelsiusToKelvin(celsius),
	})
}

func upstream(w http.ResponseWriter, err error) {
	log.Printf("upstream error: %v", err)
	writeText(w, http.StatusBadGateway, "upstream error")
}

func writeText(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(status)
	_, _ = w.Write([]byte(msg))
}
