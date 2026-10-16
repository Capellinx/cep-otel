package servicea

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"regexp"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
)

var cepPattern = regexp.MustCompile(`^\d{8}$`)

type request struct {
	Cep string `json:"cep"`
}

type Handler struct {
	serviceBURL string
	client      *http.Client
}

func New(serviceBURL string) http.Handler {
	h := &Handler{
		serviceBURL: serviceBURL,
		client:      &http.Client{Transport: otelhttp.NewTransport(http.DefaultTransport)},
	}
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

	ctx, span := otel.Tracer("service-a").Start(r.Context(), "forward-to-service-b")
	defer span.End()

	body, _ := json.Marshal(req)
	outReq, err := http.NewRequestWithContext(ctx, http.MethodPost, h.serviceBURL, bytes.NewReader(body))
	if err != nil {
		writeText(w, http.StatusInternalServerError, "internal error")
		return
	}
	outReq.Header.Set("Content-Type", "application/json")

	resp, err := h.client.Do(outReq)
	if err != nil {
		writeText(w, http.StatusBadGateway, "error contacting service B")
		return
	}
	defer resp.Body.Close()

	w.Header().Set("Content-Type", resp.Header.Get("Content-Type"))
	w.WriteHeader(resp.StatusCode)
	_, _ = io.Copy(w, resp.Body)
}

func writeText(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(status)
	_, _ = w.Write([]byte(msg))
}
