package o11y

import (
	"context"
	"time"

	"gitlab.com/neonmei/wesense_exporter/model"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/metric/global"
	controller "go.opentelemetry.io/otel/sdk/metric/controller/basic"
	processor "go.opentelemetry.io/otel/sdk/metric/processor/basic"
	"go.opentelemetry.io/otel/sdk/metric/selector/simple"

	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
)

// Take info from ConfigSpec and into resource.
func newResource(config *model.Config) *resource.Resource {
	return resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceNameKey.String(config.Instance),
		semconv.ServiceVersionKey.String(config.Version),
	)
}

func InitTracerMetrics(config *model.Config) {
	grpcClient := otlpmetricgrpc.NewClient()
	metricsExporter, err := otlpmetric.New(context.Background(), grpcClient)
	if err != nil {
		panic(err.Error())
	}

	pusherController := controller.New(
		processor.NewFactory(
			simple.NewWithHistogramDistribution(),
			metricsExporter,
		),
		controller.WithExporter(metricsExporter),
		controller.WithCollectPeriod(time.Second*time.Duration(config.Interval)),
		controller.WithResource(newResource(config)),
	)

	global.SetMeterProvider(pusherController)
	pusherController.Start(context.Background())
}
