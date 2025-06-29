# 🚀 QUICK START - Talas Platform

## Démarrage Immédiat

```bash
# 1. Configuration initiale (première fois)
./scripts/talas-admin.sh setup

# 2. Compilation de tous les services
./scripts/talas-admin.sh build

# 3. Démarrage de la plateforme
./scripts/talas-admin.sh start

# 4. Vérification
./scripts/talas-admin.sh status
```

## Accès aux Services

- **Frontend :** http://localhost:5173
- **Backend API :** http://localhost:8080
- **Chat WebSocket :** ws://localhost:3001/ws
- **Stream WebSocket :** ws://localhost:3002/ws

## Commandes Principales

```bash
./scripts/talas-admin.sh status     # État des services
./scripts/talas-admin.sh logs       # Voir les logs
./scripts/talas-admin.sh restart    # Redémarrer
./scripts/talas-admin.sh stop       # Arrêter
./scripts/talas-admin.sh test       # Tests d'intégration
./scripts/talas-admin.sh clean      # Nettoyer
```

## Structure Finale

✅ **4 modules intégrés** avec authentification JWT unifiée  
✅ **Script d'administration central** (`talas-admin.sh`)  
✅ **Documentation complète** (`docs/INTEGRATION.md`)  
✅ **Tests automatisés** et monitoring  

**Votre plateforme Talas est prête ! 🎉**
