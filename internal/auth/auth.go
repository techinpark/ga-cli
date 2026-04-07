package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/analyticsadmin/v1beta"
	"google.golang.org/api/analyticsdata/v1beta"
	"gopkg.in/yaml.v3"
)

// Scopes defines the required GA4 OAuth2 scopes.
var Scopes = []string{
	analyticsdata.AnalyticsReadonlyScope,
	analyticsadmin.AnalyticsReadonlyScope,
}

const defaultAccountName = "default"

// AccountInfo represents a registered account and its status.
type AccountInfo struct {
	Name     string
	IsActive bool
	Valid    bool // 토큰 유효 여부
}

// Manager handles OAuth2 authentication for the CLI.
type Manager struct {
	configDir string // ~/.ga-cli/
}

// NewManager creates a new auth Manager with the default config directory.
func NewManager() (*Manager, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}
	configDir := filepath.Join(home, ".ga-cli")
	if err := os.MkdirAll(configDir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create config directory: %w", err)
	}
	return &Manager{configDir: configDir}, nil
}

// AccountsDir returns the accounts directory path.
func (m *Manager) AccountsDir() string {
	return filepath.Join(m.configDir, "accounts")
}

// TokenPathForAccount returns token path for a specific account.
func (m *Manager) TokenPathForAccount(account string) string {
	return filepath.Join(m.AccountsDir(), account+".json")
}

// TokenPath returns the active account's token path.
// For backward compatibility with existing callers.
func (m *Manager) TokenPath() string {
	return m.TokenPathForAccount(m.ActiveAccount())
}

// CredentialsPath returns the path to the OAuth client credentials file.
func (m *Manager) CredentialsPath() string {
	return filepath.Join(m.configDir, "credentials.json")
}

// Login performs OAuth2 browser-based login flow for the specified account.
func (m *Manager) Login(ctx context.Context, account string) error {
	if account == "" {
		account = defaultAccountName
	}

	if err := os.MkdirAll(m.AccountsDir(), 0700); err != nil {
		return fmt.Errorf("failed to create accounts directory: %w", err)
	}

	oauthConfig, err := m.loadOAuthConfig()
	if err != nil {
		return err
	}

	// 로컬 서버로 콜백 수신 (랜덤 포트)
	listener, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		return fmt.Errorf("failed to start local server: %w", err)
	}
	port := listener.Addr().(*net.TCPAddr).Port

	oauthConfig.RedirectURL = fmt.Sprintf("http://localhost:%d/callback", port)

	codeCh := make(chan string, 1)
	errCh := make(chan error, 1)

	mux := http.NewServeMux()
	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		if code == "" {
			errMsg := r.URL.Query().Get("error")
			http.Error(w, "authentication failed: "+errMsg, http.StatusBadRequest)
			errCh <- fmt.Errorf("authentication failed: %s", errMsg)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, `<html><body><h2>Authentication complete!</h2><p>You can close this window and return to the terminal.</p></body></html>`)
		codeCh <- code
	})

	server := &http.Server{Handler: mux}
	go func() {
		if err := server.Serve(listener); err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	authURL := oauthConfig.AuthCodeURL("state", oauth2.AccessTypeOffline, oauth2.ApprovalForce)
	fmt.Printf("Opening authentication page in browser...\n\n")
	fmt.Printf("If the browser does not open automatically, visit:\n%s\n\n", authURL)

	openBrowser(authURL)

	var code string
	select {
	case code = <-codeCh:
	case err := <-errCh:
		_ = server.Shutdown(ctx)
		return err
	case <-ctx.Done():
		_ = server.Shutdown(ctx)
		return ctx.Err()
	}

	_ = server.Shutdown(ctx)

	token, err := oauthConfig.Exchange(ctx, code)
	if err != nil {
		return fmt.Errorf("failed to exchange token: %w", err)
	}

	return m.saveTokenForAccount(token, account)
}

// GetTokenSource returns a token source for the active account.
func (m *Manager) GetTokenSource(ctx context.Context) (oauth2.TokenSource, error) {
	return m.GetTokenSourceForAccount(ctx, m.ActiveAccount())
}

// GetTokenSourceForAccount returns a token source for a specific account.
func (m *Manager) GetTokenSourceForAccount(ctx context.Context, account string) (oauth2.TokenSource, error) {
	oauthConfig, err := m.loadOAuthConfig()
	if err != nil {
		return nil, err
	}

	token, err := m.loadTokenForAccount(account)
	if err != nil {
		return nil, fmt.Errorf("not logged in as %q, run 'ga auth login --account %s' first: %w", account, account, err)
	}

	return oauthConfig.TokenSource(ctx, token), nil
}

// IsLoggedIn checks if the active account has a valid token.
func (m *Manager) IsLoggedIn() bool {
	return m.IsLoggedInAs(m.ActiveAccount())
}

// IsLoggedInAs checks if a specific account has a valid token.
func (m *Manager) IsLoggedInAs(account string) bool {
	token, err := m.loadTokenForAccount(account)
	if err != nil {
		return false
	}
	return token.Valid() || token.RefreshToken != ""
}

// Logout removes the active account's token.
func (m *Manager) Logout() error {
	return m.LogoutAccount(m.ActiveAccount())
}

// LogoutAccount removes a specific account's token.
func (m *Manager) LogoutAccount(account string) error {
	path := m.TokenPathForAccount(account)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil
	}
	if err := os.Remove(path); err != nil {
		return fmt.Errorf("failed to remove token for account %q: %w", account, err)
	}

	// 활성 계정을 삭제한 경우 default로 전환
	if m.ActiveAccount() == account {
		_ = m.SetActiveAccount(defaultAccountName)
	}
	return nil
}

// LogoutAll removes all account tokens.
func (m *Manager) LogoutAll() error {
	accountsDir := m.AccountsDir()
	if _, err := os.Stat(accountsDir); os.IsNotExist(err) {
		return nil
	}

	entries, err := os.ReadDir(accountsDir)
	if err != nil {
		return fmt.Errorf("failed to read accounts directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		if err := os.Remove(filepath.Join(accountsDir, entry.Name())); err != nil {
			return fmt.Errorf("failed to remove %s: %w", entry.Name(), err)
		}
	}

	_ = m.SetActiveAccount(defaultAccountName)
	return nil
}

// ListAccounts returns all registered accounts.
func (m *Manager) ListAccounts() ([]AccountInfo, error) {
	accountsDir := m.AccountsDir()
	if _, err := os.Stat(accountsDir); os.IsNotExist(err) {
		return nil, nil
	}

	entries, err := os.ReadDir(accountsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read accounts directory: %w", err)
	}

	active := m.ActiveAccount()
	var accounts []AccountInfo

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		name := strings.TrimSuffix(entry.Name(), ".json")
		token, err := m.loadTokenForAccount(name)
		valid := err == nil && (token.Valid() || token.RefreshToken != "")

		accounts = append(accounts, AccountInfo{
			Name:     name,
			IsActive: name == active,
			Valid:    valid,
		})
	}

	return accounts, nil
}

// ActiveAccount reads the active_account from config.yaml.
func (m *Manager) ActiveAccount() string {
	configPath := filepath.Join(m.configDir, "config.yaml")
	data, err := os.ReadFile(configPath)
	if err != nil {
		return defaultAccountName
	}

	var cfg map[string]interface{}
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return defaultAccountName
	}

	if account, ok := cfg["active_account"].(string); ok && account != "" {
		return account
	}
	return defaultAccountName
}

// SetActiveAccount writes the active_account to config.yaml.
func (m *Manager) SetActiveAccount(account string) error {
	configPath := filepath.Join(m.configDir, "config.yaml")

	data, _ := os.ReadFile(configPath)

	var cfg map[string]interface{}
	if err := yaml.Unmarshal(data, &cfg); err != nil || cfg == nil {
		cfg = make(map[string]interface{})
	}

	cfg["active_account"] = account

	out, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	return os.WriteFile(configPath, out, 0600)
}

// MigrateTokenIfNeeded moves legacy token.json to accounts/default.json.
func (m *Manager) MigrateTokenIfNeeded() error {
	oldPath := filepath.Join(m.configDir, "token.json")
	newPath := m.TokenPathForAccount(defaultAccountName)

	if _, err := os.Stat(oldPath); os.IsNotExist(err) {
		return nil
	}

	if _, err := os.Stat(newPath); err == nil {
		// accounts/default.json already exists, skip migration
		return nil
	}

	if err := os.MkdirAll(m.AccountsDir(), 0700); err != nil {
		return fmt.Errorf("failed to create accounts directory: %w", err)
	}

	if err := os.Rename(oldPath, newPath); err != nil {
		return fmt.Errorf("failed to migrate token: %w", err)
	}

	return nil
}

// HasCredentials checks if OAuth credentials file exists.
func (m *Manager) HasCredentials() bool {
	_, err := os.Stat(m.CredentialsPath())
	return err == nil
}

func (m *Manager) loadOAuthConfig() (*oauth2.Config, error) {
	// 1. ~/.ga-cli/credentials.json이 있으면 우선 사용
	data, err := os.ReadFile(m.CredentialsPath())
	if err == nil {
		config, err := google.ConfigFromJSON(data, Scopes...)
		if err != nil {
			return nil, fmt.Errorf("failed to parse credentials: %w", err)
		}
		return config, nil
	}

	// 2. 내장 credentials 사용 (brew install 등 배포 바이너리)
	if hasEmbeddedCredentials() {
		return embeddedOAuthConfig(Scopes), nil
	}

	return nil, fmt.Errorf("OAuth credentials not available\n\n" +
		"Homebrew로 설치한 경우 최신 버전으로 업데이트하세요:\n" +
		"  brew upgrade ga-cli\n\n" +
		"소스에서 빌드한 경우 다음 중 하나를 선택하세요:\n" +
		"  1. credentials.json 직접 제공: %s\n" +
		"  2. ADC 사용: gcloud auth application-default login",
		m.CredentialsPath())
}

func (m *Manager) saveTokenForAccount(token *oauth2.Token, account string) error {
	path := m.TokenPathForAccount(account)
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("failed to create token file: %w", err)
	}
	defer f.Close()

	return json.NewEncoder(f).Encode(token)
}

func (m *Manager) loadTokenForAccount(account string) (*oauth2.Token, error) {
	data, err := os.ReadFile(m.TokenPathForAccount(account))
	if err != nil {
		return nil, err
	}

	var token oauth2.Token
	if err := json.Unmarshal(data, &token); err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	return &token, nil
}

// openBrowser opens a URL in the default browser.
func openBrowser(url string) {
	_ = exec.Command("open", url).Start()
}
