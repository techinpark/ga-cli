package client

import (
	"context"
	"fmt"
	"strings"

	"github.com/techinpark/ga-cli/internal/model"
	"google.golang.org/api/analyticsadmin/v1beta"
	"google.golang.org/api/option"
)

type adminClient struct {
	service *analyticsadmin.Service
	aliases map[string]string // property ID -> alias (reverse mapping)
}

// NewAdminClient creates a new Admin API client.
// apiOpts configures authentication (e.g. WithCredentialsFile, WithTokenSource).
// If apiOpts is empty, Application Default Credentials (ADC) are used.
// aliases is a map from alias name to property ID; it is reversed internally
// so properties can be looked up by ID.
func NewAdminClient(apiOpts []option.ClientOption, aliases map[string]string) (AdminClient, error) {
	ctx := context.Background()

	service, err := analyticsadmin.NewService(ctx, apiOpts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create admin service: %w", err)
	}

	// Build reverse mapping: property ID -> alias
	reversed := make(map[string]string, len(aliases))
	for alias, id := range aliases {
		reversed[id] = alias
	}

	return &adminClient{
		service: service,
		aliases: reversed,
	}, nil
}

// ListProperties returns all GA4 properties accessible by the authenticated account.
func (c *adminClient) ListProperties(ctx context.Context) ([]model.Property, error) {
	var properties []model.Property
	pageToken := ""

	for {
		call := c.service.AccountSummaries.List().Context(ctx)
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}

		resp, err := call.Do()
		if err != nil {
			return nil, fmt.Errorf("failed to list account summaries: %w", err)
		}

		for _, account := range resp.AccountSummaries {
			for _, ps := range account.PropertySummaries {
				id := extractPropertyID(ps.Property)
				prop := model.Property{
					ID:          id,
					DisplayName: ps.DisplayName,
				}
				if alias, ok := c.aliases[id]; ok {
					prop.Alias = alias
				}
				properties = append(properties, prop)
			}
		}

		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}

	return properties, nil
}

// extractPropertyID extracts the numeric ID from "properties/12345" format.
func extractPropertyID(property string) string {
	parts := strings.SplitN(property, "/", 2)
	if len(parts) == 2 {
		return parts[1]
	}
	return property
}
