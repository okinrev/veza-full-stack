# 🚀 Guide de Déploiement Veza - Système Unifié

## 📋 Vue d'ensemble

Ce guide couvre le **nouveau système de déploiement unifié** de Veza, qui remplace tous les anciens scripts redondants par une solution simplifiée et cohérente.

### ✨ Améliorations apportées

- **Nettoyage complet** : 9 scripts redondants supprimés  
- **Script unifié** : Un seul point d'entrée pour tous les déploiements
- **Tests automatisés** : Suite de tests complète pour validation
- **Makefile synchronisé** : Commandes cohérentes avec le nouveau système
- **Documentation claire** : Instructions simplifiées

---

## 🛠️ Prérequis

```bash
# Installation d'Incus (si pas déjà installé)
sudo snap install incus --channel=latest/stable
sudo incus admin init

# Vérification
incus version
```

---

## 🚀 Déploiement Rapide

### Option 1 : Via Make (Recommandé)

```bash
# Déploiement complet en production
make deploy

# Déploiement en mode développement  
make deploy-dev

# Déploiement de l'infrastructure uniquement
make deploy-infrastructure

# Déploiement des applications uniquement
make deploy-apps

# Tests du déploiement
make deploy-test
```

### Option 2 : Via Script Direct

```bash
# Rendre le script exécutable
chmod +x scripts/deploy.sh

# Déploiement complet
./scripts/deploy.sh deploy

# Mode développement
./scripts/deploy.sh deploy --dev

# Voir toutes les options
./scripts/deploy.sh --help
```

---

## 🧪 Tests et Validation

### Tests Rapides (2-3 minutes)

```bash
# Via Make
make deploy-test-quick

# Via script
./scripts/test.sh --quick
```

### Tests Complets (5-10 minutes)

```bash
# Via Make  
make deploy-test

# Via script
./scripts/test.sh --full
```

### Tests Couverts

- ✅ Infrastructure Incus (réseau, profils)
- ✅ État des containers (statut, IPs)
- ✅ Connectivité réseau inter-containers
- ✅ Services (PostgreSQL, Redis, NFS)
- ✅ Applications (Backend, Frontend, Chat, Stream)
- ✅ Performance basique (mémoire, CPU)
- ✅ Sécurité (containers non-privilégiés)
- ✅ Sauvegarde/Restauration
- ✅ Monitoring et logs
- ✅ Tests de régression

---

## 🌐 Architecture Déployée

### 8 Containers Incus

| Container | IP | Service | Port |
|-----------|-----|---------|------|
| `veza-postgres` | 10.100.0.15 | PostgreSQL | 5432 |
| `veza-redis` | 10.100.0.17 | Redis | 6379 |
| `veza-storage` | 10.100.0.18 | NFS Storage | 2049 |
| `veza-backend` | 10.100.0.12 | API Go | 8080 |
| `veza-chat` | 10.100.0.13 | Chat Rust | 8081 |
| `veza-stream` | 10.100.0.14 | Stream Rust | 8082 |
| `veza-frontend` | 10.100.0.11 | React App | 5173 |
| `veza-haproxy` | 10.100.0.16 | Load Balancer | 80/8404 |

### Points d'Accès

- **🌐 Application principale** : http://10.100.0.16
- **📊 HAProxy Stats** : http://10.100.0.16:8404/stats
- **🎨 Frontend (dev)** : http://10.100.0.11:5173
- **⚙️ Backend API** : http://10.100.0.12:8080

---

## 🔧 Commandes Utiles

### Gestion des Containers

```bash
# Statut de tous les containers
incus list

# Logs d'un container spécifique
make logs-backend
make logs-frontend  
make logs-postgres

# Redémarrer un container
make incus-restart CONTAINER=veza-backend

# Arrêter tous les containers
make incus-stop

# Démarrer tous les containers  
make incus-start
```

### Base de Données

```bash
# Connexion à PostgreSQL
incus exec veza-postgres -- sudo -u postgres psql veza_db

# Sauvegarde
make backup

# Migration base existante
make db-migrate-existing
```

### ZFS Storage

```bash
# Statut du stockage ZFS
make zfs-status

# Créer des snapshots
make zfs-snapshot

# Monitoring temps réel
make zfs-monitor
```

---

## 🔄 Workflows Courants

### Premier Déploiement

```bash
# 1. Cloner le repository et aller dans le dossier
cd veza-full-stack

# 2. Déploiement complet
make deploy

# 3. Tester le déploiement
make deploy-test

# 4. Accéder à l'application
# http://10.100.0.16
```

### Développement

```bash
# 1. Déploiement en mode dev
make deploy-dev

# 2. Tests rapides
make deploy-test-quick

# 3. Logs en temps réel
make logs
```

### Redéploiement

```bash
# Reconstruction complète (efface tout)
make deploy-rebuild

# Ou redéploiement des apps uniquement
make deploy-apps --force
```

### Nettoyage

```bash
# Nettoyage complet de l'environnement
make incus-clean

# Nettoyage des fichiers de build
make clean
```

---

## 🚨 Dépannage

### Container ne démarre pas

```bash
# Vérifier les logs
incus info veza-backend
incus exec veza-backend -- journalctl -n 50

# Vérifier la configuration réseau
incus network show veza-network
```

### Service ne répond pas

```bash
# Vérifier si le service écoute
incus exec veza-backend -- netstat -ln | grep 8080

# Redémarrer le service
incus exec veza-backend -- systemctl restart veza-backend
```

### Problème de connectivité

```bash
# Tester la connectivité réseau
incus exec veza-frontend -- ping 10.100.0.12

# Vérifier les IPs des containers
incus list -c n,4
```

### Espace disque

```bash
# Vérifier l'espace ZFS
make zfs-status

# Nettoyer les anciens snapshots
make zfs-cleanup
```

---

## 📊 Scripts Nettoyés

### ❌ Scripts Supprimés (Redondants)

- `veza-deploy-final.sh` → Remplacé par `deploy.sh`
- `veza-deploy-apps-fixed.sh` → Intégré dans `deploy.sh`
- `veza-deploy-apps.sh` → Intégré dans `deploy.sh`
- `veza-manage.sh` → Fonctionnalités dans `deploy.sh`
- `incus-deploy-final.sh` → Remplacé par `deploy.sh`
- `incus-deploy-optimized.sh` → Remplacé par `deploy.sh`
- `incus-deploy-ultimate.sh` → Remplacé par `deploy.sh`
- `incus-deploy-final-auto.sh` → Remplacé par `deploy.sh`
- `incus-deploy-simple-auto.sh` → Remplacé par `deploy.sh`

### ✅ Scripts Conservés et Améliorés

- `deploy.sh` → **Script unifié principal**
- `test.sh` → **Nouveau script de tests complets**
- `incus-setup.sh` → Conservé pour la configuration initiale
- `incus-status.sh` → Conservé pour le monitoring
- `incus-logs.sh` → Conservé pour les logs
- `incus-restart.sh` → Conservé pour la gestion des containers

---

## 🎯 Prochaines Étapes

1. **✅ Tester le déploiement** : `make deploy-test`
2. **🔧 Personnaliser .env** : Adapter selon vos besoins
3. **📈 Monitoring** : Utiliser `make zfs-monitor` et `make health`
4. **🔄 Automatisation** : Intégrer en CI/CD si nécessaire

---

## 💡 Conseils

- **Toujours tester** après un déploiement avec `make deploy-test`
- **Surveiller l'espace disque** avec `make zfs-status`
- **Sauvegardes régulières** avec `make backup`
- **Logs centralisés** avec `make logs`

---

## 📞 Support

En cas de problème :

1. Exécuter `make deploy-test` pour un diagnostic complet
2. Vérifier les logs avec `make logs`
3. Consulter le statut avec `make status`
4. En dernier recours : `make deploy-rebuild`

---

**🎉 Félicitations ! Votre environnement Veza est maintenant propre, unifié et entièrement testable !** 