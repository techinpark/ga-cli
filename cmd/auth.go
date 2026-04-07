package cmd

import (
	"fmt"

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

	return authCmd
}

func newAuthLoginCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "login",
		Short: "Google 계정으로 로그인",
		RunE: func(cmd *cobra.Command, args []string) error {
			mgr, err := auth.NewManager()
			if err != nil {
				return err
			}

			if !mgr.HasCredentials() {
				return fmt.Errorf("OAuth credentials not found\n\n1. Google Cloud Console에서 OAuth 2.0 Client ID를 생성하세요\n2. 'Desktop App' 유형으로 생성\n3. JSON 다운로드 후 %s에 저장하세요", mgr.CredentialsPath())
			}

			return mgr.Login(cmd.Context())
		},
	}
}

func newAuthStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "인증 상태 확인",
		RunE: func(cmd *cobra.Command, args []string) error {
			mgr, err := auth.NewManager()
			if err != nil {
				return err
			}

			if mgr.IsLoggedIn() {
				fmt.Fprintln(cmd.OutOrStdout(), "Logged in")
				fmt.Fprintf(cmd.OutOrStdout(), "Token path: %s\n", mgr.TokenPath())
			} else {
				fmt.Fprintln(cmd.OutOrStdout(), "Not logged in")
				fmt.Fprintln(cmd.OutOrStdout(), "Run 'ga auth login' to log in")
			}
			return nil
		},
	}
}

func newAuthLogoutCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "logout",
		Short: "로그아웃 (토큰 삭제)",
		RunE: func(cmd *cobra.Command, args []string) error {
			mgr, err := auth.NewManager()
			if err != nil {
				return err
			}

			if err := mgr.Logout(); err != nil {
				return fmt.Errorf("failed to logout: %w", err)
			}

			fmt.Fprintln(cmd.OutOrStdout(), "Logged out")
			return nil
		},
	}
}
