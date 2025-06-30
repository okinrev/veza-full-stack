package entities

import (
	"errors"
	"regexp"
	"strings"
	"time"
)

// User représente un utilisateur dans le domaine métier
type User struct {
	ID          int64      `json:"id"`
	Username    string     `json:"username"`
	Email       string     `json:"email"`
	Password    string     `json:"-"` // Jamais sérialisé
	FirstName   string     `json:"first_name,omitempty"`
	LastName    string     `json:"last_name,omitempty"`
	Bio         string     `json:"bio,omitempty"`
	Avatar      string     `json:"avatar,omitempty"`
	Role        UserRole   `json:"role"`
	Status      UserStatus `json:"status"`
	IsActive    bool       `json:"is_active"`
	IsVerified  bool       `json:"is_verified"`
	LastLoginAt *time.Time `json:"last_login_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// UserRole énumération des rôles utilisateur
type UserRole string

const (
	RoleUser       UserRole = "user"
	RolePremium    UserRole = "premium"
	RoleModerator  UserRole = "moderator"
	RoleAdmin      UserRole = "admin"
	RoleSuperAdmin UserRole = "super_admin"
)

// UserStatus énumération des statuts utilisateur
type UserStatus string

const (
	StatusActive    UserStatus = "active"
	StatusInactive  UserStatus = "inactive"
	StatusSuspended UserStatus = "suspended"
	StatusBanned    UserStatus = "banned"
)

// NewUser crée un nouvel utilisateur avec validation
func NewUser(username, email, password string) (*User, error) {
	user := &User{
		Username:  strings.TrimSpace(username),
		Email:     strings.TrimSpace(strings.ToLower(email)),
		Password:  password,
		Role:      RoleUser,
		Status:    StatusActive,
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := user.Validate(); err != nil {
		return nil, err
	}

	return user, nil
}

// Validate valide les données de l'utilisateur
func (u *User) Validate() error {
	if err := u.ValidateUsername(); err != nil {
		return err
	}
	if err := u.ValidateEmail(); err != nil {
		return err
	}
	if err := u.ValidatePassword(); err != nil {
		return err
	}
	return nil
}

// ValidateUsername valide le nom d'utilisateur
func (u *User) ValidateUsername() error {
	if u.Username == "" {
		return errors.New("nom d'utilisateur requis")
	}
	if len(u.Username) < 3 {
		return errors.New("nom d'utilisateur doit contenir au moins 3 caractères")
	}
	if len(u.Username) > 30 {
		return errors.New("nom d'utilisateur ne peut pas dépasser 30 caractères")
	}

	// Vérifier les caractères autorisés
	validUsername := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	if !validUsername.MatchString(u.Username) {
		return errors.New("nom d'utilisateur ne peut contenir que des lettres, chiffres, _ et -")
	}

	return nil
}

// ValidateEmail valide l'adresse email
func (u *User) ValidateEmail() error {
	if u.Email == "" {
		return errors.New("email requis")
	}

	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(u.Email) {
		return errors.New("format email invalide")
	}

	return nil
}

// ValidatePassword valide le mot de passe
func (u *User) ValidatePassword() error {
	if u.Password == "" {
		return errors.New("mot de passe requis")
	}
	if len(u.Password) < 8 {
		return errors.New("mot de passe doit contenir au moins 8 caractères")
	}
	if len(u.Password) > 128 {
		return errors.New("mot de passe ne peut pas dépasser 128 caractères")
	}

	// Vérifier la complexité
	hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(u.Password)
	hasLower := regexp.MustCompile(`[a-z]`).MatchString(u.Password)
	hasDigit := regexp.MustCompile(`[0-9]`).MatchString(u.Password)
	hasSpecial := regexp.MustCompile(`[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?]`).MatchString(u.Password)

	if !hasUpper || !hasLower || !hasDigit || !hasSpecial {
		return errors.New("mot de passe doit contenir au moins une majuscule, une minuscule, un chiffre et un caractère spécial")
	}

	return nil
}

// HasRole vérifie si l'utilisateur a le rôle spécifié ou supérieur
func (u *User) HasRole(role UserRole) bool {
	roleHierarchy := map[UserRole]int{
		RoleUser:       1,
		RolePremium:    2,
		RoleModerator:  3,
		RoleAdmin:      4,
		RoleSuperAdmin: 5,
	}

	userLevel := roleHierarchy[u.Role]
	requiredLevel := roleHierarchy[role]

	return userLevel >= requiredLevel
}

// IsAdmin vérifie si l'utilisateur est administrateur
func (u *User) IsAdmin() bool {
	return u.Role == RoleAdmin || u.Role == RoleSuperAdmin
}

// IsPremium vérifie si l'utilisateur est premium ou supérieur
func (u *User) IsPremium() bool {
	return u.HasRole(RolePremium)
}

// CanModerate vérifie si l'utilisateur peut modérer
func (u *User) CanModerate() bool {
	return u.HasRole(RoleModerator)
}

// IsAllowedToLogin vérifie si l'utilisateur peut se connecter
func (u *User) IsAllowedToLogin() bool {
	return u.IsActive && (u.Status == StatusActive || u.Status == StatusInactive)
}

// UpdateLastLogin met à jour la dernière connexion
func (u *User) UpdateLastLogin() {
	now := time.Now()
	u.LastLoginAt = &now
	u.UpdatedAt = now
}

// UpdateProfile met à jour le profil utilisateur
func (u *User) UpdateProfile(firstName, lastName, bio string) {
	u.FirstName = strings.TrimSpace(firstName)
	u.LastName = strings.TrimSpace(lastName)
	u.Bio = strings.TrimSpace(bio)
	u.UpdatedAt = time.Now()
}

// SetAvatar définit l'avatar de l'utilisateur
func (u *User) SetAvatar(avatarURL string) {
	u.Avatar = strings.TrimSpace(avatarURL)
	u.UpdatedAt = time.Now()
}

// Activate active l'utilisateur
func (u *User) Activate() {
	u.IsActive = true
	u.Status = StatusActive
	u.UpdatedAt = time.Now()
}

// Deactivate désactive l'utilisateur
func (u *User) Deactivate() {
	u.IsActive = false
	u.Status = StatusInactive
	u.UpdatedAt = time.Now()
}

// Suspend suspend l'utilisateur
func (u *User) Suspend() {
	u.Status = StatusSuspended
	u.UpdatedAt = time.Now()
}

// Ban bannit l'utilisateur
func (u *User) Ban() {
	u.IsActive = false
	u.Status = StatusBanned
	u.UpdatedAt = time.Now()
}

// Verify vérifie l'utilisateur
func (u *User) Verify() {
	u.IsVerified = true
	u.UpdatedAt = time.Now()
}

// ToPublic retourne une version publique de l'utilisateur (sans données sensibles)
func (u *User) ToPublic() *UserPublic {
	return &UserPublic{
		ID:         u.ID,
		Username:   u.Username,
		FirstName:  u.FirstName,
		LastName:   u.LastName,
		Bio:        u.Bio,
		Avatar:     u.Avatar,
		Role:       u.Role,
		IsVerified: u.IsVerified,
		CreatedAt:  u.CreatedAt,
	}
}

// UserPublic représente les données publiques d'un utilisateur
type UserPublic struct {
	ID         int64     `json:"id"`
	Username   string    `json:"username"`
	FirstName  string    `json:"first_name,omitempty"`
	LastName   string    `json:"last_name,omitempty"`
	Bio        string    `json:"bio,omitempty"`
	Avatar     string    `json:"avatar,omitempty"`
	Role       UserRole  `json:"role"`
	IsVerified bool      `json:"is_verified"`
	CreatedAt  time.Time `json:"created_at"`
}

// UserSession représente une session utilisateur
type UserSession struct {
	UserID    int64     `json:"user_id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Role      UserRole  `json:"role"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
}

// Alias pour compatibilité avec le code existant
type PublicUser = UserPublic

// ToSession convertit l'utilisateur en session pour JWT
func (u *User) ToSession() *UserSession {
	return &UserSession{
		UserID:    u.ID,
		Username:  u.Username,
		Email:     u.Email,
		Role:      u.Role,
		IsActive:  u.IsActive,
		CreatedAt: u.CreatedAt,
	}
}

// ToJWTSession convertit vers le format JWT session
func (u *User) ToJWTSession() *JWTSession {
	return &JWTSession{
		UserID:   u.ID,
		Username: u.Username,
		Email:    u.Email,
		Role:     string(u.Role), // Conversion en string pour JWT
	}
}

// JWTSession structure pour compatibilité JWT
type JWTSession struct {
	UserID   int64  `json:"user_id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Role     string `json:"role"`
}
