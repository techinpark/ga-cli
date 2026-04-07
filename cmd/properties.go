package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func newPropertiesCmd(deps func() *Dependencies) *cobra.Command {
	return &cobra.Command{
		Use:   "properties",
		Short: "GA4 속성 목록 조회",
		RunE: func(cmd *cobra.Command, args []string) error {
			d := deps()
			ctx := cmd.Context()

			properties, err := d.Admin.ListProperties(ctx)
			if err != nil {
				return fmt.Errorf("failed to list properties: %w", err)
			}

			return d.Formatter.FormatProperties(os.Stdout, "GA4 Properties", properties)
		},
	}
}
