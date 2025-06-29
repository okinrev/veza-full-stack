# Archive Talas

Ce dossier contient les anciens composants archivés du projet Talas.

## Contenu

### veza-basic-frontend/
Ancien frontend HTML/Alpine.js - **À CONSERVER UNIQUEMENT COMME RÉFÉRENCE**

**⚠️ IMPORTANT :** Ce frontend est obsolète et ne doit plus être utilisé.
Le frontend actif est maintenant `veza-frontend/` (React + TypeScript).

### Historique
- Archivé le : $(date)
- Raison : Migration vers architecture unifiée Talas
- Status : Archive de référence uniquement

## Frontend Actif

Le frontend principal est maintenant :
- **Chemin :** `../veza-frontend/`
- **Technologie :** React + TypeScript + Vite
- **Authentification :** JWT unifiée avec tous les services
- **WebSocket :** Intégration directe avec Chat et Stream servers

## Migration

L'ancien système HTML/Alpine.js a été remplacé par une architecture moderne :

| Ancien (Archivé) | Nouveau (Actif) |
|-------------------|-----------------|
| HTML statique | React SPA |
| Alpine.js | TypeScript |
| Auth basique | JWT unifiée |
| WebSocket simple | WebSocket intégrée |

---
*Archivé automatiquement par le système de migration Talas* 