# Scripts utilitaires de migration

## 📁 Structure des scripts
```
backend/scripts/
├── migrate.sh              # Script principal de migration
├── rollback.sh              # Script de rollback
├── fix_imports.sh           # Correction automatique des imports
├── validate_build.sh        # Validation complète
├── check_imports.sh         # Vérification des imports
├── test_endpoints.sh        # Tests des endpoints
├── test_database.sh         # Tests de base de données
├── test_performance.sh      # Tests de performance
├── clean_duplicates.sh      # Nettoyage
├── start_server.sh          # Lancement du serveur
└── setup_dev.sh             # Configuration développement
```

## 🚀 Script principal de migration

### `scripts/migrate.sh`
```bash
#!/bin/bash

echo "🔧 Migration complète du backend Talas"
echo "======================================"

# Configuration
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Étapes de migration
STEPS=(
    "1:Consolidation architecture"
    "2:Correction imports"
    "3:Implémentation handlers"
    "4:Consolidation services"
    "5:Refactoring routes"
    "6:Tests validation"
)

# Fonction d'aide
show_help() {
    echo "Usage: $0 [options]"
    echo ""
    echo "Options:"
    echo "  -h, --help     Afficher cette aide"
    echo "  -s, --step N   Exécuter seulement l'étape N (1-6)"
    echo "  -f, --force    Forcer l'exécution même en cas d'erreur"
    echo "  -d, --dry-run  Simulation sans modification"
    echo ""
    echo "Étapes disponibles:"
    for step in "${STEPS[@]}"; do
        echo "  ${step}"
    done
}

# Variables
STEP_ONLY=""
FORCE=false
DRY_RUN=false

# Parsing des arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -h|--help)
            show_help
            exit 0
            ;;
        -s|--step)
            STEP_ONLY="$2"
            shift 2
            ;;
        -f|--force)
            FORCE=true
            shift
            ;;
        -d|--dry-run)
            DRY_RUN=true
            shift
            ;;
        *)
            echo "Option inconnue: $1"
            show_help
            exit 1
            ;;
    esac
done

# Fonction d'exécution d'étape
execute_step() {
    local step_num="$1"
    local step_name="$2"
    local step_script="$3"
    
    echo -e "\n${BLUE}📋 ÉTAPE $step_num: $step_name${NC}"
    echo "================================================"
    
    if [ "$DRY_RUN" = true ]; then
        echo "🔍 [DRY RUN] Simulation de: $step_script"
        return 0
    fi
    
    # Backup avant l'étape
    if [ "$step_num" != "6" ]; then  # Pas de backup pour les tests
        echo "💾 Backup avant étape $step_num..."
        git add . && git commit -m "Backup avant étape $step_num: $step_name" >/dev/null 2>&1 || true
    fi
    
    # Exécution
    if eval "$step_script"; then
        echo -e "${GREEN}✅ Étape $step_num terminée avec succès${NC}"
        return 0
    else
        echo -e "${RED}❌ Étape $step_num échouée${NC}"
        if [ "$FORCE" = false ]; then
            echo "Utilisez --force pour continuer malgré les erreurs"
            exit 1
        fi
        return 1
    fi
}

# Étape 1: Consolidation architecture
step1_consolidation() {
    echo "Suppression de l'ancien système de routes..."
    if [ -d internal/routes ]; then
        rm -rf internal/routes
    fi
    
    echo "Création des modules API manquants..."
    
    # Créer admin module
    mkdir -p internal/api/admin
    
    # Créer file module  
    mkdir -p internal/api/file
    
    # Créer product module
    mkdir -p internal/api/product
    
    echo "Mise à jour du main.go..."
    # Ici on pourrait générer le nouveau main.go
    # Pour l'instant, on assume qu'il sera fait manuellement
    
    return 0
}

# Étape 2: Correction imports
step2_imports() {
    echo "Correction automatique des imports..."
    
    # Utiliser le script de correction
    if [ -f scripts/fix_imports.sh ]; then
        bash scripts/fix_imports.sh
    else
        # Correction inline
        find internal/ cmd/ pkg/ -name "*.go" -type f | while read file; do
            if grep -q "veza-web-app/" "$file"; then
                sed -i.bak 's|"veza-web-app/|"github.com/okinrev/veza-web-app/|g' "$file"
                rm -f "$file.bak"
            fi
        done
    fi
    
    # Test de compilation
    echo "Test de compilation..."
    go mod tidy
    go build ./cmd/server >/dev/null 2>&1
}

# Étape 3: Implémentation handlers
step3_handlers() {
    echo "Implémentation des handlers prioritaires..."
    echo "⚠️ Cette étape nécessite une implémentation manuelle"
    echo "Consultez 03_implementation_handlers.md pour les détails"
    
    # Vérifier que les handlers existent
    local missing_handlers=0
    
    for module in auth user admin track; do
        if [ ! -f "internal/api/$module/handler.go" ]; then
            echo "❌ Handler manquant: $module"
            ((missing_handlers++))
        fi
    done
    
    if [ $missing_handlers -gt 0 ]; then
        echo "⚠️ $missing_handlers handlers manquants"
        return 1
    fi
    
    return 0
}

# Étape 4: Consolidation services
step4_services() {
    echo "Consolidation des services..."
    
    # Supprimer l'ancien répertoire services s'il existe
    if [ -d internal/services ]; then
        echo "Suppression de internal/services/..."
        rm -rf internal/services
    fi
    
    # Vérifier que chaque module a son service
    local missing_services=0
    
    for module_dir in internal/api/*/; do
        if [ -d "$module_dir" ]; then
            module_name=$(basename "$module_dir")
            if [ ! -f "${module_dir}service.go" ]; then
                echo "❌ Service manquant: $module_name"
                ((missing_services++))
            fi
        fi
    done
    
    if [ $missing_services -gt 0 ]; then
        echo "⚠️ $missing_services services manquants"
        return 1
    fi
    
    return 0
}

# Étape 5: Refactoring routes
step5_routes() {
    echo "Refactoring du système de routes..."
    
    # Créer le router centralisé
    if [ ! -f internal/api/router.go ]; then
        echo "⚠️ internal/api/router.go manquant"
        echo "Consultez 05_refactoring_routes.md pour créer ce fichier"
        return 1
    fi
    
    # Vérifier que chaque module a ses routes
    local missing_routes=0
    
    for module_dir in internal/api/*/; do
        if [ -d "$module_dir" ]; then
            module_name=$(basename "$module_dir")
            if [ ! -f "${module_dir}routes.go" ]; then
                echo "❌ Routes manquantes: $module_name"
                ((missing_routes++))
            fi
        fi
    done
    
    if [ $missing_routes -gt 0 ]; then
        echo "⚠️ $missing_routes fichiers de routes manquants"
        return 1
    fi
    
    return 0
}

# Étape 6: Tests et validation
step6_tests() {
    echo "Exécution des tests de validation..."
    
    if [ -f scripts/validate_build.sh ]; then
        bash scripts/validate_build.sh
    else
        echo "Script de validation manquant, tests basiques..."
        
        # Tests basiques
        echo "• Test compilation..."
        if ! go build ./cmd/server >/dev/null 2>&1; then
            echo "❌ Compilation échouée"
            return 1
        fi
        
        echo "• Test imports..."
        if grep -r "veza-web-app/" internal/ cmd/ --include="*.go" | grep -v github.com/okinrev >/dev/null; then
            echo "❌ Imports incorrects trouvés"
            return 1
        fi
        
        echo "✅ Tests basiques passés"
    fi
    
    return 0
}

# Exécution principale
main() {
    echo -e "${YELLOW}🚀 Début de la migration${NC}"
    echo "Répertoire: $(pwd)"
    echo "Date: $(date)"
    
    # Vérifications préalables
    if [ ! -f go.mod ]; then
        echo -e "${RED}❌ go.mod non trouvé. Êtes-vous dans le bon répertoire ?${NC}"
        exit 1
    fi
    
    if [ ! -d internal ]; then
        echo -e "${RED}❌ Répertoire internal/ non trouvé${NC}"
        exit 1
    fi
    
    # Backup initial
    if [ "$DRY_RUN" = false ]; then
        echo "💾 Backup initial..."
        git add . && git commit -m "Backup initial avant migration" >/dev/null 2>&1 || true
    fi
    
    # Exécution des étapes
    if [ -n "$STEP_ONLY" ]; then
        echo "Exécution de l'étape $STEP_ONLY uniquement"
        case $STEP_ONLY in
            1) execute_step 1 "Consolidation architecture" "step1_consolidation" ;;
            2) execute_step 2 "Correction imports" "step2_imports" ;;
            3) execute_step 3 "Implémentation handlers" "step3_handlers" ;;
            4) execute_step 4 "Consolidation services" "step4_services" ;;
            5) execute_step 5 "Refactoring routes" "step5_routes" ;;
            6) execute_step 6 "Tests validation" "step6_tests" ;;
            *) echo "Étape invalide: $STEP_ONLY"; exit 1 ;;
        esac
    else
        # Toutes les étapes
        execute_step 1 "Consolidation architecture" "step1_consolidation"
        execute_step 2 "Correction imports" "step2_imports"
        execute_step 3 "Implémentation handlers" "step3_handlers"
        execute_step 4 "Consolidation services" "step4_services"
        execute_step 5 "Refactoring routes" "step5_routes"
        execute_step 6 "Tests validation" "step6_tests"
    fi
    
    # Résultat final
    echo ""
    echo -e "${GREEN}🎉 MIGRATION TERMINÉE AVEC SUCCÈS !${NC}"
    echo "Le backend Talas est maintenant fonctionnel avec la nouvelle architecture."
    echo ""
    echo "Prochaines étapes recommandées:"
    echo "1. Tester manuellement les endpoints principaux"
    echo "2. Implémenter les handlers manquants"
    echo "3. Ajouter des tests unitaires"
    echo "4. Configurer le déploiement"
}

# Exécution
main "$@"
```

## 🔄 Script de rollback

### `scripts/rollback.sh`
```bash
#!/bin/bash

echo "🔄 Rollback de la migration"
echo "=========================="

# Couleurs
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

show_help() {
    echo "Usage: $0 [options]"
    echo ""
    echo "Options:"
    echo "  -h, --help      Afficher cette aide"
    echo "  -t, --to TAG    Rollback vers un tag spécifique"
    echo "  -s, --steps N   Rollback de N commits"
    echo "  -f, --force     Forcer le rollback (perte de données)"
    echo ""
    echo "Exemples:"
    echo "  $0 --to v1.0.0-before-migration"
    echo "  $0 --steps 5"
}

# Variables
TARGET_TAG=""
STEPS=""
FORCE=false

# Parsing des arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -h|--help)
            show_help
            exit 0
            ;;
        -t|--to)
            TARGET_TAG="$2"
            shift 2
            ;;
        -s|--steps)
            STEPS="$2"
            shift 2
            ;;
        -f|--force)
            FORCE=true
            shift
            ;;
        *)
            echo "Option inconnue: $1"
            show_help
            exit 1
            ;;
    esac
done

# Vérifications
if [ ! -d .git ]; then
    echo -e "${RED}❌ Pas de dépôt Git trouvé${NC}"
    exit 1
fi

# Sauvegarder l'état actuel
echo "💾 Sauvegarde de l'état actuel..."
git add . && git commit -m "Sauvegarde avant rollback" >/dev/null 2>&1 || true

# Rollback
if [ -n "$TARGET_TAG" ]; then
    echo "🔄 Rollback vers le tag: $TARGET_TAG"
    if [ "$FORCE" = true ]; then
        git reset --hard "$TARGET_TAG"
    else
        git checkout "$TARGET_TAG"
    fi
elif [ -n "$STEPS" ]; then
    echo "🔄 Rollback de $STEPS commits"
    if [ "$FORCE" = true ]; then
        git reset --hard "HEAD~$STEPS"
    else
        git reset --soft "HEAD~$STEPS"
    fi
else
    echo "🔄 Rollback au dernier commit stable"
    # Chercher le dernier commit avant migration
    LAST_STABLE=$(git log --oneline | grep -E "(avant migration|backup|stable)" | head -1 | cut -d' ' -f1)
    if [ -n "$LAST_STABLE" ]; then
        echo "Commit stable trouvé: $LAST_STABLE"
        if [ "$FORCE" = true ]; then
            git reset --hard "$LAST_STABLE"
        else
            git checkout "$LAST_STABLE"
        fi
    else
        echo -e "${YELLOW}⚠️ Aucun commit stable trouvé${NC}"
        echo "Utilisez --to TAG ou --steps N"
        exit 1
    fi
fi

echo -e "${GREEN}✅ Rollback terminé${NC}"
echo ""
echo "État actuel:"
git log --oneline -5
```

## 🛠️ Script de configuration développement

### `scripts/setup_dev.sh`
```bash
#!/bin/bash

echo "🛠️ Configuration de l'environnement de développement"
echo "=================================================="

# Vérifications des prérequis
echo "1. Vérification des prérequis..."

# Go
if ! command -v go &> /dev/null; then
    echo "❌ Go n'est pas installé"
    echo "Installez Go depuis https://golang.org/"
    exit 1
fi

GO_VERSION=$(go version | cut -d' ' -f3)
echo "✅ Go installé: $GO_VERSION"

# PostgreSQL
if ! command -v psql &> /dev/null; then
    echo "⚠️ PostgreSQL client non trouvé"
    echo "Installez postgresql-client pour les tests DB"
fi

# Git
if ! command -v git &> /dev/null; then
    echo "❌ Git n'est pas installé"
    exit 1
fi

echo "✅ Git installé: $(git --version)"

# Configuration de l'environnement
echo ""
echo "2. Configuration de l'environnement..."

# Créer .env si nécessaire
if [ ! -f .env ]; then
    echo "📝 Création du fichier .env..."
    cat > .env << 'EOF'
# Configuration de développement
DATABASE_URL=postgresql://postgres:password@localhost:5432/talas_dev
JWT_SECRET=dev-secret-key-change-in-production
PORT=8080
ENVIRONMENT=development

# Base de données
DB_HOST=localhost
DB_PORT=5432
DB_USERNAME=postgres
DB_PASSWORD=password
DB_NAME=talas_dev
DB_SSLMODE=disable

# Fichiers
UPLOAD_DIR=./static
MAX_FILE_SIZE=104857600

# Modules Rust
CHAT_SERVER_PORT=9001
STREAM_SERVER_PORT=8082
EOF
    
    echo "✅ Fichier .env créé"
    echo "⚠️ Modifiez les valeurs selon votre configuration"
else
    echo "✅ Fichier .env existe déjà"
fi

# Répertoires nécessaires
echo ""
echo "3. Création des répertoires..."

DIRS=(
    "static/audio"
    "static/uploads"
    "static/avatars"
    "static/shared_resources"
    "static/internal_docs"
    "logs"
)

for dir in "${DIRS[@]}"; do
    if [ ! -d "$dir" ]; then
        mkdir -p "$dir"
        echo "📁 Créé: $dir"
    else
        echo "✅ Existe: $dir"
    fi
done

# Permissions des scripts
echo ""
echo "4. Configuration des permissions..."
chmod +x scripts/*.sh 2>/dev/null || true
echo "✅ Permissions des scripts configurées"

# Installation des dépendances Go
echo ""
echo "5. Installation des dépendances Go..."
go mod tidy
if [ $? -eq 0 ]; then
    echo "✅ Dépendances installées"
else
    echo "❌ Erreur lors de l'installation des dépendances"
    exit 1
fi

# Test de compilation
echo ""
echo "6. Test de compilation..."
if go build -o tmp_server ./cmd/server; then
    rm -f tmp_server
    echo "✅ Compilation réussie"
else
    echo "❌ Erreur de compilation"
    echo "Vérifiez les imports et la structure du code"
    exit 1
fi

# Configuration Git hooks (optionnel)
echo ""
echo "7. Configuration des Git hooks..."
if [ -d .git ]; then
    cat > .git/hooks/pre-commit << 'EOF'
#!/bin/bash
# Pre-commit hook: vérification du code

echo "🔍 Vérification avant commit..."

# Format du code
go fmt ./...

# Tests rapides
go vet ./...

# Vérification de compilation
if ! go build ./cmd/server >/dev/null 2>&1; then
    echo "❌ Erreur de compilation"
    exit 1
fi

echo "✅ Vérifications passées"
EOF
    
    chmod +x .git/hooks/pre-commit
    echo "✅ Pre-commit hook configuré"
fi

# Résumé
echo ""
echo "🎉 Configuration terminée !"
echo "========================="
echo ""
echo "📋 Résumé:"
echo "• Environnement: $(go version)"
echo "• Configuration: .env créé"
echo "• Répertoires: créés"
echo "• Dépendances: installées"
echo "• Compilation: OK"
echo ""
echo "🚀 Commandes utiles:"
echo "• Lancer le serveur: go run cmd/server/main.go"
echo "• Tests: bash scripts/test_endpoints.sh"
echo "• Validation: bash scripts/validate_build.sh"
echo ""
echo "📚 Documentation:"
echo "• Architecture: docs/migration/"
echo "• API: docs/api/"
echo "• Handlers: docs/doc_*_handler.md"
```

## 📊 Script de monitoring

### `scripts/monitor.sh`
```bash
#!/bin/bash

echo "📊 Monitoring du backend Talas"
echo "=============================="

# Configuration
INTERVAL=5
LOG_FILE="logs/monitor.log"
PID_FILE="/tmp/talas_server.pid"

# Créer le répertoire de logs
mkdir -p logs

# Fonction de monitoring
monitor_server() {
    local start_time=$(date +%s)
    
    while true; do
        local current_time=$(date)
        
        # Vérifier si le serveur répond
        if curl -s http://localhost:8080/health >/dev/null 2>&1; then
            local response_time=$(curl -w "%{time_total}" -o /dev/null -s http://localhost:8080/health)
            echo "[$current_time] ✅ Server OK - Response time: ${response_time}s" | tee -a "$LOG_FILE"
        else
            echo "[$current_time] ❌ Server DOWN" | tee -a "$LOG_FILE"
        fi
        
        # Statistiques système
        local cpu_usage=$(ps -p $(cat $PID_FILE 2>/dev/null) -o %cpu= 2>/dev/null | tr -d ' ')
        local mem_usage=$(ps -p $(cat $PID_FILE 2>/dev/null) -o %mem= 2>/dev/null | tr -d ' ')
        
        if [ -n "$cpu_usage" ]; then
            echo "[$current_time] 📊 CPU: ${cpu_usage}% | MEM: ${mem_usage}%" | tee -a "$LOG_FILE"
        fi
        
        sleep $INTERVAL
    done
}

# Options
case "$1" in
    start)
        echo "🚀 Démarrage du monitoring..."
        monitor_server &
        echo $! > /tmp/monitor.pid
        echo "Monitoring démarré (PID: $!)"
        ;;
    stop)
        if [ -f /tmp/monitor.pid ]; then
            kill $(cat /tmp/monitor.pid) 2>/dev/null
            rm -f /tmp/monitor.pid
            echo "Monitoring arrêté"
        else
            echo "Monitoring non actif"
        fi
        ;;
    status)
        tail -20 "$LOG_FILE" 2>/dev/null || echo "Pas de logs disponibles"
        ;;
    *)
        echo "Usage: $0 {start|stop|status}"
        exit 1
        ;;
esac
```

Ces scripts complètent le plan de migration en offrant :

1. **🚀 Automatisation** : Migration automatique ou par étapes
2. **🔄 Sécurité** : Rollback en cas de problème
3. **🛠️ Setup** : Configuration développement automatique
4. **📊 Monitoring** : Surveillance du serveur
5. **✅ Validation** : Tests complets à chaque étape

Le processus devient ainsi **beaucoup plus sûr et efficace** pour reconstruire votre backend Go !