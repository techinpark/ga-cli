package client

import (
	"fmt"
	"unicode"
)

// PropertyResolver resolves property names or aliases to property IDs.
type PropertyResolver struct {
	aliases map[string]string
}

// NewPropertyResolver creates a new PropertyResolver with the given alias map.
func NewPropertyResolver(aliases map[string]string) *PropertyResolver {
	return &PropertyResolver{aliases: aliases}
}

// Resolve returns the property ID for the given name or alias.
// If nameOrID is a numeric string, it is returned as-is.
// If nameOrID matches an alias, the corresponding ID is returned.
// Otherwise, an error is returned.
func (r *PropertyResolver) Resolve(nameOrID string) (string, error) {
	if isNumeric(nameOrID) {
		return nameOrID, nil
	}

	if id, ok := r.aliases[nameOrID]; ok {
		return id, nil
	}

	return "", fmt.Errorf("unknown property alias: %s", nameOrID)
}

func isNumeric(s string) bool {
	if s == "" {
		return false
	}
	for _, ch := range s {
		if !unicode.IsDigit(ch) {
			return false
		}
	}
	return true
}
