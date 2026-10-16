package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	"github.com/capella/cep-otel/internal/observability"
	"github.com/capella/cep-otel/internal/serviceb"
	"github.com/capella/cep-otel/internal/viacep"
	"github.com/capella/cep-otel/internal/weather"
)

func main() {
	ctx := context.Background()

	shutdown, err := observability.InitTracer(ctx, env("OTEL_SERVICE_NAME", "service-b"), env("OTEL_EXPORTER_OTLP_ENDPOINT", "otel-collector:4318"))
	if err != nil {
		log.Fatalf("erro ao inicializar OpenTelemetry: %v", err)
	}
	defer func() { _ = shutdown(context.Background()) }()

	apiKey := os.Getenv("WEATHER_API_KEY")
	if apiKey == "" {
		log.Println("aviso: WEATHER_API_KEY não definida; a consulta de clima falhará")
	}

	handler := otelhttp.NewHandler(
		serviceb.New(viacep.NewClient(), weather.NewClient(apiKey)),
		"service-b",
	)

	port := env("HTTP_PORT", "8081")
	server := &http.Server{
		Addr:         ":" + port,
		Handler:      handler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	log.Printf("Serviço B (orquestração) ouvindo na porta %s", port)
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
