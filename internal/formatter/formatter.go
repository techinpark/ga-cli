package formatter

import (
	"fmt"
	"io"

	"github.com/techinpark/ga-cli/internal/model"
)

// Format represents an output format type.
type Format string

const (
	FormatTable Format = "table"
	FormatJSON  Format = "json"
	FormatCSV   Format = "csv"
)

// Formatter defines the interface for formatting analytics data.
type Formatter interface {
	FormatProperties(w io.Writer, title string, properties []model.Property) error
	FormatDAU(w io.Writer, title string, records []model.DAURecord) error
	FormatDAUSummary(w io.Writer, summaries []model.DAUSummary) error
	FormatEvents(w io.Writer, title string, records []model.EventRecord) error
	FormatCountries(w io.Writer, title string, records []model.CountryRecord) error
	FormatPlatforms(w io.Writer, title string, records []model.PlatformRecord) error
	FormatRealtime(w io.Writer, title string, report *model.RealtimeReport) error
}

// New creates a Formatter for the given format.
func New(format Format) Formatter {
	switch format {
	case FormatJSON:
		return &jsonFormatter{}
	case FormatCSV:
		return &csvFormatter{}
	default:
		return &tableFormatter{}
	}
}

// ParseFormat converts a string to a Format value.
func ParseFormat(s string) (Format, error) {
	switch s {
	case "table":
		return FormatTable, nil
	case "json":
		return FormatJSON, nil
	case "csv":
		return FormatCSV, nil
	default:
		return "", fmt.Errorf("unsupported format: %s (supported: table, json, csv)", s)
	}
}
