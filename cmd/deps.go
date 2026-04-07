package cmd

import (
	"github.com/techinpark/ga-cli/internal/client"
	"github.com/techinpark/ga-cli/internal/config"
	"github.com/techinpark/ga-cli/internal/formatter"
)

// Dependencies holds shared dependencies for all subcommands.
type Dependencies struct {
	Admin     client.AdminClient
	Data      client.DataClient
	Resolver  *client.PropertyResolver
	Formatter formatter.Formatter
	Config    *config.Config
}
