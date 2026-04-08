package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/techinpark/ga-cli/internal/auth"
)

func newAuthCmd() *cobra.Command {
	authCmd := &cobra.Command{
		Use:   "auth",
		Short: "Google 계정 인증 관리",
	}

	authCmd.AddCommand(newAuthLoginCmd())
	authCmd.AddCommand(newAuthStatusCmd())
	authCmd.AddCommand(newAuthLogoutCmd())
	authCmd.AddCommand(newAuthListCmd())
	authCmd.AddCommand(newAuthSwitchCmd())

	return authCmd
}

func newAuthLoginCmd() *cobra.Command {
	var account string

	cmd := &cobra.Command{
		Use:   "login",
		Short: "Google 계정으로 로그인",
		RunE: func(cmd *cobra.Command, args []string) error {
			mgr, err := auth.NewManager()
			if err != nil {
				return err
			}

			if err := mgr.MigrateTokenIfNeeded(); err != nil {
				return fmt.Errorf("failed to migrate token: %w", err)
			}

			accountSpecified := cmd.Flag("account").Changed

			if mgr.IsLoggedInAs(account) {
				fmt.Fprintf(cmd.OutOrStdout(), "Account %q is already logged in.\n", account)
				fmt.Fprintf(cmd.OutOrStdout(), "Logging in again will overwrite the existing token.\n")
				fmt.Fprintf(cmd.OutOrStdout(), "To add a different Google account, use: ga auth login --account <name>\n\n")
				fmt.Fprint(cmd.OutOrStdout(), "Continue? [y/N]: ")

				reader := bufio.NewReader(os.Stdin)
				input, _ := reader.ReadString('\n')
				input = strings.TrimSpace(strings.ToLower(input))
				if input != "y" && input != "yes" {
					fmt.Fprintln(cmd.OutOrStdout(), "Cancelled.")
					return nil
				}
			}

			if err := mgr.Login(cmd.Context(), account); err != nil {
				return err
			}

			// --account 미지정 시 로그인 후 이름 지정 기회 제공
			if !accountSpecified && account == "default" {
				fmt.Fprint(cmd.OutOrStdout(), "\nGive this account a name (enter to keep \"default\"): ")
				reader := bufio.NewReader(os.Stdin)
				input, _ := reader.ReadString('\n')
				input = strings.TrimSpace(input)

				if input != "" && input != "default" {
					// default 토큰을 새 이름으로 이동
					if err := mgr.RenameAccount("default", input); err != nil {
						fmt.Fprintf(cmd.ErrOrStderr(), "warning: failed to rename account: %v\n", err)
					} else {
						account = input
					}
				}
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Logged in as account %q\n", account)
			return nil
		},
	}

	cmd.Flags().StringVar(&account, "account", "default", "계정 이름")
	return cmd
}

func newAuthStatusCmd() *cobra.Command {
	var account string

	cmd := &cobra.Command{
		Use:   "status",
		Short: "인증 상태 확인",
		RunE: func(cmd *cobra.Command, args []string) error {
			mgr, err := auth.NewManager()
			if err != nil {
				return err
			}

			if err := mgr.MigrateTokenIfNeeded(); err != nil {
				return fmt.Errorf("failed to migrate token: %w", err)
			}

			targetAccount := account
			if targetAccount == "" {
				targetAccount = mgr.ActiveAccount()
			}

			if mgr.IsLoggedInAs(targetAccount) {
				fmt.Fprintf(cmd.OutOrStdout(), "Account %q: Logged in\n", targetAccount)
				fmt.Fprintf(cmd.OutOrStdout(), "Token path: %s\n", mgr.TokenPathForAccount(targetAccount))
			} else {
				fmt.Fprintf(cmd.OutOrStdout(), "Account %q: Not logged in\n", targetAccount)
				fmt.Fprintf(cmd.OutOrStdout(), "Run 'ga auth login --account %s' to log in\n", targetAccount)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&account, "account", "", "확인할 계정 이름 (기본: 활성 계정)")
	return cmd
}

func newAuthLogoutCmd() *cobra.Command {
	var (
		account string
		all     bool
	)

	cmd := &cobra.Command{
		Use:   "logout",
		Short: "로그아웃 (토큰 삭제)",
		RunE: func(cmd *cobra.Command, args []string) error {
			mgr, err := auth.NewManager()
			if err != nil {
				return err
			}

			if err := mgr.MigrateTokenIfNeeded(); err != nil {
				return fmt.Errorf("failed to migrate token: %w", err)
			}

			if all {
				if err := mgr.LogoutAll(); err != nil {
					return fmt.Errorf("failed to logout: %w", err)
				}
				fmt.Fprintln(cmd.OutOrStdout(), "Logged out from all accounts")
				return nil
			}

			targetAccount := account
			if targetAccount == "" {
				targetAccount = mgr.ActiveAccount()
			}

			if err := mgr.LogoutAccount(targetAccount); err != nil {
				return fmt.Errorf("failed to logout: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Logged out from account %q\n", targetAccount)
			return nil
		},
	}

	cmd.Flags().StringVar(&account, "account", "", "로그아웃할 계정 이름 (기본: 활성 계정)")
	cmd.Flags().BoolVar(&all, "all", false, "모든 계정 로그아웃")
	return cmd
}

func newAuthListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "등록된 계정 목록",
		RunE: func(cmd *cobra.Command, args []string) error {
			mgr, err := auth.NewManager()
			if err != nil {
				return err
			}

			if err := mgr.MigrateTokenIfNeeded(); err != nil {
				return fmt.Errorf("failed to migrate token: %w", err)
			}

			accounts, err := mgr.ListAccounts()
			if err != nil {
				return err
			}

			if len(accounts) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No accounts found. Run 'ga-cli auth login' to add one.")
				return nil
			}

			for _, a := range accounts {
				marker := "  "
				if a.IsActive {
					marker = "* "
				}
				status := "valid"
				if !a.Valid {
					status = "expired"
				}
				fmt.Fprintf(cmd.OutOrStdout(), "%s%s (%s)\n", marker, a.Name, status)
			}
			return nil
		},
	}
}

func newAuthSwitchCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "switch <account>",
		Short: "활성 계정 전환",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			mgr, err := auth.NewManager()
			if err != nil {
				return err
			}

			if err := mgr.MigrateTokenIfNeeded(); err != nil {
				return fmt.Errorf("failed to migrate token: %w", err)
			}

			account := args[0]
			if !mgr.IsLoggedInAs(account) {
				return fmt.Errorf("account %q not found, run 'ga-cli auth login --account %s' first", account, account)
			}

			if err := mgr.SetActiveAccount(account); err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Switched to account %q\n", account)
			return nil
		},
	}
}
