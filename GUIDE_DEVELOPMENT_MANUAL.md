# 🚀 Guide d'Infrastructure de Développement Veza

Guide complet pour l'infrastructure de développement manuelle avec containers Incus, services systemd et synchronisation rsync automatique.

## 📋 Vue d'ensemble

Cette infrastructure offre :
- **8 containers Incus** spécialisés sur le réseau par défaut
- **Services systemd** pour chaque composant
- **Synchronisation rsync automatique** du code local vers les containers
- **Redémarrage intelligent** des services après modifications
- **Surveillance en temps réel** des changements de fichiers

## 🏗️ Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   veza-postgres │    │    veza-redis   │    │   veza-storage  │
│   PostgreSQL    │    │     Redis       │    │   NFS Server    │
│   Port: 5432    │    │   Port: 6379    │    │   Exports: /    │
└─────────────────┘    └─────────────────┘    └─────────────────┘

┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   veza-backend  │    │    veza-chat    │    │   veza-stream   │
│   API Go        │    │  Chat Server    │    │ Stream Server   │
│   Port: 8080    │    │ Rust + WebSocket│    │ Rust + Audio    │
│   rsync: ✅     │    │   Port: 8081    │    │   Port: 8082    │
└─────────────────┘    │   rsync: ✅     │    │   rsync: ✅     │
                       └─────────────────┘    └─────────────────┘

┌─────────────────┐    ┌─────────────────┐
│  veza-frontend  │    │   veza-haproxy  │
│   React + Vite  │    │ Load Balancer   │
│   Port: 3000    │    │   Port: 80      │
│   rsync: ✅     │    │ Stats: 8404     │
└─────────────────┘    └─────────────────┘
```

## 🚀 Installation Rapide

### 1. Créer les containers

```bash
# Créer tous les containers avec configuration complète
./scripts/setup-manual-containers.sh
```

### 2. Configurer les services systemd

```bash
# Installer tous les services systemd
./scripts/setup-systemd-services.sh
```

### 3. Configurer rsync et SSH

```bash
# Configuration complète rsync + clés SSH
./scripts/setup-rsync.sh
```

### 4. Démarrer l'infrastructure

```bash
# Démarrer tous les services
./scripts/start-all-services.sh

# Vérifier le statut
./scripts/status-all-services.sh
```

## 🔄 Workflow de Développement

### Synchronisation Manuelle

```bash
# Synchroniser tout
./scripts/quick-sync.sh

# Synchroniser un composant spécifique
./scripts/quick-sync.sh backend
./scripts/quick-sync.sh frontend
./scripts/quick-sync.sh chat
./scripts/quick-sync.sh stream

# Synchroniser + build + redémarrer
./scripts/quick-sync.sh backend --build --restart
./scripts/quick-sync.sh frontend --restart
```

### Surveillance Automatique

```bash
# Surveiller tous les composants (recommandé)
./scripts/watch-and-sync.sh

# Surveiller un composant spécifique
./scripts/watch-and-sync.sh backend
./scripts/watch-and-sync.sh frontend
```

### Développement TypeScript

Quand vous modifiez du code dans :
- `veza-backend-api/` → Sync automatique vers `veza-backend:/opt/veza/backend/`
- `veza-frontend/` → Sync automatique vers `veza-frontend:/opt/veza/frontend/`
- `veza-chat-server/` → Sync automatique vers `veza-chat:/opt/veza/chat/`
- `veza-stream-server/` → Sync automatique vers `veza-stream:/opt/veza/stream/`

## 🛠️ Commandes Utiles

### Gestion des Services

```bash
# Démarrer tout
./scripts/start-all-services.sh

# Arrêter tout
./scripts/stop-all-services.sh

# Voir le statut
./scripts/status-all-services.sh

# Redémarrer un service spécifique
incus exec veza-backend -- systemctl restart veza-backend
incus exec veza-frontend -- systemctl restart veza-frontend
```

### Logs et Débogage

```bash
# Logs d'un service
incus exec veza-backend -- journalctl -u veza-backend -f
incus exec veza-frontend -- journalctl -u veza-frontend -f

# Debug connexions rsync
./scripts/debug-connections.sh

# Tester manuellement une connexion SSH
ssh -i ~/.ssh/veza_rsa root@$(incus ls veza-backend -c 4 --format csv | cut -d' ' -f1)
```

### Accès aux Applications

```bash
# Obtenir l'IP du load balancer
HAPROXY_IP=$(incus ls veza-haproxy -c 4 --format csv | cut -d' ' -f1)

# URLs d'accès
echo "Application: http://$HAPROXY_IP"
echo "HAProxy Stats: http://$HAPROXY_IP:8404/stats"
echo "API Direct: http://$(incus ls veza-backend -c 4 --format csv | cut -d' ' -f1):8080"
echo "Frontend Direct: http://$(incus ls veza-frontend -c 4 --format csv | cut -d' ' -f1):3000"
```

## 🏛️ Structure des Containers

### Backend Go (`veza-backend`)
- **Path**: `/opt/veza/backend/`
- **Service**: `veza-backend.service`
- **Build**: `cd /opt/veza/backend && ./build.sh`
- **Binary**: `/opt/veza/backend/main`
- **Port**: 8080

### Frontend React (`veza-frontend`)
- **Path**: `/opt/veza/frontend/`
- **Service**: `veza-frontend.service`
- **Dev Server**: `npm run dev -- --host 0.0.0.0 --port 3000`
- **Build**: `cd /opt/veza/frontend && npm run build`
- **Port**: 3000

### Chat Server Rust (`veza-chat`)
- **Path**: `/opt/veza/chat/`
- **Service**: `veza-chat.service`
- **Build**: `cd /opt/veza/chat && cargo build --release`
- **Binary**: `/opt/veza/chat/target/release/veza-chat-server`
- **Port**: 8081

### Stream Server Rust (`veza-stream`)
- **Path**: `/opt/veza/stream/`
- **Service**: `veza-stream.service`
- **Build**: `cd /opt/veza/stream && cargo build --release`
- **Binary**: `/opt/veza/stream/target/release/veza-stream-server`
- **Port**: 8082

## 🔧 Configuration des Services

### Variables d'Environnement

Chaque container a ses variables dans `/etc/environment` :

**Backend Go:**
```env
DATABASE_URL=postgres://veza_user:veza_password@veza-postgres:5432/veza_db?sslmode=disable
REDIS_URL=redis://veza-redis:6379
JWT_SECRET=your-super-secret-jwt-key-change-in-production
PORT=8080
```

**Frontend React:**
```env
NODE_ENV=development
VITE_API_URL=http://veza-backend:8080/api/v1
VITE_WS_CHAT_URL=ws://veza-chat:8081/ws
VITE_WS_STREAM_URL=ws://veza-stream:8082/ws
```

### Modification des Configurations

```bash
# Éditer la configuration d'un service
incus exec veza-backend -- nano /etc/systemd/system/veza-backend.service

# Recharger après modification
incus exec veza-backend -- systemctl daemon-reload
incus exec veza-backend -- systemctl restart veza-backend
```

## 🚨 Dépannage

### Problèmes de Synchronisation

```bash
# Vérifier les connexions
./scripts/debug-connections.sh

# Tester manuellement
rsync -avz --dry-run -e "ssh -i ~/.ssh/veza_rsa" veza-backend-api/ root@IP_CONTAINER:/opt/veza/backend/

# Régénérer les clés SSH
rm ~/.ssh/veza_rsa*
./scripts/setup-rsync.sh
```

### Services qui ne Démarrent Pas

```bash
# Vérifier les logs
incus exec veza-backend -- journalctl -u veza-backend -n 50

# Vérifier la configuration
incus exec veza-backend -- systemctl status veza-backend

# Redémarrer les dépendances
incus exec veza-postgres -- systemctl restart postgresql
incus exec veza-redis -- systemctl restart redis-server
```

### Problèmes de Réseau

```bash
# Vérifier la connectivité entre containers
incus exec veza-backend -- ping veza-postgres
incus exec veza-backend -- telnet veza-postgres 5432

# Vérifier les ports ouverts
incus exec veza-backend -- netstat -tlnp
```

## 🎯 Bonnes Pratiques

### Développement Efficace

1. **Utilisez la surveillance automatique** : `./scripts/watch-and-sync.sh`
2. **Synchronisez avant de tester** : `./scripts/quick-sync.sh component --restart`
3. **Surveillez les logs** : `incus exec container -- journalctl -u service -f`
4. **Testez les builds localement** avant de synchroniser

### Optimisations

1. **Exclusions rsync** : Les fichiers suivants sont automatiquement exclus :
   - `.git`, `node_modules`, `target`, `dist`, `build`, `.next`, `*.log`

2. **Build intelligent** : 
   - Frontend : Seulement `npm install` en dev
   - Rust : Build release automatique
   - Go : Build optimisé

3. **Debounce** : La surveillance attend 2 secondes avant de synchroniser

## 📊 Monitoring

### HAProxy Stats
- URL : `http://HAPROXY_IP:8404/stats`
- Monitoring en temps réel des backends
- Statistiques de charge et santé

### Logs Centralisés
```bash
# Tous les logs dans journald
incus exec veza-backend -- journalctl -u veza-backend -f
incus exec veza-frontend -- journalctl -u veza-frontend -f
incus exec veza-chat -- journalctl -u veza-chat -f
incus exec veza-stream -- journalctl -u veza-stream -f
```

---

## 🎉 Prêt pour le Développement !

Votre infrastructure est maintenant configurée pour un développement rapide et efficace. 

**Commandes de démarrage rapide :**

```bash
# 1. Démarrer tout
./scripts/start-all-services.sh

# 2. Synchroniser le code
./scripts/quick-sync.sh

# 3. Activer la surveillance automatique
./scripts/watch-and-sync.sh
```

**Workflow quotidien :**
1. Modifiez votre code localement
2. Les changements sont automatiquement synchronisés
3. Les services redémarrent automatiquement
4. Testez immédiatement vos modifications

**Bonne programmation ! 🚀** 