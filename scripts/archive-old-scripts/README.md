# 🚀 Scripts Talas - Infrastructure Incus Simplifiée

## 📋 Vue d'ensemble

Le dossier `scripts` a été entièrement nettoyé et simplifié pour ne contenir que les scripts essentiels à l'administration de l'infrastructure de test avec containers Incus.

## 🗂️ Structure Actuelle

```
scripts/
├── talas-admin.sh            # 🎯 Administration services locaux
├── talas-incus.sh            # 🏗️ ADMINISTRATION INFRASTRUCTURE INCUS
├── talas-cleanup.sh          # 🧹 Nettoyage de l'environnement
├── deploy-base-containers.sh # 📦 Déploiement containers Incus
├── update-source-code.sh     # 📝 Mise à jour code source
├── network-fix.sh            # 🌐 Réparation réseau
├── README.md                 # 📖 Ce fichier
└── archive-old-scripts/      # 📁 Scripts archivés (référence)
```

## 🎯 Scripts Principaux

### `talas-admin.sh` - Services Locaux
Gestion des services en développement local (sans containers).

```bash
./scripts/talas-admin.sh setup    # Configuration JWT
./scripts/talas-admin.sh build    # Compilation locale
./scripts/talas-admin.sh start    # Services locaux
```

### `talas-incus.sh` - Infrastructure Containers ⭐ **NOUVEAU**
**LE SCRIPT COMPLET** pour administrer l'infrastructure Incus avec toutes les fonctionnalités demandées.

## 🏗️ Administration Infrastructure Incus Complète

### **✅ DÉMARRER 8 CONTAINERS INCUS**
```bash
./scripts/talas-incus.sh deploy
# Crée et configure automatiquement :
# - veza-postgres (PostgreSQL)
# - veza-redis (Redis)  
# - veza-storage (NFS)
# - veza-backend (API Go)
# - veza-chat (WebSocket Rust)
# - veza-stream (Audio Rust)
# - veza-frontend (React)
# - veza-haproxy (Load Balancer)
```

### **✅ INSTALLER DÉPENDANCES & SERVICES SYSTEMD**
```bash
./scripts/talas-incus.sh setup
# Configure automatiquement :
# - Services systemd pour chaque container
# - Variables d'environnement
# - Configuration auto-restart
# - Dépendances inter-services
```

### **✅ RSYNC LE CODE FACILEMENT**
```bash
./scripts/talas-incus.sh update
# Workflow automatique :
# 1. Rsync de tout le code source
# 2. Compilation dans les containers
# 3. Redémarrage des services
# 4. Vérification de santé
```

### **✅ COMPILER LE CODE**
```bash
./scripts/talas-incus.sh compile
# Compile automatiquement :
# - Backend Go (bin/server)
# - Chat Server Rust (cargo build --release)
# - Stream Server Rust (cargo build --release)
# - Frontend React (npm install)
```

### **✅ REDÉMARRER LES SERVICES**
```bash
# Redémarrer tous les services
./scripts/talas-incus.sh restart

# Redémarrer un service spécifique
./scripts/talas-incus.sh restart backend
./scripts/talas-incus.sh restart chat
./scripts/talas-incus.sh restart frontend
```

### **✅ VOIR LES LOGS**
```bash
# Logs en temps réel d'un service
./scripts/talas-incus.sh logs backend
./scripts/talas-incus.sh logs chat
./scripts/talas-incus.sh logs frontend

# Logs avec journalctl dans le container
```

### **✅ SUPPRIMER TOUTE L'INFRASTRUCTURE**
```bash
./scripts/talas-incus.sh clean
# Supprime complètement :
# - Tous les 8 containers
# - Données et configurations
# - Avec confirmation de sécurité
```

### **✅ DEBUGGER**
```bash
# Vérification de santé complète
./scripts/talas-incus.sh health

# Mode debug avancé
./scripts/talas-incus.sh debug

# État détaillé de l'infrastructure
./scripts/talas-incus.sh status

# Réparation réseau automatique
./scripts/talas-incus.sh network-fix
```

## 🚀 Workflow Complet Infrastructure Incus

### 1. Déploiement Initial
```bash
# Déploiement complet (8 containers + dépendances)
./scripts/talas-incus.sh deploy

# Configuration des services systemd
./scripts/talas-incus.sh setup

# Première compilation
./scripts/talas-incus.sh compile

# Démarrage des services
./scripts/talas-incus.sh start
```

### 2. Développement Quotidien
```bash
# Après chaque modification de code
./scripts/talas-incus.sh update

# Vérifier l'état
./scripts/talas-incus.sh status

# Voir les logs
./scripts/talas-incus.sh logs backend
```

### 3. Debugging
```bash
# Problème réseau
./scripts/talas-incus.sh network-fix

# Diagnostic complet
./scripts/talas-incus.sh health

# Debug avancé
./scripts/talas-incus.sh debug
```

### 4. Export/Import (Optimisation)
```bash
# Sauvegarder les containers configurés
./scripts/talas-incus.sh export

# Restaurer rapidement
./scripts/talas-incus.sh import
```

## 🛠️ Scripts Spécialisés

### `talas-cleanup.sh`
Nettoyage complet de l'environnement de développement.
```bash
./scripts/talas-cleanup.sh
```

### `deploy-base-containers.sh`
Déploiement de l'infrastructure complète avec containers Incus (appelé par talas-incus.sh).
```bash
./scripts/deploy-base-containers.sh
```

### `update-source-code.sh`
Mise à jour du code source dans les containers (appelé par talas-incus.sh).
```bash
./scripts/update-source-code.sh rsync backend
./scripts/update-source-code.sh all
```

### `network-fix.sh`
Résolution des problèmes de réseau et DNS.
```bash
./scripts/network-fix.sh
```

## 📊 Comparaison des Scripts

| Fonctionnalité | talas-admin.sh | **talas-incus.sh** |
|----------------|----------------|-------------------|
| **Déploiement** | ❌ Local seulement | ✅ 8 containers complets |
| **Services systemd** | ❌ Manuel | ✅ Configuration automatique |
| **Rsync automatique** | ❌ Non | ✅ Avec compilation |
| **Compilation** | ✅ Locale | ✅ Dans containers |
| **Logs** | ✅ Locaux | ✅ Containers + journalctl |
| **Debug** | ❌ Basique | ✅ Mode debug avancé |
| **Infrastructure** | ❌ Non | ✅ Gestion complète |
| **Export/Import** | ❌ Non | ✅ Containers configurés |

## 💡 Avantages de la Nouvelle Structure

✅ **Infrastructure Complète** - 8 containers avec toutes les dépendances  
✅ **Workflow Automatisé** - Rsync + compilation + restart en une commande  
✅ **Services systemd** - Configuration et gestion automatique  
✅ **Monitoring Avancé** - Logs, health checks, debug mode  
✅ **Maintenance Simple** - Export/import, nettoyage complet  
✅ **Development Ready** - Modification de code → test immédiat  

## 🆘 Support

```bash
# Aide complète infrastructure Incus
./scripts/talas-incus.sh help

# Aide services locaux
./scripts/talas-admin.sh help

# État de l'infrastructure
./scripts/talas-incus.sh status

# Tests de santé
./scripts/talas-incus.sh health
```

---

**🎯 Infrastructure Talas complète avec administration automatisée pour containers Incus !** 