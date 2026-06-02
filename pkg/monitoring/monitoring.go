package monitoring

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
)

const meterName = "mongodb-migrator"

var (
	commonMeter metric.Meter
)

// Init initializes the global OpenTelemetry MeterProvider and sets up the common meter.
func Init(ctx context.Context) (func(), error) {
	// Create a basic MeterProvider. In this initial implementation, we setup a basic
	// MeterProvider without additional exporters (it acts as a basic/noop metric collector
	// until readers/exporters are configured).
	provider := sdkmetric.NewMeterProvider()

	otel.SetMeterProvider(provider)

	// Initialize the common meter
	commonMeter = provider.Meter(meterName)

	shutdown := func() {
		if err := provider.Shutdown(ctx); err != nil {
			otel.Handle(err)
		}
	}

	return shutdown, nil
}

// Meter returns the common meter instance.
func Meter() metric.Meter {
	if commonMeter == nil {
		// Fallback to global meter if not initialized yet
		return otel.Meter(meterName)
	}
	return commonMeter
}

// NewDocumentsReadCounter initializes and returns the documents_read_total counter.
func NewDocumentsReadCounter() (metric.Int64Counter, error) {
	return Meter().Int64Counter("documents_read_total",
		metric.WithDescription("Total number of documents read during backfill"),
		metric.WithUnit("{document}"),
	)
}

