# 📚 Documentation Veza Full-Stack Backend

## 🎯 Vue d'ensemble

Cette documentation couvre l'architecture complète du backend Veza, comprenant :
- **Backend API Go** : API REST principale avec authentification et gestion des utilisateurs
- **Chat Server Rust** : Serveur WebSocket haute performance pour la messagerie en temps réel
- **Stream Server Rust** : Serveur de streaming audio avec codecs multiples et métadonnées

## 📁 Structure de la Documentation

```
docs/
├── architecture/           # Architecture système et design patterns
├── api/                   # Documentation des APIs (REST, gRPC, WebSocket)
├── modules/               # Documentation détaillée par module
├── deployment/            # Guide de déploiement et configuration
├── development/           # Guide de développement et contribution
├── troubleshooting/       # Résolution de problèmes et debugging
└── generated/             # Documentation générée automatiquement
```

## 🚀 Démarrage Rapide

### Prérequis
- Go 1.23+
- Rust 1.70+
- PostgreSQL 15+
- Redis 7+
- Docker & Docker Compose

### Installation
```bash
# Cloner le repository
git clone <repository-url>
cd veza-full-stack

# Installer les dépendances
make install-deps

# Configurer l'environnement
cp .env.example .env
# Éditer .env avec vos configurations

# Lancer les services
make up
```

## 📖 Navigation

### Pour les Développeurs Frontend
- [Architecture API](./architecture/api-overview.md)
- [Endpoints REST](./api/rest/)
- [WebSocket Events](./api/websocket/)
- [Authentification](./api/auth/)

### Pour les DevOps
- [Architecture Système](./architecture/system-design.md)
- [Déploiement](./deployment/)
- [Monitoring](./deployment/monitoring.md)

### Pour les Développeurs Backend
- [Architecture Modulaire](./architecture/modular-design.md)
- [Patterns de Code](./development/patterns.md)
- [Tests](./development/testing.md)

## 🔧 Outils de Documentation

### Génération Automatique
- **Go** : `godoc`, `swag` (Swagger)
- **Rust** : `cargo doc`, `rustdoc`
- **API** : OpenAPI 3.0, AsyncAPI
- **Diagrammes** : Mermaid, PlantUML

### Commandes Utiles
```bash
# Générer la documentation Go
make docs-go

# Générer la documentation Rust
make docs-rust

# Générer les diagrammes
make docs-diagrams

# Lancer le serveur de documentation
make docs-serve
```

## 📊 Métriques et Monitoring

- **Prometheus** : Métriques système et business
- **Grafana** : Dashboards de monitoring
- **Jaeger** : Distributed tracing
- **ELK Stack** : Logs centralisés

## 🔒 Sécurité

- [Politique de Sécurité](./security/policy.md)
- [Audit de Sécurité](./security/audit.md)
- [Bonnes Pratiques](./security/best-practices.md)

## 🤝 Contribution

Voir le [Guide de Contribution](./development/contributing.md) pour les détails sur :
- Standards de code
- Processus de review
- Tests requis
- Documentation des changements

## 📞 Support

- **Issues** : [GitHub Issues](https://github.com/veza/issues)
- **Discussions** : [GitHub Discussions](https://github.com/veza/discussions)
- **Documentation** : Cette documentation

---

*Dernière mise à jour : $(date)*
*Version : 0.2.0* 