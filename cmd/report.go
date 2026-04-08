package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/techinpark/ga-cli/internal/model"
)

func newReportCmd(deps func() *Dependencies) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "report",
		Short: "일간/주간 종합 리포트",
	}

	cmd.AddCommand(newReportDailyCmd(deps))
	cmd.AddCommand(newReportWeeklyCmd(deps))
	cmd.AddCommand(newReportCompareCmd(deps))

	return cmd
}

func newReportDailyCmd(deps func() *Dependencies) *cobra.Command {
	var all bool

	cmd := &cobra.Command{
		Use:   "daily [property]",
		Short: "일간 리포트 (DAU 7일, 이벤트 Top 10, 플랫폼, 실시간)",
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
			d := deps()
			ctx := cmd.Context()

			if all {
				return runReportAll(ctx, d, runDailyReport)
			}

			propertyID, err := d.Resolver.Resolve(args[0])
			if err != nil {
				return fmt.Errorf("failed to resolve property: %w", err)
			}

			return runDailyReport(ctx, d, args[0], propertyID)
		},
	}

	cmd.Flags().BoolVar(&all, "all", false, "모든 속성 리포트")

	return cmd
}

func newReportWeeklyCmd(deps func() *Dependencies) *cobra.Command {
	var all bool

	cmd := &cobra.Command{
		Use:   "weekly [property]",
		Short: "주간 리포트 (DAU 14일, 이벤트 Top 20, 국가별, 플랫폼)",
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
			d := deps()
			ctx := cmd.Context()

			if all {
				return runReportAll(ctx, d, runWeeklyReport)
			}

			propertyID, err := d.Resolver.Resolve(args[0])
			if err != nil {
				return fmt.Errorf("failed to resolve property: %w", err)
			}

			return runWeeklyReport(ctx, d, args[0], propertyID)
		},
	}

	cmd.Flags().BoolVar(&all, "all", false, "모든 속성 리포트")

	return cmd
}

func runDailyReport(ctx context.Context, d *Dependencies, name, propertyID string) error {
	w := os.Stdout
	title := strings.ToUpper(name)

	fmt.Fprintf(w, "\n=== Daily Report: %s ===\n\n", title)

	// DAU (7일)
	dauRecords, err := d.Data.GetDAU(ctx, propertyID, 7)
	if err != nil {
		return fmt.Errorf("failed to get DAU: %w", err)
	}
	if err := d.Formatter.FormatDAU(w, fmt.Sprintf("%s - Daily Active Users (Last 7 days)", title), dauRecords); err != nil {
		return err
	}
	fmt.Fprintln(w)

	// Events Top 10 (1일)
	events, err := d.Data.GetEvents(ctx, propertyID, 1, 10)
	if err != nil {
		return fmt.Errorf("failed to get events: %w", err)
	}
	if err := d.Formatter.FormatEvents(w, fmt.Sprintf("%s - Top Events (Yesterday)", title), events); err != nil {
		return err
	}
	fmt.Fprintln(w)

	// Platforms (1일)
	platforms, err := d.Data.GetPlatforms(ctx, propertyID, 1)
	if err != nil {
		return fmt.Errorf("failed to get platforms: %w", err)
	}
	if err := d.Formatter.FormatPlatforms(w, fmt.Sprintf("%s - Platforms (Yesterday)", title), platforms); err != nil {
		return err
	}
	fmt.Fprintln(w)

	// Realtime
	realtime, err := d.Data.GetRealtime(ctx, propertyID)
	if err != nil {
		return fmt.Errorf("failed to get realtime: %w", err)
	}
	return d.Formatter.FormatRealtime(w, fmt.Sprintf("%s - Realtime", title), realtime)
}

func runWeeklyReport(ctx context.Context, d *Dependencies, name, propertyID string) error {
	w := os.Stdout
	title := strings.ToUpper(name)

	fmt.Fprintf(w, "\n=== Weekly Report: %s ===\n\n", title)

	// DAU (14일, week-over-week 비교)
	dauRecords, err := d.Data.GetDAU(ctx, propertyID, 14)
	if err != nil {
		return fmt.Errorf("failed to get DAU: %w", err)
	}
	if err := d.Formatter.FormatDAU(w, fmt.Sprintf("%s - Daily Active Users (Last 14 days)", title), dauRecords); err != nil {
		return err
	}
	fmt.Fprintln(w)

	// Events Top 20 (7일)
	events, err := d.Data.GetEvents(ctx, propertyID, 7, 20)
	if err != nil {
		return fmt.Errorf("failed to get events: %w", err)
	}
	if err := d.Formatter.FormatEvents(w, fmt.Sprintf("%s - Top Events (Last 7 days)", title), events); err != nil {
		return err
	}
	fmt.Fprintln(w)

	// Countries (7일)
	countries, err := d.Data.GetCountries(ctx, propertyID, 7)
	if err != nil {
		return fmt.Errorf("failed to get countries: %w", err)
	}
	if err := d.Formatter.FormatCountries(w, fmt.Sprintf("%s - Users by Country (Last 7 days)", title), countries); err != nil {
		return err
	}
	fmt.Fprintln(w)

	// Platforms (7일)
	platforms, err := d.Data.GetPlatforms(ctx, propertyID, 7)
	if err != nil {
		return fmt.Errorf("failed to get platforms: %w", err)
	}
	return d.Formatter.FormatPlatforms(w, fmt.Sprintf("%s - Platforms (Last 7 days)", title), platforms)
}

func newReportCompareCmd(deps func() *Dependencies) *cobra.Command {
	var all bool

	cmd := &cobra.Command{
		Use:   "compare [property]",
		Short: "전일/전주 대비 비교 리포트",
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
			d := deps()
			ctx := cmd.Context()

			if all {
				return runReportAll(ctx, d, runCompareReport)
			}

			propertyID, err := d.Resolver.Resolve(args[0])
			if err != nil {
				return fmt.Errorf("failed to resolve property: %w", err)
			}

			return runCompareReport(ctx, d, args[0], propertyID)
		},
	}

	cmd.Flags().BoolVar(&all, "all", false, "모든 속성 비교 리포트")

	return cmd
}

func runCompareReport(ctx context.Context, d *Dependencies, name, propertyID string) error {
	now := time.Now()
	today := now.Format("2006-01-02")
	yesterday := now.AddDate(0, 0, -1).Format("2006-01-02")
	twoDaysAgo := now.AddDate(0, 0, -2).Format("2006-01-02")

	thisWeekStart := now.AddDate(0, 0, -6).Format("2006-01-02")
	lastWeekEnd := now.AddDate(0, 0, -7).Format("2006-01-02")
	lastWeekStart := now.AddDate(0, 0, -13).Format("2006-01-02")

	// Day over Day: today vs yesterday
	todaySummary, err := d.Data.GetMetricsSummary(ctx, propertyID, today, today)
	if err != nil {
		return fmt.Errorf("failed to get today's metrics: %w", err)
	}

	yesterdaySummary, err := d.Data.GetMetricsSummary(ctx, propertyID, twoDaysAgo, yesterday)
	if err != nil {
		return fmt.Errorf("failed to get yesterday's metrics: %w", err)
	}

	// Week over Week: this week (7 days) vs last week (7 days)
	thisWeekSummary, err := d.Data.GetMetricsSummary(ctx, propertyID, thisWeekStart, today)
	if err != nil {
		return fmt.Errorf("failed to get this week's metrics: %w", err)
	}

	lastWeekSummary, err := d.Data.GetMetricsSummary(ctx, propertyID, lastWeekStart, lastWeekEnd)
	if err != nil {
		return fmt.Errorf("failed to get last week's metrics: %w", err)
	}

	report := &model.CompareReport{
		PropertyName: name,
		DayOverDay:   buildCompareRecords(todaySummary, yesterdaySummary),
		WeekOverWeek: buildCompareRecords(thisWeekSummary, lastWeekSummary),
	}

	return d.Formatter.FormatCompare(os.Stdout, report)
}

func buildCompareRecords(current, previous *model.MetricsSummary) []model.CompareRecord {
	calc := func(metric string, cur, prev int64) model.CompareRecord {
		var change float64
		if prev > 0 {
			change = float64(cur-prev) / float64(prev) * 100
		}
		return model.CompareRecord{
			Metric:        metric,
			Current:       cur,
			Previous:      prev,
			ChangePercent: change,
		}
	}

	return []model.CompareRecord{
		calc("DAU", current.ActiveUsers, previous.ActiveUsers),
		calc("Events", current.Events, previous.Events),
		calc("Sessions", current.Sessions, previous.Sessions),
	}
}

type reportFunc func(ctx context.Context, d *Dependencies, name, propertyID string) error

func runReportAll(ctx context.Context, d *Dependencies, fn reportFunc) error {
	aliases := d.Config.Aliases
	if len(aliases) == 0 {
		return fmt.Errorf("no aliases configured; add aliases to ~/.ga-cli/config.yaml")
	}

	first := true
	for name, id := range aliases {
		if !first {
			fmt.Fprintln(os.Stdout, "\n"+strings.Repeat("─", 60))
		}
		first = false

		if err := fn(ctx, d, name, id); err != nil {
			fmt.Fprintf(os.Stderr, "warning: %s skipped: %v\n", name, err)
			continue
		}
	}

	return nil
}
