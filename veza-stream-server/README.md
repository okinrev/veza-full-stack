# Stream Server

Serveur de streaming audio haute performance écrit en Rust, conçu pour la diffusion sécurisée de fichiers audio avec authentification, analytics en temps réel et fonctionnalités avancées.

## 🚀 Fonctionnalités

- **Streaming audio sécurisé** avec support des Range Requests
- **Authentification JWT** avec gestion des rôles et permissions
- **Analytics en temps réel** des sessions d'écoute
- **Cache intelligent** pour optimiser les performances
- **Rate limiting** pour prévenir les abus
- **Health checks** avec monitoring système
- **WebSocket** pour les événements temps réel
- **Compression audio** avec profils de qualité multiples
- **Notifications** multi-canal (WebSocket, Email, SMS, Push)
- **Architecture modulaire** prête pour la production

## 📋 Prérequis

- **Rust** 1.70 ou supérieur
- **SQLite** (intégré)
- **FFmpeg** (optionnel, pour la compression avancée)

## ⚡ Installation Rapide

1. **Clone du projet**
```bash
git clone https://github.com/your-org/veza-stream-server.git
cd veza-stream-server
```

2. **Configuration**
```bash
cp env.example .env
# Éditez .env avec vos paramètres
```

3. **Compilation et lancement**
```bash
cargo build --release
./target/release/stream_server
```

## 🔧 Configuration

### Variables d'environnement principales

```bash
# Configuration de base
STREAM_SERVER_PORT=8082
AUDIO_DIR=./audio
SECRET_KEY=your-32-chars-secret-key-here

# Base de données
DATABASE_URL=sqlite:stream_server.db

# Sécurité
JWT_SECRET=your-jwt-secret-key
ALLOWED_ORIGINS=https://yourapp.com,https://admin.yourapp.com

# Performance
MAX_FILE_SIZE=104857600  # 100MB
MAX_RANGE_SIZE=10485760  # 10MB
CACHE_MAX_SIZE_MB=256
```

### Structure des fichiers

```
audio/
├── track1.mp3
├── track2.mp3
└── subfolder/
    └── track3.mp3
```

## 🌐 API Endpoints

### Streaming Audio
```
GET /stream/{filename}?expires={timestamp}&sig={signature}
```

### Health Checks
```
GET /health              # Santé basique
GET /health/detailed     # Diagnostics complets
GET /metrics            # Métriques Prometheus
```

### Analytics
```
GET /admin/stats?token={admin-token}  # Statistiques administrateur
```

## 🐳 Déploiement Docker

```bash
# Build de l'image
docker build -t stream-server .

# Lancement avec Docker Compose
docker-compose up -d
```

## 🛠️ Outils Utilitaires

Le projet inclut des outils dans le répertoire `tools/` :

```bash
# Compilation des outils
cd tools
cargo build --release

# Génération de waveforms
./target/release/waveform_generator --input-dir ../audio --output-dir ../waveforms

# Transcodage audio
./target/release/transcoder --input-dir ../audio --output-dir ../transcoded
```

## 📊 Monitoring

### Health Checks
```bash
curl http://localhost:8082/health
curl http://localhost:8082/health/detailed
```

### Métriques
```bash
curl http://localhost:8082/metrics
```

## 🔐 Sécurité

- URLs signées avec expiration temporelle
- Authentification JWT avec refresh tokens
- Rate limiting par IP
- Headers de sécurité automatiques
- Validation stricte des entrées
- Protection contre les attaques courantes

## 📈 Performance

- Architecture asynchrone complète
- Streaming par chunks optimisé
- Cache LRU pour les métadonnées
- Compression HTTP automatique
- Pool de connexions base de données
- Workers parallèles pour le traitement

## 🏗️ Architecture

```
┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│   Client    │◄──►│   Nginx     │◄──►│ Stream      │
│ (Frontend)  │    │ (Reverse    │    │ Server      │
│             │    │  Proxy)     │    │ (Rust)      │
└─────────────┘    └─────────────┘    └─────────────┘
                                            │
        ┌───────────────────────────────────┼───────────────────────────────────┐
        │                                   │                                   │
        ▼                                   ▼                                   ▼
┌─────────────┐                   ┌─────────────┐                   ┌─────────────┐
│   SQLite    │                   │   File      │                   │   Redis     │
│ (Analytics) │                   │   Cache     │                   │ (Optional)  │
└─────────────┘                   └─────────────┘                   └─────────────┘
```

## 📄 Licence

Ce projet est sous licence MIT. Voir le fichier [LICENSE](LICENSE) pour plus de détails.

## 🤝 Contribution

1. Fork le projet
2. Créez une branche feature (`git checkout -b feature/amazing-feature`)
3. Committez vos changements (`git commit -m 'Add amazing feature'`)
4. Push vers la branche (`git push origin feature/amazing-feature`)
5. Ouvrez une Pull Request

## 📞 Support

- **Issues** : [GitHub Issues](https://github.com/your-org/veza-stream-server/issues)
- **Documentation** : Consultez les archives pour la documentation complète
- **Email** : team@streamserver.com

---

**Développé avec ❤️ en Rust pour des performances maximales** 🦀 