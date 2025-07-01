#!/bin/bash

# Script de génération des dashboards Grafana production
echo "📊 Génération des dashboards Grafana..."

mkdir -p dashboards

# Dashboard System Overview
cat > dashboards/system-overview.json << 'DASH1'
{
  "dashboard": {
    "title": "🖥️ System Overview - Veza Stream Server",
    "tags": ["veza", "system"],
    "refresh": "5s",
    "panels": [
      {
        "id": 1,
        "title": "CPU Usage",
        "type": "gauge",
        "targets": [{"expr": "system_cpu_usage_percent"}],
        "fieldConfig": {
          "defaults": {
            "min": 0, "max": 100, "unit": "percent",
            "thresholds": {"steps": [
              {"color": "green", "value": 0},
              {"color": "yellow", "value": 70},
              {"color": "red", "value": 85}
            ]}
          }
        }
      }
    ]
  }
}
DASH1

echo "✅ Dashboard system-overview.json créé"
echo "📁 Dashboards sauvés dans: dashboards/"
