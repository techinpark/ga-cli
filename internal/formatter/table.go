package formatter

import (
	"fmt"
	"io"
	"strings"

	"github.com/techinpark/ga-cli/internal/model"
)

type tableFormatter struct{}

// formatNumber formats an integer with thousand separators.
func formatNumber(n int) string {
	return formatNumber64(int64(n))
}

// formatNumber64 formats an int64 with thousand separators.
func formatNumber64(n int64) string {
	if n < 0 {
		return "-" + formatNumber64(-n)
	}

	s := fmt.Sprintf("%d", n)
	if len(s) <= 3 {
		return s
	}

	var result strings.Builder
	remainder := len(s) % 3
	if remainder > 0 {
		result.WriteString(s[:remainder])
	}
	for i := remainder; i < len(s); i += 3 {
		if result.Len() > 0 {
			result.WriteByte(',')
		}
		result.WriteString(s[i : i+3])
	}
	return result.String()
}

// formatPercent formats a percentage with +/- sign.
func formatPercent(p float64) string {
	if p > 0 {
		return fmt.Sprintf("+%.1f%%", p)
	}
	if p < 0 {
		return fmt.Sprintf("%.1f%%", p)
	}
	return "0.0%"
}

// renderTable writes a formatted table to the writer.
// It auto-calculates column widths and renders without borders.
func renderTable(w io.Writer, headers []string, rows [][]string) {
	if len(headers) == 0 {
		return
	}

	// Calculate column widths
	widths := make([]int, len(headers))
	for i, h := range headers {
		widths[i] = len(h)
	}
	for _, row := range rows {
		for i, cell := range row {
			if i < len(widths) && len(cell) > widths[i] {
				widths[i] = len(cell)
			}
		}
	}

	// Build format string with right-padding
	writeRow := func(cells []string) {
		for i, cell := range cells {
			if i > 0 {
				fmt.Fprint(w, "  ")
			}
			if i < len(widths) {
				fmt.Fprintf(w, "%-*s", widths[i], cell)
			}
		}
		fmt.Fprintln(w)
	}

	// Write header
	writeRow(headers)

	// Write separator
	sep := make([]string, len(headers))
	for i, width := range widths {
		sep[i] = strings.Repeat("-", width)
	}
	writeRow(sep)

	// Write data rows
	for _, row := range rows {
		writeRow(row)
	}
}

func (f *tableFormatter) FormatProperties(w io.Writer, title string, properties []model.Property) error {
	fmt.Fprintf(w, "\n%s\n\n", title)

	headers := []string{"ALIAS", "PROPERTY ID", "DISPLAY NAME"}
	rows := make([][]string, len(properties))
	for i, p := range properties {
		rows[i] = []string{p.Alias, p.ID, p.DisplayName}
	}

	renderTable(w, headers, rows)
	return nil
}

func (f *tableFormatter) FormatDAU(w io.Writer, title string, records []model.DAURecord) error {
	fmt.Fprintf(w, "\n%s\n\n", title)

	headers := []string{"DATE", "DAU", "CHANGE"}
	rows := make([][]string, len(records))
	totalUsers := 0
	for i, r := range records {
		totalUsers += r.ActiveUsers
		rows[i] = []string{r.Date, formatNumber(r.ActiveUsers), formatPercent(r.ChangePercent)}
	}

	renderTable(w, headers, rows)

	// Print average
	if len(records) > 0 {
		avg := totalUsers / len(records)
		fmt.Fprintf(w, "\nAvg: %s\n", formatNumber(avg))
	}
	return nil
}

func (f *tableFormatter) FormatDAUSummary(w io.Writer, summaries []model.DAUSummary) error {
	headers := []string{"PROPERTY", "DAU (TODAY)"}
	rows := make([][]string, len(summaries))
	for i, s := range summaries {
		name := s.PropertyName
		if s.PropertyID != "" {
			name = fmt.Sprintf("%s (%s)", s.PropertyName, s.PropertyID)
		}
		rows[i] = []string{name, formatNumber(s.ActiveUsers)}
	}

	renderTable(w, headers, rows)
	return nil
}

func (f *tableFormatter) FormatEvents(w io.Writer, title string, records []model.EventRecord) error {
	fmt.Fprintf(w, "\n%s\n\n", title)

	headers := []string{"#", "EVENT", "COUNT", "USERS"}
	rows := make([][]string, len(records))
	for i, r := range records {
		rows[i] = []string{
			fmt.Sprintf("%d", i+1),
			r.EventName,
			formatNumber64(r.EventCount),
			formatNumber(r.TotalUsers),
		}
	}

	renderTable(w, headers, rows)
	return nil
}

func (f *tableFormatter) FormatCountries(w io.Writer, title string, records []model.CountryRecord) error {
	fmt.Fprintf(w, "\n%s\n\n", title)

	headers := []string{"#", "COUNTRY", "USERS", "SESSIONS", "VIEWS/USER"}
	rows := make([][]string, len(records))
	for i, r := range records {
		rows[i] = []string{
			fmt.Sprintf("%d", i+1),
			r.Country,
			formatNumber(r.ActiveUsers),
			formatNumber(r.Sessions),
			fmt.Sprintf("%.1f", r.ScreenPageViewsPerUser),
		}
	}

	renderTable(w, headers, rows)
	return nil
}

func (f *tableFormatter) FormatPlatforms(w io.Writer, title string, records []model.PlatformRecord) error {
	fmt.Fprintf(w, "\n%s\n\n", title)

	headers := []string{"PLATFORM", "USERS", "SESSIONS", "ENGAGED", "RATE", "SESSIONS/USER"}
	rows := make([][]string, len(records))
	for i, r := range records {
		rows[i] = []string{
			r.Platform,
			formatNumber(r.ActiveUsers),
			formatNumber(r.Sessions),
			formatNumber(r.EngagedSessions),
			fmt.Sprintf("%.1f%%", r.EngagementRate),
			fmt.Sprintf("%.2f", r.SessionsPerUser),
		}
	}

	renderTable(w, headers, rows)
	return nil
}

func (f *tableFormatter) FormatRealtime(w io.Writer, title string, report *model.RealtimeReport) error {
	fmt.Fprintf(w, "\n%s\n\n", title)
	fmt.Fprintf(w, "Active Users: %s\n\n", formatNumber(report.ActiveUsers))

	if len(report.Events) > 0 {
		headers := []string{"#", "EVENT", "COUNT"}
		rows := make([][]string, len(report.Events))
		for i, e := range report.Events {
			rows[i] = []string{
				fmt.Sprintf("%d", i+1),
				e.EventName,
				formatNumber(e.EventCount),
			}
		}
		renderTable(w, headers, rows)
	}
	return nil
}

func (f *tableFormatter) FormatCompare(w io.Writer, report *model.CompareReport) error {
	fmt.Fprintf(w, "\n%s - Comparison Report\n", strings.ToUpper(report.PropertyName))

	if len(report.DayOverDay) > 0 {
		fmt.Fprintf(w, "\nDay over Day (Today vs Yesterday)\n\n")
		headers := []string{"METRIC", "TODAY", "YESTERDAY", "CHANGE"}
		rows := make([][]string, len(report.DayOverDay))
		for i, r := range report.DayOverDay {
			rows[i] = []string{
				r.Metric,
				formatNumber64(r.Current),
				formatNumber64(r.Previous),
				formatPercent(r.ChangePercent),
			}
		}
		renderTable(w, headers, rows)
	}

	if len(report.WeekOverWeek) > 0 {
		fmt.Fprintf(w, "\nWeek over Week (This Week vs Last Week)\n\n")
		headers := []string{"METRIC", "THIS WEEK", "LAST WEEK", "CHANGE"}
		rows := make([][]string, len(report.WeekOverWeek))
		for i, r := range report.WeekOverWeek {
			rows[i] = []string{
				r.Metric,
				formatNumber64(r.Current),
				formatNumber64(r.Previous),
				formatPercent(r.ChangePercent),
			}
		}
		renderTable(w, headers, rows)
	}

	return nil
}
