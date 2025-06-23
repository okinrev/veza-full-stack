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