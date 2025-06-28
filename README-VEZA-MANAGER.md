# 🚀 Veza Manager - Infrastructure Unifiée

## 📖 Description

**Veza Manager** est le script unifié pour gérer complètement l'infrastructure Veza avec 8 containers Incus. Toutes les fonctionnalités de build, synchronisation, déploiement et monitoring sont intégrées dans un seul outil.

## 🎯 Utilisation Rapide

```bash
# Script principal
./scripts/veza-manager.sh <commande> [options]

# Raccourci simplifié
./veza <commande> [options]
```

## 📋 Commandes Principales

### 🚀 **Déploiement Initial**
```bash
./veza setup       # Configuration initiale
./veza deploy      # Déploiement complet (8 containers + services)
./veza status      # Vérifier l'état
```

### 🔨 **Build et Synchronisation**
```bash
./veza sync                # Synchroniser le code avec rsync
./veza build               # Compiler tous les projets
./veza build backend       # Compiler uniquement le backend
./veza build-start         # Build complet + démarrage des services
./veza watch               # Surveillance automatique + sync temps réel
```

### ⚙️ **Gestion des Services**
```bash
./veza start               # Démarrer tous les services
./veza start backend       # Démarrer uniquement le backend
./veza stop                # Arrêter tous les services
./veza restart chat        # Redémarrer le chat
./veza logs backend        # Voir les logs du backend
./veza logs chat -f        # Logs du chat en temps réel
```

### 🏥 **Monitoring et Maintenance**
```bash
./veza health              # Vérification complète de santé
./veza fix-deps            # Réparer/installer les dépendances
./veza network-fix         # Réparer les problèmes réseau
```

### 🧹 **Nettoyage et Export**
```bash
./veza export              # Exporter les containers
./veza import              # Importer les containers
./veza clean               # Nettoyage complet
```

## 🏗️ Architecture Infrastructure

### **8 Containers Incus Déployés :**

| Service | Container | IP | Port | Description |
|---------|-----------|----|----|-------------|
| **PostgreSQL** | `veza-postgres` | `10.5.191.154` | `5432` | Base de données principale |
| **Redis** | `veza-redis` | `10.5.191.95` | `6379` | Cache en mémoire |
| **Storage** | `veza-storage` | `10.5.191.144` | `2049` | Serveur NFS |
| **Backend** | `veza-backend` | `10.5.191.175` | `8080` | API REST Go |
| **Chat** | `veza-chat` | `10.5.191.108` | `8081` | Serveur WebSocket Rust |
| **Stream** | `veza-stream` | `10.5.191.188` | `8082` | Serveur streaming Rust |
| **Frontend** | `veza-frontend` | `10.5.191.121` | `3000` | Interface React |
| **HAProxy** | `veza-haproxy` | `10.5.191.29` | `80` | Load balancer |

### **🌐 Points d'Accès :**
- **Application complète :** http://10.5.191.29
- **Frontend direct :** http://10.5.191.121:3000
- **API Backend :** http://10.5.191.175:8080
- **HAProxy Stats :** http://10.5.191.29:8404/stats

## ⚡ Workflow de Développement

### **1. Setup Initial (une seule fois)**
```bash
./veza setup              # Configuration réseau et profils
./veza deploy             # Création containers + installation
```

### **2. Développement Quotidien**
```bash
./veza sync               # Synchroniser les modifications
./veza build backend      # Compiler après modification
./veza restart backend    # Redémarrer après build
./veza logs backend       # Vérifier les logs
```

### **3. Surveillance Continue**
```bash
./veza watch              # Mode surveillance automatique
# → Détecte les changements et synchronise automatiquement
```

### **4. Build et Démarrage Complet**
```bash
./veza build-start        # Compile tout + démarre tous les services
./veza status             # Vérifier que tout fonctionne
./veza health             # Tests de santé complets
```

## 🔧 Dépannage

### **Problèmes Courants :**

**Services non démarrés :**
```bash
./veza fix-deps           # Réparer les dépendances
./veza restart <service>  # Redémarrer le service
```

**Problèmes de compilation :**
```bash
./veza sync               # Re-synchroniser le code
./veza build <service>    # Re-compiler
```

**Problèmes réseau :**
```bash
./veza network-fix        # Réparer la configuration réseau
./veza status             # Vérifier la connectivité
```

**Réinitialisation complète :**
```bash
./veza clean              # Nettoyage complet
./veza deploy             # Redéploiement complet
```

## 📁 Scripts Intégrés

Le manager unifie tous les scripts précédemment créés :

- `setup-manual-containers.sh` → intégré dans `deploy`
- `setup-systemd-services.sh` → intégré dans `deploy`
- `setup-rsync.sh` → intégré dans `deploy`
- `quick-sync.sh` → commande `sync`
- `watch-and-sync.sh` → commande `watch`
- `build-and-start.sh` → commande `build-start`
- `status-all-services.sh` → commande `status`
- `start-all-services.sh` → commande `start`

## 🎉 Avantages

✅ **Un seul script** pour toute l'infrastructure  
✅ **Synchronisation automatique** du code  
✅ **Build et déploiement** intégrés  
✅ **Monitoring en temps réel**  
✅ **Gestion complète** des services  
✅ **Export/Import** des containers  
✅ **Scripts d'aide** et dépannage  

---

**Veza Manager v2.0.0** - Infrastructure complète en un seul outil ! 🚀 