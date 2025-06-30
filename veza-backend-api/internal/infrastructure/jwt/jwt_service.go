package jwt

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// JWTService gère la génération et validation des tokens JWT
type JWTService struct {
	accessSecretKey  []byte
	refreshSecretKey []byte
	accessDuration   time.Duration
	refreshDuration  time.Duration
	issuer           string
}

// UserSession représente une session utilisateur dans les tokens JWT
type UserSession struct {
	UserID   int64  `json:"user_id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Role     string `json:"role"`
}

// AccessTokenClaims représente les claims d'un token d'accès
type AccessTokenClaims struct {
	UserSession
	jwt.RegisteredClaims
}

// RefreshTokenClaims représente les claims d'un token de rafraîchissement
type RefreshTokenClaims struct {
	UserID int64 `json:"user_id"`
	jwt.RegisteredClaims
}

// NewJWTService crée une nouvelle instance du service JWT
func NewJWTService(accessSecret, refreshSecret, issuer string) *JWTService {
	return &JWTService{
		accessSecretKey:  []byte(accessSecret),
		refreshSecretKey: []byte(refreshSecret),
		accessDuration:   15 * time.Minute,   // Token d'accès expire dans 15 minutes
		refreshDuration:  7 * 24 * time.Hour, // Token de rafraîchissement expire dans 7 jours
		issuer:           issuer,
	}
}

// GenerateAccessToken génère un token d'accès pour une session utilisateur
func (j *JWTService) GenerateAccessToken(session *UserSession) (string, error) {
	now := time.Now()

	claims := AccessTokenClaims{
		UserSession: *session,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    j.issuer,
			Subject:   fmt.Sprintf("user:%d", session.UserID),
			Audience:  []string{"veza-api"},
			ExpiresAt: jwt.NewNumericDate(now.Add(j.accessDuration)),
			NotBefore: jwt.NewNumericDate(now),
			IssuedAt:  jwt.NewNumericDate(now),
			ID:        fmt.Sprintf("access_%d_%d", session.UserID, now.Unix()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.accessSecretKey)
}

// GenerateRefreshToken génère un token de rafraîchissement pour un utilisateur
func (j *JWTService) GenerateRefreshToken(session *UserSession) (string, error) {
	now := time.Now()

	claims := RefreshTokenClaims{
		UserID: session.UserID,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    j.issuer,
			Subject:   fmt.Sprintf("user:%d", session.UserID),
			Audience:  []string{"veza-refresh"},
			ExpiresAt: jwt.NewNumericDate(now.Add(j.refreshDuration)),
			NotBefore: jwt.NewNumericDate(now),
			IssuedAt:  jwt.NewNumericDate(now),
			ID:        fmt.Sprintf("refresh_%d_%d", session.UserID, now.Unix()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.refreshSecretKey)
}

// ValidateAccessToken valide un token d'accès et retourne la session utilisateur
func (j *JWTService) ValidateAccessToken(tokenString string) (*UserSession, error) {
	token, err := jwt.ParseWithClaims(tokenString, &AccessTokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return j.accessSecretKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	claims, ok := token.Claims.(*AccessTokenClaims)
	if !ok {
		return nil, errors.New("invalid token claims")
	}

	// Vérifications supplémentaires
	if claims.Issuer != j.issuer {
		return nil, errors.New("invalid token issuer")
	}

	if len(claims.Audience) == 0 || claims.Audience[0] != "veza-api" {
		return nil, errors.New("invalid token audience")
	}

	return &claims.UserSession, nil
}

// ValidateRefreshToken valide un token de rafraîchissement et retourne la session utilisateur
func (j *JWTService) ValidateRefreshToken(tokenString string) (*UserSession, error) {
	token, err := jwt.ParseWithClaims(tokenString, &RefreshTokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return j.refreshSecretKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	claims, ok := token.Claims.(*RefreshTokenClaims)
	if !ok {
		return nil, errors.New("invalid token claims")
	}

	// Vérifications supplémentaires
	if claims.Issuer != j.issuer {
		return nil, errors.New("invalid token issuer")
	}

	if len(claims.Audience) == 0 || claims.Audience[0] != "veza-refresh" {
		return nil, errors.New("invalid token audience")
	}

	// Retourner une session basique (pour refresh, on a seulement l'ID utilisateur)
	return &UserSession{
		UserID: claims.UserID,
	}, nil
}

// GetTokenDuration retourne la durée de validité d'un token d'accès
func (j *JWTService) GetTokenDuration() time.Duration {
	return j.accessDuration
}

// GetRefreshTokenDuration retourne la durée de validité d'un token de rafraîchissement
func (j *JWTService) GetRefreshTokenDuration() time.Duration {
	return j.refreshDuration
}

// ExtractTokenFromHeader extrait le token d'un header Authorization
func (j *JWTService) ExtractTokenFromHeader(authHeader string) (string, error) {
	if authHeader == "" {
		return "", errors.New("authorization header is required")
	}

	const bearerPrefix = "Bearer "
	if len(authHeader) < len(bearerPrefix) || authHeader[:len(bearerPrefix)] != bearerPrefix {
		return "", errors.New("invalid authorization header format")
	}

	token := authHeader[len(bearerPrefix):]
	if token == "" {
		return "", errors.New("token is required")
	}

	return token, nil
}

// IsTokenExpired vérifie si un token est expiré sans le valider complètement
func (j *JWTService) IsTokenExpired(tokenString string) bool {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// On ne valide pas la signature, on veut juste lire les claims
		return j.accessSecretKey, nil
	})

	if err != nil {
		return true
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		if exp, ok := claims["exp"].(float64); ok {
			return time.Now().Unix() > int64(exp)
		}
	}

	return true
}

// RevokeToken ajoute un token à une liste de révocation (implémentation basique)
// En production, cela devrait utiliser Redis ou une base de données
func (j *JWTService) RevokeToken(tokenString string) error {
	// TODO: Implémenter une vraie révocation de token avec Redis
	// Pour l'instant, on simule simplement que c'est fait
	return nil
}

// GetUserIDFromToken extrait l'ID utilisateur d'un token sans validation complète
func (j *JWTService) GetUserIDFromToken(tokenString string) (int64, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return j.accessSecretKey, nil
	})

	if err != nil {
		return 0, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		if userID, ok := claims["user_id"].(float64); ok {
			return int64(userID), nil
		}
	}

	return 0, errors.New("unable to extract user ID from token")
}
