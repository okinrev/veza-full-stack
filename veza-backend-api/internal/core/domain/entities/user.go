package entities

import (
	"errors"
	"time"
	"unicode"

	"golang.org/x/crypto/bcrypt"
)

// UserRole définit les rôles possibles pour un utilisateur
type UserRole string

const (
	RoleGuest          UserRole = "guest"
	UserRoleGuest      UserRole = "guest" // Alias pour compatibilité
	RoleUser           UserRole = "user"
	UserRoleUser       UserRole = "user" // Alias pour compatibilité
	RolePremium        UserRole = "premium"
	UserRolePremium    UserRole = "premium" // Alias pour compatibilité
	RoleModerator      UserRole = "moderator"
	UserRoleModerator  UserRole = "moderator" // Alias pour compatibilité
	RoleAdmin          UserRole = "admin"
	UserRoleAdmin      UserRole = "admin" // Alias pour compatibilité
	RoleSuperAdmin     UserRole = "superadmin"
	UserRoleSuperAdmin UserRole = "superadmin" // Alias pour compatibilité
)

// UserStatus définit les statuts possibles pour un utilisateur
type UserStatus string

const (
	StatusActive     UserStatus = "active"
	UserStatusActive UserStatus = "active" // Alias pour compatibilité
	StatusInactive   UserStatus = "inactive"
	StatusSuspended  UserStatus = "suspended"
	StatusBanned     UserStatus = "banned"
	StatusDeleted    UserStatus = "deleted"
)

// User représente l'entité utilisateur avec toutes les propriétés
type User struct {
	// Identifiants
	ID       int64  `json:"id" db:"id"`
	UUID     string `json:"uuid" db:"uuid"`
	Username string `json:"username" db:"username"`
	Email    string `json:"email" db:"email"`

	// Authentification
	PasswordHash string `json:"-" db:"password_hash"`
	Salt         string `json:"-" db:"salt"`

	// Profil
	FirstName   string `json:"first_name" db:"first_name"`
	LastName    string `json:"last_name" db:"last_name"`
	DisplayName string `json:"display_name" db:"display_name"`
	Avatar      string `json:"avatar" db:"avatar"`
	Bio         string `json:"bio" db:"bio"`

	// Autorisation et sécurité
	Role                   UserRole   `json:"role" db:"role"`
	Status                 UserStatus `json:"status" db:"status"`
	EmailVerified          bool       `json:"email_verified" db:"email_verified"`
	EmailVerificationToken string     `json:"-" db:"email_verification_token"`
	TwoFactorEnabled       bool       `json:"two_factor_enabled" db:"two_factor_enabled"`
	TwoFactorSecret        string     `json:"-" db:"two_factor_secret"`

	// Activité et présence
	IsOnline      bool       `json:"is_online" db:"is_online"`
	LastSeen      time.Time  `json:"last_seen" db:"last_seen"`
	LastLoginIP   string     `json:"last_login_ip" db:"last_login_ip"`
	LoginAttempts int        `json:"login_attempts" db:"login_attempts"`
	LockedUntil   *time.Time `json:"locked_until,omitempty" db:"locked_until"`

	// Préférences
	Timezone string `json:"timezone" db:"timezone"`
	Language string `json:"language" db:"language"`
	Theme    string `json:"theme" db:"theme"`

	// Métadonnées
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`

	// Statistiques (calculées)
	MessageCount int64 `json:"message_count,omitempty" db:"message_count"`
	StreamCount  int64 `json:"stream_count,omitempty" db:"stream_count"`
}

// NewUser crée une nouvelle entité User avec validation
func NewUser(username, email, password string) (*User, error) {
	user := &User{
		Username:      username,
		Email:         email,
		Role:          RoleUser,
		Status:        StatusActive,
		EmailVerified: false,
		IsOnline:      false,
		LoginAttempts: 0,
		Timezone:      "UTC",
		Language:      "en",
		Theme:         "light",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	// Validation
	if err := user.Validate(); err != nil {
		return nil, err
	}

	// Hash du mot de passe
	if err := user.SetPassword(password); err != nil {
		return nil, err
	}

	return user, nil
}

// Validate vérifie que l'entité User est valide
func (u *User) Validate() error {
	if err := u.ValidateUsername(); err != nil {
		return err
	}

	if err := u.ValidateEmail(); err != nil {
		return err
	}

	if err := u.ValidateRole(); err != nil {
		return err
	}

	if err := u.ValidateStatus(); err != nil {
		return err
	}

	return nil
}

// ValidateUsername vérifie que le nom d'utilisateur est valide
func (u *User) ValidateUsername() error {
	if u.Username == "" {
		return errors.New("username is required")
	}

	if len(u.Username) < 3 {
		return errors.New("username must be at least 3 characters long")
	}

	if len(u.Username) > 50 {
		return errors.New("username must be less than 50 characters long")
	}

	// Validation des caractères (alphanumériques + underscore + tiret)
	for _, r := range u.Username {
		if !unicode.IsLetter(r) && !unicode.IsNumber(r) && r != '_' && r != '-' {
			return errors.New("username can only contain letters, numbers, underscores and hyphens")
		}
	}

	return nil
}

// ValidateEmail vérifie que l'email est valide
func (u *User) ValidateEmail() error {
	if u.Email == "" {
		return errors.New("email is required")
	}

	// Validation basique d'email
	if !isValidEmail(u.Email) {
		return errors.New("invalid email format")
	}

	return nil
}

// ValidateRole vérifie que le rôle est valide
func (u *User) ValidateRole() error {
	switch u.Role {
	case RoleGuest, RoleUser, RolePremium, RoleModerator, RoleAdmin, RoleSuperAdmin:
		return nil
	default:
		return errors.New("invalid user role")
	}
}

// ValidateStatus vérifie que le statut est valide
func (u *User) ValidateStatus() error {
	switch u.Status {
	case StatusActive, StatusInactive, StatusSuspended, StatusBanned, StatusDeleted:
		return nil
	default:
		return errors.New("invalid user status")
	}
}

// SetPassword hash et définit le mot de passe
func (u *User) SetPassword(password string) error {
	if err := ValidatePassword(password); err != nil {
		return err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	u.PasswordHash = string(hash)
	u.UpdatedAt = time.Now()
	return nil
}

// CheckPassword vérifie si le mot de passe fourni correspond
func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password))
	return err == nil
}

// IsActive vérifie si l'utilisateur est actif
func (u *User) IsActive() bool {
	return u.Status == StatusActive && !u.IsLocked()
}

// IsLocked vérifie si le compte est verrouillé
func (u *User) IsLocked() bool {
	if u.LockedUntil == nil {
		return false
	}
	return time.Now().Before(*u.LockedUntil)
}

// CanLogin vérifie si l'utilisateur peut se connecter
func (u *User) CanLogin() bool {
	return u.IsActive() && !u.IsLocked()
}

// HasRole vérifie si l'utilisateur a un rôle spécifique
func (u *User) HasRole(role UserRole) bool {
	return u.Role == role
}

// HasAnyRole vérifie si l'utilisateur a un des rôles spécifiés
func (u *User) HasAnyRole(roles ...UserRole) bool {
	for _, role := range roles {
		if u.Role == role {
			return true
		}
	}
	return false
}

// IsAdmin vérifie si l'utilisateur est admin ou superadmin
func (u *User) IsAdmin() bool {
	return u.HasAnyRole(RoleAdmin, RoleSuperAdmin)
}

// IsModerator vérifie si l'utilisateur peut modérer
func (u *User) IsModerator() bool {
	return u.HasAnyRole(RoleModerator, RoleAdmin, RoleSuperAdmin)
}

// IsPremium vérifie si l'utilisateur a un accès premium
func (u *User) IsPremium() bool {
	return u.HasAnyRole(RolePremium, RoleModerator, RoleAdmin, RoleSuperAdmin)
}

// IncrementLoginAttempts incrémente le nombre de tentatives d'authentification
func (u *User) IncrementLoginAttempts() {
	u.LoginAttempts++
	u.UpdatedAt = time.Now()

	// Verrouiller le compte après 5 tentatives
	if u.LoginAttempts >= 5 {
		lockDuration := time.Duration(u.LoginAttempts-4) * 15 * time.Minute
		lockUntil := time.Now().Add(lockDuration)
		u.LockedUntil = &lockUntil
	}
}

// ResetLoginAttempts remet à zéro les tentatives d'authentification
func (u *User) ResetLoginAttempts() {
	u.LoginAttempts = 0
	u.LockedUntil = nil
	u.UpdatedAt = time.Now()
}

// SetOnline définit le statut en ligne de l'utilisateur
func (u *User) SetOnline() {
	u.IsOnline = true
	u.LastSeen = time.Now()
	u.UpdatedAt = time.Now()
}

// SetOffline définit le statut hors ligne de l'utilisateur
func (u *User) SetOffline() {
	u.IsOnline = false
	u.LastSeen = time.Now()
	u.UpdatedAt = time.Now()
}

// GetDisplayName retourne le nom d'affichage préféré
func (u *User) GetDisplayName() string {
	if u.DisplayName != "" {
		return u.DisplayName
	}
	if u.FirstName != "" && u.LastName != "" {
		return u.FirstName + " " + u.LastName
	}
	if u.FirstName != "" {
		return u.FirstName
	}
	return u.Username
}

// GetFullName retourne le nom complet
func (u *User) GetFullName() string {
	if u.FirstName != "" && u.LastName != "" {
		return u.FirstName + " " + u.LastName
	}
	return u.GetDisplayName()
}

// ToPublic retourne une version publique de l'utilisateur (sans données sensibles)
func (u *User) ToPublic() *User {
	public := *u
	public.PasswordHash = ""
	public.Salt = ""
	public.EmailVerificationToken = ""
	public.TwoFactorSecret = ""
	public.LastLoginIP = ""
	return &public
}

// ValidatePassword valide la complexité du mot de passe
func ValidatePassword(password string) error {
	if len(password) < 8 {
		return errors.New("password must be at least 8 characters long")
	}

	if len(password) > 128 {
		return errors.New("password must be less than 128 characters long")
	}

	var (
		hasUpper   = false
		hasLower   = false
		hasNumber  = false
		hasSpecial = false
	)

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsNumber(char):
			hasNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	if !hasUpper {
		return errors.New("password must contain at least one uppercase letter")
	}

	if !hasLower {
		return errors.New("password must contain at least one lowercase letter")
	}

	if !hasNumber {
		return errors.New("password must contain at least one number")
	}

	if !hasSpecial {
		return errors.New("password must contain at least one special character")
	}

	return nil
}

// isValidEmail vérifie le format basique d'un email
func isValidEmail(email string) bool {
	// Validation basique pour l'exemple - en production, utiliser une regex plus complète
	return len(email) > 3 &&
		len(email) < 254 &&
		unicode.IsLetter(rune(email[0])) &&
		containsAtAndDot(email)
}

// containsAtAndDot vérifie la présence d'@ et d'un point
func containsAtAndDot(email string) bool {
	hasAt := false
	hasDot := false

	for _, char := range email {
		if char == '@' {
			hasAt = true
		}
		if char == '.' && hasAt {
			hasDot = true
		}
	}

	return hasAt && hasDot
}

// ============================================================================
// MÉTHODES MANQUANTES POUR COMPATIBILITÉ
// ============================================================================

// ValidatePassword vérifie si le mot de passe fourni correspond au hash
func (u *User) ValidatePassword(password string) bool {
	return u.CheckPassword(password)
}

// HashPassword hash le mot de passe et le définit
func (u *User) HashPassword(password string) error {
	return u.SetPassword(password)
}

// EnableTwoFactor active l'authentification à deux facteurs
func (u *User) EnableTwoFactor() {
	u.TwoFactorEnabled = true
	u.UpdatedAt = time.Now()
}

// DisableTwoFactor désactive l'authentification à deux facteurs
func (u *User) DisableTwoFactor() {
	u.TwoFactorEnabled = false
	u.TwoFactorSecret = ""
	u.UpdatedAt = time.Now()
}

// UpdateUser met à jour les informations de l'utilisateur
func (u *User) UpdateUser() {
	u.UpdatedAt = time.Now()
}
