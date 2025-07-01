use crate::error::{ChatError, Result};
use serde::{Deserialize, Serialize};
use std::collections::HashSet;

/// Rôles disponibles dans le système de chat
#[derive(Debug, Clone, PartialEq, Eq, Hash, Serialize, Deserialize)]
pub enum Role {
    /// Utilisateur standard
    User,
    /// Modérateur avec permissions étendues
    Moderator,
    /// Administrateur avec tous les pouvoirs
    Admin,
    /// Super administrateur
    SuperAdmin,
}

/// Permissions granulaires dans le système
#[derive(Debug, Clone, PartialEq, Eq, Hash, Serialize, Deserialize)]
pub enum Permission {
    // Messages
    SendMessage,
    EditMessage,
    DeleteMessage,
    PinMessage,
    
    // Modération
    ModerateMessages,
    BanUsers,
    KickUsers,
    MuteUsers,
    
    // Administration
    ManageRoles,
    ManageChannels,
    ManageServer,
    ViewAuditLog,
    
    // Avancé
    ManageWebhooks,
    BypassRateLimit,
}

impl Role {
    /// Retourne les permissions par défaut pour un rôle
    pub fn default_permissions(&self) -> HashSet<Permission> {
        match self {
            Role::User => [
                Permission::SendMessage,
                Permission::EditMessage,
            ].into_iter().collect(),
            
            Role::Moderator => [
                Permission::SendMessage,
                Permission::EditMessage,
                Permission::DeleteMessage,
                Permission::PinMessage,
                Permission::ModerateMessages,
                Permission::KickUsers,
                Permission::MuteUsers,
            ].into_iter().collect(),
            
            Role::Admin => [
                Permission::SendMessage,
                Permission::EditMessage,
                Permission::DeleteMessage,
                Permission::PinMessage,
                Permission::ModerateMessages,
                Permission::BanUsers,
                Permission::KickUsers,
                Permission::MuteUsers,
                Permission::ManageRoles,
                Permission::ManageChannels,
                Permission::ViewAuditLog,
            ].into_iter().collect(),
            
            Role::SuperAdmin => {
                // Toutes les permissions
                [
                    Permission::SendMessage,
                    Permission::EditMessage,
                    Permission::DeleteMessage,
                    Permission::PinMessage,
                    Permission::ModerateMessages,
                    Permission::BanUsers,
                    Permission::KickUsers,
                    Permission::MuteUsers,
                    Permission::ManageRoles,
                    Permission::ManageChannels,
                    Permission::ManageServer,
                    Permission::ViewAuditLog,
                    Permission::ManageWebhooks,
                    Permission::BypassRateLimit,
                ].into_iter().collect()
            }
        }
    }

    pub fn get_permissions(&self) -> Vec<Permission> {
        self.default_permissions().into_iter().collect()
    }

    pub fn has_permission(&self, permission: &Permission) -> bool {
        self.default_permissions().contains(permission)
    }

    pub fn from_string(role_str: &str) -> Result<Self> {
        match role_str.to_lowercase().as_str() {
            "admin" => Ok(Role::Admin),
            "moderator" | "mod" => Ok(Role::Moderator),
            "user" => Ok(Role::User),
            "superadmin" => Ok(Role::SuperAdmin),
            _ => Err(ChatError::configuration_error(&format!("Rôle invalide: {}", role_str))),
        }
    }
}

/// Structure représentant les permissions d'un utilisateur
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct UserPermissions {
    pub user_id: i64,
    pub roles: HashSet<Role>,
    pub custom_permissions: HashSet<Permission>,
}

impl UserPermissions {
    /// Crée une nouvelle instance avec des permissions utilisateur de base
    pub fn new_user(user_id: i64) -> Self {
        Self {
            user_id,
            roles: [Role::User].into_iter().collect(),
            custom_permissions: HashSet::new(),
        }
    }
    
    /// Vérifie si l'utilisateur possède une permission spécifique
    pub fn has_permission(&self, permission: &Permission) -> bool {
        // Vérifier les permissions custom
        if self.custom_permissions.contains(permission) {
            return true;
        }
        
        // Vérifier les permissions des rôles
        self.roles.iter().any(|role| {
            role.default_permissions().contains(permission)
        })
    }
    
    /// Ajoute un rôle à l'utilisateur
    pub fn add_role(&mut self, role: Role) {
        self.roles.insert(role);
    }
    
    /// Retire un rôle de l'utilisateur
    pub fn remove_role(&mut self, role: &Role) {
        self.roles.remove(role);
    }
    
    /// Ajoute une permission custom
    pub fn grant_permission(&mut self, permission: Permission) {
        self.custom_permissions.insert(permission);
    }
    
    /// Retire une permission custom
    pub fn revoke_permission(&mut self, permission: &Permission) {
        self.custom_permissions.remove(permission);
    }
}

/// Fonction utilitaire pour vérifier les permissions
pub fn check_permission(user_permissions: &UserPermissions, required_permission: &Permission) -> bool {
    user_permissions.has_permission(required_permission)
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_user_permissions() {
        let mut perms = UserPermissions::new_user(123);
        
        // Utilisateur de base peut envoyer des messages
        assert!(perms.has_permission(&Permission::SendMessage));
        
        // Mais ne peut pas bannir
        assert!(!perms.has_permission(&Permission::BanUsers));
        
        // Ajouter le rôle modérateur
        perms.add_role(Role::Moderator);
        assert!(perms.has_permission(&Permission::KickUsers));
        
        // Ajouter permission custom
        perms.grant_permission(Permission::ManageServer);
        assert!(perms.has_permission(&Permission::ManageServer));
    }
    
    #[test]
    fn test_role_permissions() {
        let admin_perms = Role::Admin.default_permissions();
        assert!(admin_perms.contains(&Permission::ManageRoles));
        assert!(admin_perms.contains(&Permission::BanUsers));
        
        let user_perms = Role::User.default_permissions();
        assert!(!user_perms.contains(&Permission::BanUsers));
        assert!(user_perms.contains(&Permission::SendMessage));
    }
} 