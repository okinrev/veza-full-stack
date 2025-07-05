package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/okinrev/veza-web-app/internal/response"
	"github.com/okinrev/veza-web-app/internal/utils"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
	"golang.org/x/oauth2/google"
)

// OAuthHandler gestionnaire pour les authentifications OAuth2
type OAuthHandler struct {
	service       *Service
	googleConfig  *oauth2.Config
	githubConfig  *oauth2.Config
	discordConfig *oauth2.Config
}

// NewOAuthHandler crée une nouvelle instance du handler OAuth2
func NewOAuthHandler(service *Service) *OAuthHandler {
	return &OAuthHandler{
		service:       service,
		googleConfig:  setupGoogleOAuth(),
		githubConfig:  setupGitHubOAuth(),
		discordConfig: setupDiscordOAuth(),
	}
}

// GoogleUserInfo structure des données utilisateur Google
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

// GitHubUserInfo structure des données utilisateur GitHub
type GitHubUserInfo struct {
	ID        int    `json:"id"`
	Login     string `json:"login"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	AvatarURL string `json:"avatar_url"`
	Company   string `json:"company"`
	Bio       string `json:"bio"`
}

// GitHubEmail structure pour les emails GitHub
type GitHubEmail struct {
	Email      string `json:"email"`
	Primary    bool   `json:"primary"`
	Verified   bool   `json:"verified"`
	Visibility string `json:"visibility"`
}

// DiscordUserInfo structure des données utilisateur Discord
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

// OAuth2StateStore stockage temporaire des states OAuth2
var oauthStateStore = make(map[string]time.Time)

// setupGoogleOAuth configure le client Google OAuth2
func setupGoogleOAuth() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     getEnvOrDefault("OAUTH_GOOGLE_CLIENT_ID", ""),
		ClientSecret: getEnvOrDefault("OAUTH_GOOGLE_CLIENT_SECRET", ""),
		RedirectURL:  getEnvOrDefault("OAUTH_GOOGLE_REDIRECT_URL", "http://localhost:8080/api/v1/auth/oauth/google/callback"),
		Scopes:       []string{"openid", "profile", "email"},
		Endpoint:     google.Endpoint,
	}
}

// setupGitHubOAuth configure le client GitHub OAuth2
func setupGitHubOAuth() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     getEnvOrDefault("OAUTH_GITHUB_CLIENT_ID", ""),
		ClientSecret: getEnvOrDefault("OAUTH_GITHUB_CLIENT_SECRET", ""),
		RedirectURL:  getEnvOrDefault("OAUTH_GITHUB_REDIRECT_URL", "http://localhost:8080/api/v1/auth/oauth/github/callback"),
		Scopes:       []string{"user:email"},
		Endpoint:     github.Endpoint,
	}
}

// setupDiscordOAuth configure le client Discord OAuth2
func setupDiscordOAuth() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     getEnvOrDefault("OAUTH_DISCORD_CLIENT_ID", ""),
		ClientSecret: getEnvOrDefault("OAUTH_DISCORD_CLIENT_SECRET", ""),
		RedirectURL:  getEnvOrDefault("OAUTH_DISCORD_REDIRECT_URL", "http://localhost:8080/api/v1/auth/oauth/discord/callback"),
		Scopes:       []string{"identify", "email"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://discord.com/api/oauth2/authorize",
			TokenURL: "https://discord.com/api/oauth2/token",
		},
	}
}

// GoogleOAuthURL génère l'URL d'authentification Google
func (h *OAuthHandler) GoogleOAuthURL(c *gin.Context) {
	state := h.generateSecureState()
	oauthStateStore[state] = time.Now().Add(10 * time.Minute)

	url := h.googleConfig.AuthCodeURL(state, oauth2.AccessTypeOffline)

	response.Success(c, map[string]string{
		"auth_url": url,
		"state":    state,
	}, "URL d'authentification Google générée")
}

// GoogleOAuthCallback traite le callback Google OAuth2
func (h *OAuthHandler) GoogleOAuthCallback(c *gin.Context) {
	code := c.Query("code")
	state := c.Query("state")

	if !h.validateState(state) {
		response.BadRequest(c, "State OAuth2 invalide")
		return
	}

	// Échanger le code contre un token
	token, err := h.googleConfig.Exchange(context.Background(), code)
	if err != nil {
		response.InternalServerError(c, "Erreur échange token Google")
		return
	}

	// Récupérer les informations utilisateur
	userInfo, err := h.fetchGoogleUserInfo(token.AccessToken)
	if err != nil {
		response.InternalServerError(c, "Erreur récupération profil Google")
		return
	}

	// Authentifier ou créer l'utilisateur
	authResponse, err := h.authenticateOAuthUser(userInfo.Email, userInfo.Name, "google", userInfo)
	if err != nil {
		response.InternalServerError(c, "Erreur authentification OAuth")
		return
	}

	response.Success(c, authResponse, "Connexion Google réussie")
}

// GitHubOAuthURL génère l'URL d'authentification GitHub
func (h *OAuthHandler) GitHubOAuthURL(c *gin.Context) {
	state := h.generateSecureState()
	oauthStateStore[state] = time.Now().Add(10 * time.Minute)

	url := h.githubConfig.AuthCodeURL(state, oauth2.AccessTypeOffline)

	response.Success(c, map[string]string{
		"auth_url": url,
		"state":    state,
	}, "URL d'authentification GitHub générée")
}

// GitHubOAuthCallback traite le callback GitHub OAuth2
func (h *OAuthHandler) GitHubOAuthCallback(c *gin.Context) {
	code := c.Query("code")
	state := c.Query("state")

	if !h.validateState(state) {
		response.BadRequest(c, "State OAuth2 invalide")
		return
	}

	// Échanger le code contre un token
	token, err := h.githubConfig.Exchange(context.Background(), code)
	if err != nil {
		response.InternalServerError(c, "Erreur échange token GitHub")
		return
	}

	// Récupérer les informations utilisateur
	userInfo, err := h.fetchGitHubUserInfo(token.AccessToken)
	if err != nil {
		response.InternalServerError(c, "Erreur récupération profil GitHub")
		return
	}

	// Si pas d'email public, récupérer les emails
	if userInfo.Email == "" {
		email, err := h.fetchGitHubEmail(token.AccessToken)
		if err != nil {
			response.InternalServerError(c, "Erreur récupération email GitHub")
			return
		}
		userInfo.Email = email
	}

	// Authentifier ou créer l'utilisateur
	authResponse, err := h.authenticateOAuthUser(userInfo.Email, userInfo.Login, "github", userInfo)
	if err != nil {
		response.InternalServerError(c, "Erreur authentification OAuth")
		return
	}

	response.Success(c, authResponse, "Connexion GitHub réussie")
}

// DiscordOAuthURL génère l'URL d'authentification Discord
func (h *OAuthHandler) DiscordOAuthURL(c *gin.Context) {
	state := h.generateSecureState()
	oauthStateStore[state] = time.Now().Add(10 * time.Minute)

	url := h.discordConfig.AuthCodeURL(state, oauth2.AccessTypeOffline)

	response.Success(c, map[string]string{
		"auth_url": url,
		"state":    state,
	}, "URL d'authentification Discord générée")
}

// DiscordOAuthCallback traite le callback Discord OAuth2
func (h *OAuthHandler) DiscordOAuthCallback(c *gin.Context) {
	code := c.Query("code")
	state := c.Query("state")

	if !h.validateState(state) {
		response.BadRequest(c, "State OAuth2 invalide")
		return
	}

	// Échanger le code contre un token
	token, err := h.discordConfig.Exchange(context.Background(), code)
	if err != nil {
		response.InternalServerError(c, "Erreur échange token Discord")
		return
	}

	// Récupérer les informations utilisateur
	userInfo, err := h.fetchDiscordUserInfo(token.AccessToken)
	if err != nil {
		response.InternalServerError(c, "Erreur récupération profil Discord")
		return
	}

	// Authentifier ou créer l'utilisateur
	username := fmt.Sprintf("%s_%s", userInfo.Username, userInfo.Discriminator)
	authResponse, err := h.authenticateOAuthUser(userInfo.Email, username, "discord", userInfo)
	if err != nil {
		response.InternalServerError(c, "Erreur authentification OAuth")
		return
	}

	response.Success(c, authResponse, "Connexion Discord réussie")
}

// fetchGoogleUserInfo récupère les informations utilisateur Google
func (h *OAuthHandler) fetchGoogleUserInfo(accessToken string) (*GoogleUserInfo, error) {
	resp, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + accessToken)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("erreur API Google: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var userInfo GoogleUserInfo
	if err := json.Unmarshal(body, &userInfo); err != nil {
		return nil, err
	}

	if !userInfo.VerifiedEmail {
		return nil, fmt.Errorf("email Google non vérifié")
	}

	return &userInfo, nil
}

// fetchGitHubUserInfo récupère les informations utilisateur GitHub
func (h *OAuthHandler) fetchGitHubUserInfo(accessToken string) (*GitHubUserInfo, error) {
	req, err := http.NewRequest("GET", "https://api.github.com/user", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "token "+accessToken)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("erreur API GitHub: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var userInfo GitHubUserInfo
	if err := json.Unmarshal(body, &userInfo); err != nil {
		return nil, err
	}

	return &userInfo, nil
}

// fetchGitHubEmail récupère l'email principal GitHub
func (h *OAuthHandler) fetchGitHubEmail(accessToken string) (string, error) {
	req, err := http.NewRequest("GET", "https://api.github.com/user/emails", nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", "token "+accessToken)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("erreur API GitHub emails: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var emails []GitHubEmail
	if err := json.Unmarshal(body, &emails); err != nil {
		return "", err
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
func (h *OAuthHandler) fetchDiscordUserInfo(accessToken string) (*DiscordUserInfo, error) {
	req, err := http.NewRequest("GET", "https://discord.com/api/users/@me", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("erreur API Discord: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var userInfo DiscordUserInfo
	if err := json.Unmarshal(body, &userInfo); err != nil {
		return nil, err
	}

	if !userInfo.Verified {
		return nil, fmt.Errorf("email Discord non vérifié")
	}

	return &userInfo, nil
}

// authenticateOAuthUser authentifie ou crée un utilisateur OAuth
func (h *OAuthHandler) authenticateOAuthUser(email, username, provider string, userInfo interface{}) (*LoginResponse, error) {
	// Chercher utilisateur existant par email
	user, err := h.service.GetUserByEmail(email)
	if err == nil && user != nil {
		// Utilisateur existe, générer les tokens
		accessToken, refreshToken, expiresIn, err := utils.GenerateTokenPair(user.ID, user.Username, user.Role, "your-jwt-secret") // TODO: Injecter le secret
		if err != nil {
			return nil, fmt.Errorf("génération tokens: %w", err)
		}

		return &LoginResponse{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
			User:         user,
			ExpiresIn:    expiresIn,
		}, nil
	}

	// Utilisateur n'existe pas, le créer
	registerReq := RegisterRequest{
		Username: h.generateUniqueUsername(username),
		Email:    email,
		Password: h.generateRandomPassword(), // Mot de passe aléatoire (OAuth)
	}

	newUser, err := h.service.Register(registerReq)
	if err != nil {
		return nil, fmt.Errorf("création utilisateur OAuth: %w", err)
	}

	// Générer les tokens pour le nouvel utilisateur
	accessToken, refreshToken, expiresIn, err := utils.GenerateTokenPair(newUser.ID, newUser.Username, newUser.Role, "your-jwt-secret") // TODO: Injecter le secret
	if err != nil {
		return nil, fmt.Errorf("génération tokens nouvel utilisateur: %w", err)
	}

	return &LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         newUser,
		ExpiresIn:    expiresIn,
	}, nil
}

// generateSecureState génère un state OAuth2 sécurisé
func (h *OAuthHandler) generateSecureState() string {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		panic(err)
	}
	return base64.URLEncoding.EncodeToString(bytes)
}

// validateState valide le state OAuth2
func (h *OAuthHandler) validateState(state string) bool {
	if state == "" {
		return false
	}

	expiry, exists := oauthStateStore[state]
	if !exists {
		return false
	}

	if time.Now().After(expiry) {
		delete(oauthStateStore, state)
		return false
	}

	delete(oauthStateStore, state)
	return true
}

// generateUniqueUsername génère un nom d'utilisateur unique
func (h *OAuthHandler) generateUniqueUsername(base string) string {
	// Nettoyer le nom de base
	cleaned := strings.ToLower(strings.ReplaceAll(base, " ", "_"))

	// Vérifier s'il est unique
	if !h.service.UsernameExists(cleaned) {
		return cleaned
	}

	// Ajouter un suffix numérique
	for i := 1; i <= 999; i++ {
		candidate := fmt.Sprintf("%s_%d", cleaned, i)
		if !h.service.UsernameExists(candidate) {
			return candidate
		}
	}

	// Fallback avec timestamp
	return fmt.Sprintf("%s_%d", cleaned, time.Now().Unix())
}

// generateRandomPassword génère un mot de passe aléatoire pour OAuth
func (h *OAuthHandler) generateRandomPassword() string {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		panic(err)
	}
	return base64.URLEncoding.EncodeToString(bytes)[:32]
}

// getEnvOrDefault récupère une variable d'environnement avec valeur par défaut
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
