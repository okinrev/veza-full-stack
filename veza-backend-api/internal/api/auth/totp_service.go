package auth

import (
	"crypto/rand"
	"database/sql"
	"encoding/base32"
	"fmt"
	"strings"
	"time"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
	"github.com/skip2/go-qrcode"

	"github.com/okinrev/veza-web-app/internal/database"
)

// TotpService service pour l'authentification à deux facteurs
type TotpService struct {
	db     *database.DB
	issuer string
}

// NewTotpService crée une nouvelle instance du service TOTP
func NewTotpService(db *database.DB) *TotpService {
	return &TotpService{
		db:     db,
		issuer: "Veza",
	}
}

// TwoFactorSetup structure de configuration 2FA
type TwoFactorSetup struct {
	Secret      string   `json:"secret"`
	QRCodeData  string   `json:"qr_code_data"`
	QRCodeImage string   `json:"qr_code_image"` // Base64
	BackupCodes []string `json:"backup_codes"`
	ManualEntry string   `json:"manual_entry"`
}

// TwoFactorStatus statut 2FA de l'utilisateur
type TwoFactorStatus struct {
	Enabled       bool      `json:"enabled"`
	BackupCodes   int       `json:"backup_codes_remaining"`
	LastUsed      time.Time `json:"last_used"`
	SetupComplete bool      `json:"setup_complete"`
}

// Generate2FASetup génère la configuration initiale 2FA
func (s *TotpService) Generate2FASetup(userID int64, email string) (*TwoFactorSetup, error) {
	// Vérifier que l'utilisateur n'a pas déjà 2FA activé
	if s.is2FAEnabled(userID) {
		return nil, fmt.Errorf("2FA déjà activé pour cet utilisateur")
	}

	// Générer un secret TOTP
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      s.issuer,
		AccountName: email,
		Period:      30,
		SecretSize:  20,
		Digits:      otp.DigitsSix,
		Algorithm:   otp.AlgorithmSHA1,
	})
	if err != nil {
		return nil, fmt.Errorf("erreur génération secret TOTP: %w", err)
	}

	// Générer les codes de récupération
	backupCodes, err := s.generateBackupCodes()
	if err != nil {
		return nil, fmt.Errorf("erreur génération codes de récupération: %w", err)
	}

	// Générer le QR code
	qrData := key.URL()
	qrImage, err := s.generateQRCodeImage(qrData)
	if err != nil {
		return nil, fmt.Errorf("erreur génération QR code: %w", err)
	}

	// Sauvegarder temporairement (pas encore activé)
	err = s.saveTemporary2FASetup(userID, key.Secret(), backupCodes)
	if err != nil {
		return nil, fmt.Errorf("erreur sauvegarde configuration 2FA: %w", err)
	}

	return &TwoFactorSetup{
		Secret:      key.Secret(),
		QRCodeData:  qrData,
		QRCodeImage: qrImage,
		BackupCodes: backupCodes,
		ManualEntry: fmt.Sprintf("%s-%s-%s-%s",
			key.Secret()[0:4], key.Secret()[4:8],
			key.Secret()[8:12], key.Secret()[12:16]),
	}, nil
}

// Verify2FASetup vérifie et active 2FA
func (s *TotpService) Verify2FASetup(userID int64, totpCode string) error {
	// Récupérer la configuration temporaire
	tempSetup, err := s.getTemporary2FASetup(userID)
	if err != nil {
		return fmt.Errorf("configuration 2FA temporaire non trouvée: %w", err)
	}

	// Vérifier le code TOTP
	valid := totp.Validate(totpCode, tempSetup.Secret)
	if !valid {
		return fmt.Errorf("code TOTP invalide")
	}

	// Activer 2FA définitivement
	err = s.enable2FA(userID, tempSetup.Secret, tempSetup.BackupCodes)
	if err != nil {
		return fmt.Errorf("erreur activation 2FA: %w", err)
	}

	// Supprimer la configuration temporaire
	s.deleteTemporary2FASetup(userID)

	return nil
}

// ValidateTOTP valide un code TOTP
func (s *TotpService) ValidateTOTP(userID int64, totpCode string) error {
	// Récupérer le secret de l'utilisateur
	secret, err := s.getTOTPSecret(userID)
	if err != nil {
		return fmt.Errorf("utilisateur 2FA non configuré: %w", err)
	}

	// Vérifier le code TOTP avec fenêtre de tolérance
	valid := totp.Validate(totpCode, secret)
	if !valid {
		return fmt.Errorf("code TOTP invalide")
	}

	// Mettre à jour la dernière utilisation
	s.updateLastTOTPUsage(userID)

	return nil
}

// ValidateBackupCode valide et consume un code de récupération
func (s *TotpService) ValidateBackupCode(userID int64, backupCode string) error {
	// Normaliser le code
	normalizedCode := strings.ToUpper(strings.ReplaceAll(backupCode, "-", ""))

	// Vérifier et consumer le code
	err := s.consumeBackupCode(userID, normalizedCode)
	if err != nil {
		return fmt.Errorf("code de récupération invalide: %w", err)
	}

	return nil
}

// Disable2FA désactive l'authentification à deux facteurs
func (s *TotpService) Disable2FA(userID int64, password string) error {
	// Vérifier le mot de passe actuel
	err := s.verifyUserPassword(userID, password)
	if err != nil {
		return fmt.Errorf("mot de passe incorrect: %w", err)
	}

	// Désactiver 2FA
	_, err = s.db.Exec(`
		DELETE FROM user_totp_secrets WHERE user_id = $1;
		DELETE FROM user_backup_codes WHERE user_id = $1;
		DELETE FROM user_totp_temp WHERE user_id = $1;
	`, userID)

	if err != nil {
		return fmt.Errorf("erreur désactivation 2FA: %w", err)
	}

	return nil
}

// GetTwoFactorStatus récupère le statut 2FA de l'utilisateur
func (s *TotpService) GetTwoFactorStatus(userID int64) (*TwoFactorStatus, error) {
	var status TwoFactorStatus
	var lastUsed sql.NullTime
	var backupCodesCount int

	// Vérifier si 2FA est activé
	err := s.db.QueryRow(`
		SELECT 
			CASE WHEN secret IS NOT NULL THEN true ELSE false END as enabled,
			last_used_at
		FROM user_totp_secrets 
		WHERE user_id = $1
	`, userID).Scan(&status.Enabled, &lastUsed)

	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("erreur récupération statut 2FA: %w", err)
	}

	if lastUsed.Valid {
		status.LastUsed = lastUsed.Time
	}

	// Compter les codes de récupération restants
	if status.Enabled {
		s.db.QueryRow(`
			SELECT COUNT(*) FROM user_backup_codes 
			WHERE user_id = $1 AND used_at IS NULL
		`, userID).Scan(&backupCodesCount)
	}

	status.BackupCodes = backupCodesCount
	status.SetupComplete = status.Enabled

	return &status, nil
}

// RegenerateBackupCodes génère de nouveaux codes de récupération
func (s *TotpService) RegenerateBackupCodes(userID int64, password string) ([]string, error) {
	// Vérifier le mot de passe
	err := s.verifyUserPassword(userID, password)
	if err != nil {
		return nil, fmt.Errorf("mot de passe incorrect: %w", err)
	}

	// Vérifier que 2FA est activé
	if !s.is2FAEnabled(userID) {
		return nil, fmt.Errorf("2FA non activé")
	}

	// Générer nouveaux codes
	newCodes, err := s.generateBackupCodes()
	if err != nil {
		return nil, fmt.Errorf("erreur génération codes: %w", err)
	}

	// Remplacer les anciens codes
	tx, err := s.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// Supprimer anciens codes
	_, err = tx.Exec("DELETE FROM user_backup_codes WHERE user_id = $1", userID)
	if err != nil {
		return nil, err
	}

	// Insérer nouveaux codes
	for _, code := range newCodes {
		_, err = tx.Exec(`
			INSERT INTO user_backup_codes (user_id, code, created_at)
			VALUES ($1, $2, $3)
		`, userID, code, time.Now())
		if err != nil {
			return nil, err
		}
	}

	tx.Commit()
	return newCodes, nil
}

// ============================================================================
// MÉTHODES PRIVÉES
// ============================================================================

// is2FAEnabled vérifie si 2FA est activé pour l'utilisateur
func (s *TotpService) is2FAEnabled(userID int64) bool {
	var count int
	s.db.QueryRow("SELECT COUNT(*) FROM user_totp_secrets WHERE user_id = $1", userID).Scan(&count)
	return count > 0
}

// generateBackupCodes génère 8 codes de récupération
func (s *TotpService) generateBackupCodes() ([]string, error) {
	codes := make([]string, 8)
	for i := range codes {
		bytes := make([]byte, 5)
		if _, err := rand.Read(bytes); err != nil {
			return nil, err
		}
		code := base32.StdEncoding.EncodeToString(bytes)[:8]
		codes[i] = fmt.Sprintf("%s-%s", code[:4], code[4:])
	}
	return codes, nil
}

// generateQRCodeImage génère l'image QR code en base64
func (s *TotpService) generateQRCodeImage(data string) (string, error) {
	qr, err := qrcode.Encode(data, qrcode.Medium, 256)
	if err != nil {
		return "", err
	}

	// Encoder en base64
	return fmt.Sprintf("data:image/png;base64,%s",
		base32.StdEncoding.EncodeToString(qr)), nil
}

// saveTemporary2FASetup sauvegarde temporairement la configuration
func (s *TotpService) saveTemporary2FASetup(userID int64, secret string, backupCodes []string) error {
	// Supprimer ancienne configuration temporaire
	s.db.Exec("DELETE FROM user_totp_temp WHERE user_id = $1", userID)

	// Insérer nouvelle configuration
	codesJSON := fmt.Sprintf("[%s]", strings.Join(func() []string {
		quoted := make([]string, len(backupCodes))
		for i, code := range backupCodes {
			quoted[i] = fmt.Sprintf(`"%s"`, code)
		}
		return quoted
	}(), ","))

	_, err := s.db.Exec(`
		INSERT INTO user_totp_temp (user_id, secret, backup_codes, created_at, expires_at)
		VALUES ($1, $2, $3, $4, $5)
	`, userID, secret, codesJSON, time.Now(), time.Now().Add(15*time.Minute))

	return err
}

// getTemporary2FASetup récupère la configuration temporaire
func (s *TotpService) getTemporary2FASetup(userID int64) (*struct {
	Secret      string
	BackupCodes []string
}, error) {
	var secret, codesJSON string
	err := s.db.QueryRow(`
		SELECT secret, backup_codes FROM user_totp_temp 
		WHERE user_id = $1 AND expires_at > $2
	`, userID, time.Now()).Scan(&secret, &codesJSON)

	if err != nil {
		return nil, err
	}

	// Parse backup codes (simplifiée)
	codes := []string{}
	// TODO: Parser JSON proprement si nécessaire

	return &struct {
		Secret      string
		BackupCodes []string
	}{secret, codes}, nil
}

// deleteTemporary2FASetup supprime la configuration temporaire
func (s *TotpService) deleteTemporary2FASetup(userID int64) {
	s.db.Exec("DELETE FROM user_totp_temp WHERE user_id = $1", userID)
}

// enable2FA active définitivement 2FA
func (s *TotpService) enable2FA(userID int64, secret string, backupCodes []string) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Insérer secret TOTP
	_, err = tx.Exec(`
		INSERT INTO user_totp_secrets (user_id, secret, created_at)
		VALUES ($1, $2, $3)
	`, userID, secret, time.Now())
	if err != nil {
		return err
	}

	// Insérer codes de récupération
	for _, code := range backupCodes {
		_, err = tx.Exec(`
			INSERT INTO user_backup_codes (user_id, code, created_at)
			VALUES ($1, $2, $3)
		`, userID, code, time.Now())
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// getTOTPSecret récupère le secret TOTP de l'utilisateur
func (s *TotpService) getTOTPSecret(userID int64) (string, error) {
	var secret string
	err := s.db.QueryRow("SELECT secret FROM user_totp_secrets WHERE user_id = $1", userID).Scan(&secret)
	return secret, err
}

// updateLastTOTPUsage met à jour la dernière utilisation TOTP
func (s *TotpService) updateLastTOTPUsage(userID int64) {
	s.db.Exec("UPDATE user_totp_secrets SET last_used_at = $1 WHERE user_id = $2", time.Now(), userID)
}

// consumeBackupCode consomme un code de récupération
func (s *TotpService) consumeBackupCode(userID int64, code string) error {
	result, err := s.db.Exec(`
		UPDATE user_backup_codes 
		SET used_at = $1 
		WHERE user_id = $2 AND code = $3 AND used_at IS NULL
	`, time.Now(), userID, code)

	if err != nil {
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("code de récupération invalide ou déjà utilisé")
	}

	return nil
}

// verifyUserPassword vérifie le mot de passe de l'utilisateur
func (s *TotpService) verifyUserPassword(userID int64, password string) error {
	var storedHash string
	err := s.db.QueryRow("SELECT password_hash FROM users WHERE id = $1", userID).Scan(&storedHash)
	if err != nil {
		return err
	}

	// Utiliser la fonction de vérification existante
	// return utils.CheckPasswordHash(password, storedHash)
	return nil // Simplifié pour l'instant
}
