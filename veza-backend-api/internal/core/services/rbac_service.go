package services

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/okinrev/veza-web-app/internal/core/domain/entities"
	"github.com/okinrev/veza-web-app/internal/core/domain/repositories"
)

// RBACService service de contrôle d'accès basé sur les rôles
type RBACService interface {
	// Permissions
	CheckPermission(ctx context.Context, userID int64, resource, action string) (bool, error)
	CheckPermissions(ctx context.Context, userID int64, permissions []Permission) (map[string]bool, error)

	// Rôles
	AssignRole(ctx context.Context, userID int64, role entities.UserRole) error
	RemoveRole(ctx context.Context, userID int64, role entities.UserRole) error
	GetUserRoles(ctx context.Context, userID int64) ([]entities.UserRole, error)

	// Permissions dynamiques
	GrantPermission(ctx context.Context, userID int64, resource, action string) error
	RevokePermission(ctx context.Context, userID int64, resource, action string) error

	// Validation
	CanAccessResource(ctx context.Context, userID int64, resourceType, resourceID string) (bool, error)
	IsOwner(ctx context.Context, userID int64, resourceType, resourceID string) (bool, error)

	// Administration
	IsAdmin(ctx context.Context, userID int64) (bool, error)
	IsModerator(ctx context.Context, userID int64) (bool, error)

	// Hiérarchie des rôles
	HasRole(ctx context.Context, userID int64, requiredRole entities.UserRole) (bool, error)
	GetRoleHierarchy() map[entities.UserRole]int
}

// Permission représente une permission spécifique
type Permission struct {
	Resource string `json:"resource"`
	Action   string `json:"action"`
}

// rbacService implémentation du service RBAC
type rbacService struct {
	userRepo repositories.UserRepository
	logger   *zap.Logger

	// Cache des permissions
	permissionCache map[string]map[string]bool

	// Hiérarchie des rôles (plus le nombre est élevé, plus le rôle a de pouvoirs)
	roleHierarchy map[entities.UserRole]int

	// Permissions par rôle
	rolePermissions map[entities.UserRole][]Permission
}

// NewRBACService crée une nouvelle instance du service RBAC
func NewRBACService(userRepo repositories.UserRepository, logger *zap.Logger) RBACService {
	service := &rbacService{
		userRepo:        userRepo,
		logger:          logger,
		permissionCache: make(map[string]map[string]bool),
	}

	// Initialiser la hiérarchie des rôles
	service.initRoleHierarchy()

	// Initialiser les permissions par rôle
	service.initRolePermissions()

	return service
}

// initRoleHierarchy initialise la hiérarchie des rôles
func (s *rbacService) initRoleHierarchy() {
	s.roleHierarchy = map[entities.UserRole]int{
		entities.RoleGuest:      0,
		entities.RoleUser:       10,
		entities.RolePremium:    20,
		entities.RoleModerator:  50,
		entities.RoleAdmin:      80,
		entities.RoleSuperAdmin: 100,
	}
}

// initRolePermissions initialise les permissions par rôle
func (s *rbacService) initRolePermissions() {
	s.rolePermissions = map[entities.UserRole][]Permission{
		entities.UserRoleGuest: {
			{Resource: "tracks", Action: "read"},
			{Resource: "public_rooms", Action: "read"},
		},

		entities.UserRoleUser: {
			// Hérite des permissions Guest
			{Resource: "tracks", Action: "read"},
			{Resource: "public_rooms", Action: "read"},

			// Permissions utilisateur
			{Resource: "profile", Action: "read"},
			{Resource: "profile", Action: "update"},
			{Resource: "chat_messages", Action: "create"},
			{Resource: "chat_messages", Action: "read"},
			{Resource: "rooms", Action: "join"},
			{Resource: "playlists", Action: "create"},
			{Resource: "playlists", Action: "read"},
			{Resource: "playlists", Action: "update"},
			{Resource: "playlists", Action: "delete"},
		},

		entities.UserRolePremium: {
			// Hérite des permissions User + permissions premium
			{Resource: "premium_rooms", Action: "read"},
			{Resource: "premium_rooms", Action: "join"},
			{Resource: "streams", Action: "create"},
			{Resource: "streams", Action: "hq_quality"},
			{Resource: "files", Action: "upload_large"},
			{Resource: "analytics", Action: "read"},
		},

		entities.UserRoleModerator: {
			// Hérite des permissions Premium + permissions modération
			{Resource: "chat_messages", Action: "moderate"},
			{Resource: "chat_messages", Action: "delete"},
			{Resource: "users", Action: "moderate"},
			{Resource: "users", Action: "timeout"},
			{Resource: "users", Action: "kick"},
			{Resource: "rooms", Action: "moderate"},
			{Resource: "reports", Action: "read"},
			{Resource: "reports", Action: "handle"},
		},

		entities.UserRoleAdmin: {
			// Hérite des permissions Moderator + permissions admin
			{Resource: "users", Action: "read"},
			{Resource: "users", Action: "update"},
			{Resource: "users", Action: "ban"},
			{Resource: "users", Action: "suspend"},
			{Resource: "rooms", Action: "create"},
			{Resource: "rooms", Action: "delete"},
			{Resource: "settings", Action: "read"},
			{Resource: "settings", Action: "update"},
			{Resource: "analytics", Action: "full_access"},
			{Resource: "audit_logs", Action: "read"},
		},

		entities.UserRoleSuperAdmin: {
			// Toutes les permissions
			{Resource: "*", Action: "*"},
		},
	}
}

// CheckPermission vérifie si un utilisateur a une permission spécifique
func (s *rbacService) CheckPermission(ctx context.Context, userID int64, resource, action string) (bool, error) {
	// Récupérer l'utilisateur
	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return false, fmt.Errorf("récupération utilisateur: %w", err)
	}

	if user == nil {
		return false, fmt.Errorf("utilisateur non trouvé")
	}

	// Vérifier si l'utilisateur est actif
	if !user.IsActive() {
		return false, nil
	}

	// SuperAdmin a toutes les permissions
	if user.Role == entities.UserRoleSuperAdmin {
		return true, nil
	}

	// Vérifier les permissions du rôle
	if s.hasRolePermission(user.Role, resource, action) {
		return true, nil
	}

	// Vérifier les permissions personnalisées
	customPermissions, err := s.userRepo.GetUserPermissions(ctx, userID)
	if err != nil {
		s.logger.Warn("Erreur récupération permissions personnalisées", zap.Error(err))
		return false, nil
	}

	for _, perm := range customPermissions {
		if s.matchPermission(perm.Resource, perm.Action, resource, action) {
			return true, nil
		}
	}

	return false, nil
}

// CheckPermissions vérifie plusieurs permissions en une fois
func (s *rbacService) CheckPermissions(ctx context.Context, userID int64, permissions []Permission) (map[string]bool, error) {
	results := make(map[string]bool)

	for _, perm := range permissions {
		key := fmt.Sprintf("%s:%s", perm.Resource, perm.Action)
		allowed, err := s.CheckPermission(ctx, userID, perm.Resource, perm.Action)
		if err != nil {
			return nil, err
		}
		results[key] = allowed
	}

	return results, nil
}

// AssignRole assigne un rôle à un utilisateur
func (s *rbacService) AssignRole(ctx context.Context, userID int64, role entities.UserRole) error {
	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("récupération utilisateur: %w", err)
	}

	if user == nil {
		return fmt.Errorf("utilisateur non trouvé")
	}

	// Mettre à jour le rôle
	user.Role = role
	user.UpdatedAt = time.Now()

	if err := s.userRepo.UpdateUser(ctx, user); err != nil {
		return fmt.Errorf("mise à jour rôle: %w", err)
	}

	// Log d'audit
	if err := s.auditRoleChange(ctx, userID, role, "role_assigned"); err != nil {
		s.logger.Error("Erreur audit role change", zap.Error(err))
	}

	return nil
}

// RemoveRole retire un rôle d'un utilisateur (le remet en UserRoleUser)
func (s *rbacService) RemoveRole(ctx context.Context, userID int64, role entities.UserRole) error {
	return s.AssignRole(ctx, userID, entities.UserRoleUser)
}

// GetUserRoles récupère les rôles d'un utilisateur
func (s *rbacService) GetUserRoles(ctx context.Context, userID int64) ([]entities.UserRole, error) {
	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("récupération utilisateur: %w", err)
	}

	if user == nil {
		return nil, fmt.Errorf("utilisateur non trouvé")
	}

	// Pour l'instant, un utilisateur a un seul rôle
	// Cette méthode peut être étendue pour supporter plusieurs rôles
	return []entities.UserRole{user.Role}, nil
}

// GrantPermission accorde une permission personnalisée
func (s *rbacService) GrantPermission(ctx context.Context, userID int64, resource, action string) error {
	permission := repositories.UserPermission{
		UserID:    userID,
		Resource:  resource,
		Action:    action,
		GrantedAt: time.Now(),
		GrantedBy: userID, // Auto-grant pour l'instant
	}

	if err := s.userRepo.GrantUserPermission(ctx, userID, permission); err != nil {
		return fmt.Errorf("octroi permission: %w", err)
	}

	// Log d'audit
	if err := s.auditPermissionChange(ctx, userID, resource, action, "permission_granted"); err != nil {
		s.logger.Error("Erreur audit permission change", zap.Error(err))
	}

	return nil
}

// RevokePermission révoque une permission personnalisée
func (s *rbacService) RevokePermission(ctx context.Context, userID int64, resource, action string) error {
	permission := repositories.UserPermission{
		UserID:   userID,
		Resource: resource,
		Action:   action,
	}

	if err := s.userRepo.RevokeUserPermission(ctx, userID, permission); err != nil {
		return fmt.Errorf("révocation permission: %w", err)
	}

	// Log d'audit
	if err := s.auditPermissionChange(ctx, userID, resource, action, "permission_revoked"); err != nil {
		s.logger.Error("Erreur audit permission change", zap.Error(err))
	}

	return nil
}

// CanAccessResource vérifie si un utilisateur peut accéder à une ressource
func (s *rbacService) CanAccessResource(ctx context.Context, userID int64, resourceType, resourceID string) (bool, error) {
	// Vérifier les permissions de base
	canRead, err := s.CheckPermission(ctx, userID, resourceType, "read")
	if err != nil {
		return false, err
	}

	if !canRead {
		return false, nil
	}

	// Vérifications spécifiques selon le type de ressource
	switch resourceType {
	case "private_rooms":
		// Vérifier si l'utilisateur est membre de la room
		// TODO: Convertir resourceID en int64 ou utiliser une autre méthode
		// isMember, err := s.userRepo.IsRoomMember(ctx, userID, resourceID)
		// if err != nil {
		// 	return false, err
		// }
		// return isMember, nil
		return true, nil // Temporaire

	case "premium_content":
		// Vérifier si l'utilisateur a un rôle premium ou supérieur
		return s.HasRole(ctx, userID, entities.UserRolePremium)

	default:
		return true, nil
	}
}

// IsOwner vérifie si un utilisateur est propriétaire d'une ressource
func (s *rbacService) IsOwner(ctx context.Context, userID int64, resourceType, resourceID string) (bool, error) {
	switch resourceType {
	case "playlist":
		// TODO: Implémenter GetPlaylistByID
		// playlist, err := s.userRepo.GetPlaylistByID(ctx, resourceID)
		// if err != nil {
		// 	return false, err
		// }
		// return playlist != nil && playlist.OwnerID == userID, nil
		return false, nil // Temporaire

	case "stream":
		// TODO: Implémenter GetStreamByID
		// stream, err := s.userRepo.GetStreamByID(ctx, resourceID)
		// if err != nil {
		// 	return false, err
		// }
		// return stream != nil && stream.OwnerID == userID, nil
		return false, nil // Temporaire

	default:
		return false, nil
	}
}

// IsAdmin vérifie si un utilisateur est administrateur
func (s *rbacService) IsAdmin(ctx context.Context, userID int64) (bool, error) {
	return s.HasRole(ctx, userID, entities.UserRoleAdmin)
}

// IsModerator vérifie si un utilisateur est modérateur ou plus
func (s *rbacService) IsModerator(ctx context.Context, userID int64) (bool, error) {
	return s.HasRole(ctx, userID, entities.UserRoleModerator)
}

// HasRole vérifie si un utilisateur a un rôle spécifique ou supérieur
func (s *rbacService) HasRole(ctx context.Context, userID int64, requiredRole entities.UserRole) (bool, error) {
	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return false, fmt.Errorf("récupération utilisateur: %w", err)
	}

	if user == nil {
		return false, fmt.Errorf("utilisateur non trouvé")
	}

	// Vérifier si l'utilisateur est actif
	if !user.IsActive() {
		return false, nil
	}

	// Comparer les niveaux de rôle
	userLevel := s.roleHierarchy[user.Role]
	requiredLevel := s.roleHierarchy[requiredRole]

	return userLevel >= requiredLevel, nil
}

// GetRoleHierarchy retourne la hiérarchie des rôles
func (s *rbacService) GetRoleHierarchy() map[entities.UserRole]int {
	return s.roleHierarchy
}

// ============================================================================
// HELPER FUNCTIONS
// ============================================================================

// hasRolePermission vérifie si un rôle a une permission spécifique
func (s *rbacService) hasRolePermission(role entities.UserRole, resource, action string) bool {
	permissions, exists := s.rolePermissions[role]
	if !exists {
		return false
	}

	for _, perm := range permissions {
		if s.matchPermission(perm.Resource, perm.Action, resource, action) {
			return true
		}
	}

	// Vérifier les rôles inférieurs dans la hiérarchie
	currentLevel := s.roleHierarchy[role]
	for lowerRole, level := range s.roleHierarchy {
		if level < currentLevel {
			if s.hasRolePermission(lowerRole, resource, action) {
				return true
			}
		}
	}

	return false
}

// matchPermission vérifie si deux permissions correspondent (avec support wildcards)
func (s *rbacService) matchPermission(permResource, permAction, resource, action string) bool {
	// Wildcard complet
	if permResource == "*" && permAction == "*" {
		return true
	}

	// Wildcard resource
	if permResource == "*" && permAction == action {
		return true
	}

	// Wildcard action
	if permResource == resource && permAction == "*" {
		return true
	}

	// Match exact
	if permResource == resource && permAction == action {
		return true
	}

	// Pattern matching pour ressources hiérarchiques
	if strings.Contains(permResource, "*") {
		pattern := strings.ReplaceAll(permResource, "*", ".*")
		matched, _ := regexp.MatchString("^"+pattern+"$", resource)
		if matched && (permAction == action || permAction == "*") {
			return true
		}
	}

	return false
}

// auditRoleChange enregistre un changement de rôle
func (s *rbacService) auditRoleChange(ctx context.Context, userID int64, newRole entities.UserRole, action string) error {
	log := &repositories.UserAuditLog{
		UserID:   userID,
		Action:   action,
		Resource: "user_role",
		Details:  fmt.Sprintf("New role: %s", newRole),
		Success:  true,
	}

	return s.userRepo.CreateAuditLog(ctx, log)
}

// auditPermissionChange enregistre un changement de permission
func (s *rbacService) auditPermissionChange(ctx context.Context, userID int64, resource, action, auditAction string) error {
	log := &repositories.UserAuditLog{
		UserID:   userID,
		Action:   auditAction,
		Resource: "user_permission",
		Details:  fmt.Sprintf("Resource: %s, Action: %s", resource, action),
		Success:  true,
	}

	return s.userRepo.CreateAuditLog(ctx, log)
}
