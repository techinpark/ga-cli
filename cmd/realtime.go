package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func newRealtimeCmd(deps func() *Dependencies) *cobra.Command {
	return &cobra.Command{
		Use:   "realtime <property>",
		Short: "실시간 사용자 통계 조회",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			d := deps()
			ctx := cmd.Context()

			propertyID, err := d.Resolver.Resolve(args[0])
			if err != nil {
				return fmt.Errorf("failed to resolve property: %w", err)
			}

			report, err := d.Data.GetRealtime(ctx, propertyID)
			if err != nil {
				return fmt.Errorf("failed to get realtime data: %w", err)
			}

			title := fmt.Sprintf("%s - Realtime", propertyTitle(args[0]))
			return d.Formatter.FormatRealtime(os.Stdout, title, report)
		},
	}
}
