package model

// Property represents a GA4 property.
type Property struct {
	ID          string `json:"id"`
	DisplayName string `json:"display_name"`
	Alias       string `json:"alias"`
}

// DAURecord represents daily active users for a single date.
type DAURecord struct {
	Date          string  `json:"date"`
	ActiveUsers   int     `json:"active_users"`
	ChangePercent float64 `json:"change_percent"`
}

// DAUSummary represents DAU for a property in --all mode.
type DAUSummary struct {
	PropertyName string `json:"property_name"`
	PropertyID   string `json:"property_id"`
	ActiveUsers  int    `json:"active_users"`
}

// EventRecord represents an event with its count and total users.
type EventRecord struct {
	EventName  string `json:"event_name"`
	EventCount int64  `json:"event_count"`
	TotalUsers int    `json:"total_users"`
}

// CountryRecord represents active users and sessions by country.
type CountryRecord struct {
	Country                string  `json:"country"`
	ActiveUsers            int     `json:"active_users"`
	Sessions               int     `json:"sessions"`
	ScreenPageViewsPerUser float64 `json:"screen_page_views_per_user"`
}

// PlatformRecord represents active users and sessions by platform.
type PlatformRecord struct {
	Platform        string  `json:"platform"`
	ActiveUsers     int     `json:"active_users"`
	Sessions        int     `json:"sessions"`
	EngagedSessions int     `json:"engaged_sessions"`
	EngagementRate  float64 `json:"engagement_rate"`
	SessionsPerUser float64 `json:"sessions_per_user"`
}

// RealtimeReport represents a realtime analytics report.
type RealtimeReport struct {
	ActiveUsers int             `json:"active_users"`
	Events      []RealtimeEvent `json:"events"`
}

// RealtimeEvent represents a single event in a realtime report.
type RealtimeEvent struct {
	EventName  string `json:"event_name"`
	EventCount int    `json:"event_count"`
}
