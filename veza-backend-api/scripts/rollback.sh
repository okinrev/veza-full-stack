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