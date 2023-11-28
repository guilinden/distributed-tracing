package main

import (
	"context"
	"log"
	"net/http"
	"fmt"
	"os"
	"io"
	"github.com/labstack/echo/v4"
	"time"
	"go.opentelemetry.io/contrib/instrumentation/github.com/labstack/echo/otelecho"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	oteltrace "go.opentelemetry.io/otel/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
)

var tracer = otel.Tracer("echo-server-mainApp")

func main() {
	ctx := context.Background()
	tp, err := initTracer(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			log.Printf("Error shutting down tracer provider: %v", err)
		}
	}()
	r := echo.New()
	r.Use(otelecho.Middleware("echo-server-mainapp"))

	r.GET("/users-delay/:id", func(c echo.Context) error {
		id := c.Param("id")
		userData := searchUser(c.Request().Context(), id, false)
		return c.JSON(http.StatusOK, userData)
	})

	r.GET("/users-delay-background/:id", func(c echo.Context) error {
		id := c.Param("id")
		userData := searchUser(c.Request().Context(), id, true)
		return c.JSON(http.StatusOK, userData)
	})
	_ = r.Start(":8080")
}

func initTracer(ctx context.Context) (*sdktrace.TracerProvider, error) {
	exporter, err := otlptracehttp.New(ctx,otlptracehttp.WithInsecure())
	if err != nil {
		return nil, err
	}
	providerResource, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName("mainapp"),
		),
	)
	if err != nil {
		return nil, err
	}
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(providerResource),
	)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))
	return tp, nil
}

func searchUser(ctx context.Context, id string, concurrentDelay bool) string {
	_, span := tracer.Start(ctx, "searchUser", oteltrace.WithAttributes(attribute.String("id", id)))
	defer span.End()
	if concurrentDelay {
		go delay(ctx,2)
	} else {
		delay(ctx,2)
	}
	req, _ := http.NewRequestWithContext(ctx,"GET","http://searchapp:8081/users/" + id,nil)
	client := &http.Client{Transport: otelhttp.NewTransport(http.DefaultTransport)}
	res, err := client.Do(req)
	if err != nil {
		fmt.Printf("error making http request: %s\n", err)
		os.Exit(1)
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	fmt.Println(string(body))
	return string(body)
}

func delay(ctx context.Context, timeInSeconds time.Duration) {
	_, span := tracer.Start(ctx, "delay", oteltrace.WithAttributes(attribute.String("delay",timeInSeconds.String())))
	defer span.End()
	time.Sleep(timeInSeconds * time.Second)
}