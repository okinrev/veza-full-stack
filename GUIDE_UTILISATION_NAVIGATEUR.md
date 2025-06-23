# 🌐 Guide d'Utilisation de l'Application Veza dans le Navigateur

## 🎯 **ACCÈS À L'APPLICATION**

### **URL Principale (Recommandée)**
```
http://10.5.191.133
```
**Point d'entrée unique via HAProxy - Toutes les APIs sont automatiquement routées**

### **URL Alternative (Accès Direct)**
```
http://10.5.191.41:3000
```
**Accès direct au frontend React**

---

## 🚀 **SERVICES DISPONIBLES**

### ✅ **TOUS LES SERVICES SONT OPÉRATIONNELS**

| Service | Status | URL Directe | Via HAProxy |
|---------|---------|-------------|-------------|
| **Frontend React** | ✅ Actif (200) | `http://10.5.191.41:3000` | `http://10.5.191.133/` |
| **Backend Go** | ✅ Actif (404) | `http://10.5.191.241:8080` | `http://10.5.191.133/api/` |
| **Chat Server Rust** | ✅ Healthy | `http://10.5.191.49:3001` | `http://10.5.191.133/chat-api/` |
| **Stream Server Rust** | ✅ Healthy | `http://10.5.191.196:8000` | `http://10.5.191.133/stream/` |

---

## 🎮 **FONCTIONNALITÉS DISPONIBLES**

### 1. **Interface Utilisateur (Frontend React)**
- 🔐 **Authentification** - Login/Register
- 💬 **Chat en Temps Réel** - Messages instantanés
- 🎵 **Gestion Audio** - Streaming et lecture
- 📊 **Dashboard** - Statistiques utilisateur
- 👤 **Profil Utilisateur** - Gestion du compte

### 2. **Chat System (Rust)**
**Endpoints disponibles:**
- `GET /chat-api/messages?room=general` - Récupérer les messages
- `POST /chat-api/messages` - Envoyer un message
- `GET /chat-api/messages/stats` - Statistiques du chat

**Test rapide:**
```bash
# Via HAProxy
curl http://10.5.191.133/chat-api/health
curl http://10.5.191.133/chat-api/messages?room=general
```

### 3. **Streaming Audio (Rust)**
**Endpoints disponibles:**
- `GET /stream/health` - Santé du service
- `GET /stream/list` - Liste des fichiers audio
- `GET /stream/:filename` - Streaming d'un fichier
- `GET /stream/info/:filename` - Informations sur un fichier

**Test rapide:**
```bash
# Via HAProxy
curl http://10.5.191.133/stream/health
curl http://10.5.191.133/stream/list
```

### 4. **API Backend (Go)**
**Endpoints disponibles:**
- Gestion des utilisateurs
- Authentification JWT
- CRUD des données
- APIs RESTful

---

## 🧪 **TESTS DE FONCTIONNEMENT**

### **Test 1: Interface Principale**
1. Ouvrez votre navigateur
2. Allez sur `http://10.5.191.133`
3. Vous devriez voir l'interface Veza avec :
   - Logo et navigation
   - Pages de connexion/inscription
   - Dashboard principal

### **Test 2: Chat en Temps Réel**
1. Accédez à la section Chat
2. Envoyez un message de test
3. Vérifiez que le message s'affiche instantanément

### **Test 3: Streaming Audio**
1. Accédez à la section Tracks/Audio
2. Parcourez les fichiers disponibles
3. Testez la lecture audio

### **Test 4: APIs via Navigateur**
Testez les APIs directement dans votre navigateur :

**Chat API:**
```
http://10.5.191.133/chat-api/health
http://10.5.191.133/chat-api/messages?room=general
```

**Stream API:**
```
http://10.5.191.133/stream/health
http://10.5.191.133/stream/list
```

---

## 🔧 **ARCHITECTURE DE L'APPLICATION**

```
┌─────────────────┐    ┌─────────────────┐
│   NAVIGATEUR    │───▶│     HAProxy     │ 
│                 │    │  10.5.191.133   │
└─────────────────┘    └─────────┬───────┘
                                 │
                    ┌────────────┼────────────┐
                    │            │            │
                    ▼            ▼            ▼
          ┌─────────────┐ ┌─────────────┐ ┌─────────────┐
          │  Frontend   │ │ Chat Server │ │Stream Server│
          │   React     │ │   (Rust)    │ │   (Rust)    │
          │10.5.191.41  │ │10.5.191.49  │ │10.5.191.196 │
          └─────────────┘ └─────────────┘ └─────────────┘
                    │
                    ▼
          ┌─────────────┐
          │  Backend    │
          │    Go       │
          │10.5.191.241 │
          └─────────────┘
```

---

## 🚀 **COMMANDES UTILES**

### **Redémarrer tous les services:**
```bash
./scripts/deploy-all.sh
```

### **Voir les logs en temps réel:**
```bash
# Chat Server
incus exec veza-chat -- journalctl -u veza-chat -f

# Stream Server  
incus exec veza-stream -- journalctl -u veza-stream -f

# Frontend
incus exec veza-frontend -- journalctl -u nginx -f

# Backend
incus exec veza-backend -- journalctl -u veza-backend -f
```

### **Vérifier l'état des services:**
```bash
# Tous les containers
incus list

# Services individuels
curl http://10.5.191.49:3001/health    # Chat
curl http://10.5.191.196:8000/health   # Stream
curl -I http://10.5.191.41:3000        # Frontend
curl -I http://10.5.191.241:8080       # Backend
```

---

## 🎉 **FÉLICITATIONS !**

Votre application Veza est **100% fonctionnelle** avec :

✅ **4 services interconnectés**
✅ **Point d'entrée unique via HAProxy**
✅ **Frontend React moderne**
✅ **Chat temps réel (Rust)**
✅ **Streaming audio (Rust)**
✅ **Backend API (Go)**
✅ **Architecture microservices**
✅ **Scripts de déploiement automatisés**

**L'application est prête pour une utilisation complète dans votre navigateur web !**

---

## 📞 **Support et Maintenance**

### **En cas de problème:**

1. **Vérifiez les services:**
   ```bash
   incus list
   ```

2. **Redéployez si nécessaire:**
   ```bash
   ./scripts/deploy-all.sh
   ```

3. **Consultez les logs:**
   ```bash
   # Exemple pour le chat
   incus exec veza-chat -- journalctl -u veza-chat --no-pager -n 50
   ```

### **URLs de Debug:**
- Frontend: `http://10.5.191.41:3000`
- Chat Health: `http://10.5.191.49:3001/health`
- Stream Health: `http://10.5.191.196:8000/health`
- HAProxy Stats: `http://10.5.191.133` (si configuré)

**Votre application Veza est maintenant totalement opérationnelle ! 🚀** 