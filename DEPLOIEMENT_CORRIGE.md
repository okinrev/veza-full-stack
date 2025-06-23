# 🚀 Veza - Déploiement Automatisé Corrigé

Ce guide présente le déploiement automatisé corrigé de l'infrastructure Veza avec les bonnes configurations d'IP et l'interconnexion des services.

## 🔧 Problèmes Identifiés et Corrigés

### 1. Problème de Configuration IP
- **Problème**: Incohérence entre les IPs configurées dans les scripts (`10.100.0.x`) et les IPs réelles (`10.5.191.x`)
- **Solution**: Utilisation des vraies IPs du fichier `configs/infrastructure.yaml`

### 2. Problème de Connexion Database
- **Problème**: Backend essayait de se connecter à `10.5.191.47` au lieu de `10.5.191.134`
- **Solution**: Script `fix-backend-connection.sh` qui détecte automatiquement la bonne IP

### 3. Problème HAProxy
- **Problème**: Mauvais ports configurés (3000 au lieu de 5173, 3001 au lieu de 8081)
- **Solution**: Configuration corrigée dans `deploy-all.sh`

## 🎯 Scripts de Déploiement Corrigés

### Scripts Principaux

| Script | Description | Usage |
|--------|-------------|-------|
| `veza-deploy.sh` | **Script orchestrateur principal** | `./scripts/veza-deploy.sh deploy` |
| `deploy-unified.sh` | **Déploiement applications avec bonnes IPs** | `./scripts/deploy-unified.sh` |
| `fix-backend-connection.sh` | **Correction rapide connexion DB** | `./scripts/fix-backend-connection.sh` |
| `validate-deployment.sh` | **Validation du déploiement** | `./scripts/validate-deployment.sh` |

### Scripts Existants (corrigés)

| Script | Description | Modifications |
|--------|-------------|---------------|
| `deploy-all.sh` | Configuration HAProxy finale | ✅ Ports corrigés |
| `incus-setup.sh` | Configuration initiale | ✅ Fonctionne |
| `incus-deploy.sh` | Déploiement infrastructure | ✅ Fonctionne |
| `test-complete.sh` | Tests complets | ✅ Fonctionne |

## 🚀 Déploiement Rapide

### Option 1: Déploiement Complet Automatique
```bash
# Déploiement complet de A à Z
./scripts/veza-deploy.sh deploy
```

### Option 2: Correction Rapide (si infrastructure existe)
```bash
# Si vous avez déjà l'infrastructure mais problème de connexion
./scripts/fix-backend-connection.sh
```

### Option 3: Étape par Étape
```bash
# 1. Configuration initiale
./scripts/incus-setup.sh

# 2. Infrastructure
./scripts/incus-deploy.sh

# 3. Applications
./scripts/deploy-unified.sh

# 4. Configuration finale
./scripts/deploy-all.sh

# 5. Validation
./scripts/validate-deployment.sh
```

## 🔍 Validation et Tests

### Validation Rapide
```bash
./scripts/validate-deployment.sh
```

### Tests Complets
```bash
./scripts/test-complete.sh
```

### Vérification Manuelle
```bash
# Statut des containers
incus list

# Test des services
curl http://10.5.191.133          # HAProxy (principal)
curl http://10.5.191.41:5173      # Frontend
curl http://10.5.191.241:8080     # Backend API
curl http://10.5.191.49:8081      # Chat Server
curl http://10.5.191.196:8082     # Stream Server
```

## 🌐 Configuration Réseau Finale

### IPs et Ports des Services

| Service | Container | IP | Port | URL |
|---------|-----------|----|----- |-----|
| **HAProxy** | veza-haproxy | 10.5.191.133 | 80 | http://10.5.191.133 |
| **Frontend** | veza-frontend | 10.5.191.41 | 5173 | http://10.5.191.41:5173 |
| **Backend** | veza-backend | 10.5.191.241 | 8080 | http://10.5.191.241:8080 |
| **Chat** | veza-chat | 10.5.191.49 | 8081 | http://10.5.191.49:8081 |
| **Stream** | veza-stream | 10.5.191.196 | 8082 | http://10.5.191.196:8082 |
| **PostgreSQL** | veza-postgres | 10.5.191.134 | 5432 | - |
| **Redis** | veza-redis | 10.5.191.186 | 6379 | - |
| **Storage** | veza-storage | 10.5.191.206 | 2049 | - |

### Routage HAProxy

| Chemin | Service de Destination |
|--------|----------------------|
| `/` | Frontend React (10.5.191.41:5173) |
| `/api/` | Backend Go (10.5.191.241:8080) |
| `/chat-api/` | Chat Server (10.5.191.49:8081) |
| `/stream/` | Stream Server (10.5.191.196:8082) |

## 🔧 Configuration Backend

### Variables d'Environnement (.env)
```bash
DATABASE_URL=postgres://veza:veza_password@10.5.191.134:5432/veza_db?sslmode=disable
REDIS_URL=redis://10.5.191.186:6379
JWT_SECRET=veza_jwt_secret_key_2025_production
SERVER_PORT=8080
SERVER_HOST=0.0.0.0
LOG_LEVEL=info
ENVIRONMENT=production
```

## 🐛 Dépannage

### Problème de Connexion Database
```bash
# Vérifier l'IP PostgreSQL
incus list veza-postgres -c 4

# Corriger automatiquement
./scripts/fix-backend-connection.sh
```

### Problème HAProxy
```bash
# Redémarrer HAProxy avec la bonne config
./scripts/deploy-all.sh
```

### Problème de Service
```bash
# Vérifier les logs
./scripts/incus-logs.sh veza-backend
./scripts/incus-logs.sh veza-chat

# Redémarrer un service
incus exec veza-backend -- systemctl restart veza-backend
```

### Nettoyage Complet
```bash
# Si tout va mal, recommencer
./scripts/incus-clean.sh
./scripts/veza-deploy.sh deploy
```

## 📊 Points de Vérification

### ✅ Checklist de Déploiement Réussi

- [ ] Tous les containers sont `RUNNING`
- [ ] PostgreSQL répond à `pg_isready`
- [ ] Redis répond à `PING`
- [ ] Backend répond sur port 8080
- [ ] Frontend accessible sur port 5173
- [ ] HAProxy route correctement
- [ ] Chat/Stream servers accessibles
- [ ] Pas d'erreurs dans les logs

### 🎯 URLs à Tester

- **Application principale**: http://10.5.191.133
- **Frontend direct**: http://10.5.191.41:5173
- **API Backend**: http://10.5.191.241:8080
- **HAProxy Stats**: http://10.5.191.133:8404/stats (si configuré)

## 🚀 Résultat Final

Après le déploiement correct, vous devriez avoir :

1. **Infrastructure complète** avec 8 containers interconnectés
2. **Services communicants** avec les bonnes IPs
3. **Application web accessible** via HAProxy
4. **API fonctionnelle** avec base de données
5. **Chat et Stream** prêts pour utilisation

**Accès principal**: http://10.5.191.133 🌐 