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