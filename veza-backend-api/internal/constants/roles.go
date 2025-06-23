//file: internal/constants/roles.go

package constants

type Role string
type Permission string

const (
	// Rôles
	RoleUser       Role = "user"
	RoleAdmin      Role = "admin"
	RoleSuperAdmin Role = "super_admin"
	RoleModerator  Role = "moderator"
)

const (
	// Permissions produits
	PermReadProducts   Permission = "products:read"
	PermWriteProducts  Permission = "products:write"
	PermDeleteProducts Permission = "products:delete"
	
	// Permissions catégories
	PermReadCategories   Permission = "categories:read"
	PermWriteCategories  Permission = "categories:write"
	PermDeleteCategories Permission = "categories:delete"
	
	// Permissions utilisateurs
	PermReadUsers   Permission = "users:read"
	PermWriteUsers  Permission = "users:write"
	PermDeleteUsers Permission = "users:delete"
	
	// Permissions analytics
	PermReadAnalytics Permission = "analytics:read"
	
	// Permissions fichiers
	PermUploadFiles Permission = "files:upload"
	PermDeleteFiles Permission = "files:delete"
)

// RolePermissions définit les permissions par rôle
var RolePermissions = map[Role][]Permission{
	RoleUser: {
		PermReadProducts,
	},
	RoleModerator: {
		PermReadProducts, PermWriteProducts,
		PermReadCategories,
		PermReadUsers,
		PermUploadFiles,
	},
	RoleAdmin: {
		PermReadProducts, PermWriteProducts, PermDeleteProducts,
		PermReadCategories, PermWriteCategories, PermDeleteCategories,
		PermReadUsers, PermWriteUsers,
		PermReadAnalytics,
		PermUploadFiles, PermDeleteFiles,
	},
	RoleSuperAdmin: {
		PermReadProducts, PermWriteProducts, PermDeleteProducts,
		PermReadCategories, PermWriteCategories, PermDeleteCategories,
		PermReadUsers, PermWriteUsers, PermDeleteUsers,
		PermReadAnalytics,
		PermUploadFiles, PermDeleteFiles,
	},
}