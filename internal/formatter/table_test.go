package formatter

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/techinpark/ga-cli/internal/model"
)

func TestFormatNumber64(t *testing.T) {
	tests := []struct {
		name  string
		input int64
		want  string
	}{
		{"zero", 0, "0"},
		{"below thousand", 999, "999"},
		{"exactly thousand", 1000, "1,000"},
		{"millions", 1234567, "1,234,567"},
		{"negative", -1234, "-1,234"},
		{"large number", 1000000000, "1,000,000,000"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatNumber64(tt.input)
			if got != tt.want {
				t.Errorf("formatNumber64(%d) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestFormatNumber(t *testing.T) {
	tests := []struct {
		name  string
		input int
		want  string
	}{
		{"delegates to formatNumber64", 1234, "1,234"},
		{"zero", 0, "0"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatNumber(tt.input)
			if got != tt.want {
				t.Errorf("formatNumber(%d) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestFormatPercent(t *testing.T) {
	tests := []struct {
		name  string
		input float64
		want  string
	}{
		{"positive", 5.5, "+5.5%"},
		{"negative", -3.2, "-3.2%"},
		{"zero", 0.0, "0.0%"},
		{"large positive", 100.0, "+100.0%"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatPercent(tt.input)
			if got != tt.want {
				t.Errorf("formatPercent(%f) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestTableFormatProperties(t *testing.T) {
	f := &tableFormatter{}
	properties := []model.Property{
		{ID: "123456", DisplayName: "My App", Alias: "app"},
		{ID: "789012", DisplayName: "My Site", Alias: "site"},
	}

	var buf bytes.Buffer
	err := f.FormatProperties(&buf, "Properties", properties)
	if err != nil {
		t.Fatalf("FormatProperties returned error: %v", err)
	}

	output := buf.String()
	for _, want := range []string{"ALIAS", "PROPERTY ID", "DISPLAY NAME", "app", "123456", "My App", "site", "789012", "My Site"} {
		if !strings.Contains(output, want) {
			t.Errorf("output missing %q\noutput:\n%s", want, output)
		}
	}
}

func TestTableFormatDAU(t *testing.T) {
	f := &tableFormatter{}
	records := []model.DAURecord{
		{Date: "2026-01-01", ActiveUsers: 1500, ChangePercent: 5.0},
		{Date: "2026-01-02", ActiveUsers: 2500, ChangePercent: -3.0},
	}

	var buf bytes.Buffer
	err := f.FormatDAU(&buf, "DAU Report", records)
	if err != nil {
		t.Fatalf("FormatDAU returned error: %v", err)
	}

	output := buf.String()
	for _, want := range []string{"DAU Report", "1,500", "2,500", "Avg:"} {
		if !strings.Contains(output, want) {
			t.Errorf("output missing %q\noutput:\n%s", want, output)
		}
	}
}

func TestTableFormatRealtime(t *testing.T) {
	f := &tableFormatter{}
	report := &model.RealtimeReport{
		ActiveUsers: 42,
		Events: []model.RealtimeEvent{
			{EventName: "page_view", EventCount: 100},
		},
	}

	var buf bytes.Buffer
	err := f.FormatRealtime(&buf, "Realtime", report)
	if err != nil {
		t.Fatalf("FormatRealtime returned error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Active Users:") {
		t.Errorf("output missing 'Active Users:'\noutput:\n%s", output)
	}
	if !strings.Contains(output, "42") {
		t.Errorf("output missing active user count\noutput:\n%s", output)
	}
}

func TestTableFormatDAUSummary(t *testing.T) {
	f := &tableFormatter{}
	summaries := []model.DAUSummary{
		{PropertyName: "My App", PropertyID: "123", ActiveUsers: 5000},
		{PropertyName: "My Site", PropertyID: "456", ActiveUsers: 3000},
	}

	var buf bytes.Buffer
	err := f.FormatDAUSummary(&buf, summaries)
	if err != nil {
		t.Fatalf("FormatDAUSummary returned error: %v", err)
	}

	output := buf.String()
	for _, want := range []string{"PROPERTY", "DAU (TODAY)", "My App (123)", "5,000", "My Site (456)", "3,000"} {
		if !strings.Contains(output, want) {
			t.Errorf("output missing %q\noutput:\n%s", want, output)
		}
	}
}

// --- JSON formatter: FormatProperties ---

func TestJSONFormatProperties(t *testing.T) {
	f := &jsonFormatter{}
	properties := []model.Property{
		{ID: "123", DisplayName: "Test App", Alias: "test"},
	}

	var buf bytes.Buffer
	err := f.FormatProperties(&buf, "ignored title", properties)
	if err != nil {
		t.Fatalf("FormatProperties returned error: %v", err)
	}

	output := buf.Bytes()
	if !json.Valid(output) {
		t.Fatalf("output is not valid JSON:\n%s", output)
	}

	outputStr := string(output)
	for _, want := range []string{"123", "Test App", "test"} {
		if !strings.Contains(outputStr, want) {
			t.Errorf("JSON output missing %q", want)
		}
	}
}

func TestJSONFormatDAU(t *testing.T) {
	f := &jsonFormatter{}
	records := []model.DAURecord{
		{Date: "2026-01-01", ActiveUsers: 100, ChangePercent: 2.5},
	}

	var buf bytes.Buffer
	err := f.FormatDAU(&buf, "ignored", records)
	if err != nil {
		t.Fatalf("FormatDAU returned error: %v", err)
	}

	if !json.Valid(buf.Bytes()) {
		t.Fatalf("output is not valid JSON:\n%s", buf.String())
	}
}

// --- CSV formatter: FormatProperties ---

func TestCSVFormatProperties(t *testing.T) {
	f := &csvFormatter{}
	properties := []model.Property{
		{ID: "123", DisplayName: "Test App", Alias: "test"},
		{ID: "456", DisplayName: "Other App", Alias: "other"},
	}

	var buf bytes.Buffer
	err := f.FormatProperties(&buf, "ignored", properties)
	if err != nil {
		t.Fatalf("FormatProperties returned error: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) < 3 {
		t.Fatalf("expected at least 3 lines (header + 2 rows), got %d", len(lines))
	}

	// header
	if !strings.Contains(lines[0], "alias") || !strings.Contains(lines[0], "property_id") || !strings.Contains(lines[0], "display_name") {
		t.Errorf("header row missing expected columns: %s", lines[0])
	}

	// data rows
	if !strings.Contains(lines[1], "test") || !strings.Contains(lines[1], "123") {
		t.Errorf("first data row missing expected values: %s", lines[1])
	}
	if !strings.Contains(lines[2], "other") || !strings.Contains(lines[2], "456") {
		t.Errorf("second data row missing expected values: %s", lines[2])
	}
}

func TestCSVFormatDAU(t *testing.T) {
	f := &csvFormatter{}
	records := []model.DAURecord{
		{Date: "2026-01-01", ActiveUsers: 500, ChangePercent: 1.5},
	}

	var buf bytes.Buffer
	err := f.FormatDAU(&buf, "ignored", records)
	if err != nil {
		t.Fatalf("FormatDAU returned error: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) < 2 {
		t.Fatalf("expected at least 2 lines, got %d", len(lines))
	}

	if !strings.Contains(lines[0], "date") || !strings.Contains(lines[0], "active_users") {
		t.Errorf("header row missing expected columns: %s", lines[0])
	}
	if !strings.Contains(lines[1], "2026-01-01") || !strings.Contains(lines[1], "500") {
		t.Errorf("data row missing expected values: %s", lines[1])
	}
}
