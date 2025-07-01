package auth

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/okinrev/veza-web-app/internal/response"
)

// MagicLinkHandler gestionnaire pour l'authentification par liens magiques
type MagicLinkHandler struct {
	service          *Service
	magicLinkService *MagicLinkService
}

// NewMagicLinkHandler crée une nouvelle instance du handler Magic Link
func NewMagicLinkHandler(service *Service, magicLinkService *MagicLinkService) *MagicLinkHandler {
	return &MagicLinkHandler{
		service:          service,
		magicLinkService: magicLinkService,
	}
}

// SendMagicLinkRequest requête d'envoi de Magic Link
type SendMagicLinkRequest struct {
	Email       string `json:"email" binding:"required,email"`
	RedirectURL string `json:"redirect_url,omitempty"`
}

// VerifyMagicLinkRequest requête de vérification Magic Link
type VerifyMagicLinkRequest struct {
	Token       string `json:"token" binding:"required"`
	RedirectURL string `json:"redirect_url,omitempty"`
}

// SendMagicLink envoie un lien magique par email
func (h *MagicLinkHandler) SendMagicLink(c *gin.Context) {
	var req SendMagicLinkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Données de requête invalides: "+err.Error())
		return
	}

	// Récupérer les informations de la requête
	ipAddress := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	// Envoyer le Magic Link
	err := h.magicLinkService.SendMagicLink(req.Email, req.RedirectURL, ipAddress, userAgent)
	if err != nil {
		response.InternalServerError(c, "Erreur envoi Magic Link: "+err.Error())
		return
	}

	response.Success(c, map[string]string{
		"message": "Si votre adresse email est associée à un compte, vous recevrez un lien de connexion.",
	}, "Magic Link envoyé")
}

// VerifyMagicLink vérifie un Magic Link et connecte l'utilisateur
func (h *MagicLinkHandler) VerifyMagicLink(c *gin.Context) {
	// Récupérer le token depuis les paramètres de requête (pour les liens dans les emails)
	token := c.Query("token")
	redirectURL := c.Query("redirect_url")

	// Si pas de token dans l'URL, essayer dans le body JSON
	if token == "" {
		var req VerifyMagicLinkRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			response.BadRequest(c, "Token requis")
			return
		}
		token = req.Token
		if redirectURL == "" {
			redirectURL = req.RedirectURL
		}
	}

	if token == "" {
		response.BadRequest(c, "Token Magic Link requis")
		return
	}

	// Récupérer l'adresse IP
	ipAddress := c.ClientIP()

	// Valider le Magic Link
	loginResponse, err := h.magicLinkService.ValidateMagicLink(token, ipAddress)
	if err != nil {
		response.BadRequest(c, "Magic Link invalide: "+err.Error())
		return
	}

	// Si c'est une requête GET (clic sur lien email), rediriger
	if c.Request.Method == "GET" {
		if redirectURL != "" {
			c.Redirect(http.StatusTemporaryRedirect, redirectURL+"?token="+loginResponse.AccessToken)
		} else {
			// Page de succès par défaut
			c.HTML(http.StatusOK, "magic_link_success.html", gin.H{
				"access_token":  loginResponse.AccessToken,
				"refresh_token": loginResponse.RefreshToken,
				"user":          loginResponse.User,
			})
		}
		return
	}

	// Réponse JSON pour les requêtes API
	response.Success(c, loginResponse, "Connexion Magic Link réussie")
}

// CheckMagicLink vérifie le statut d'un Magic Link
func (h *MagicLinkHandler) CheckMagicLink(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		response.BadRequest(c, "Token requis")
		return
	}

	status, err := h.magicLinkService.GetMagicLinkStatus(token)
	if err != nil {
		response.InternalServerError(c, "Erreur vérification Magic Link")
		return
	}

	response.Success(c, status, "Statut Magic Link récupéré")
}

// GetMagicLinkHistory récupère l'historique des Magic Links pour un utilisateur
func (h *MagicLinkHandler) GetMagicLinkHistory(c *gin.Context) {
	// Cette route nécessite une authentification
	userID, exists := h.getUserIDFromContext(c)
	if !exists {
		response.Unauthorized(c, "Authentification requise")
		return
	}

	history, err := h.getMagicLinkHistory(userID)
	if err != nil {
		response.InternalServerError(c, "Erreur récupération historique")
		return
	}

	response.Success(c, map[string]interface{}{
		"history": history,
	}, "Historique Magic Link récupéré")
}

// ============================================================================
// MÉTHODES HELPER
// ============================================================================

// getUserIDFromContext récupère l'ID utilisateur depuis le contexte
func (h *MagicLinkHandler) getUserIDFromContext(c *gin.Context) (int, bool) {
	userID, exists := c.Get("user_id")
	if !exists {
		return 0, false
	}

	id, ok := userID.(int)
	return id, ok
}

// getMagicLinkHistory récupère l'historique des Magic Links
func (h *MagicLinkHandler) getMagicLinkHistory(userID int) ([]MagicLinkHistoryItem, error) {
	rows, err := h.magicLinkService.db.Query(`
		SELECT 
			email, 
			created_at, 
			expires_at, 
			used_at, 
			ip_address,
			CASE WHEN used_at IS NOT NULL THEN 'used'
				 WHEN expires_at < NOW() THEN 'expired'
				 ELSE 'active' END as status
		FROM magic_links 
		WHERE user_id = $1 
		ORDER BY created_at DESC 
		LIMIT 50
	`, userID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var history []MagicLinkHistoryItem
	for rows.Next() {
		var item MagicLinkHistoryItem
		err := rows.Scan(
			&item.Email,
			&item.CreatedAt,
			&item.ExpiresAt,
			&item.UsedAt,
			&item.IPAddress,
			&item.Status,
		)
		if err != nil {
			continue
		}
		history = append(history, item)
	}

	return history, nil
}

// MagicLinkHistoryItem élément de l'historique Magic Link
type MagicLinkHistoryItem struct {
	Email     string  `json:"email"`
	CreatedAt string  `json:"created_at"`
	ExpiresAt string  `json:"expires_at"`
	UsedAt    *string `json:"used_at"`
	IPAddress string  `json:"ip_address"`
	Status    string  `json:"status"` // active, used, expired
}

// EmailSenderImpl implémentation basique de l'envoyeur d'emails
type EmailSenderImpl struct {
	// Configuration SMTP
}

// SendMagicLink implémentation de l'envoi d'email Magic Link
func (e *EmailSenderImpl) SendMagicLink(email, token, loginURL string) error {
	// TODO: Implémenter l'envoi d'email réel avec SMTP
	// Pour l'instant, log le lien pour le développement
	fmt.Printf("🔗 Magic Link pour %s: %s\n", email, loginURL)

	// Template email simplifié
	emailContent := fmt.Sprintf(`
		Bonjour,
		
		Cliquez sur le lien suivant pour vous connecter à Veza :
		%s
		
		Ce lien expire dans 15 minutes et ne peut être utilisé qu'une seule fois.
		
		Si vous n'avez pas demandé cette connexion, ignorez cet email.
		
		L'équipe Veza
	`, loginURL)

	// Log de l'email pour le développement
	fmt.Printf("📧 Email Magic Link envoyé à %s:\n%s\n", email, emailContent)

	return nil
}

// NewEmailSender crée une nouvelle instance de l'envoyeur d'emails
func NewEmailSender() EmailSender {
	return &EmailSenderImpl{}
}
