package jwt

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
)

// SharedJWTService service JWT partagé entre tous les services
type SharedJWTService struct {
	secret        []byte
	expiration    time.Duration
	refreshExpiry time.Duration
	issuer        string
	logger        *zap.Logger
}

// ServiceClaims représente les claims JWT pour validation inter-services
type ServiceClaims struct {
	UserID      int64    `json:"user_id"`
	Username    string   `json:"username"`
	Email       string   `json:"email"`
	Role        string   `json:"role"`
	IsActive    bool     `json:"is_active"`
	IsVerified  bool     `json:"is_verified"`
	Services    []string `json:"services"`    // Services autorisés
	Permissions []string `json:"permissions"` // Permissions spécifiques
	jwt.RegisteredClaims
}

// NewSharedJWTService crée un nouveau service JWT partagé
func NewSharedJWTService(secret string, expiration time.Duration) *SharedJWTService {
	logger, _ := zap.NewProduction()

	return &SharedJWTService{
		secret:        []byte(secret),
		expiration:    expiration,
		refreshExpiry: 7 * 24 * time.Hour, // 7 jours
		issuer:        "veza-backend",
		logger:        logger,
	}
}

// GenerateServiceToken génère un token JWT pour un service spécifique
func (s *SharedJWTService) GenerateServiceToken(userID int64, username, email, role string, services []string, permissions []string) (string, error) {
	now := time.Now()
	claims := ServiceClaims{
		UserID:      userID,
		Username:    username,
		Email:       email,
		Role:        role,
		IsActive:    true,
		IsVerified:  true,
		Services:    services,
		Permissions: permissions,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    s.issuer,
			Subject:   fmt.Sprintf("user:%d", userID),
			Audience:  []string{"veza-chat", "veza-stream", "veza-backend"},
			ExpiresAt: jwt.NewNumericDate(now.Add(s.expiration)),
			NotBefore: jwt.NewNumericDate(now),
			IssuedAt:  jwt.NewNumericDate(now),
			ID:        fmt.Sprintf("%d_%d", userID, now.Unix()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(s.secret)
	if err != nil {
		s.logger.Error("Failed to sign JWT token", zap.Error(err))
		return "", err
	}

	return tokenString, nil
}

// ValidateForService valide un token JWT pour un service spécifique
func (s *SharedJWTService) ValidateForService(tokenString, service string) (*ServiceClaims, error) {
	// Parse du token
	token, err := jwt.ParseWithClaims(tokenString, &ServiceClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.secret, nil
	})

	if err != nil {
		s.logger.Warn("JWT validation failed", zap.Error(err), zap.String("service", service))
		return nil, err
	}

	claims, ok := token.Claims.(*ServiceClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token claims")
	}

	// Vérifier l'expiration
	if claims.ExpiresAt != nil && claims.ExpiresAt.Time.Before(time.Now()) {
		return nil, errors.New("token expired")
	}

	// Vérifier l'issuer
	if claims.Issuer != s.issuer {
		return nil, errors.New("invalid token issuer")
	}

	// Vérifier les permissions pour le service
	if !s.hasServiceAccess(claims, service) {
		return nil, fmt.Errorf("insufficient permissions for service: %s", service)
	}

	return claims, nil
}

// hasServiceAccess vérifie si l'utilisateur a accès au service
func (s *SharedJWTService) hasServiceAccess(claims *ServiceClaims, service string) bool {
	// Si pas de services spécifiés, accès global autorisé
	if len(claims.Services) == 0 {
		return true
	}

	// Vérifier si le service est dans la liste autorisée
	for _, allowedService := range claims.Services {
		if allowedService == service || allowedService == "*" {
			return true
		}
	}

	return false
}

// GetUserInfo extrait les informations utilisateur du token
func (s *SharedJWTService) GetUserInfo(tokenString string) (*UserInfo, error) {
	claims, err := s.ValidateForService(tokenString, "*")
	if err != nil {
		return nil, err
	}

	return &UserInfo{
		ID:          claims.UserID,
		Username:    claims.Username,
		Email:       claims.Email,
		Role:        claims.Role,
		IsActive:    claims.IsActive,
		IsVerified:  claims.IsVerified,
		Services:    claims.Services,
		Permissions: claims.Permissions,
		IssuedAt:    claims.IssuedAt.Time,
		ExpiresAt:   claims.ExpiresAt.Time,
	}, nil
}

// RefreshToken génère un nouveau token à partir d'un token valide
func (s *SharedJWTService) RefreshToken(tokenString string) (string, error) {
	claims, err := s.ValidateForService(tokenString, "*")
	if err != nil {
		return "", err
	}

	// Générer un nouveau token avec les mêmes claims
	return s.GenerateServiceToken(
		claims.UserID,
		claims.Username,
		claims.Email,
		claims.Role,
		claims.Services,
		claims.Permissions,
	)
}

// RevokeToken ajoute un token à la blacklist (implémentation future avec Redis)
func (s *SharedJWTService) RevokeToken(tokenString, reason string) error {
	// TODO: Implémenter la blacklist Redis
	s.logger.Info("Token revoked",
		zap.String("token", tokenString[:10]+"..."),
		zap.String("reason", reason),
	)
	return nil
}

// ValidatePermission vérifie si l'utilisateur a une permission spécifique
func (s *SharedJWTService) ValidatePermission(claims *ServiceClaims, resource, action string) bool {
	// Administrateurs ont tous les droits
	if claims.Role == "admin" || claims.Role == "super_admin" {
		return true
	}

	// Vérifier les permissions spécifiques
	requiredPermission := fmt.Sprintf("%s:%s", resource, action)
	for _, permission := range claims.Permissions {
		if permission == requiredPermission || permission == resource+":*" || permission == "*" {
			return true
		}
	}

	return false
}

// UserInfo représente les informations utilisateur extraites du JWT
type UserInfo struct {
	ID          int64     `json:"id"`
	Username    string    `json:"username"`
	Email       string    `json:"email"`
	Role        string    `json:"role"`
	IsActive    bool      `json:"is_active"`
	IsVerified  bool      `json:"is_verified"`
	Services    []string  `json:"services"`
	Permissions []string  `json:"permissions"`
	IssuedAt    time.Time `json:"issued_at"`
	ExpiresAt   time.Time `json:"expires_at"`
}

// GetDefaultPermissions retourne les permissions par défaut selon le rôle
func GetDefaultPermissions(role string) []string {
	switch role {
	case "super_admin":
		return []string{"*"}
	case "admin":
		return []string{
			"chat:*", "stream:*", "user:*", "room:*",
			"moderation:*", "analytics:read",
		}
	case "moderator":
		return []string{
			"chat:read", "chat:write", "chat:moderate",
			"stream:read", "user:read", "room:read", "room:moderate",
		}
	case "premium":
		return []string{
			"chat:read", "chat:write", "stream:read", "stream:write",
			"room:read", "room:write", "room:create_premium",
		}
	case "user":
		return []string{
			"chat:read", "chat:write", "stream:read",
			"room:read", "room:write", "room:create_public",
		}
	default:
		return []string{"chat:read", "stream:read", "room:read"}
	}
}

// GetDefaultServices retourne les services par défaut selon le rôle
func GetDefaultServices(role string) []string {
	switch role {
	case "super_admin", "admin":
		return []string{"*"}
	default:
		return []string{"chat", "stream", "api"}
	}
}
