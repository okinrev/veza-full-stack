#!/bin/bash

# veza-chat-test.sh
# Script d'administration complet pour le projet Veza
# Nettoyage, setup, déploiement, test, ouverture du chat, logs, status, etc.

set -e

# Couleurs
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
BOLD='\033[1m'
NC='\033[0m'

WORKSPACE_DIR="$(cd "$(dirname "$0")/.." && pwd)"
SCRIPTS_DIR="$WORKSPACE_DIR/scripts"
LOG_FILE="$WORKSPACE_DIR/veza-admin.log"
CHAT_URL="http://10.5.191.133/chat"

log() { echo -e "${BLUE}[$(date '+%H:%M:%S')]${NC} $1" | tee -a "$LOG_FILE"; }
success() { echo -e "${GREEN}✅ $1${NC}" | tee -a "$LOG_FILE"; }
warning() { echo -e "${YELLOW}⚠️  $1${NC}" | tee -a "$LOG_FILE"; }
error() { echo -e "${RED}❌ $1${NC}" | tee -a "$LOG_FILE"; exit 1; }
header() { echo -e "\n${PURPLE}${BOLD}==== $1 ====\n${NC}" | tee -a "$LOG_FILE"; }

# Nettoyage complet
task_clean() {
  header "Nettoyage complet de l'environnement"
  bash "$SCRIPTS_DIR/veza-deploy.sh" clean --force || warning "Nettoyage partiel."
  bash "$SCRIPTS_DIR/incus-clean.sh" || true
  rm -rf "$WORKSPACE_DIR/logs" "$WORKSPACE_DIR/data" "$WORKSPACE_DIR/test-results"
  success "Nettoyage terminé."
}

# Setup infra
setup_infra() {
  header "Configuration initiale de l'infrastructure"
  bash "$SCRIPTS_DIR/incus-setup.sh"
  success "Setup infra OK."
}

# Déploiement complet
task_deploy() {
  header "Déploiement complet de l'infrastructure et des apps"
  bash "$SCRIPTS_DIR/incus-deploy.sh"
  bash "$SCRIPTS_DIR/deploy-all.sh"
  success "Déploiement terminé."
}

# Vérification du statut
task_status() {
  header "Statut des containers et services"
  bash "$SCRIPTS_DIR/incus-status.sh"
  echo "\n--- CURL endpoints ---" | tee -a "$LOG_FILE"
  curl -s -o /dev/null -w "Frontend: %{http_code}\n" "$CHAT_URL" || warning "Frontend KO"
  curl -s "$CHAT_URL-api/health" | head -1 || warning "Chat API KO"
  curl -s "$CHAT_URL/../api/v1/health" | head -1 || warning "Backend API KO"
  success "Statut vérifié."
}

# Lancer les tests
task_test() {
  header "Tests d'intégration et de santé"
  bash "$SCRIPTS_DIR/test-complete.sh"
  success "Tests terminés."
}

# Ouvrir le chat dans le navigateur
open_chat() {
  header "Ouverture de la page de chat dans le navigateur"
  if command -v xdg-open &> /dev/null; then
    xdg-open "$CHAT_URL"
    success "Chat ouvert dans le navigateur."
  else
    warning "Impossible d'ouvrir automatiquement le navigateur. Ouvre : $CHAT_URL"
  fi
}

# Logs récents
task_logs() {
  header "Logs récents des services principaux"
  for c in veza-backend veza-chat veza-frontend veza-haproxy; do
    echo -e "\n--- Logs $c ---" | tee -a "$LOG_FILE"
    incus exec $c -- tail -n 30 /var/log/* 2>/dev/null || echo "Pas de logs pour $c"
  done
}

# Menu principal
case "$1" in
  clean)
    task_clean
    ;;
  setup)
    setup_infra
    ;;
  deploy)
    task_deploy
    ;;
  status)
    task_status
    ;;
  test)
    task_test
    ;;
  logs)
    task_logs
    ;;
  open)
    open_chat
    ;;
  all|"")
    task_clean
    setup_infra
    task_deploy
    task_status
    task_test
    open_chat
    task_logs
    header "✅ Administration complète terminée."
    ;;
  *)
    echo -e "${CYAN}Usage: $0 [all|clean|setup|deploy|status|test|logs|open]${NC}"
    ;;
esac
