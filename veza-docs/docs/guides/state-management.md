# Guide de Gestion d'État - Veza Platform

## Vue d'ensemble

Ce guide présente les stratégies, outils et bonnes pratiques pour la gestion d'état dans les applications frontend (React) et backend (Go) de la plateforme Veza.

## 🗂️ Stratégies de Gestion d'État
- State local (useState, useReducer)
- State global (Redux, Zustand, Context API)
- State persistant (localStorage, IndexedDB)
- State distribué (WebSocket, EventBus)

## ✍️ Exemples

### State local React
```tsx
const [count, setCount] = useState(0);
```

### State global avec Redux
```ts
import { createSlice } from '@reduxjs/toolkit';

const userSlice = createSlice({
  name: 'user',
  initialState: { loggedIn: false },
  reducers: {
    login: state => { state.loggedIn = true; },
    logout: state => { state.loggedIn = false; }
  }
});
```

### State distribué via WebSocket
```ts
// Réception d'un message de synchronisation
dispatch(updateStateFromServer(payload));
```

## ✅ Bonnes Pratiques
- Privilégier le state local pour les composants isolés
- Centraliser le state partagé (Redux, Zustand)
- Documenter la structure du state ([documentation-standards.md](./documentation-standards.md))
- Synchroniser le state avec le backend si besoin
- Gérer la persistance et la réhydratation
- Tester les reducers et selectors

## ⚠️ Pièges à Éviter
- State global incontrôlé
- Mutations directes du state
- Absence de normalisation des données
- Oublier la gestion des edge cases (loading, error)
- Fuites de mémoire (listeners non nettoyés)

## 🔗 Liens Utiles
- [component-design.md](./component-design.md)
- [frontend-development.md](./frontend-development.md)
- [api-testing.md](./api-testing.md)

---

**Dernière mise à jour** : $(date)
**Version** : 1.0.0 