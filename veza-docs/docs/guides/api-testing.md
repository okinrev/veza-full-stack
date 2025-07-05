# Guide de Tests d'API - Veza Platform

## Vue d'ensemble

Ce guide présente les méthodes, outils et bonnes pratiques pour tester les API REST et WebSocket de la plateforme Veza.

## 🧪 Types de Tests
- Tests unitaires (handlers, services)
- Tests d'intégration (end-to-end)
- Tests de charge et de performance
- Tests de sécurité ([security-guidelines.md](./security-guidelines.md))

## ✍️ Exemples

### Test unitaire Go (Gin)
```go
func TestCreateUser(t *testing.T) {
    router := setupRouter()
    w := httptest.NewRecorder()
    req, _ := http.NewRequest("POST", "/users", bytes.NewBuffer([]byte(`{"email":"test@ex.com"}`)))
    router.ServeHTTP(w, req)
    assert.Equal(t, 201, w.Code)
}
```

### Test d'intégration avec Postman
```json
{
  "info": {"name": "Veza API Test"},
  "item": [
    {
      "name": "Login",
      "request": {
        "method": "POST",
        "url": "{{baseUrl}}/auth/login",
        "body": {"mode": "raw", "raw": "{\"email\":\"user@ex.com\",\"password\":\"pass\"}"}
      },
      "response": []
    }
  ]
}
```

## ✅ Bonnes Pratiques
- Automatiser les tests (CI/CD)
- Couvrir les cas d'erreur et limites
- Utiliser des jeux de données isolés
- Vérifier la conformité OpenAPI ([documentation-standards.md](./documentation-standards.md))
- Monitorer la couverture de code

## ⚠️ Pièges à Éviter
- Tester uniquement les cas nominaux
- Oublier les tests de sécurité (XSS, injection)
- Ne pas nettoyer les données de test
- Laisser des tokens de test en production

## 🔗 Liens Utiles
- [debugging.md](./debugging.md)
- [performance-profiling.md](./performance-profiling.md)
- [security-guidelines.md](./security-guidelines.md)

---

**Dernière mise à jour** : $(date)
**Version** : 1.0.0 