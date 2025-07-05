---
id: security-authorization
sidebar_label: Authorization
---

# Authorization - Veza Platform

Ce dossier documente la gestion des permissions et des rôles sur la plateforme.

## Index
- À compléter : ajouter la documentation sur les rôles, scopes, RBAC, etc.

## Navigation
- [Retour au schéma principal](../../diagrams/architecture-overview.md)

## 🚦 Autorisation et Permissions - Veza

## 📋 Vue d'ensemble

Ce guide détaille le système d'autorisation et de gestion des permissions de la plateforme Veza.

## 👥 Rôles et Permissions

### Rôles Système
```javascript
const ROLES = {
  ADMIN: 'admin',
  MODERATOR: 'moderator',
  USER: 'user',
  GUEST: 'guest'
};

const PERMISSIONS = {
  // Gestion des utilisateurs
  USER_CREATE: 'user:create',
  USER_READ: 'user:read',
  USER_UPDATE: 'user:update',
  USER_DELETE: 'user:delete',
  
  // Gestion du chat
  MESSAGE_SEND: 'message:send',
  MESSAGE_READ: 'message:read',
  MESSAGE_DELETE: 'message:delete',
  
  // Gestion des streams
  STREAM_CREATE: 'stream:create',
  STREAM_READ: 'stream:read',
  STREAM_UPDATE: 'stream:update',
  STREAM_DELETE: 'stream:delete'
};
```

### Mapping Rôles-Permissions
```javascript
const ROLE_PERMISSIONS = {
  [ROLES.ADMIN]: [
    PERMISSIONS.USER_CREATE,
    PERMISSIONS.USER_READ,
    PERMISSIONS.USER_UPDATE,
    PERMISSIONS.USER_DELETE,
    PERMISSIONS.MESSAGE_SEND,
    PERMISSIONS.MESSAGE_READ,
    PERMISSIONS.MESSAGE_DELETE,
    PERMISSIONS.STREAM_CREATE,
    PERMISSIONS.STREAM_READ,
    PERMISSIONS.STREAM_UPDATE,
    PERMISSIONS.STREAM_DELETE
  ],
  [ROLES.MODERATOR]: [
    PERMISSIONS.USER_READ,
    PERMISSIONS.MESSAGE_SEND,
    PERMISSIONS.MESSAGE_READ,
    PERMISSIONS.MESSAGE_DELETE,
    PERMISSIONS.STREAM_READ
  ],
  [ROLES.USER]: [
    PERMISSIONS.USER_READ,
    PERMISSIONS.MESSAGE_SEND,
    PERMISSIONS.MESSAGE_READ,
    PERMISSIONS.STREAM_READ
  ],
  [ROLES.GUEST]: [
    PERMISSIONS.MESSAGE_READ,
    PERMISSIONS.STREAM_READ
  ]
};
```

## 🔐 Middleware d'Autorisation

```javascript
// Middleware de vérification des permissions
function requirePermission(permission) {
  return (req, res, next) => {
    const user = req.user;
    
    if (!user) {
      return res.status(401).json({ error: 'Utilisateur non authentifié' });
    }
    
    const userPermissions = ROLE_PERMISSIONS[user.role] || [];
    
    if (!userPermissions.includes(permission)) {
      return res.status(403).json({ 
        error: 'Permissions insuffisantes',
        required: permission,
        user_permissions: userPermissions
      });
    }
    
    next();
  };
}

// Utilisation
app.get('/api/users', 
  requirePermission(PERMISSIONS.USER_READ),
  (req, res) => {
    // Logique de récupération des utilisateurs
  }
);
```

## 🎯 Vérification Granulaire

```javascript
// Vérification de propriété (ownership)
function requireOwnership(resourceType) {
  return async (req, res, next) => {
    const resourceId = req.params.id;
    const userId = req.user.id;
    
    const resource = await getResourceById(resourceType, resourceId);
    
    if (!resource) {
      return res.status(404).json({ error: 'Ressource introuvable' });
    }
    
    if (resource.user_id !== userId && req.user.role !== ROLES.ADMIN) {
      return res.status(403).json({ error: 'Accès refusé' });
    }
    
    req.resource = resource;
    next();
  };
}

// Utilisation
app.put('/api/messages/:id',
  requirePermission(PERMISSIONS.MESSAGE_UPDATE),
  requireOwnership('message'),
  (req, res) => {
    // Logique de mise à jour du message
  }
);
```

## 📚 Ressources

- [Guide de Sécurité](../README.md)
- [Authentification](../authentication/README.md)
- [Audit de Sécurité](../audit/README.md)

---

**Dernière mise à jour** : $(date)
**Version** : 1.0.0 