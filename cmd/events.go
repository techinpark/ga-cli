package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func newEventsCmd(deps func() *Dependencies) *cobra.Command {
	var (
		days int
		top  int
	)

	cmd := &cobra.Command{
		Use:   "events <property>",
		Short: "이벤트 목록 조회",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if days <= 0 {
				return fmt.Errorf("--days must be a positive integer, got %d", days)
			}
			if top <= 0 {
				return fmt.Errorf("--top must be a positive integer, got %d", top)
			}

			d := deps()
			ctx := cmd.Context()

			propertyID, err := d.Resolver.Resolve(args[0])
			if err != nil {
				return fmt.Errorf("failed to resolve property: %w", err)
			}

			records, err := d.Data.GetEvents(ctx, propertyID, days, top)
			if err != nil {
				return fmt.Errorf("failed to get events: %w", err)
			}

			title := fmt.Sprintf("%s - Top Events (Last %d days)", propertyTitle(args[0]), days)
			return d.Formatter.FormatEvents(os.Stdout, title, records)
		},
	}

	cmd.Flags().IntVarP(&days, "days", "d", 180, "조회 기간 (일)")
	cmd.Flags().IntVarP(&top, "top", "t", 20, "상위 이벤트 수")

	return cmd
}
