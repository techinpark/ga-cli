package formatter

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/techinpark/ga-cli/internal/model"
)

type jsonFormatter struct{}

func writeJSON(w io.Writer, data any) error {
	b, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}
	_, err = fmt.Fprintf(w, "%s\n", b)
	if err != nil {
		return fmt.Errorf("failed to write JSON: %w", err)
	}
	return nil
}

func (f *jsonFormatter) FormatProperties(w io.Writer, _ string, properties []model.Property) error {
	return writeJSON(w, properties)
}

func (f *jsonFormatter) FormatDAU(w io.Writer, _ string, records []model.DAURecord) error {
	return writeJSON(w, records)
}

func (f *jsonFormatter) FormatDAUSummary(w io.Writer, summaries []model.DAUSummary) error {
	return writeJSON(w, summaries)
}

func (f *jsonFormatter) FormatEvents(w io.Writer, _ string, records []model.EventRecord) error {
	return writeJSON(w, records)
}

func (f *jsonFormatter) FormatCountries(w io.Writer, _ string, records []model.CountryRecord) error {
	return writeJSON(w, records)
}

func (f *jsonFormatter) FormatPlatforms(w io.Writer, _ string, records []model.PlatformRecord) error {
	return writeJSON(w, records)
}

func (f *jsonFormatter) FormatRealtime(w io.Writer, _ string, report *model.RealtimeReport) error {
	return writeJSON(w, report)
}
