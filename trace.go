package tracero

import (
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/contrib/propagators/b3"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.20.0"
)

type TraceConfig struct {
	AgentHost      string
	AgentPort      string
	ServiceName    string
	ServiceEnv     string
	ServiceVersion string
}

func ConfigureWithConfig(conf TraceConfig) *trace.TracerProvider {
	exporter, err := jaeger.New(
		jaeger.WithAgentEndpoint(
			jaeger.WithAgentHost(conf.AgentHost),
			jaeger.WithAgentPort(conf.AgentPort),
		),
	)
	if err != nil {
		logrus.Fatalf("unable to create jaeger client : %s", err.Error())
	}

	tp := trace.NewTracerProvider(
		trace.WithBatcher(exporter),
		trace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(conf.ServiceName),
			attribute.String("env", conf.ServiceEnv),
			attribute.String("version", conf.ServiceVersion),
		)),
	)

	pp := b3.New(b3.WithInjectEncoding(b3.B3MultipleHeader | b3.B3SingleHeader))

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(pp)

	return tp
}

func Configure() *trace.TracerProvider {
	return ConfigureWithConfig(TraceConfig{
		AgentHost:      "localhost",
		AgentPort:      "6831",
		ServiceName:    "",
		ServiceEnv:     "",
		ServiceVersion: "",
	})
}
