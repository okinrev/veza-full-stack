package services

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"go.uber.org/zap"
)

// OAuthService service pour la gestion OAuth2
type OAuthService struct {
	logger *zap.Logger
	client *http.Client
}

// NewOAuthService crée une nouvelle instance du service OAuth2
func NewOAuthService(logger *zap.Logger) *OAuthService {
	return &OAuthService{
		logger: logger,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// OAuthUserInfo informations utilisateur OAuth2
type OAuthUserInfo struct {
	Email       string `json:"email"`
	Username    string `json:"username"`
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	DisplayName string `json:"display_name"`
	Avatar      string `json:"avatar"`
	Provider    string `json:"provider"`
	ProviderID  string `json:"provider_id"`
}

// GoogleUserInfo structure réponse Google
type GoogleUserInfo struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Picture       string `json:"picture"`
	Locale        string `json:"locale"`
}

// GitHubUserInfo structure réponse GitHub
type GitHubUserInfo struct {
	ID        int    `json:"id"`
	Login     string `json:"login"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	AvatarURL string `json:"avatar_url"`
	Company   string `json:"company"`
	Location  string `json:"location"`
	Bio       string `json:"bio"`
}

// GitHubEmail structure pour les emails GitHub
type GitHubEmail struct {
	Email      string `json:"email"`
	Primary    bool   `json:"primary"`
	Verified   bool   `json:"verified"`
	Visibility string `json:"visibility"`
}

// DiscordUserInfo structure réponse Discord
type DiscordUserInfo struct {
	ID            string `json:"id"`
	Username      string `json:"username"`
	Discriminator string `json:"discriminator"`
	Avatar        string `json:"avatar"`
	Email         string `json:"email"`
	Verified      bool   `json:"verified"`
	Locale        string `json:"locale"`
	MfaEnabled    bool   `json:"mfa_enabled"`
}

// FetchOAuthUserInfo récupère les informations utilisateur OAuth2
func (s *OAuthService) FetchOAuthUserInfo(ctx context.Context, provider, accessToken string) (*OAuthUserInfo, error) {
	switch provider {
	case "google":
		return s.fetchGoogleUserInfo(ctx, accessToken)
	case "github":
		return s.fetchGitHubUserInfo(ctx, accessToken)
	case "discord":
		return s.fetchDiscordUserInfo(ctx, accessToken)
	default:
		return nil, fmt.Errorf("provider OAuth2 non supporté: %s", provider)
	}
}

// fetchGoogleUserInfo récupère les informations utilisateur Google
func (s *OAuthService) fetchGoogleUserInfo(ctx context.Context, accessToken string) (*OAuthUserInfo, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://www.googleapis.com/oauth2/v2/userinfo", nil)
	if err != nil {
		return nil, fmt.Errorf("création requête Google: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("requête Google API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("erreur Google API: status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("lecture réponse Google: %w", err)
	}

	var googleUser GoogleUserInfo
	if err := json.Unmarshal(body, &googleUser); err != nil {
		return nil, fmt.Errorf("parsing réponse Google: %w", err)
	}

	// Valider les données requises
	if googleUser.Email == "" || !googleUser.VerifiedEmail {
		return nil, fmt.Errorf("email Google non vérifié")
	}

	// Créer le nom d'utilisateur à partir de l'email si pas de name
	username := strings.Split(googleUser.Email, "@")[0]

	return &OAuthUserInfo{
		Email:       googleUser.Email,
		Username:    username,
		FirstName:   googleUser.GivenName,
		LastName:    googleUser.FamilyName,
		DisplayName: googleUser.Name,
		Avatar:      googleUser.Picture,
		Provider:    "google",
		ProviderID:  googleUser.ID,
	}, nil
}

// fetchGitHubUserInfo récupère les informations utilisateur GitHub
func (s *OAuthService) fetchGitHubUserInfo(ctx context.Context, accessToken string) (*OAuthUserInfo, error) {
	// Récupérer les informations de base
	userInfo, err := s.fetchGitHubUser(ctx, accessToken)
	if err != nil {
		return nil, err
	}

	// Si pas d'email public, récupérer les emails
	if userInfo.Email == "" {
		email, err := s.fetchGitHubPrimaryEmail(ctx, accessToken)
		if err != nil {
			return nil, err
		}
		userInfo.Email = email
	}

	// Séparer le nom complet
	firstName, lastName := s.splitFullName(userInfo.Name)

	return &OAuthUserInfo{
		Email:       userInfo.Email,
		Username:    userInfo.Login,
		FirstName:   firstName,
		LastName:    lastName,
		DisplayName: userInfo.Name,
		Avatar:      userInfo.AvatarURL,
		Provider:    "github",
		ProviderID:  fmt.Sprintf("%d", userInfo.ID),
	}, nil
}

// fetchGitHubUser récupère les informations utilisateur GitHub
func (s *OAuthService) fetchGitHubUser(ctx context.Context, accessToken string) (*GitHubUserInfo, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://api.github.com/user", nil)
	if err != nil {
		return nil, fmt.Errorf("création requête GitHub: %w", err)
	}

	req.Header.Set("Authorization", "token "+accessToken)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("requête GitHub API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("erreur GitHub API: status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("lecture réponse GitHub: %w", err)
	}

	var githubUser GitHubUserInfo
	if err := json.Unmarshal(body, &githubUser); err != nil {
		return nil, fmt.Errorf("parsing réponse GitHub: %w", err)
	}

	return &githubUser, nil
}

// fetchGitHubPrimaryEmail récupère l'email principal GitHub
func (s *OAuthService) fetchGitHubPrimaryEmail(ctx context.Context, accessToken string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://api.github.com/user/emails", nil)
	if err != nil {
		return "", fmt.Errorf("création requête emails GitHub: %w", err)
	}

	req.Header.Set("Authorization", "token "+accessToken)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := s.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("requête emails GitHub API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("erreur emails GitHub API: status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("lecture emails GitHub: %w", err)
	}

	var emails []GitHubEmail
	if err := json.Unmarshal(body, &emails); err != nil {
		return "", fmt.Errorf("parsing emails GitHub: %w", err)
	}

	// Chercher l'email principal vérifié
	for _, email := range emails {
		if email.Primary && email.Verified {
			return email.Email, nil
		}
	}

	// Si pas d'email principal, prendre le premier vérifié
	for _, email := range emails {
		if email.Verified {
			return email.Email, nil
		}
	}

	return "", fmt.Errorf("aucun email vérifié trouvé")
}

// fetchDiscordUserInfo récupère les informations utilisateur Discord
func (s *OAuthService) fetchDiscordUserInfo(ctx context.Context, accessToken string) (*OAuthUserInfo, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://discord.com/api/users/@me", nil)
	if err != nil {
		return nil, fmt.Errorf("création requête Discord: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("requête Discord API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("erreur Discord API: status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("lecture réponse Discord: %w", err)
	}

	var discordUser DiscordUserInfo
	if err := json.Unmarshal(body, &discordUser); err != nil {
		return nil, fmt.Errorf("parsing réponse Discord: %w", err)
	}

	// Valider email vérifié
	if discordUser.Email == "" || !discordUser.Verified {
		return nil, fmt.Errorf("email Discord non vérifié")
	}

	// Construire l'URL de l'avatar Discord
	avatarURL := ""
	if discordUser.Avatar != "" {
		avatarURL = fmt.Sprintf("https://cdn.discordapp.com/avatars/%s/%s.png", discordUser.ID, discordUser.Avatar)
	}

	// Créer nom d'utilisateur unique (Discord peut avoir des doublons)
	username := fmt.Sprintf("%s_%s", discordUser.Username, discordUser.Discriminator)

	return &OAuthUserInfo{
		Email:       discordUser.Email,
		Username:    username,
		FirstName:   discordUser.Username,
		LastName:    "",
		DisplayName: discordUser.Username,
		Avatar:      avatarURL,
		Provider:    "discord",
		ProviderID:  discordUser.ID,
	}, nil
}

// splitFullName sépare un nom complet en prénom et nom
func (s *OAuthService) splitFullName(fullName string) (firstName, lastName string) {
	if fullName == "" {
		return "", ""
	}

	parts := strings.Fields(fullName)
	if len(parts) == 0 {
		return "", ""
	}

	firstName = parts[0]
	if len(parts) > 1 {
		lastName = strings.Join(parts[1:], " ")
	}

	return firstName, lastName
}

// ValidateOAuthState valide le state OAuth2 pour la sécurité
func (s *OAuthService) ValidateOAuthState(receivedState, expectedState string) error {
	if receivedState == "" || expectedState == "" {
		return fmt.Errorf("state OAuth2 manquant")
	}

	if receivedState != expectedState {
		return fmt.Errorf("state OAuth2 invalide")
	}

	return nil
}

// GenerateOAuthState génère un state OAuth2 sécurisé
func (s *OAuthService) GenerateOAuthState() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("génération state OAuth2: %w", err)
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}
