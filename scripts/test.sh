#!/bin/bash

# ╭──────────────────────────────────────────────────────────────╮
# │               🧪 Veza - Tests de Déploiement                │
# │            Script de validation et de tests                 │
# ╰──────────────────────────────────────────────────────────────╯

set -e

# Couleurs pour les messages
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
BOLD='\033[1m'
NC='\033[0m'

# Fonctions utilitaires
log() { echo -e "${BLUE}[$(date '+%H:%M:%S')]${NC} $1"; }
success() { echo -e "${GREEN}✅ $1${NC}"; }
warning() { echo -e "${YELLOW}⚠️  $1${NC}"; }
error() { echo -e "${RED}❌ $1${NC}"; }
header() { 
    echo -e "${PURPLE}${BOLD}"
    echo "╭──────────────────────────────────────────────────────────────╮"
    echo "│ $1"
    echo "╰──────────────────────────────────────────────────────────────╯"
    echo -e "${NC}"
}

# Variables de test
CONTAINERS=("veza-postgres" "veza-redis" "veza-storage" "veza-backend" "veza-chat" "veza-stream" "veza-frontend" "veza-haproxy")
EXPECTED_IPS=(
    "veza-postgres:10.100.0.15"
    "veza-redis:10.100.0.17"
    "veza-storage:10.100.0.18"
    "veza-backend:10.100.0.12"
    "veza-chat:10.100.0.13"
    "veza-stream:10.100.0.14"
    "veza-frontend:10.100.0.11"
    "veza-haproxy:10.100.0.16"
)

# Compteurs de tests
TESTS_PASSED=0
TESTS_FAILED=0
TESTS_TOTAL=0

# Fonction pour exécuter un test
run_test() {
    local test_name="$1"
    local test_command="$2"
    
    ((TESTS_TOTAL++))
    
    log "Test: $test_name"
    
    if eval "$test_command" >/dev/null 2>&1; then
        success "$test_name"
        ((TESTS_PASSED++))
        return 0
    else
        error "$test_name"
        ((TESTS_FAILED++))
        return 1
    fi
}

# Test 1: Vérifier qu'Incus est installé et fonctionne
test_incus_installation() {
    header "🔧 Tests d'Infrastructure"
    
    run_test "Incus installé" "command -v incus"
    run_test "Incus accessible" "incus version"
    run_test "Réseau veza-network existe" "incus network show veza-network"
    run_test "Profils Veza créés" "incus profile list | grep -q veza"
}

# Test 2: Vérifier l'état des containers
test_containers_status() {
    header "📦 Tests des Containers"
    
    for container in "${CONTAINERS[@]}"; do
        run_test "Container $container existe" "incus info $container"
        run_test "Container $container en cours d'exécution" "incus list $container -c s | grep -q RUNNING"
    done
}

# Test 3: Vérifier les IPs des containers
test_containers_ips() {
    header "🌐 Tests des Adresses IP"
    
    for ip_pair in "${EXPECTED_IPS[@]}"; do
        container_name=$(echo "$ip_pair" | cut -d: -f1)
        expected_ip=$(echo "$ip_pair" | cut -d: -f2)
        
        run_test "IP de $container_name ($expected_ip)" "incus exec $container_name -- ip addr show eth0 | grep -q $expected_ip"
    done
}

# Test 4: Tester la connectivité réseau
test_network_connectivity() {
    header "🔗 Tests de Connectivité Réseau"
    
    # Test ping entre containers
    run_test "Backend peut ping PostgreSQL" "incus exec veza-backend -- ping -c 1 10.100.0.15"
    run_test "Backend peut ping Redis" "incus exec veza-backend -- ping -c 1 10.100.0.17"
    run_test "Frontend peut ping Backend" "incus exec veza-frontend -- ping -c 1 10.100.0.12"
}

# Test 5: Tester les services
test_services() {
    header "🚀 Tests des Services"
    
    # PostgreSQL
    run_test "PostgreSQL écoute sur le port 5432" "incus exec veza-postgres -- netstat -ln | grep :5432"
    run_test "Base de données veza_db existe" "incus exec veza-postgres -- sudo -u postgres psql -l | grep -q veza_db"
    
    # Redis
    run_test "Redis écoute sur le port 6379" "incus exec veza-redis -- netstat -ln | grep :6379"
    run_test "Redis répond aux commandes" "incus exec veza-redis -- redis-cli ping | grep -q PONG"
    
    # NFS Storage
    run_test "NFS Server actif" "incus exec veza-storage -- systemctl is-active nfs-kernel-server"
    run_test "Exports NFS configurés" "incus exec veza-storage -- exportfs -v | grep -q /storage"
}

# Test 6: Tester les applications
test_applications() {
    header "🎯 Tests des Applications"
    
    # Backend (si déployé)
    if incus info veza-backend >/dev/null 2>&1; then
        run_test "Backend service actif" "incus exec veza-backend -- systemctl is-active veza-backend"
        run_test "Backend écoute sur le port 8080" "incus exec veza-backend -- netstat -ln | grep :8080"
    fi
    
    # Frontend (si déployé)
    if incus info veza-frontend >/dev/null 2>&1; then
        run_test "Frontend Node.js installé" "incus exec veza-frontend -- node --version"
    fi
}

# Test 7: Tests de performances basiques
test_performance() {
    header "⚡ Tests de Performance"
    
    # Test de mémoire et CPU
    for container in "${CONTAINERS[@]}"; do
        if incus info "$container" >/dev/null 2>&1; then
            run_test "Mémoire disponible $container" "incus exec $container -- free -m | grep -q Mem"
            run_test "CPU accessible $container" "incus exec $container -- nproc | grep -q '[0-9]'"
        fi
    done
}

# Test 8: Tests de sécurité basiques
test_security() {
    header "🔒 Tests de Sécurité"
    
    # Vérifier que les containers ne sont pas en mode privilégié (sauf storage)
    for container in "${CONTAINERS[@]}"; do
        if [[ "$container" != "veza-storage" ]] && incus info "$container" >/dev/null 2>&1; then
            run_test "Container $container non privilégié" "! incus config get $container security.privileged | grep -q true"
        fi
    done
}

# Test 9: Tests de sauvegarde et récupération
test_backup_recovery() {
    header "💾 Tests de Sauvegarde"
    
    # Test de snapshot
    run_test "Création snapshot PostgreSQL" "incus snapshot veza-postgres test-snapshot"
    run_test "Liste des snapshots" "incus info veza-postgres | grep -q test-snapshot"
    run_test "Suppression snapshot test" "incus delete veza-postgres/test-snapshot"
}

# Test 10: Tests de monitoring
test_monitoring() {
    header "📊 Tests de Monitoring"
    
    # Vérifier les logs
    for container in "${CONTAINERS[@]}"; do
        if incus info "$container" >/dev/null 2>&1; then
            run_test "Logs accessibles $container" "incus exec $container -- journalctl --no-pager -n 1"
        fi
    done
}

# Test de régression complet
test_regression() {
    header "🔄 Tests de Régression"
    
    # Redémarrer un container et vérifier qu'il fonctionne toujours
    local test_container="veza-redis"
    
    log "Redémarrage de $test_container pour test de régression..."
    incus restart "$test_container"
    sleep 10
    
    run_test "Container $test_container redémarré correctement" "incus list $test_container -c s | grep -q RUNNING"
    run_test "Service Redis toujours actif après redémarrage" "incus exec $test_container -- systemctl is-active redis-server"
}

# Fonction principale
main() {
    header "🧪 Veza - Suite de Tests Complète"
    
    log "Démarrage des tests de déploiement Veza..."
    echo ""
    
    # Exécuter toutes les suites de tests
    test_incus_installation
    test_containers_status
    test_containers_ips
    test_network_connectivity
    test_services
    test_applications  
    test_performance
    test_security
    test_backup_recovery
    test_monitoring
    test_regression
    
    # Rapport final
    header "📊 Rapport Final des Tests"
    
    echo -e "${CYAN}Tests Exécutés: ${BOLD}$TESTS_TOTAL${NC}"
    echo -e "${GREEN}Tests Réussis: ${BOLD}$TESTS_PASSED${NC}"
    echo -e "${RED}Tests Échoués: ${BOLD}$TESTS_FAILED${NC}"
    
    if [ $TESTS_FAILED -eq 0 ]; then
        echo ""
        success "🎉 Tous les tests sont passés ! Déploiement validé."
        echo ""
        echo -e "${CYAN}🌐 Accès aux services :${NC}"
        echo -e "  • Application principale: ${GREEN}http://10.100.0.16${NC}"
        echo -e "  • HAProxy Stats: ${GREEN}http://10.100.0.16:8404/stats${NC}"
        echo -e "  • Frontend (dev): ${GREEN}http://10.100.0.11:5173${NC}"
        echo -e "  • Backend API: ${GREEN}http://10.100.0.12:8080${NC}"
        echo ""
        exit 0
    else
        echo ""
        error "❌ Certains tests ont échoué. Vérifiez le déploiement."
        echo ""
        echo -e "${YELLOW}💡 Commandes de diagnostic :${NC}"
        echo -e "  • Statut containers: ${CYAN}incus list${NC}"
        echo -e "  • Logs d'un container: ${CYAN}incus exec <container> -- journalctl -n 50${NC}"
        echo -e "  • Réseau: ${CYAN}incus network show veza-network${NC}"
        echo ""
        exit 1
    fi
}

# Fonction d'aide
show_help() {
    echo -e "${CYAN}${BOLD}Veza - Script de Tests${NC}"
    echo ""
    echo -e "${BLUE}Usage:${NC} $0 [OPTIONS]"
    echo ""
    echo -e "${YELLOW}Options:${NC}"
    echo -e "  ${GREEN}--quick${NC}      - Tests rapides uniquement"
    echo -e "  ${GREEN}--full${NC}       - Tests complets (défaut)"
    echo -e "  ${GREEN}--help${NC}       - Afficher cette aide"
    echo ""
}

# Parser les arguments
if [[ $# -gt 0 ]]; then
    case $1 in
        --help|-h)
            show_help
            exit 0
            ;;
        --quick)
            log "Mode tests rapides"
            test_incus_installation
            test_containers_status
            test_containers_ips
            exit 0
            ;;
        --full)
            log "Mode tests complets"
            ;;
        *)
            error "Option inconnue: $1"
            ;;
    esac
fi

# Exécution
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi 