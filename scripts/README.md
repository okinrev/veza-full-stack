# 🚀 Scripts Veza - Système Unifié et Optimisé

## 📋 Vue d'ensemble

Le dossier `scripts` a été entièrement restructuré et nettoyé pour offrir une expérience d'administration simplifiée et unifiée. Toutes les fonctionnalités sont maintenant accessibles via le script principal.

## 🗂️ Structure Actuelle

```
scripts/
├── veza-manager.sh           # 🎯 SCRIPT PRINCIPAL UNIFIÉ
├── network-fix.sh            # 🌐 Réparation réseau
├── deploy-base-containers.sh # 📦 Déploiement containers de base
├── update-source-code.sh     # 📝 Mise à jour code source
├── README-NEW-SYSTEM.md      # 📖 Documentation complète
├── SOLUTION-RESUME.md        # 📋 Résumé exécutif
└── archive-old-scripts/      # 📁 Scripts archivés
```

## 🎯 Script Principal : `veza-manager.sh`

**LE SEUL SCRIPT DONT VOUS AVEZ BESOIN** pour administrer votre infrastructure Veza.

### Commandes Principales

```bash
# Configuration et déploiement
./scripts/veza-manager.sh setup    # Configuration initiale
./scripts/veza-manager.sh deploy   # Déploiement complet
./scripts/veza-manager.sh import   # Import containers de base

# Monitoring et état
./scripts/veza-manager.sh status   # État complet infrastructure
./scripts/veza-manager.sh health   # Vérification de santé

# Gestion des services
./scripts/veza-manager.sh start [service]    # Démarrer service(s)
./scripts/veza-manager.sh stop [service]     # Arrêter service(s)
./scripts/veza-manager.sh restart [service]  # Redémarrer service(s)
./scripts/veza-manager.sh logs <service>     # Afficher logs

# Maintenance
./scripts/veza-manager.sh clean        # Nettoyage complet
./scripts/veza-manager.sh network-fix  # Réparation réseau
./scripts/veza-manager.sh update       # Mise à jour code
./scripts/veza-manager.sh export      # Export containers
```

### Services Disponibles

- `postgres` - Base de données PostgreSQL
- `redis` - Cache Redis
- `storage` - Stockage NFS
- `backend` - API Backend Go
- `chat` - WebSocket Chat Rust
- `stream` - Streaming Audio Rust
- `frontend` - Interface React
- `haproxy` - Load Balancer

## 🛠️ Scripts Spécialisés (Optionnels)

### `network-fix.sh`
Résout les problèmes de connectivité réseau et DNS.
```bash
./scripts/network-fix.sh
```

### `deploy-base-containers.sh`
Crée les containers avec toutes les dépendances installées.
```bash
./scripts/deploy-base-containers.sh
```

### `update-source-code.sh`
Met à jour le code source avec plusieurs méthodes.
```bash
./scripts/update-source-code.sh rsync backend
./scripts/update-source-code.sh all
```

## 🚀 Workflow Recommandé

### Installation Initiale (Une fois)
```bash
# 1. Configuration système
./scripts/veza-manager.sh setup

# 2. Résolution problèmes réseau
./scripts/veza-manager.sh network-fix

# 3. Déploiement infrastructure
./scripts/veza-manager.sh deploy

# 4. Export pour réutilisation
./scripts/veza-manager.sh export
```

### Utilisation Quotidienne
```bash
# État global
./scripts/veza-manager.sh status

# Redémarrer un service
./scripts/veza-manager.sh restart backend

# Consulter logs
./scripts/veza-manager.sh logs chat

# Mise à jour code
./scripts/veza-manager.sh update

# Vérification santé
./scripts/veza-manager.sh health
```

### Déploiement Rapide (Après export)
```bash
# Import containers préconfigurés
./scripts/veza-manager.sh import

# Mise à jour code source
./scripts/veza-manager.sh update

# Démarrage
./scripts/veza-manager.sh start
```

## 📁 Scripts Archivés

Les anciens scripts ont été déplacés dans `archive-old-scripts/` :
- `incus-*.sh` - Anciens scripts de gestion
- `deploy-chat-only.sh` - Déploiement spécialisé  
- `fix-dns.sh` - Ancien script DNS
- Et autres...

**Toutes leurs fonctionnalités sont maintenant intégrées dans `veza-manager.sh`.**

## 💡 Avantages du Nouveau Système

✅ **Un seul point d'entrée** - Plus besoin de mémoriser plusieurs scripts  
✅ **Interface unifiée** - Commandes cohérentes et intuitives  
✅ **Gestion complète** - Services, logs, health checks, etc.  
✅ **Auto-découverte** - Détection automatique des scripts spécialisés  
✅ **Maintenance simplifiée** - Code centralisé et organisé  
✅ **Backward compatibility** - Scripts spécialisés toujours utilisables  

## 🆘 Aide et Support

```bash
# Aide complète
./scripts/veza-manager.sh help

# Documentation détaillée
cat scripts/README-NEW-SYSTEM.md

# Résumé de la solution
cat scripts/SOLUTION-RESUME.md
```

## 🔧 Migration depuis l'Ancien Système

Si vous utilisiez les anciens scripts :

| Ancien | Nouveau |
|--------|---------|
| `./incus-status.sh` | `./veza-manager.sh status` |
| `./incus-services.sh start backend` | `./veza-manager.sh start backend` |
| `./incus-logs.sh chat` | `./veza-manager.sh logs chat` |
| `./incus-clean.sh` | `./veza-manager.sh clean` |
| `./fix-dns.sh` | `./veza-manager.sh network-fix` |

---

**🎯 Résultat : Une infrastructure Veza plus simple, plus puissante et plus facile à administrer !** 