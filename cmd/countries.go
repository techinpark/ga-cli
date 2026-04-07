package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func newCountriesCmd(deps func() *Dependencies) *cobra.Command {
	var days int

	cmd := &cobra.Command{
		Use:   "countries <property>",
		Short: "국가별 사용자 통계 조회",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if days <= 0 {
				return fmt.Errorf("--days must be a positive integer, got %d", days)
			}

			d := deps()
			ctx := cmd.Context()

			propertyID, err := d.Resolver.Resolve(args[0])
			if err != nil {
				return fmt.Errorf("failed to resolve property: %w", err)
			}

			records, err := d.Data.GetCountries(ctx, propertyID, days)
			if err != nil {
				return fmt.Errorf("failed to get countries: %w", err)
			}

			title := fmt.Sprintf("%s - Users by Country (Last %d days)", propertyTitle(args[0]), days)
			return d.Formatter.FormatCountries(os.Stdout, title, records)
		},
	}

	cmd.Flags().IntVarP(&days, "days", "d", 30, "조회 기간 (일)")

	return cmd
}
