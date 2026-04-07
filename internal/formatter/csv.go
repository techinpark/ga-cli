package formatter

import (
	"encoding/csv"
	"fmt"
	"io"
	"strconv"

	"github.com/techinpark/ga-cli/internal/model"
)

type csvFormatter struct{}

func writeCSV(w io.Writer, headers []string, rows [][]string) error {
	cw := csv.NewWriter(w)
	if err := cw.Write(headers); err != nil {
		return fmt.Errorf("failed to write CSV header: %w", err)
	}
	for _, row := range rows {
		if err := cw.Write(row); err != nil {
			return fmt.Errorf("failed to write CSV row: %w", err)
		}
	}
	cw.Flush()
	if err := cw.Error(); err != nil {
		return fmt.Errorf("failed to flush CSV: %w", err)
	}
	return nil
}

func (f *csvFormatter) FormatProperties(w io.Writer, _ string, properties []model.Property) error {
	headers := []string{"alias", "property_id", "display_name"}
	rows := make([][]string, len(properties))
	for i, p := range properties {
		rows[i] = []string{p.Alias, p.ID, p.DisplayName}
	}
	return writeCSV(w, headers, rows)
}

func (f *csvFormatter) FormatDAU(w io.Writer, _ string, records []model.DAURecord) error {
	headers := []string{"date", "active_users", "change_percent"}
	rows := make([][]string, len(records))
	for i, r := range records {
		rows[i] = []string{
			r.Date,
			strconv.Itoa(r.ActiveUsers),
			strconv.FormatFloat(r.ChangePercent, 'f', 1, 64),
		}
	}
	return writeCSV(w, headers, rows)
}

func (f *csvFormatter) FormatDAUSummary(w io.Writer, summaries []model.DAUSummary) error {
	headers := []string{"property_name", "property_id", "active_users"}
	rows := make([][]string, len(summaries))
	for i, s := range summaries {
		rows[i] = []string{
			s.PropertyName,
			s.PropertyID,
			strconv.Itoa(s.ActiveUsers),
		}
	}
	return writeCSV(w, headers, rows)
}

func (f *csvFormatter) FormatEvents(w io.Writer, _ string, records []model.EventRecord) error {
	headers := []string{"event_name", "event_count", "total_users"}
	rows := make([][]string, len(records))
	for i, r := range records {
		rows[i] = []string{
			r.EventName,
			strconv.FormatInt(r.EventCount, 10),
			strconv.Itoa(r.TotalUsers),
		}
	}
	return writeCSV(w, headers, rows)
}

func (f *csvFormatter) FormatCountries(w io.Writer, _ string, records []model.CountryRecord) error {
	headers := []string{"country", "active_users", "sessions", "screen_page_views_per_user"}
	rows := make([][]string, len(records))
	for i, r := range records {
		rows[i] = []string{
			r.Country,
			strconv.Itoa(r.ActiveUsers),
			strconv.Itoa(r.Sessions),
			strconv.FormatFloat(r.ScreenPageViewsPerUser, 'f', 1, 64),
		}
	}
	return writeCSV(w, headers, rows)
}

func (f *csvFormatter) FormatPlatforms(w io.Writer, _ string, records []model.PlatformRecord) error {
	headers := []string{"platform", "active_users", "sessions", "engaged_sessions", "engagement_rate", "sessions_per_user"}
	rows := make([][]string, len(records))
	for i, r := range records {
		rows[i] = []string{
			r.Platform,
			strconv.Itoa(r.ActiveUsers),
			strconv.Itoa(r.Sessions),
			strconv.Itoa(r.EngagedSessions),
			strconv.FormatFloat(r.EngagementRate, 'f', 4, 64),
			strconv.FormatFloat(r.SessionsPerUser, 'f', 2, 64),
		}
	}
	return writeCSV(w, headers, rows)
}

func (f *csvFormatter) FormatRealtime(w io.Writer, _ string, report *model.RealtimeReport) error {
	headers := []string{"event_name", "event_count"}
	rows := make([][]string, len(report.Events))
	for i, e := range report.Events {
		rows[i] = []string{
			e.EventName,
			strconv.Itoa(e.EventCount),
		}
	}
	return writeCSV(w, headers, rows)
}
