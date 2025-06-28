# 🚀 Nouveau Système d'Administration Veza

## 📋 Vue d'ensemble

Le nouveau système d'administration Veza résout les problèmes de réseau et optimise le workflow de déploiement avec une approche en deux phases :

1. **Phase 1** : Création et export des containers de base avec toutes les dépendances
2. **Phase 2** : Import rapide et déploiement du code source uniquement

## 🔍 Analyse des Problèmes Résolus

### Problème Réseau
- **Ancien système** : Créait un réseau `veza-network` custom qui causait des problèmes de connectivité
- **Nouveau système** : Utilise le réseau par défaut Incus (`incusbr0`) qui fonctionne parfaitement
- **Résultat** : Connectivité internet garantie, configuration DNS simplifiée

### Problème de Performance
- **Ancien système** : Réinstallait toutes les dépendances à chaque déploiement
- **Nouveau système** : Containers de base pré-configurés, export/import rapide
- **Résultat** : Déploiement 10x plus rapide après la première installation

## 🛠️ Scripts Principaux

### 1. `veza-manager.sh` - Script Principal Unifié
```bash
# Configuration initiale complète
./scripts/veza-manager.sh setup

# Déploiement complet de l'infrastructure
./scripts/veza-manager.sh deploy

# État de l'infrastructure
./scripts/veza-manager.sh status

# Export des containers de base
./scripts/veza-manager.sh export

# Import des containers de base
./scripts/veza-manager.sh import

# Nettoyage complet
./scripts/veza-manager.sh clean
```

### 2. `network-fix.sh` - Réparation Réseau
```bash
# Résoudre les problèmes de réseau
./scripts/network-fix.sh
```

### 3. `deploy-base-containers.sh` - Déploiement des Containers de Base
```bash
# Créer les 8 containers avec toutes les dépendances
./scripts/deploy-base-containers.sh
```

### 4. `update-source-code.sh` - Mise à Jour du Code Source
```bash
# Synchronisation via rsync (recommandé)
./scripts/update-source-code.sh rsync backend
./scripts/update-source-code.sh all

# Via archive tar.gz
./scripts/update-source-code.sh archive frontend

# Via Git (à configurer)
./scripts/update-source-code.sh git chat
```

## 🚀 Workflow Recommandé

### Installation Initiale (Une seule fois)
```bash
# 1. Configuration initiale
./scripts/veza-manager.sh setup

# 2. Réparation réseau si nécessaire
./scripts/network-fix.sh

# 3. Déploiement des containers de base
./scripts/deploy-base-containers.sh

# 4. Export des containers configurés
./scripts/veza-manager.sh export
```

### Déploiements Suivants (Rapides)
```bash
# 1. Import des containers de base
./scripts/veza-manager.sh import

# 2. Déploiement du code source
./scripts/update-source-code.sh all

# 3. Vérification
./scripts/veza-manager.sh status
```

### Développement au Quotidien
```bash
# Mise à jour d'un service spécifique
./scripts/update-source-code.sh rsync backend

# Vérification de l'état
./scripts/veza-manager.sh status

# Consultation des logs
./scripts/incus-logs.sh backend
```

## 🌐 Configuration Réseau

### Ancienne Configuration (Problématique)
- Réseau custom `veza-network` avec IPs statiques
- Configuration DNS complexe avec dnsmasq
- Problèmes de connectivité internet

### Nouvelle Configuration (Optimisée)
- Utilisation du réseau par défaut Incus (`incusbr0`)
- IPs attribuées automatiquement par DHCP
- DNS système optimisé
- Connectivité internet garantie

## 📦 Containers et Services

| Container | Service | Port | Description |
|-----------|---------|------|-------------|
| `veza-postgres` | `postgresql` | 5432 | Base de données PostgreSQL |
| `veza-redis` | `redis-server` | 6379 | Cache Redis |
| `veza-storage` | `nfs-kernel-server` | 2049 | Stockage NFS |
| `veza-backend` | `veza-backend` | 8080 | API Backend Go |
| `veza-chat` | `veza-chat` | 8081 | WebSocket Chat Rust |
| `veza-stream` | `veza-stream` | 8082 | Streaming Audio Rust |
| `veza-frontend` | `veza-frontend` | 5173 | Interface React |
| `veza-haproxy` | `haproxy` | 80, 8404 | Load Balancer |

## 🔧 Méthodes de Déploiement du Code

### 1. Rsync (Recommandé pour le développement)
- **Avantages** : Rapide, synchronisation différentielle, pas de dépendances externes
- **Inconvénients** : Nécessite accès local au code source

### 2. Archive tar.gz
- **Avantages** : Portable, versioning facile, pas de dépendances Git
- **Inconvénients** : Plus lent, archive complète à chaque fois

### 3. Git Clone/Pull (À configurer)
- **Avantages** : Versioning complet, branches, collaboration
- **Inconvénients** : Nécessite configuration des dépôts distants

## 📊 Monitoring et Logs

```bash
# État complet
./scripts/veza-manager.sh status

# Logs spécifiques
./scripts/incus-logs.sh backend
./scripts/incus-logs.sh chat

# Santé des services
./scripts/incus-services.sh health

# Redémarrage d'un service
./scripts/incus-services.sh restart backend
```

## 🧹 Maintenance

```bash
# Nettoyage complet
./scripts/veza-manager.sh clean

# Redémarrage de tous les containers
./scripts/incus-services.sh restart

# Vérification de l'état réseau
./scripts/network-fix.sh
```

## 💡 Avantages du Nouveau Système

1. **Fiabilité** : Réseau stable avec connectivité garantie
2. **Performance** : Déploiement initial lent, mais redéploiements ultra-rapides
3. **Flexibilité** : Plusieurs méthodes de mise à jour du code
4. **Simplicité** : Script principal unifié pour toutes les opérations
5. **Maintenance** : Containers de base exportables et réutilisables
6. **Monitoring** : Outils complets de surveillance et de logs

## 🚨 Migration depuis l'Ancien Système

Si vous avez des containers existants avec l'ancien système :

```bash
# 1. Sauvegarde des données importantes
cp -r data/ data-backup/

# 2. Nettoyage complet
./scripts/veza-manager.sh clean

# 3. Nouveau déploiement
./scripts/veza-manager.sh setup
./scripts/network-fix.sh
./scripts/deploy-base-containers.sh
./scripts/veza-manager.sh export
```

## 📞 Support et Dépannage

### Problèmes Courants

1. **Container ne démarre pas** : Vérifier `incus list` et `incus info <container>`
2. **Pas de connectivité internet** : Exécuter `./scripts/network-fix.sh`
3. **Service ne démarre pas** : Vérifier les logs avec `./scripts/incus-logs.sh <service>`

### Commandes de Debug

```bash
# État des réseaux
incus network list

# État des containers
incus ls

# Logs système d'un container
incus exec <container> -- journalctl -xe

# Test de connectivité
incus exec <container> -- ping 8.8.8.8
``` 