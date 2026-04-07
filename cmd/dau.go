package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/techinpark/ga-cli/internal/model"

	"github.com/spf13/cobra"
)

func newDAUCmd(deps func() *Dependencies) *cobra.Command {
	var (
		days int
		all  bool
	)

	cmd := &cobra.Command{
		Use:   "dau [property]",
		Short: "일일 활성 사용자(DAU) 조회",
		Args: func(cmd *cobra.Command, args []string) error {
			if all {
				return nil
			}
			if len(args) < 1 {
				return fmt.Errorf("property argument is required (or use --all)")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if days <= 0 {
				return fmt.Errorf("--days must be a positive integer, got %d", days)
			}

			d := deps()
			ctx := cmd.Context()

			if all {
				return runDAUAll(cmd, d)
			}

			propertyID, err := d.Resolver.Resolve(args[0])
			if err != nil {
				return fmt.Errorf("failed to resolve property: %w", err)
			}

			records, err := d.Data.GetDAU(ctx, propertyID, days)
			if err != nil {
				return fmt.Errorf("failed to get DAU: %w", err)
			}

			title := fmt.Sprintf("%s - Daily Active Users (Last %d days)", propertyTitle(args[0]), days)
			return d.Formatter.FormatDAU(os.Stdout, title, records)
		},
	}

	cmd.Flags().IntVarP(&days, "days", "d", 30, "조회 기간 (일)")
	cmd.Flags().BoolVar(&all, "all", false, "모든 속성의 오늘 DAU 조회")

	return cmd
}

func runDAUAll(cmd *cobra.Command, d *Dependencies) error {
	ctx := cmd.Context()
	aliases := d.Config.Aliases

	var summaries []model.DAUSummary
	for name, id := range aliases {
		records, err := d.Data.GetDAU(ctx, id, 1)
		if err != nil {
			return fmt.Errorf("failed to get DAU for %s: %w", name, err)
		}

		activeUsers := 0
		if len(records) > 0 {
			activeUsers = records[0].ActiveUsers
		}

		summaries = append(summaries, model.DAUSummary{
			PropertyName: name,
			PropertyID:   id,
			ActiveUsers:  activeUsers,
		})
	}

	return d.Formatter.FormatDAUSummary(os.Stdout, summaries)
}

func propertyTitle(nameOrID string) string {
	return strings.ToUpper(nameOrID)
}
