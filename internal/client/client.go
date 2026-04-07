package client

import (
	"context"

	"github.com/techinpark/ga-cli/internal/model"
)

// AdminClient defines operations for the GA4 Admin API.
type AdminClient interface {
	ListProperties(ctx context.Context) ([]model.Property, error)
}

// DataClient defines operations for the GA4 Data API.
type DataClient interface {
	GetDAU(ctx context.Context, propertyID string, days int) ([]model.DAURecord, error)
	GetEvents(ctx context.Context, propertyID string, days int, top int) ([]model.EventRecord, error)
	GetCountries(ctx context.Context, propertyID string, days int) ([]model.CountryRecord, error)
	GetPlatforms(ctx context.Context, propertyID string, days int) ([]model.PlatformRecord, error)
	GetRealtime(ctx context.Context, propertyID string) (*model.RealtimeReport, error)
}
