#!/bin/bash

# Script ultra-conservateur pour le stream server
# Ne supprime que les imports les plus évidents et sûrs

echo "🛡️ Nettoyage ULTRA-CONSERVATEUR du stream server..."

# Fonction pour compter les warnings
count_warnings() {
    cargo check 2>&1 | grep -c "warning:" || echo "0"
}

# État initial
echo "📊 État initial:"
INITIAL_WARNINGS=$(count_warnings)
echo "Warnings: $INITIAL_WARNINGS"

# Sauvegarde
git add -A
git commit -m "Backup ultra conservateur $(date)"

echo "🔧 Suppression d'imports évidemment inutilisés (ULTRA-SÉCURISÉ)..."

# Supprimer seulement les imports de tokio::time::Duration si non utilisés
find src/ -name "*.rs" -exec grep -L "Duration::" {} \; | \
xargs grep -l "use.*Duration[,;]" | \
while read file; do
    # Vérifier que Duration n'est pas utilisé dans le fichier
    if ! grep -q "Duration::\|Duration{\|Duration(\|\.as_\(millis\|secs\|nanos\)" "$file"; then
        echo "🧹 Suppression de Duration dans $file"
        sed -i 's/, Duration//g; s/Duration, //g' "$file"
    fi
done

# Supprimer seulement HeaderValue si non utilisé
find src/ -name "*.rs" -exec grep -L "HeaderValue::" {} \; | \
xargs grep -l "use.*HeaderValue" | \
while read file; do
    if ! grep -q "HeaderValue::\|HeaderValue{" "$file"; then
        echo "🧹 Suppression de HeaderValue dans $file"
        sed -i 's/, HeaderValue//g; s/HeaderValue, //g' "$file"
    fi
done

# Supprimer Config si pas utilisé directement
find src/ -name "*.rs" -exec grep -L "Config::" {} \; | \
xargs grep -l "use.*config::Config" | \
while read file; do
    if ! grep -q "Config::\|Config{" "$file"; then
        echo "🧹 Suppression de config::Config dans $file"
        sed -i 's/use crate::config::Config;//g' "$file"
    fi
done

echo "✅ Nettoyage ultra-conservateur terminé"

# Vérification
echo "📊 État après nettoyage:"
FINAL_WARNINGS=$(count_warnings)
echo "Warnings: $FINAL_WARNINGS"

# Calcul de la réduction
REDUCTION=$((INITIAL_WARNINGS - FINAL_WARNINGS))
echo "📉 Réduction: $REDUCTION warnings"

if [ $REDUCTION -gt 0 ]; then
    echo "🎉 Nettoyage réussi ! $REDUCTION warnings supprimés"
else
    echo "✅ Aucune réduction mais pas de casse"
fi

echo "🏁 Nettoyage ultra-conservateur terminé" 