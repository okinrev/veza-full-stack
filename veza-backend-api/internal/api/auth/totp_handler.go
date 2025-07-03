package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/okinrev/veza-web-app/internal/common"
	"github.com/okinrev/veza-web-app/internal/response"
)

// TotpHandler gestionnaire pour l'authentification à deux facteurs
type TotpHandler struct {
	service     *Service
	totpService *TotpService
}

// NewTotpHandler crée une nouvelle instance du handler TOTP
func NewTotpHandler(service *Service, totpService *TotpService) *TotpHandler {
	return &TotpHandler{
		service:     service,
		totpService: totpService,
	}
}

// Setup2FARequest requête de configuration 2FA
type Setup2FARequest struct {
	Password string `json:"password" binding:"required"`
}

// Verify2FARequest requête de vérification 2FA
type Verify2FARequest struct {
	TOTPCode string `json:"totp_code" binding:"required,len=6"`
}

// Validate2FARequest requête de validation 2FA
type Validate2FARequest struct {
	TOTPCode   string `json:"totp_code,omitempty"`
	BackupCode string `json:"backup_code,omitempty"`
}

// Disable2FARequest requête de désactivation 2FA
type Disable2FARequest struct {
	Password string `json:"password" binding:"required"`
}

// RegenerateBackupCodesRequest requête de régénération codes
type RegenerateBackupCodesRequest struct {
	Password string `json:"password" binding:"required"`
}

// Setup2FA démarre la configuration 2FA
func (h *TotpHandler) Setup2FA(c *gin.Context) {
	var req Setup2FARequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Données de requête invalides: "+err.Error())
		return
	}

	// Récupérer l'utilisateur depuis le contexte
	userID, exists := common.GetUserIDFromContext(c)
	if !exists {
		response.Unauthorized(c, "Utilisateur non authentifié")
		return
	}

	// Récupérer les informations utilisateur
	user, err := h.service.GetUserByID(userID)
	if err != nil {
		response.InternalServerError(c, "Erreur récupération utilisateur")
		return
	}

	// Vérifier le mot de passe actuel
	// TODO: Implement password verification

	// Générer la configuration 2FA
	setup, err := h.totpService.Generate2FASetup(userID, user.Email)
	if err != nil {
		response.Error(c, http.StatusConflict, "Erreur configuration 2FA: "+err.Error())
		return
	}

	response.Success(c, setup, "Configuration 2FA générée. Veuillez scanner le QR code et entrer un code pour confirmer.")
}

// Verify2FA vérifie et active 2FA
func (h *TotpHandler) Verify2FA(c *gin.Context) {
	var req Verify2FARequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Code TOTP requis")
		return
	}

	userID, exists := common.GetUserIDFromContext(c)
	if !exists {
		response.Unauthorized(c, "Utilisateur non authentifié")
		return
	}

	// Vérifier et activer 2FA
	err := h.totpService.Verify2FASetup(userID, req.TOTPCode)
	if err != nil {
		response.BadRequest(c, "Code TOTP invalide: "+err.Error())
		return
	}

	response.Success(c, map[string]bool{
		"enabled": true,
	}, "2FA activé avec succès")
}

// Validate2FA valide un code 2FA lors de la connexion
func (h *TotpHandler) Validate2FA(c *gin.Context) {
	var req Validate2FARequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Code TOTP ou code de récupération requis")
		return
	}

	userID, exists := common.GetUserIDFromContext(c)
	if !exists {
		response.Unauthorized(c, "Utilisateur non authentifié")
		return
	}

	var err error

	// Valider code TOTP ou code de récupération
	if req.TOTPCode != "" {
		err = h.totpService.ValidateTOTP(userID, req.TOTPCode)
	} else if req.BackupCode != "" {
		err = h.totpService.ValidateBackupCode(userID, req.BackupCode)
	} else {
		response.BadRequest(c, "Code TOTP ou code de récupération requis")
		return
	}

	if err != nil {
		response.BadRequest(c, "Code invalide: "+err.Error())
		return
	}

	response.Success(c, map[string]bool{
		"validated": true,
	}, "Code 2FA validé")
}

// Get2FAStatus récupère le statut 2FA de l'utilisateur
func (h *TotpHandler) Get2FAStatus(c *gin.Context) {
	userID, exists := common.GetUserIDFromContext(c)
	if !exists {
		response.Unauthorized(c, "Utilisateur non authentifié")
		return
	}

	status, err := h.totpService.GetTwoFactorStatus(userID)
	if err != nil {
		response.InternalServerError(c, "Erreur récupération statut 2FA")
		return
	}

	response.Success(c, status, "Statut 2FA récupéré")
}

// Disable2FA désactive l'authentification à deux facteurs
func (h *TotpHandler) Disable2FA(c *gin.Context) {
	var req Disable2FARequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Mot de passe requis")
		return
	}

	userID, exists := common.GetUserIDFromContext(c)
	if !exists {
		response.Unauthorized(c, "Utilisateur non authentifié")
		return
	}

	err := h.totpService.Disable2FA(userID, req.Password)
	if err != nil {
		response.BadRequest(c, "Erreur désactivation 2FA: "+err.Error())
		return
	}

	response.Success(c, map[string]bool{
		"disabled": true,
	}, "2FA désactivé avec succès")
}

// RegenerateBackupCodes génère de nouveaux codes de récupération
func (h *TotpHandler) RegenerateBackupCodes(c *gin.Context) {
	var req RegenerateBackupCodesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Mot de passe requis")
		return
	}

	userID, exists := common.GetUserIDFromContext(c)
	if !exists {
		response.Unauthorized(c, "Utilisateur non authentifié")
		return
	}

	newCodes, err := h.totpService.RegenerateBackupCodes(userID, req.Password)
	if err != nil {
		response.BadRequest(c, "Erreur régénération codes: "+err.Error())
		return
	}

	response.Success(c, map[string][]string{
		"backup_codes": newCodes,
	}, "Nouveaux codes de récupération générés")
}

// GetUserByID méthode helper pour récupérer un utilisateur par ID
func (s *Service) GetUserByID(userID int64) (*User, error) {
	var user User
	err := s.db.QueryRow(`
		SELECT id, username, email, role, created_at, updated_at
		FROM users WHERE id = $1 AND role != 'deleted'
	`, userID).Scan(
		&user.ID, &user.Username, &user.Email,
		&user.Role, &user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &user, nil
}

// User structure utilisateur simplifiée
type User struct {
	ID        int    `json:"id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	Role      string `json:"role"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}
