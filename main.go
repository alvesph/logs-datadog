package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

var tracer trace.Tracer

func initTracer() func() {
	ctx := context.Background()

	traceExporter, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithEndpoint("otel-collector.observability.svc.cluster.local:4317"),
		otlptracegrpc.WithInsecure(),
	)
	if err != nil {
		log.Fatalf("Falha ao criar o exportador de trace: %v", err)
	}

	metricExporter, err := otlpmetricgrpc.New(ctx,
		otlpmetricgrpc.WithEndpoint("otel-collector.observability.svc.cluster.local:4317"),
		otlpmetricgrpc.WithInsecure(),
	)
	if err != nil {
		log.Fatalf("Falha ao criar o exportador de métricas: %v", err)
	}

	res, err := resource.New(ctx, resource.WithAttributes(
		attribute.String("service.name", "meu-servico-go"),
	))
	if err != nil {
		log.Fatalf("Falha ao criar recurso: %v", err)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(traceExporter),
		sdktrace.WithResource(res),
	)
	otel.SetTracerProvider(tp)

	tracer = otel.Tracer("meu-servico-go")

	mp := metric.NewMeterProvider(
		metric.WithReader(metric.NewPeriodicReader(metricExporter)),
		metric.WithResource(res),
	)

	log.SetOutput(os.Stdout)

	return func() {
		if err := tp.Shutdown(ctx); err != nil {
			log.Fatalf("Erro ao encerrar o tracer: %v", err)
		}
		if err := mp.Shutdown(ctx); err != nil {
			log.Fatalf("Erro ao encerrar o provedor de métricas: %v", err)
		}
	}
}

func logHTTPMethods(w http.ResponseWriter, r *http.Request) {

	logRequest(r.Context(), r)

	fmt.Fprintf(w, "Método %s recebido!", r.Method)
}

func logRequest(ctx context.Context, r *http.Request) {

	ctx, span := tracer.Start(ctx, "logHTTPMethods")
	defer span.End()

	log.Printf("Recebido método HTTP: %s\n", r.Method)
	span.AddEvent("Método HTTP recebido", trace.WithAttributes(attribute.String("http.method", r.Method)))
}

func logEvery5Seconds() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		uid := uuid.New()
		log.Printf("Log UID: %s\n", uid)
	}
}

func main() {

	cleanup := initTracer()
	defer cleanup()

	go logEvery5Seconds()

	http.HandleFunc("/", logHTTPMethods)

	port := ":8081"
	log.Printf("Servidor ouvindo na porta %s\n", port)

	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatalf("Erro ao iniciar o servidor: %s", err)
	}
}
