package formatter

import (
	"testing"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name     string
		format   Format
		wantType string
	}{
		{"table format", FormatTable, "*formatter.tableFormatter"},
		{"json format", FormatJSON, "*formatter.jsonFormatter"},
		{"csv format", FormatCSV, "*formatter.csvFormatter"},
		{"unknown defaults to table", Format("unknown"), "*formatter.tableFormatter"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := New(tt.format)
			got := typeString(f)
			if got != tt.wantType {
				t.Errorf("New(%q) returned %s, want %s", tt.format, got, tt.wantType)
			}
		})
	}
}

func typeString(f Formatter) string {
	switch f.(type) {
	case *tableFormatter:
		return "*formatter.tableFormatter"
	case *jsonFormatter:
		return "*formatter.jsonFormatter"
	case *csvFormatter:
		return "*formatter.csvFormatter"
	default:
		return "unknown"
	}
}

func TestParseFormat(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    Format
		wantErr bool
	}{
		{"table", "table", FormatTable, false},
		{"json", "json", FormatJSON, false},
		{"csv", "csv", FormatCSV, false},
		{"xml unsupported", "xml", "", true},
		{"empty string", "", "", true},
		{"yaml unsupported", "yaml", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseFormat(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseFormat(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ParseFormat(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
