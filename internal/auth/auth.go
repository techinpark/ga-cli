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

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/analyticsadmin/v1beta"
	"google.golang.org/api/analyticsdata/v1beta"
)

// Scopes defines the required GA4 OAuth2 scopes.
var Scopes = []string{
	analyticsdata.AnalyticsReadonlyScope,
	analyticsadmin.AnalyticsReadonlyScope,
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

// TokenPath returns the path to the saved token file.
func (m *Manager) TokenPath() string {
	return filepath.Join(m.configDir, "token.json")
}

// CredentialsPath returns the path to the OAuth client credentials file.
func (m *Manager) CredentialsPath() string {
	return filepath.Join(m.configDir, "credentials.json")
}

// Login performs OAuth2 browser-based login flow.
// 1. 로컬 HTTP 서버 시작 (랜덤 포트)
// 2. 브라우저에서 Google OAuth 동의 화면 열기
// 3. 콜백으로 authorization code 수신
// 4. code를 token으로 교환
// 5. ~/.ga-cli/token.json에 저장
func (m *Manager) Login(ctx context.Context) error {
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

	return m.saveToken(token)
}

// GetTokenSource returns a token source for authenticated API calls.
// The returned TokenSource automatically handles token refresh.
func (m *Manager) GetTokenSource(ctx context.Context) (oauth2.TokenSource, error) {
	oauthConfig, err := m.loadOAuthConfig()
	if err != nil {
		return nil, err
	}

	token, err := m.loadToken()
	if err != nil {
		return nil, fmt.Errorf("not logged in, run 'ga auth login' first: %w", err)
	}

	return oauthConfig.TokenSource(ctx, token), nil
}

// IsLoggedIn checks if a valid token exists.
func (m *Manager) IsLoggedIn() bool {
	token, err := m.loadToken()
	if err != nil {
		return false
	}
	return token.Valid() || token.RefreshToken != ""
}

// Logout removes the saved token.
func (m *Manager) Logout() error {
	path := m.TokenPath()
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil
	}
	return os.Remove(path)
}

// HasCredentials checks if OAuth credentials file exists.
func (m *Manager) HasCredentials() bool {
	_, err := os.Stat(m.CredentialsPath())
	return err == nil
}

func (m *Manager) loadOAuthConfig() (*oauth2.Config, error) {
	data, err := os.ReadFile(m.CredentialsPath())
	if err != nil {
		return nil, fmt.Errorf("OAuth credentials not found at %s\nDownload from Google Cloud Console > APIs & Services > Credentials > OAuth 2.0 Client IDs > Download JSON", m.CredentialsPath())
	}

	config, err := google.ConfigFromJSON(data, Scopes...)
	if err != nil {
		return nil, fmt.Errorf("failed to parse credentials: %w", err)
	}

	return config, nil
}

func (m *Manager) saveToken(token *oauth2.Token) error {
	path := m.TokenPath()
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("failed to create token file: %w", err)
	}
	defer f.Close()

	return json.NewEncoder(f).Encode(token)
}

func (m *Manager) loadToken() (*oauth2.Token, error) {
	data, err := os.ReadFile(m.TokenPath())
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
