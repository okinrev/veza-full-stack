# Bonnes Pratiques React - Veza Platform

## Vue d'ensemble

Ce guide présente les conventions, astuces et pièges à éviter pour écrire du code React robuste, maintenable et performant dans le contexte Veza.

## 🏆 Principes de Base
- Utiliser des composants fonctionnels
- Typage strict avec TypeScript
- Découper les composants par feature
- Utiliser les hooks pour la logique d'état et d'effet
- Documenter chaque composant ([documentation-standards.md](./documentation-standards.md))

## ✍️ Exemples

### Composant contrôlé
```tsx
interface InputProps {
  value: string;
  onChange: (v: string) => void;
}

export const Input: React.FC<InputProps> = ({ value, onChange }) => (
  <input value={value} onChange={e => onChange(e.target.value)} />
);
```

### Hook personnalisé
```ts
import { useState } from 'react';

export function useToggle(initial = false): [boolean, () => void] {
  const [value, setValue] = useState(initial);
  return [value, () => setValue(v => !v)];
}
```

## ✅ Bonnes Pratiques
- Préférer les hooks aux classes
- Utiliser useMemo/useCallback pour les optimisations
- Gérer les effets de bord dans useEffect
- Tester les composants et hooks
- Séparer la logique métier de la présentation

## ⚠️ Pièges à Éviter
- State non localisé
- Effets secondaires non contrôlés
- Props non typées
- Rendu conditionnel complexe dans le JSX
- Absence de fallback UI

## 🔗 Liens Utiles
- [frontend-development.md](./frontend-development.md)
- [component-design.md](./component-design.md)
- [state-management.md](./state-management.md)
- [api-testing.md](./api-testing.md)

---

**Dernière mise à jour** : $(date)
**Version** : 1.0.0 