package client

import (
	"fmt"
	"strings"

	"google.golang.org/api/googleapi"
)

// WrapAPIError wraps a Google API error with a user-friendly message.
func WrapAPIError(context string, err error) error {
	if err == nil {
		return nil
	}

	if apiErr, ok := err.(*googleapi.Error); ok {
		switch apiErr.Code {
		case 401:
			return fmt.Errorf("%s: authentication required. Run 'ga auth login'", context)
		case 403:
			return fmt.Errorf("%s: permission denied. Check read access to the GA4 property", context)
		case 404:
			return fmt.Errorf("%s: property not found. Verify the property ID", context)
		case 429:
			return fmt.Errorf("%s: API quota exceeded. Try again later", context)
		case 500, 502, 503:
			return fmt.Errorf("%s: Google Analytics server error. Try again later", context)
		}
	}

	msg := err.Error()
	if strings.Contains(msg, "oauth2: token expired") || strings.Contains(msg, "oauth2: cannot fetch token") {
		return fmt.Errorf("%s: token expired. Run 'ga auth login' to re-authenticate", context)
	}

	if strings.Contains(msg, "connection refused") || strings.Contains(msg, "no such host") {
		return fmt.Errorf("%s: network error. Check your internet connection", context)
	}

	return fmt.Errorf("%s: %w", context, err)
}
