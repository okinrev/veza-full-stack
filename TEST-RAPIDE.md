# 🧪 Guide de Test Rapide - Veza

## ⚡ Tests en 5 minutes

### 1. Vérification Rapide des Containers

```bash
# Voir tous les containers
incus list

# Statut attendu : 8 containers RUNNING
```

### 2. Test des Services Principaux

```bash
# PostgreSQL
incus exec veza-postgres -- sudo -u postgres psql -c "SELECT 'PostgreSQL OK' as status;"

# Redis
incus exec veza-redis -- redis-cli ping

# Backend (si déployé)
incus exec veza-backend -- systemctl status veza-backend

# Frontend (si déployé) 
incus exec veza-frontend -- systemctl status veza-frontend
```

### 3. Test de Connectivité

```bash
# Backend vers PostgreSQL
incus exec veza-backend -- ping -c 1 $(incus list veza-postgres -c 4 --format csv | cut -d' ' -f1)

# Frontend vers Backend
incus exec veza-frontend -- ping -c 1 $(incus list veza-backend -c 4 --format csv | cut -d' ' -f1)
```

### 4. URLs d'Accès Web

```bash
# Récupérer les IPs
echo "HAProxy: http://$(incus list veza-haproxy -c 4 --format csv | cut -d' ' -f1)"
echo "Frontend: http://$(incus list veza-frontend -c 4 --format csv | cut -d' ' -f1):5173"
echo "Backend: http://$(incus list veza-backend -c 4 --format csv | cut -d' ' -f1):8080"
echo "HAProxy Stats: http://$(incus list veza-haproxy -c 4 --format csv | cut -d' ' -f1):8404/stats"
```

## 🚀 Déploiement et Tests Automatiques

### Déploiement Complet
```bash
# Déploiement en mode développement
./scripts/deploy.sh deploy --dev

# Tests automatiques
./scripts/test.sh --quick
```

### Résolution de Problèmes

```bash
# Logs d'un service
incus exec <container> -- journalctl -u <service> -n 20

# Redémarrer un service
incus exec <container> -- systemctl restart <service>

# Vérifier les ports
incus exec <container> -- netstat -ln | grep <port>
```

## ✅ Checklist Rapide

- [ ] 8 containers RUNNING
- [ ] PostgreSQL répond
- [ ] Redis répond  
- [ ] Services déployés
- [ ] Connectivité OK
- [ ] URLs accessibles

**Si tout est ✅ : Votre Veza fonctionne parfaitement !** 