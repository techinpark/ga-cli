package client

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/techinpark/ga-cli/internal/model"
	"google.golang.org/api/analyticsdata/v1beta"
	"google.golang.org/api/option"
)

type dataClient struct {
	service *analyticsdata.Service
}

// NewDataClient creates a new Data API client.
// apiOpts configures authentication (e.g. WithCredentialsFile, WithTokenSource).
// If apiOpts is empty, Application Default Credentials (ADC) are used.
func NewDataClient(apiOpts []option.ClientOption) (DataClient, error) {
	ctx := context.Background()

	service, err := analyticsdata.NewService(ctx, apiOpts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create data service: %w", err)
	}

	return &dataClient{service: service}, nil
}

// propertyResource returns the "properties/{id}" format required by the API.
func propertyResource(propertyID string) string {
	return "properties/" + propertyID
}

// GetDAU returns daily active users for the given property over the last N days.
func (c *dataClient) GetDAU(ctx context.Context, propertyID string, days int) ([]model.DAURecord, error) {
	startDate := time.Now().AddDate(0, 0, -days).Format("2006-01-02")

	req := &analyticsdata.RunReportRequest{
		DateRanges: []*analyticsdata.DateRange{
			{StartDate: startDate, EndDate: "today"},
		},
		Dimensions: []*analyticsdata.Dimension{
			{Name: "date"},
		},
		Metrics: []*analyticsdata.Metric{
			{Name: "activeUsers"},
		},
		OrderBys: []*analyticsdata.OrderBy{
			{Dimension: &analyticsdata.DimensionOrderBy{DimensionName: "date"}},
		},
	}

	resp, err := c.service.Properties.RunReport(propertyResource(propertyID), req).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to run DAU report: %w", err)
	}

	records := make([]model.DAURecord, 0, len(resp.Rows))
	for _, row := range resp.Rows {
		date, err := formatGADate(row.DimensionValues[0].Value)
		if err != nil {
			return nil, fmt.Errorf("failed to parse date: %w", err)
		}

		activeUsers, err := strconv.Atoi(row.MetricValues[0].Value)
		if err != nil {
			return nil, fmt.Errorf("failed to parse activeUsers: %w", err)
		}

		records = append(records, model.DAURecord{
			Date:        date,
			ActiveUsers: activeUsers,
		})
	}

	// Calculate ChangePercent for each record
	for i := 1; i < len(records); i++ {
		prev := records[i-1].ActiveUsers
		if prev > 0 {
			records[i].ChangePercent = float64(records[i].ActiveUsers-prev) / float64(prev) * 100
		}
	}

	return records, nil
}

// GetEvents returns the top events for the given property over the last N days.
func (c *dataClient) GetEvents(ctx context.Context, propertyID string, days int, top int) ([]model.EventRecord, error) {
	startDate := time.Now().AddDate(0, 0, -days).Format("2006-01-02")

	req := &analyticsdata.RunReportRequest{
		DateRanges: []*analyticsdata.DateRange{
			{StartDate: startDate, EndDate: "today"},
		},
		Dimensions: []*analyticsdata.Dimension{
			{Name: "eventName"},
		},
		Metrics: []*analyticsdata.Metric{
			{Name: "eventCount"},
			{Name: "totalUsers"},
		},
		OrderBys: []*analyticsdata.OrderBy{
			{
				Desc:   true,
				Metric: &analyticsdata.MetricOrderBy{MetricName: "eventCount"},
			},
		},
		Limit: int64(top),
	}

	resp, err := c.service.Properties.RunReport(propertyResource(propertyID), req).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to run events report: %w", err)
	}

	records := make([]model.EventRecord, 0, len(resp.Rows))
	for _, row := range resp.Rows {
		eventCount, err := strconv.ParseInt(row.MetricValues[0].Value, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("failed to parse eventCount: %w", err)
		}

		totalUsers, err := strconv.Atoi(row.MetricValues[1].Value)
		if err != nil {
			return nil, fmt.Errorf("failed to parse totalUsers: %w", err)
		}

		records = append(records, model.EventRecord{
			EventName:  row.DimensionValues[0].Value,
			EventCount: eventCount,
			TotalUsers: totalUsers,
		})
	}

	return records, nil
}

// GetCountries returns active users and sessions by country.
func (c *dataClient) GetCountries(ctx context.Context, propertyID string, days int) ([]model.CountryRecord, error) {
	startDate := time.Now().AddDate(0, 0, -days).Format("2006-01-02")

	req := &analyticsdata.RunReportRequest{
		DateRanges: []*analyticsdata.DateRange{
			{StartDate: startDate, EndDate: "today"},
		},
		Dimensions: []*analyticsdata.Dimension{
			{Name: "country"},
		},
		Metrics: []*analyticsdata.Metric{
			{Name: "activeUsers"},
			{Name: "sessions"},
			{Name: "screenPageViews"},
		},
		OrderBys: []*analyticsdata.OrderBy{
			{
				Desc:   true,
				Metric: &analyticsdata.MetricOrderBy{MetricName: "activeUsers"},
			},
		},
	}

	resp, err := c.service.Properties.RunReport(propertyResource(propertyID), req).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to run countries report: %w", err)
	}

	records := make([]model.CountryRecord, 0, len(resp.Rows))
	for _, row := range resp.Rows {
		activeUsers, err := strconv.Atoi(row.MetricValues[0].Value)
		if err != nil {
			return nil, fmt.Errorf("failed to parse activeUsers: %w", err)
		}

		sessions, err := strconv.Atoi(row.MetricValues[1].Value)
		if err != nil {
			return nil, fmt.Errorf("failed to parse sessions: %w", err)
		}

		screenPageViews, err := strconv.Atoi(row.MetricValues[2].Value)
		if err != nil {
			return nil, fmt.Errorf("failed to parse screenPageViews: %w", err)
		}

		var viewsPerUser float64
		if activeUsers > 0 {
			viewsPerUser = float64(screenPageViews) / float64(activeUsers)
		}

		records = append(records, model.CountryRecord{
			Country:                row.DimensionValues[0].Value,
			ActiveUsers:            activeUsers,
			Sessions:               sessions,
			ScreenPageViewsPerUser: viewsPerUser,
		})
	}

	return records, nil
}

// GetPlatforms returns active users and sessions by platform.
func (c *dataClient) GetPlatforms(ctx context.Context, propertyID string, days int) ([]model.PlatformRecord, error) {
	startDate := time.Now().AddDate(0, 0, -days).Format("2006-01-02")

	req := &analyticsdata.RunReportRequest{
		DateRanges: []*analyticsdata.DateRange{
			{StartDate: startDate, EndDate: "today"},
		},
		Dimensions: []*analyticsdata.Dimension{
			{Name: "platform"},
		},
		Metrics: []*analyticsdata.Metric{
			{Name: "activeUsers"},
			{Name: "sessions"},
			{Name: "engagedSessions"},
		},
		OrderBys: []*analyticsdata.OrderBy{
			{
				Desc:   true,
				Metric: &analyticsdata.MetricOrderBy{MetricName: "activeUsers"},
			},
		},
	}

	resp, err := c.service.Properties.RunReport(propertyResource(propertyID), req).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to run platforms report: %w", err)
	}

	records := make([]model.PlatformRecord, 0, len(resp.Rows))
	for _, row := range resp.Rows {
		activeUsers, err := strconv.Atoi(row.MetricValues[0].Value)
		if err != nil {
			return nil, fmt.Errorf("failed to parse activeUsers: %w", err)
		}

		sessions, err := strconv.Atoi(row.MetricValues[1].Value)
		if err != nil {
			return nil, fmt.Errorf("failed to parse sessions: %w", err)
		}

		engagedSessions, err := strconv.Atoi(row.MetricValues[2].Value)
		if err != nil {
			return nil, fmt.Errorf("failed to parse engagedSessions: %w", err)
		}

		var engagementRate float64
		if sessions > 0 {
			engagementRate = float64(engagedSessions) / float64(sessions) * 100
		}

		var sessionsPerUser float64
		if activeUsers > 0 {
			sessionsPerUser = float64(sessions) / float64(activeUsers)
		}

		records = append(records, model.PlatformRecord{
			Platform:        row.DimensionValues[0].Value,
			ActiveUsers:     activeUsers,
			Sessions:        sessions,
			EngagedSessions: engagedSessions,
			EngagementRate:  engagementRate,
			SessionsPerUser: sessionsPerUser,
		})
	}

	return records, nil
}

// GetRealtime returns a realtime analytics report for the given property.
func (c *dataClient) GetRealtime(ctx context.Context, propertyID string) (*model.RealtimeReport, error) {
	req := &analyticsdata.RunRealtimeReportRequest{
		Dimensions: []*analyticsdata.Dimension{
			{Name: "eventName"},
		},
		Metrics: []*analyticsdata.Metric{
			{Name: "eventCount"},
		},
		MetricAggregations: []string{"TOTAL"},
	}

	resp, err := c.service.Properties.RunRealtimeReport(propertyResource(propertyID), req).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to run realtime report: %w", err)
	}

	report := &model.RealtimeReport{}

	// Extract total active users from totals
	if len(resp.Totals) > 0 && len(resp.Totals[0].MetricValues) > 0 {
		total, err := strconv.Atoi(resp.Totals[0].MetricValues[0].Value)
		if err != nil {
			return nil, fmt.Errorf("failed to parse total activeUsers: %w", err)
		}
		report.ActiveUsers = total
	}

	for _, row := range resp.Rows {
		eventCount, err := strconv.Atoi(row.MetricValues[0].Value)
		if err != nil {
			return nil, fmt.Errorf("failed to parse realtime eventCount: %w", err)
		}

		report.Events = append(report.Events, model.RealtimeEvent{
			EventName:  row.DimensionValues[0].Value,
			EventCount: eventCount,
		})
	}

	return report, nil
}

// GetMetricsSummary returns aggregated metrics (activeUsers, eventCount, sessions) for a date range.
func (c *dataClient) GetMetricsSummary(ctx context.Context, propertyID string, startDate string, endDate string) (*model.MetricsSummary, error) {
	req := &analyticsdata.RunReportRequest{
		DateRanges: []*analyticsdata.DateRange{
			{StartDate: startDate, EndDate: endDate},
		},
		Metrics: []*analyticsdata.Metric{
			{Name: "activeUsers"},
			{Name: "eventCount"},
			{Name: "sessions"},
		},
	}

	resp, err := c.service.Properties.RunReport(propertyResource(propertyID), req).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to run metrics summary report: %w", err)
	}

	summary := &model.MetricsSummary{}

	if len(resp.Rows) > 0 {
		row := resp.Rows[0]

		activeUsers, err := strconv.ParseInt(row.MetricValues[0].Value, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("failed to parse activeUsers: %w", err)
		}
		summary.ActiveUsers = activeUsers

		events, err := strconv.ParseInt(row.MetricValues[1].Value, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("failed to parse eventCount: %w", err)
		}
		summary.Events = events

		sessions, err := strconv.ParseInt(row.MetricValues[2].Value, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("failed to parse sessions: %w", err)
		}
		summary.Sessions = sessions
	}

	return summary, nil
}

// formatGADate converts GA date format "20260401" to "2026-04-01".
func formatGADate(gaDate string) (string, error) {
	t, err := time.Parse("20060102", gaDate)
	if err != nil {
		return "", fmt.Errorf("failed to parse GA date %q: %w", gaDate, err)
	}
	return t.Format("2006-01-02"), nil
}
