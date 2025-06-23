# Étape 4 - Consolidation des services

## 🎯 Objectif
Éliminer les doublons de services et établir une architecture de services cohérente.

## ⏱️ Durée estimée : 30-45 minutes

## 🚨 Problèmes à résoudre
- Doublons entre `internal/services/` et `internal/api/*/service.go`
- Services partiellement implémentés
- Dépendances incohérentes
- Logique métier dispersée

## 📊 Audit des services existants

### Services dans `internal/services/`
```bash
ls -la internal/services/
# Trouvé :
- auth_service.go          # ✅ Complet mais dupliqué
- chat_service.go          # ❌ Interface seulement
- file_service.go          # ❌ Routes dans mauvais fichier
- listing_service.go       # ❌ Interface seulement
- offer_service.go         # ❌ Interface seulement
- product_service.go       # ❌ Interface seulement
- room_service.go          # ❌ Routes dans mauvais fichier
- search_service.go        # ❌ Interface seulement
- tag_service.go           # ❌ Routes dans mauvais fichier
- track_service.go         # ✅ Complet
- user_service.go          # ✅ Complet
```

### Services dans `internal/api/*/`
```bash
find internal/api/ -name "service.go"
# Trouvé :
- internal/api/user/service.go       # ✅ Implémenté
- internal/api/admin/service.go      # ✅ Partiellement implémenté
```

## 📋 Plan de consolidation

### Phase 4.1 : Migration des services complets
1. Migrer `auth_service.go` → `internal/api/auth/service.go`
2. Migrer `track_service.go` → `internal/api/track/service.go`  
3. Migrer `user_service.go` → finaliser `internal/api/user/service.go`

### Phase 4.2 : Nettoyage des fichiers incorrects
4. Corriger les services mal placés (`file_service.go`, `room_service.go`, etc.)
5. Supprimer les interfaces vides

### Phase 4.3 : Création des services manquants
6. Créer les services pour tous les modules API

## 🔧 Implémentation détaillée

### Phase 4.1 : Migration des services complets

#### Migrer `auth_service.go` → `internal/api/auth/service.go`

**L'auth service a déjà été implémenté dans l'étape 3, vérifier qu'il est complet :**

```bash
# Comparer le contenu
diff internal/services/auth_service.go internal/api/auth/service.go
```

**Si différences, consolider vers `internal/api/auth/service.go` :**
```go
// Prendre la version la plus complète et ajouter les méthodes manquantes
```

#### Migrer `track_service.go` → `internal/api/track/service.go`

**Contenu de `internal/api/track/service.go` :**
```go
package track

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/lib/pq"
	"github.com/okinrev/veza-web-app/internal/database"
	"github.com/okinrev/veza-web-app/internal/models"
	"github.com/okinrev/veza-web-app/internal/utils"
)

type Service struct {
	db        *database.DB
	jwtSecret string
}

func NewService(db *database.DB, jwtSecret string) *Service {
	return &Service{
		db:        db,
		jwtSecret: jwtSecret,
	}
}

const (
	MaxAudioSize     = 100 << 20 // 100MB
	MaxAudioDuration = 600       // 10 minutes
)

// CreateTrack crée un nouveau track (basé sur track_service.go)
func (s *Service) CreateTrack(req CreateTrackRequest) (*models.Track, error) {
	// Validation audio
	if err := s.ValidateAudioFile(req.Filename, 0); err != nil {
		return nil, fmt.Errorf("invalid audio file: %w", err)
	}

	var track models.Track
	err := s.db.QueryRow(`
		INSERT INTO tracks (title, artist, filename, duration_seconds, tags, is_public, uploader_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW())
		RETURNING id, title, artist, filename, duration_seconds, tags, is_public, uploader_id, created_at, updated_at
	`, req.Title, req.Artist, req.Filename, req.DurationSeconds, pq.Array(req.Tags), req.IsPublic, req.UploaderID).Scan(
		&track.ID, &track.Title, &track.Artist, &track.Filename,
		&track.DurationSeconds, pq.Array(&track.Tags), &track.IsPublic,
		&track.UploaderID, &track.CreatedAt, &track.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create track: %w", err)
	}

	return &track, nil
}

// GetTrack récupère un track avec contrôle d'accès
func (s *Service) GetTrack(trackID, userID int) (*models.Track, error) {
	var track models.Track
	err := s.db.QueryRow(`
		SELECT t.id, t.title, t.artist, t.filename, t.duration_seconds, t.tags, 
		       t.is_public, t.uploader_id, t.created_at, t.updated_at
		FROM tracks t
		WHERE t.id = $1 AND (t.is_public = true OR t.uploader_id = $2)
	`, trackID, userID).Scan(
		&track.ID, &track.Title, &track.Artist, &track.Filename,
		&track.DurationSeconds, pq.Array(&track.Tags), &track.IsPublic,
		&track.UploaderID, &track.CreatedAt, &track.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("track not found: %w", err)
	}

	return &track, nil
}

// ListTracks liste les tracks avec pagination
func (s *Service) ListTracks(page, limit int, showPrivate bool, userID int) ([]models.Track, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	offset := (page - 1) * limit

	baseQuery := `
		SELECT t.id, t.title, t.artist, t.filename, t.duration_seconds, t.tags, 
		       t.is_public, t.uploader_id, t.created_at, t.updated_at
		FROM tracks t
	`
	countQuery := `SELECT COUNT(*) FROM tracks t`

	whereClause := ""
	args := []interface{}{}

	if showPrivate && userID > 0 {
		whereClause = " WHERE t.uploader_id = $1"
		args = append(args, userID)
	} else {
		whereClause = " WHERE t.is_public = true"
	}

	// Total
	var total int
	err := s.db.QueryRow(countQuery+whereClause, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count tracks: %w", err)
	}

	// Données
	orderClause := " ORDER BY t.created_at DESC LIMIT $" + strconv.Itoa(len(args)+1) + " OFFSET $" + strconv.Itoa(len(args)+2)
	args = append(args, limit, offset)

	rows, err := s.db.Query(baseQuery+whereClause+orderClause, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to retrieve tracks: %w", err)
	}
	defer rows.Close()

	var tracks []models.Track
	for rows.Next() {
		var track models.Track
		err := rows.Scan(
			&track.ID, &track.Title, &track.Artist, &track.Filename,
			&track.DurationSeconds, pq.Array(&track.Tags), &track.IsPublic,
			&track.UploaderID, &track.CreatedAt, &track.UpdatedAt,
		)
		if err != nil {
			continue
		}
		tracks = append(tracks, track)
	}

	return tracks, total, nil
}

// ValidateAudioFile valide un fichier audio
func (s *Service) ValidateAudioFile(filename string, size int64) error {
	// Extensions autorisées
	ext := strings.ToLower(filepath.Ext(filename))
	allowedExts := []string{".mp3", ".wav", ".flac", ".ogg", ".m4a", ".aac"}
	
	validExt := false
	for _, allowedExt := range allowedExts {
		if ext == allowedExt {
			validExt = true
			break
		}
	}

	if !validExt {
		return fmt.Errorf("unsupported audio format: %s", ext)
	}

	// Taille
	if size > 0 && size > MaxAudioSize {
		return fmt.Errorf("file size exceeds maximum: %d bytes", MaxAudioSize)
	}

	return nil
}

// GenerateStreamURL génère une URL signée pour le streaming
func (s *Service) GenerateStreamURL(filename string, userID int) (string, error) {
	// Vérifier l'accès au track
	var trackExists, isPublic bool
	var uploaderID int
	err := s.db.QueryRow(`
		SELECT EXISTS(SELECT 1 FROM tracks WHERE filename = $1), is_public, uploader_id
		FROM tracks WHERE filename = $1
	`, filename).Scan(&trackExists, &isPublic, &uploaderID)

	if !trackExists {
		return "", fmt.Errorf("track not found")
	}

	if !isPublic && uploaderID != userID {
		return "", fmt.Errorf("access denied")
	}

	// Générer URL signée
	signedURL, err := utils.GenerateSignedURL(filename, userID, s.jwtSecret)
	if err != nil {
		return "", fmt.Errorf("failed to generate signed URL: %w", err)
	}

	return signedURL, nil
}

// Types de requêtes
type CreateTrackRequest struct {
	Title           string   `json:"title" binding:"required"`
	Artist          string   `json:"artist" binding:"required"`
	Filename        string   `json:"filename" binding:"required"`
	DurationSeconds *int     `json:"duration_seconds"`
	Tags            []string `json:"tags"`
	IsPublic        bool     `json:"is_public"`
	UploaderID      int      `json:"uploader_id" binding:"required"`
}

type UpdateTrackRequest struct {
	Title    *string   `json:"title,omitempty"`
	Artist   *string   `json:"artist,omitempty"`
	Tags     *[]string `json:"tags,omitempty"`
	IsPublic *bool     `json:"is_public,omitempty"`
}

// TODO: Implémenter UpdateTrack, DeleteTrack, SearchTracks basé sur track_service.go
```

#### Finaliser `internal/api/user/service.go`

**S'assurer que toutes les méthodes de `user_service.go` sont présentes :**
```go
// Ajouter les méthodes manquantes si nécessaire
// Basé sur le contenu de internal/services/user_service.go
```

### Phase 4.2 : Nettoyage des fichiers incorrects

#### Corriger `internal/services/file_service.go`
```bash
# Ce fichier contient des routes, pas un service
cat internal/services/file_service.go
```

**Supprimer et créer le vrai service :**
```bash
rm internal/services/file_service.go
```

**Créer `internal/api/file/service.go` :**
```go
package file

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/okinrev/veza-web-app/internal/database"
	"github.com/okinrev/veza-web-app/internal/models"
)

type Service struct {
	db *database.DB
}

func NewService(db *database.DB) *Service {
	return &Service{db: db}
}

const (
	MaxFileSize         = 10 << 20  // 10MB
	MaxInternalDocSize  = 50 << 20  // 50MB
)

// UploadFile upload un fichier pour un produit utilisateur
func (s *Service) UploadFile(userID, productID int, filename, fileType string, size int64) (*models.FileResponse, error) {
	// Vérifier propriété du produit
	var ownerID int
	err := s.db.QueryRow("SELECT user_id FROM user_products WHERE id = $1", productID).Scan(&ownerID)
	if err != nil {
		return nil, fmt.Errorf("product not found")
	}
	if ownerID != userID {
		return nil, fmt.Errorf("not authorized")
	}

	// Valider le fichier
	if err := s.validateFile(filename, fileType, size); err != nil {
		return nil, err
	}

	// Insérer en base
	var file models.FileResponse
	err = s.db.QueryRow(`
		INSERT INTO files (product_id, filename, url, type, size, uploaded_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
		RETURNING id, product_id, filename, url, type, size, uploaded_at, updated_at
	`, productID, filename, "/uploads/"+filename, fileType, size).Scan(
		&file.ID, &file.ProductID, &file.Filename, &file.URL, 
		&file.Type, &file.Size, &file.UploadedAt, &file.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to save file: %w", err)
	}

	return &file, nil
}

// validateFile valide un fichier selon son type
func (s *Service) validateFile(filename, fileType string, size int64) error {
	// Types autorisés par type
	allowedExtensions := map[string][]string{
		"manual":   {".pdf", ".doc", ".docx", ".txt"},
		"warranty": {".pdf", ".jpg", ".jpeg", ".png"},
		"invoice":  {".pdf", ".jpg", ".jpeg", ".png"},
		"image":    {".jpg", ".jpeg", ".png", ".gif", ".webp"},
		"document": {".pdf", ".doc", ".docx", ".txt", ".rtf"},
	}

	ext := strings.ToLower(filepath.Ext(filename))
	if allowed, exists := allowedExtensions[fileType]; exists {
		valid := false
		for _, allowedExt := range allowed {
			if ext == allowedExt {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("invalid file extension %s for type %s", ext, fileType)
		}
	}

	// Taille
	maxSize := int64(MaxFileSize)
	if fileType == "document" {
		maxSize = MaxInternalDocSize
	}
	
	if size > maxSize {
		return fmt.Errorf("file too large: %d bytes (max: %d)", size, maxSize)
	}

	return nil
}

// TODO: Autres méthodes basées sur doc_file_handler.md
```

#### Nettoyer les autres fichiers incorrects
```bash
# Supprimer les fichiers qui contiennent des routes au lieu de services
rm internal/services/room_service.go
rm internal/services/tag_service.go

# Supprimer les interfaces vides
rm internal/services/chat_service.go
rm internal/services/listing_service.go
rm internal/services/offer_service.go
rm internal/services/product_service.go
rm internal/services/search_service.go
```

### Phase 4.3 : Création des services manquants

#### Créer `internal/api/listing/service.go`
```go
package listing

import (
	"fmt"

	"github.com/lib/pq"
	"github.com/okinrev/veza-web-app/internal/database"
	"github.com/okinrev/veza-web-app/internal/models"
)

type Service struct {
	db *database.DB
}

func NewService(db *database.DB) *Service {
	return &Service{db: db}
}

// CreateListing crée un nouveau listing
func (s *Service) CreateListing(userID int, req CreateListingRequest) (*models.ListingResponse, error) {
	// Vérifier propriété du produit
	var ownerID int
	err := s.db.QueryRow("SELECT user_id FROM user_products WHERE id = $1", req.ProductID).Scan(&ownerID)
	if err != nil {
		return nil, fmt.Errorf("product not found")
	}
	if ownerID != userID {
		return nil, fmt.Errorf("not authorized")
	}

	// Créer le listing
	var listing models.ListingResponse
	err = s.db.QueryRow(`
		INSERT INTO listings (user_id, product_id, description, state, price, exchange_for, images, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, 'open', NOW(), NOW())
		RETURNING id, user_id, product_id, description, state, price, exchange_for, images, status, created_at, updated_at
	`, userID, req.ProductID, req.Description, req.State, req.Price, req.ExchangeFor, pq.Array(req.Images)).Scan(
		&listing.ID, &listing.UserID, &listing.ProductID, &listing.Description,
		&listing.State, &listing.Price, &listing.ExchangeFor, pq.Array(&listing.Images),
		&listing.Status, &listing.CreatedAt, &listing.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create listing: %w", err)
	}

	return &listing, nil
}

// TODO: Autres méthodes basées sur doc_listing_handler.md

type CreateListingRequest struct {
	ProductID   int      `json:"product_id" binding:"required"`
	Description string   `json:"description" binding:"required"`
	State       string   `json:"state" binding:"required"`
	Price       *int     `json:"price"`
	ExchangeFor *string  `json:"exchange_for"`
	Images      []string `json:"images"`
}
```

#### Template pour les autres services manquants

**Pattern à répéter pour :** `offer`, `message`, `room`, `search`, `tag`, `shared_resources`, `product`

```bash
# Créer la structure
mkdir -p internal/api/[module]

# Créer service.go
cat > internal/api/[module]/service.go << 'EOF'
package [module]

import (
	"github.com/okinrev/veza-web-app/internal/database"
)

type Service struct {
	db *database.DB
}

func NewService(db *database.DB) *Service {
	return &Service{db: db}
}

// TODO: Méthodes basées sur doc_[module]_handler.md
EOF
```

### Phase 4.4 : Suppression de l'ancien répertoire

```bash
# Une fois tous les services migrés
rm -rf internal/services/
```

### Phase 4.5 : Mise à jour des imports

**Mettre à jour `cmd/server/main.go` :**
```go
// Remplacer les imports vers internal/services/ par internal/api/*/
import (
	"github.com/okinrev/veza-web-app/internal/api/auth"
	"github.com/okinrev/veza-web-app/internal/api/user"
	"github.com/okinrev/veza-web-app/internal/api/admin"
	"github.com/okinrev/veza-web-app/internal/api/track"
	// etc.
)

func setupAPIRoutes(api *gin.RouterGroup, db *database.DB, jwtSecret string) {
	// Auth
	authService := auth.NewService(db, jwtSecret)
	authHandler := auth.NewHandler(authService)
	auth.SetupRoutes(api, authHandler, jwtSecret)

	// User
	userService := user.NewService(db)
	userHandler := user.NewHandler(userService)
	user.SetupRoutes(api, userHandler, jwtSecret)

	// Admin
	adminService := admin.NewService(db)
	adminHandler := admin.NewHandler(adminService)
	admin.SetupRoutes(api, adminHandler, jwtSecret)

	// Track
	trackService := track.NewService(db, jwtSecret)
	trackHandler := track.NewHandler(trackService)
	track.SetupRoutes(api, trackHandler, jwtSecret)

	// TODO: Autres modules
}
```

## ✅ Checklist de validation

```bash
# 1. Structure des services
find internal/api/ -name "service.go" | sort
# Attendu : service.go dans chaque module

# 2. Suppression ancien répertoire
ls internal/services/ 2>/dev/null || echo "✅ internal/services/ supprimé"

# 3. Compilation
go build ./cmd/server
echo $?
# Attendu : 0

# 4. Pas de doublons
grep -r "type.*Service struct" internal/
# Attendu : un seul service par module

# 5. Imports corrects
grep -r "internal/services/" internal/ cmd/
# Attendu : aucun résultat
```

## 🚨 Points d'attention

1. **Migration progressive** : Ne pas supprimer avant d'avoir migré
2. **Dépendances** : Vérifier que tous les imports sont mis à jour
3. **Tests** : Tester la compilation après chaque migration
4. **Sauvegarde** : Garder une copie des services avant suppression

## ⏭️ Étape suivante
Une fois les services consolidés → `05_refactoring_routes.md`

---

**💾 IMPORTANT** : Commit après cette étape
```bash
git add .
git commit -m "Étape 4: Consolidation services - architecture unifiée"
```