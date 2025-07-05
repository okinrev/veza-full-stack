# Guide Développement Frontend - Veza Platform

## Vue d'ensemble

Ce guide présente les standards, outils et bonnes pratiques pour le développement frontend sur la plateforme Veza (React, TypeScript, Next.js, etc.).

## 🖥️ Architecture Frontend
- Utilisation de React/Next.js
- Structure par feature/module
- Gestion d'état centralisée (Redux, Zustand)
- Séparation claire UI/Logique

## ✍️ Exemples

### Composant React typé
```tsx
interface UserCardProps {
  username: string;
  avatarUrl?: string;
}

/**
 * Affiche la carte d'un utilisateur.
 */
export const UserCard: React.FC<UserCardProps> = ({ username, avatarUrl }) => (
  <div className="user-card">
    <img src={avatarUrl} alt={username} />
    <span>{username}</span>
  </div>
);
```

### Appel API typé
```ts
import axios from 'axios';

export async function fetchUser(id: string): Promise<User> {
  const { data } = await axios.get(`/api/users/${id}`);
  return data;
}
```

## ✅ Bonnes Pratiques
- Typage strict avec TypeScript
- Découper les composants réutilisables
- Documenter chaque composant ([documentation-standards.md](./documentation-standards.md))
- Tester les composants (Jest, Testing Library)
- Gérer les erreurs d'API et d'état
- Utiliser des hooks personnalisés pour la logique partagée

## ⚠️ Pièges à Éviter
- State global non contrôlé
- Props non documentées
- Absence de gestion d'erreur
- CSS non modulaire
- Appels API non typés

## 🔗 Liens Utiles
- [react-best-practices.md](./react-best-practices.md)
- [component-design.md](./component-design.md)
- [state-management.md](./state-management.md)
- [api-testing.md](./api-testing.md)

---

**Dernière mise à jour** : $(date)
**Version** : 1.0.0 