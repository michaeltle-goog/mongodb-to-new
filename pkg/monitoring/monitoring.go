package monitoring

import (
	"context"
	"fmt"
	"log"
	"time"

	mexporter "github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/metric"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
)

const meterName = "mongodb-migrator"

var (
	commonMeter metric.Meter
)

// Init initializes the global OpenTelemetry MeterProvider and sets up the common meter.
// If gcpProjectID is provided, it exports metrics to Google Cloud Monitoring.
func Init(ctx context.Context, gcpProjectID string) (func(), error) {
	var provider *sdkmetric.MeterProvider

	if gcpProjectID != "" {
		// Initialize the GCP Monitoring Exporter
		exporter, err := mexporter.New(mexporter.WithProjectID(gcpProjectID))
		if err != nil {
			return nil, fmt.Errorf("failed to create GCP metric exporter: %w", err)
		}

		// Create the Meter Provider with the GCP exporter
		provider = sdkmetric.NewMeterProvider(
			sdkmetric.WithReader(sdkmetric.NewPeriodicReader(exporter)),
		)
		log.Printf("OpenTelemetry monitoring initialized (exporting to GCP project %q)", gcpProjectID)
	} else {
		// Create a basic MeterProvider without exporters (acts as a basic/noop metric collector)
		provider = sdkmetric.NewMeterProvider()
		log.Println("OpenTelemetry monitoring initialized (local/no-op mode)")
	}

	otel.SetMeterProvider(provider)

	// Initialize the common meter
	commonMeter = provider.Meter(meterName)

	shutdown := func() {
		log.Println("OpenTelemetry monitoring shutting down and flushing metrics...")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := provider.Shutdown(shutdownCtx); err != nil {
			otel.Handle(err)
			log.Printf("Failed to shut down OpenTelemetry monitoring: %v", err)
		} else {
			log.Println("OpenTelemetry monitoring successfully shut down and metrics flushed")
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

