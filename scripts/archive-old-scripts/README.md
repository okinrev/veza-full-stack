# 🚀 Scripts Veza - Architecture Corrigée et Complète

## Scripts Fonctionnels et Prêts

Les scripts ont été complètement réécrits et sont maintenant **entièrement fonctionnels** :

### Utilisation

```bash
# 1. Configuration initiale (une seule fois)
./incus-setup.sh

# 2. Déploiement complet avec compilation et services
./incus-deploy.sh

# 3. Gestion des services
./incus-services.sh status              # État de tous les services
./incus-services.sh start backend       # Démarrer un service
./incus-services.sh restart chat        # Redémarrer un service
./incus-services.sh logs frontend       # Logs en temps réel
./incus-services.sh health              # Vérification santé

# 4. Monitoring
./incus-status.sh                       # État infrastructure
./incus-logs.sh [container]             # Logs containers

# 5. Nettoyage complet
./incus-clean.sh
```

### Architecture Déployée

Après `./incus-deploy.sh`, votre infrastructure complète sera opérationnelle :

| Container | IP | Service | Description |
|-----------|----|---------| ------------|
| `veza-frontend` | 10.100.0.11:5173 | React | Interface utilisateur moderne |
| `veza-backend` | 10.100.0.12:8080 | Go API | API REST backend |
| `veza-chat` | 10.100.0.13:8081 | Rust WebSocket | Serveur de chat temps réel |
| `veza-stream` | 10.100.0.14:8082 | Rust Stream | Serveur de streaming audio |
| `veza-postgres` | 10.100.0.15:5432 | PostgreSQL | Base de données principale |
| `veza-haproxy` | 10.100.0.16 | HAProxy | Load balancer et reverse proxy |
| `veza-redis` | 10.100.0.17:6379 | Redis | Cache et sessions |
| `veza-storage` | 10.100.0.18 | NFS | Stockage partagé |

### Points d'Accès

- **Application complète** : http://10.100.0.16 (via HAProxy)
- **Frontend de développement** : http://10.100.0.11:5173
- **API Backend** : http://10.100.0.12:8080
- **Stats HAProxy** : http://10.100.0.16:8404/stats

### Nouvelles Fonctionnalités

✅ **Compilation automatique** - Go et Rust compilés dans les containers  
✅ **Services systemd** - Démarrage automatique et gestion robuste  
✅ **Configuration DNS** - Communication inter-services par nom  
✅ **Montages NFS** - Stockage partagé pour uploads et audio  
✅ **Monitoring intégré** - Tests de santé et connectivité  
✅ **Logs centralisés** - Journalisation systemd pour tous les services  

### Gestionnaire de Services

Le nouveau script `incus-services.sh` vous permet de gérer facilement tous les services :

- **status** : État complet de l'infrastructure
- **start/stop/restart** : Gestion individuelle ou globale
- **logs** : Suivi en temps réel des logs
- **health** : Tests de connectivité et santé système

---

*Anciens scripts archivés dans `../archive/scripts-old/` - Les nouveaux scripts sont entièrement fonctionnels et prêts pour la production.* 