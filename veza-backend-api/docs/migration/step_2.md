# Étape 2 - Correction des imports

## 🎯 Objectif
Corriger tous les imports pour utiliser le bon module path et permettre la compilation.

## ⏱️ Durée estimée : 15-20 minutes

## 🚨 Problèmes à résoudre
- Imports mixtes : `veza-web-app/` vs `github.com/okinrev/veza-web-app/`
- Références vers des packages inexistants
- Cycles d'import potentiels

## 📋 Diagnostic initial

### Vérifier le module actuel
```bash
# Vérifier go.mod
head -3 go.mod
# Attendu : module github.com/okinrev/veza-web-app

# Lister tous les imports problématiques
grep -r "veza-web-app/" internal/ --include="*.go" | grep import
```

### Pattern des imports incorrects
```go
// ❌ Incorrect (trouvé dans le code)
import "veza-web-app/internal/utils/response"
import "veza-web-app/internal/common"
import "veza-web-app/internal/middleware"

// ✅ Correct 
import "github.com/okinrev/veza-web-app/internal/utils/response"
import "github.com/okinrev/veza-web-app/internal/common"
import "github.com/okinrev/veza-web-app/internal/middleware"
```

## 📋 Actions détaillées

### Action 2.1 : Script de correction automatique

**Créer le script : `scripts/fix_imports.sh`**
```bash
#!/bin/bash

echo "🔧 Correction des imports..."

# Fonction pour corriger les imports dans un fichier
fix_file() {
    local file="$1"
    echo "Correction: $file"
    
    # Remplacer les imports incorrects
    sed -i.bak 's|"veza-web-app/|"github.com/okinrev/veza-web-app/|g' "$file"
    
    # Supprimer le backup si la correction a réussi
    if [ $? -eq 0 ]; then
        rm -f "$file.bak"
    fi
}

# Trouver tous les fichiers Go et les corriger
find internal/ cmd/ pkg/ -name "*.go" -type f | while read file; do
    if grep -q "veza-web-app/" "$file"; then
        fix_file "$file"
    fi
done

echo "✅ Correction des imports terminée"
```

**Exécuter le script :**
```bash
chmod +x scripts/fix_imports.sh
./scripts/fix_imports.sh
```

### Action 2.2 : Corrections manuelles spécifiques

**Fichiers nécessitant une attention particulière :**

#### `internal/api/user/handler.go`
```go
// AVANT (incorrect)
import (
    "veza-web-app/internal/utils/response"
    "veza-web-app/internal/common"
)

// APRÈS (correct)
import (
    "net/http"
    "strconv"
    
    "github.com/gin-gonic/gin"
    "github.com/okinrev/veza-web-app/internal/common"
    "github.com/okinrev/veza-web-app/internal/utils/response"
)
```

#### `internal/api/*/handler.go` (tous les modules)
**Pattern de correction standard :**
```go
package [module]

import (
    "net/http"
    "strconv"
    
    "github.com/gin-gonic/gin"
    "github.com/okinrev/veza-web-app/internal/common"
    "github.com/okinrev/veza-web-app/internal/middleware"
    "github.com/okinrev/veza-web-app/internal/utils/response"
)
```

#### `internal/middleware/auth.go`
```go
package middleware

import (
    "net/http"
    "strings"

    "github.com/gin-gonic/gin"
    "github.com/okinrev/veza-web-app/internal/utils"
)
```

#### `cmd/server/main.go`
```go
package main

import (
    "log"
    "net/http"
    "os"
    "time"

    "github.com/gin-contrib/cors"
    "github.com/gin-gonic/gin"
    "github.com/joho/godotenv"
    
    "github.com/okinrev/veza-web-app/internal/api/admin"
    "github.com/okinrev/veza-web-app/internal/api/auth"
    "github.com/okinrev/veza-web-app/internal/api/user"
    "github.com/okinrev/veza-web-app/internal/database"
)
```

### Action 2.3 : Vérification et correction des dépendances

**Nettoyer le cache des modules :**
```bash
go clean -modcache
go mod tidy
```

**Vérifier les dépendances manquantes :**
```bash
go mod download
go list -m all
```

### Action 2.4 : Corrections spécifiques par fichier

#### Corriger `internal/api/user/service.go`
```go
package user

import (
    "fmt"
    "strconv"
    "strings"

    "github.com/okinrev/veza-web-app/internal/database"
    "github.com/okinrev/veza-web-app/internal/models"
    "github.com/okinrev/veza-web-app/internal/utils"
)
```

#### Corriger `internal/utils/auth.go`
```go
package utils

import (
    "crypto/hmac"
    "crypto/sha256"
    "encoding/hex"
    "fmt"
    "time"

    "github.com/golang-jwt/jwt/v5"
    "golang.org/x/crypto/bcrypt"
)
```

#### Corriger `internal/database/connection.go`
```go
package database

import (
    "database/sql"
    "fmt"
    "log"
    "io/ioutil"
    "path/filepath"
    "sort"
    "strings"

    _ "github.com/lib/pq"
)
```

### Action 2.5 : Test de compilation

**Test progressif :**
```bash
# 1. Vérifier les modules Go
go mod verify

# 2. Test de compilation par module
go build github.com/okinrev/veza-web-app/internal/database
go build github.com/okinrev/veza-web-app/internal/utils
go build github.com/okinrev/veza-web-app/internal/models
go build github.com/okinrev/veza-web-app/internal/middleware

# 3. Test compilation API modules
go build github.com/okinrev/veza-web-app/internal/api/user
go build github.com/okinrev/veza-web-app/internal/api/admin

# 4. Test compilation complète
go build ./cmd/server
```

### Action 2.6 : Corrections d'imports circulaires potentiels

**Vérifier les cycles :**
```bash
go list -json ./... | jq -r '.ImportPath + " imports " + (.Imports[]? // empty)'
```

**Pattern typique de cycle à éviter :**
```go
// ❌ Éviter : A imports B, B imports A
// internal/api/user/handler.go imports internal/services/user
// internal/services/user imports internal/api/user/models

// ✅ Correct : Dependencies claires
// handler -> service -> database
// models <- tous (sans cycles)
```

## 🔧 Script de validation complète

**Créer `scripts/validate_imports.sh` :**
```bash
#!/bin/bash

echo "🔍 Validation des imports..."

# Vérifier qu'aucun import incorrect ne reste
echo "1. Recherche d'imports incorrects:"
if grep -r "veza-web-app/" internal/ cmd/ pkg/ --include="*.go" | grep -v "github.com/okinrev"; then
    echo "❌ Imports incorrects trouvés"
    exit 1
else
    echo "✅ Tous les imports sont corrects"
fi

# Test de compilation
echo "2. Test de compilation:"
if go build ./cmd/server; then
    echo "✅ Compilation réussie"
else
    echo "❌ Erreurs de compilation"
    exit 1
fi

# Nettoyage
rm -f cmd/server/server

echo "3. Vérification des modules:"
go mod tidy
go mod verify

echo "✅ Validation terminée avec succès"
```

**Exécuter :**
```bash
chmod +x scripts/validate_imports.sh
./scripts/validate_imports.sh
```

## ✅ Checklist de validation

Après cette étape, vérifier :

```bash
# 1. Aucun import incorrect
grep -r "veza-web-app/" internal/ cmd/ --include="*.go" | grep -v github.com/okinrev
# Attendu : aucun résultat

# 2. Compilation réussie
go build ./cmd/server
echo $?
# Attendu : 0 (succès)

# 3. Modules propres
go mod tidy
go list -m all | head -5
# Attendu : liste des dépendances

# 4. Pas de fichiers .bak
find . -name "*.bak"
# Attendu : aucun résultat
```

## 🚨 Résolution de problèmes courants

### Erreur : "package not found"
```bash
# Solution : Vérifier le nom du module
go mod edit -module=github.com/okinrev/veza-web-app
go mod tidy
```

### Erreur : "import cycle not allowed"
```bash
# Solution : Réorganiser les imports
# Déplacer les types partagés vers internal/models/
# Éviter les imports bidirectionnels
```

### Erreur : "undefined function/type"
```bash
# Solution : Vérifier que le package est bien importé
# Ajouter les imports manquants
```

## ⏭️ Étape suivante
Une fois tous les imports corrigés et la compilation réussie → `03_implementation_handlers.md`

---

**💾 IMPORTANT** : Commit après cette étape
```bash
git add .
git commit -m "Étape 2: Correction imports - compilation OK"
```