package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	"github.com/capella/cep-otel/internal/observability"
	"github.com/capella/cep-otel/internal/servicea"
)

func main() {
	ctx := context.Background()

	shutdown, err := observability.InitTracer(ctx, env("OTEL_SERVICE_NAME", "service-a"), env("OTEL_EXPORTER_OTLP_ENDPOINT", "otel-collector:4318"))
	if err != nil {
		log.Fatalf("erro ao inicializar OpenTelemetry: %v", err)
	}
	defer func() { _ = shutdown(context.Background()) }()

	handler := otelhttp.NewHandler(servicea.New(env("SERVICE_B_URL", "http://service-b:8081/")), "service-a")

	port := env("HTTP_PORT", "8080")
	server := &http.Server{
		Addr:         ":" + port,
		Handler:      handler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	log.Printf("Serviço A (input) ouvindo na porta %s", port)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("erro no servidor: %v", err)
	}
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
