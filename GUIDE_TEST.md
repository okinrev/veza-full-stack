# 🧪 Guide de Tests Complet - Veza

## 📋 Vue d'ensemble

Ce guide vous explique comment **tester complètement** votre déploiement Veza pour vous assurer que tout fonctionne comme attendu.

---

## 🚀 Tests Automatisés

### 1. Tests Rapides (2-3 minutes)

```bash
# Via Make
make deploy-test-quick

# Via script direct
./scripts/test.sh --quick
```

**Ce qui est testé :**
- ✅ Infrastructure Incus (réseau, profils)
- ✅ État des containers (existence, statut RUNNING)
- ✅ Adresses IP correctes

### 2. Tests Complets (5-10 minutes)

```bash
# Via Make
make deploy-test

# Via script direct
./scripts/test.sh --full
```

**Ce qui est testé :**
- ✅ Tous les tests rapides +
- ✅ Connectivité réseau inter-containers
- ✅ Services (PostgreSQL, Redis, NFS)
- ✅ Applications (Backend, Frontend, Chat, Stream)
- ✅ Performance basique
- ✅ Sécurité
- ✅ Sauvegarde/Restauration
- ✅ Monitoring
- ✅ Tests de régression

---

## 🔍 Tests Manuels Détaillés

### 1. Vérification de l'Infrastructure

```bash
# Statut des containers
incus list

# Vérification du réseau
incus network show veza-network

# Vérification des profils
incus profile list | grep veza
```

**Résultat attendu :**
- 8 containers en état `RUNNING`
- Réseau `veza-network` configuré
- 4 profils Veza (`veza-base`, `veza-app`, `veza-database`, `veza-storage`)

### 2. Test des Services de Base

#### PostgreSQL
```bash
# Connexion à la base
incus exec veza-postgres -- sudo -u postgres psql veza_db -c "\l"

# Test de connexion réseau
incus exec veza-backend -- pg_isready -h $(incus list veza-postgres -c 4 --format csv | cut -d' ' -f1) -p 5432 -U veza_user
```

#### Redis
```bash
# Test de connexion
incus exec veza-redis -- redis-cli ping

# Test depuis un autre container
incus exec veza-backend -- redis-cli -h $(incus list veza-redis -c 4 --format csv | cut -d' ' -f1) ping
```

#### NFS Storage
```bash
# Vérification du service NFS
incus exec veza-storage -- systemctl status nfs-kernel-server

# Vérification des exports
incus exec veza-storage -- exportfs -v
```

### 3. Test des Applications

#### Backend Go
```bash
# Statut du service
incus exec veza-backend -- systemctl status veza-backend

# Test de l'API
curl -f http://$(incus list veza-backend -c 4 --format csv | cut -d' ' -f1):8080/health || echo "Endpoint health non disponible"

# Logs du backend
incus exec veza-backend -- journalctl -u veza-backend -n 20
```

#### Frontend React
```bash
# Statut du service
incus exec veza-frontend -- systemctl status veza-frontend

# Test d'accès
curl -f http://$(incus list veza-frontend -c 4 --format csv | cut -d' ' -f1):5173 || echo "Frontend non accessible"
```

#### Chat Server Rust
```bash
# Statut du service
incus exec veza-chat -- systemctl status veza-chat

# Test du port
incus exec veza-chat -- netstat -ln | grep :8081
```

#### Stream Server Rust
```bash
# Statut du service
incus exec veza-stream -- systemctl status veza-stream

# Test du port
incus exec veza-stream -- netstat -ln | grep :8082
```

#### HAProxy
```bash
# Statut du service
incus exec veza-haproxy -- systemctl status haproxy

# Accès aux stats HAProxy
curl -f http://$(incus list veza-haproxy -c 4 --format csv | cut -d' ' -f1):8404/stats
```

---

## 🌐 Tests d'Accès Web

### 1. Via le Navigateur

**URL principales à tester :**
- **Application HAProxy** : `http://$(incus list veza-haproxy -c 4 --format csv | cut -d' ' -f1)`
- **HAProxy Stats** : `http://$(incus list veza-haproxy -c 4 --format csv | cut -d' ' -f1):8404/stats`
- **Frontend (dev)** : `http://$(incus list veza-frontend -c 4 --format csv | cut -d' ' -f1):5173`
- **Backend API** : `http://$(incus list veza-backend -c 4 --format csv | cut -d' ' -f1):8080`

### 2. Tests avec curl

```bash
# Récupérer les IPs des containers
HAPROXY_IP=$(incus list veza-haproxy -c 4 --format csv | cut -d' ' -f1)
FRONTEND_IP=$(incus list veza-frontend -c 4 --format csv | cut -d' ' -f1)
BACKEND_IP=$(incus list veza-backend -c 4 --format csv | cut -d' ' -f1)

# Test des endpoints
echo "=== Test HAProxy ==="
curl -I http://$HAPROXY_IP

echo "=== Test Frontend ==="
curl -I http://$FRONTEND_IP:5173

echo "=== Test Backend ==="
curl -I http://$BACKEND_IP:8080

echo "=== Test HAProxy Stats ==="
curl -I http://$HAPROXY_IP:8404/stats
```

---

## 🔗 Tests de Connectivité

### 1. Tests de Ping Inter-Containers

```bash
# Backend vers PostgreSQL
incus exec veza-backend -- ping -c 3 $(incus list veza-postgres -c 4 --format csv | cut -d' ' -f1)

# Backend vers Redis
incus exec veza-backend -- ping -c 3 $(incus list veza-redis -c 4 --format csv | cut -d' ' -f1)

# Frontend vers Backend
incus exec veza-frontend -- ping -c 3 $(incus list veza-backend -c 4 --format csv | cut -d' ' -f1)

# HAProxy vers tous les services
for container in veza-frontend veza-backend veza-chat veza-stream; do
    echo "=== HAProxy vers $container ==="
    incus exec veza-haproxy -- ping -c 1 $(incus list $container -c 4 --format csv | cut -d' ' -f1)
done
```

### 2. Tests de Ports

```bash
# Vérifier que les services écoutent sur les bons ports
echo "=== Ports en écoute ==="
echo "PostgreSQL (5432):"
incus exec veza-postgres -- netstat -ln | grep :5432

echo "Redis (6379):"
incus exec veza-redis -- netstat -ln | grep :6379

echo "Backend (8080):"
incus exec veza-backend -- netstat -ln | grep :8080

echo "Chat (8081):"
incus exec veza-chat -- netstat -ln | grep :8081

echo "Stream (8082):"
incus exec veza-stream -- netstat -ln | grep :8082

echo "Frontend (5173):"
incus exec veza-frontend -- netstat -ln | grep :5173

echo "HAProxy (80, 8404):"
incus exec veza-haproxy -- netstat -ln | grep -E ':(80|8404)'
```

---

## 📊 Tests de Performance

### 1. Utilisation des Ressources

```bash
# CPU et mémoire de chaque container
for container in veza-postgres veza-redis veza-storage veza-backend veza-chat veza-stream veza-frontend veza-haproxy; do
    echo "=== $container ==="
    incus exec $container -- free -h
    incus exec $container -- nproc
    echo ""
done
```

### 2. Tests de Charge Basiques

```bash
# Test de charge simple sur le backend (si curl est disponible)
BACKEND_IP=$(incus list veza-backend -c 4 --format csv | cut -d' ' -f1)

echo "=== Test de charge Backend ==="
for i in {1..10}; do
    curl -s -o /dev/null -w "Requête $i: %{http_code} - %{time_total}s\n" http://$BACKEND_IP:8080/ || echo "Requête $i: ÉCHEC"
done
```

### 3. Test de la Base de Données

```bash
# Test de performance simple PostgreSQL
incus exec veza-postgres -- sudo -u postgres psql veza_db -c "
SELECT 
    'PostgreSQL fonctionne' as status,
    version() as version,
    now() as timestamp;
"
```

---

## 🔄 Tests de Régression

### 1. Test de Redémarrage

```bash
# Redémarrer un service et vérifier qu'il revient
echo "=== Test de redémarrage Redis ==="
incus exec veza-redis -- systemctl restart redis-server
sleep 5
incus exec veza-redis -- systemctl is-active redis-server
incus exec veza-redis -- redis-cli ping
```

### 2. Test de Montée en Charge

```bash
# Créer quelques connexions simultanées
BACKEND_IP=$(incus list veza-backend -c 4 --format csv | cut -d' ' -f1)

echo "=== Test de connexions multiples ==="
for i in {1..5}; do
    curl -s http://$BACKEND_IP:8080/ &
done
wait
echo "Test terminé"
```

---

## 🛠️ Scripts de Test Rapides

### Script de Test Global

```bash
#!/bin/bash
# Créer un fichier test-global.sh

# Test rapide de tous les services
echo "🧪 Test Global Veza"
echo "==================="

# 1. Containers
echo "1. Vérification des containers..."
if [ $(incus list --format csv | grep veza | grep RUNNING | wc -l) -eq 8 ]; then
    echo "✅ Tous les containers sont actifs"
else
    echo "❌ Certains containers ne sont pas actifs"
    incus list | grep veza
fi

# 2. Services
echo "2. Vérification des services..."
services=("veza-postgres:postgresql" "veza-redis:redis-server" "veza-storage:nfs-kernel-server" "veza-backend:veza-backend" "veza-chat:veza-chat" "veza-stream:veza-stream" "veza-frontend:veza-frontend" "veza-haproxy:haproxy")

for service in "${services[@]}"; do
    container=$(echo $service | cut -d: -f1)
    service_name=$(echo $service | cut -d: -f2)
    
    if incus exec $container -- systemctl is-active $service_name >/dev/null 2>&1; then
        echo "✅ $container:$service_name"
    else
        echo "❌ $container:$service_name"
    fi
done

# 3. Connectivité
echo "3. Test de connectivité..."
if incus exec veza-backend -- ping -c 1 $(incus list veza-postgres -c 4 --format csv | cut -d' ' -f1) >/dev/null 2>&1; then
    echo "✅ Backend → PostgreSQL"
else
    echo "❌ Backend → PostgreSQL"
fi

echo ""
echo "🎉 Test terminé ! Consultez les résultats ci-dessus."
```

### Script d'Accès Rapide

```bash
#!/bin/bash
# Créer un fichier acces-services.sh

echo "🌐 Accès aux Services Veza"
echo "========================="

# Récupérer les IPs
HAPROXY_IP=$(incus list veza-haproxy -c 4 --format csv | cut -d' ' -f1)
FRONTEND_IP=$(incus list veza-frontend -c 4 --format csv | cut -d' ' -f1)
BACKEND_IP=$(incus list veza-backend -c 4 --format csv | cut -d' ' -f1)

echo "🌍 URLs d'accès:"
echo "• Application HAProxy: http://$HAPROXY_IP"
echo "• HAProxy Stats: http://$HAPROXY_IP:8404/stats"
echo "• Frontend React: http://$FRONTEND_IP:5173"
echo "• Backend API: http://$BACKEND_IP:8080"
echo ""
echo "📋 Commandes utiles:"
echo "• Logs Backend: incus exec veza-backend -- journalctl -u veza-backend -f"
echo "• Logs Frontend: incus exec veza-frontend -- journalctl -u veza-frontend -f"
echo "• Statut global: make status"
```

---

## 🚨 Résolution de Problèmes

### Problèmes Courants et Solutions

#### 1. Service ne démarre pas
```bash
# Vérifier les logs
incus exec <container> -- journalctl -u <service> -n 50

# Vérifier la configuration
incus exec <container> -- systemctl cat <service>

# Redémarrer le service
incus exec <container> -- systemctl restart <service>
```

#### 2. Problème de connectivité
```bash
# Vérifier les IPs
incus list -c n,4

# Tester la connectivité
incus exec <container1> -- ping <ip_container2>

# Vérifier les ports
incus exec <container> -- netstat -ln | grep <port>
```

#### 3. Frontend ne se charge pas
```bash
# Vérifier Node.js
incus exec veza-frontend -- node --version

# Vérifier les dépendances
incus exec veza-frontend -- npm list

# Reconstruire si nécessaire
incus exec veza-frontend -- npm run build
```

#### 4. Backend ne répond pas
```bash
# Vérifier Go
incus exec veza-backend -- go version

# Vérifier la base de données
incus exec veza-backend -- pg_isready -h <postgres_ip> -p 5432 -U veza_user

# Vérifier les variables d'environnement
incus exec veza-backend -- cat /app/.env
```

---

## ✅ Checklist de Validation Finale

**Infrastructure :**
- [ ] 8 containers en état RUNNING
- [ ] Réseau veza-network configuré
- [ ] Profils Veza créés

**Services de Base :**
- [ ] PostgreSQL actif et accessible
- [ ] Redis actif et accessible
- [ ] NFS Storage configuré

**Applications :**
- [ ] Backend Go déployé et actif
- [ ] Frontend React accessible
- [ ] Chat Server déployé
- [ ] Stream Server déployé
- [ ] HAProxy configuré et actif

**Connectivité :**
- [ ] Inter-containers fonctionnelle
- [ ] Ports ouverts correctement
- [ ] Services web accessibles

**Performance :**
- [ ] Ressources suffisantes
- [ ] Temps de réponse acceptables
- [ ] Pas d'erreurs dans les logs

---

**🎯 Si tous les tests passent, votre déploiement Veza est parfaitement fonctionnel !** 