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
	ot "go.opentelemetry.io/otel/trace"
)

var (
	tp *trace.TracerProvider
	tr ot.Tracer
)

type TraceConfig struct {
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

	// set tracer name
	tr = otel.Tracer(conf.ServiceName)

	// set trace provider
	tp = trace.NewTracerProvider(
		trace.WithBatcher(exporter),
		trace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			setupAttributes(conf)...,
		)),
	)

	pp := b3.New(b3.WithInjectEncoding(b3.B3MultipleHeader | b3.B3SingleHeader))

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(pp)

	return tp
}

func Configure() *trace.TracerProvider {
	return ConfigureWithConfig(TraceConfig{
		AgentHost: "localhost",
		AgentPort: "6831",
	})
}

func setupAttributes(conf TraceConfig) []attribute.KeyValue {
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
