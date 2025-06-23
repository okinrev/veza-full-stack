# Veza Frontend

Application React + TypeScript + Vite avec architecture modulaire propre.

## Structure du Projet

```
src/
├── app/                    # Configuration de l'application
│   ├── pages/             # Pages de l'app principale  
│   ├── providers/         # Providers React (Auth, Theme, WebSocket)
│   ├── Router.tsx         # Configuration du routage
│   └── routes.tsx         # Définition des routes
├── components/            # Composants UI partagés
│   ├── dev/              # Composants de développement
│   ├── layout/           # Composants de layout
│   └── ui/               # Composants UI de base (shadcn/ui)
├── features/             # Fonctionnalités métier
│   ├── auth/            # Authentification
│   ├── chat/            # Chat et messages
│   ├── products/        # Gestion des produits
│   ├── profile/         # Profil utilisateur
│   ├── resources/       # Ressources partagées
│   └── tracks/          # Pistes audio
├── shared/              # Code partagé
│   ├── api/            # Services API
│   ├── components/     # Composants partagés
│   ├── hooks/          # Hooks personnalisés
│   ├── lib/            # Utilitaires et services
│   ├── stores/         # Stores Zustand
│   └── utils/          # Fonctions utilitaires
├── test/               # Configuration des tests
├── App.tsx             # Composant racine
└── main.tsx           # Point d'entrée
```

## Technologies

- **React 18** - Framework UI
- **TypeScript** - Typage statique
- **Vite** - Build tool
- **Tailwind CSS** - Styling
- **shadcn/ui** - Composants UI
- **Zustand** - State management
- **React Router** - Routing
- **Axios** - HTTP client
- **Framer Motion** - Animations
- **Vitest** - Tests unitaires
- **Playwright** - Tests E2E

## Installation

```bash
npm install
```

## Développement

```bash
npm run dev
```

## Build

```bash
npm run build
```

## Tests

```bash
# Tests unitaires
npm run test

# Tests E2E
npm run test:e2e
``` 