package tracero

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/contrib/propagators/b3"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	ot "go.opentelemetry.io/otel/trace"
)

var (
	tp *trace.TracerProvider
	tr ot.Tracer
	pr propagation.TextMapPropagator
)

type Config struct {
	AgentHost       string
	AgentPort       string
	ServiceName     string
	ServiceEnv      string
	ServiceVersion  string
	TraceAttributes []attribute.KeyValue
}

func Tracer() ot.Tracer {
	return tr
}

func Connect(ctx context.Context, conf Config) *trace.TracerProvider {
	exporter, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithEndpoint(fmt.Sprintf("%s:%s", conf.AgentHost, conf.AgentPort)),
		otlptracegrpc.WithInsecure(),
		otlptracegrpc.WithCompressor("gzip"),
	)
	if err != nil {
		logrus.Fatalf("unable to create otel client : %s", err.Error())
	}

	// set trace provider
	tp = trace.NewTracerProvider(
		trace.WithBatcher(exporter),
		trace.WithResource(resource.NewWithAttributes(semconv.SchemaURL, setupAttributes(conf)...)),
	)

	// set tracer name
	tr = tp.Tracer(conf.ServiceName)
	pr = b3.New(b3.WithInjectEncoding(b3.B3MultipleHeader))

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(pr)

	return tp
}

func setupAttributes(conf Config) []attribute.KeyValue {
	if conf.ServiceName == "" {
		logrus.Panic("trace service name is empty")
	} else {
		conf.TraceAttributes = append(conf.TraceAttributes, semconv.ServiceNameKey.String(conf.ServiceName))
	}

	if conf.ServiceEnv != "" {
		conf.TraceAttributes = append(conf.TraceAttributes, attribute.String("env", conf.ServiceEnv))
	}

	if conf.ServiceVersion != "" {
		conf.TraceAttributes = append(conf.TraceAttributes, attribute.String("version", conf.ServiceVersion))
	}

	return conf.TraceAttributes
}
