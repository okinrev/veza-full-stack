# 📁 Archive - Anciens Scripts Veza

## Scripts Archivés

Ces scripts ont été archivés car leurs fonctionnalités ont été intégrées dans le script principal `veza-manager.sh`.

### Scripts de Configuration
- `incus-setup.sh` - Configuration initiale Incus (intégré dans `veza-manager.sh setup`)

### Scripts de Déploiement  
- `incus-deploy.sh` - Ancien déploiement complet (remplacé par `deploy-base-containers.sh` + `update-source-code.sh`)
- `incus-deploy-simple.sh` - Version simplifiée (obsolète)
- `deploy-chat-only.sh` - Déploiement spécialisé chat (remplacé par options modulaires)

### Scripts de Gestion
- `incus-services.sh` - Gestion des services (intégré dans `veza-manager.sh`)
- `incus-status.sh` - État des services (intégré dans `veza-manager.sh status`)
- `incus-logs.sh` - Consultation des logs (intégré dans `veza-manager.sh logs`)
- `incus-clean.sh` - Nettoyage (intégré dans `veza-manager.sh clean`)
- `incus-clean-simple.sh` - Version simple nettoyage (obsolète)

### Scripts Réseau
- `fix-dns.sh` - Ancien fix DNS (remplacé par `network-fix.sh`)
- `test-dns.sh` - Tests DNS (intégré dans `network-fix.sh`)

### Documentation
- `README.md` - Ancien README (remplacé par `README-NEW-SYSTEM.md`)

## Migration vers le Nouveau Système

Toutes les fonctionnalités de ces scripts sont maintenant disponibles via :

```bash
# Script principal unifié
./veza-manager.sh [commande]

# Scripts spécialisés conservés
./network-fix.sh           # Réparation réseau
./deploy-base-containers.sh # Déploiement containers de base
./update-source-code.sh    # Mise à jour code source
```

Ces scripts archivés sont conservés pour référence historique uniquement. 