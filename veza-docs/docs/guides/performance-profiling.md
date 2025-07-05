---
title: Profilage de Performance
sidebar_label: Profilage
---

# 🚀 Profilage de Performance

Ce guide explique comment profiler les performances sur Veza.

## Vue d'ensemble

Ce guide détaille les techniques et outils de profiling de performance pour la plateforme Veza, couvrant l'analyse CPU, mémoire, réseau et les optimisations.

## Table des matières

- [Métriques de Performance](#métriques-de-performance)
- [Outils de Profiling](#outils-de-profiling)
- [Techniques d'Optimisation](#techniques-doptimisation)
- [Bonnes Pratiques](#bonnes-pratiques)
- [Pièges à Éviter](#pièges-à-éviter)
- [Ressources](#ressources)

## Métriques de Performance

### 1. Métriques Système

```yaml
# performance-profiling/metrics/system-metrics.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: veza-system-metrics
  namespace: veza
data:
  # Métriques CPU
  cpu_metrics:
    - "CPU Usage: Pourcentage d'utilisation"
    - "CPU Load: Charge moyenne"
    - "CPU Time: Temps CPU par processus"
    - "Context Switches: Changements de contexte"
  
  # Métriques Mémoire
  memory_metrics:
    - "Memory Usage: Utilisation mémoire"
    - "Memory Allocation: Allocations par seconde"
    - "Garbage Collection: Fréquence et durée"
    - "Memory Leaks: Fuites mémoire"
  
  # Métriques Réseau
  network_metrics:
    - "Network I/O: Entrées/sorties réseau"
    - "Connection Pool: Pool de connexions"
    - "Latency: Latence réseau"
    - "Throughput: Débit réseau"
  
  # Métriques Disque
  disk_metrics:
    - "Disk I/O: Entrées/sorties disque"
    - "Disk Space: Espace disque"
    - "Disk Latency: Latence disque"
    - "File Operations: Opérations fichiers"
```

### 2. Métriques Application

```go
// performance-profiling/metrics/app-metrics.go
package metrics

import (
    "runtime"
    "time"
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

// Métriques Prometheus
var (
    // Métriques HTTP
    httpRequestsTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "http_requests_total",
            Help: "Total des requêtes HTTP",
        },
        []string{"method", "endpoint", "status"},
    )

    httpRequestDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "http_request_duration_seconds",
            Help:    "Durée des requêtes HTTP",
            Buckets: prometheus.DefBuckets,
        },
        []string{"method", "endpoint"},
    )

    // Métriques Base de données
    dbQueryDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "db_query_duration_seconds",
            Help:    "Durée des requêtes base de données",
            Buckets: prometheus.DefBuckets,
        },
        []string{"query_type", "table"},
    )

    // Métriques Mémoire
    memoryAllocations = promauto.NewGauge(
        prometheus.GaugeOpts{
            Name: "memory_allocations_bytes",
            Help: "Allocations mémoire en bytes",
        },
    )

    // Métriques Goroutines
    goroutinesCount = promauto.NewGauge(
        prometheus.GaugeOpts{
            Name: "goroutines_count",
            Help: "Nombre de goroutines actives",
        },
    )
)

// Collecteur de métriques système
type SystemMetricsCollector struct {
    lastCollection time.Time
}

func NewSystemMetricsCollector() *SystemMetricsCollector {
    return &SystemMetricsCollector{
        lastCollection: time.Now(),
    }
}

func (c *SystemMetricsCollector) Collect() {
    // Métriques mémoire
    var m runtime.MemStats
    runtime.ReadMemStats(&m)
    memoryAllocations.Set(float64(m.Alloc))

    // Métriques goroutines
    goroutinesCount.Set(float64(runtime.NumGoroutine()))

    c.lastCollection = time.Now()
}

// Middleware pour métriques HTTP
func MetricsMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()

        // Wrapper pour capturer le statut
        rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
        next.ServeHTTP(rw, r)

        // Enregistrement des métriques
        duration := time.Since(start).Seconds()
        httpRequestsTotal.WithLabelValues(r.Method, r.URL.Path, string(rw.statusCode)).Inc()
        httpRequestDuration.WithLabelValues(r.Method, r.URL.Path).Observe(duration)
    })
}

type responseWriter struct {
    http.ResponseWriter
    statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
    rw.statusCode = code
    rw.ResponseWriter.WriteHeader(code)
}
```

## Outils de Profiling

### 1. Configuration pprof

```go
// performance-profiling/pprof/pprof-server.go
package pprof

import (
    "net/http"
    _ "net/http/pprof"
    "runtime"
    "time"
)

// Serveur pprof
func StartPprofServer(addr string) {
    go func() {
        log.Printf("Pprof server started on %s", addr)
        log.Fatal(http.ListenAndServe(addr, nil))
    }()
}

// Profiling CPU
func StartCPUProfiling(duration time.Duration) {
    f, err := os.Create("cpu.prof")
    if err != nil {
        log.Fatal(err)
    }
    defer f.Close()

    if err := pprof.StartCPUProfile(f); err != nil {
        log.Fatal(err)
    }

    time.Sleep(duration)
    pprof.StopCPUProfile()
}

// Profiling Mémoire
func StartMemoryProfiling() {
    f, err := os.Create("memory.prof")
    if err != nil {
        log.Fatal(err)
    }
    defer f.Close()

    if err := pprof.WriteHeapProfile(f); err != nil {
        log.Fatal(err)
    }
}

// Profiling Goroutines
func StartGoroutineProfiling() {
    f, err := os.Create("goroutine.prof")
    if err != nil {
        log.Fatal(err)
    }
    defer f.Close()

    if err := pprof.Lookup("goroutine").WriteTo(f, 0); err != nil {
        log.Fatal(err)
    }
}
```

### 2. Scripts de Profiling

```bash
#!/bin/bash
# performance-profiling/scripts/profile.sh

# Configuration
PROFILE_DURATION=30
PROFILE_OUTPUT="profiles"
SERVICE_URL="http://localhost:8080"

# Création du répertoire de sortie
mkdir -p $PROFILE_OUTPUT

# Profiling CPU
cpu_profile() {
    echo "Démarrage du profiling CPU..."
    
    # Démarrage du profiling
    curl -X POST http://localhost:6060/debug/pprof/profile?seconds=$PROFILE_DURATION \
        -o $PROFILE_OUTPUT/cpu.prof
    
    echo "Profiling CPU terminé: $PROFILE_OUTPUT/cpu.prof"
}

# Profiling Mémoire
memory_profile() {
    echo "Démarrage du profiling mémoire..."
    
    # Profiling heap
    curl http://localhost:6060/debug/pprof/heap \
        -o $PROFILE_OUTPUT/heap.prof
    
    # Profiling allocations
    curl http://localhost:6060/debug/pprof/allocs \
        -o $PROFILE_OUTPUT/allocs.prof
    
    echo "Profiling mémoire terminé"
}

# Profiling Goroutines
goroutine_profile() {
    echo "Démarrage du profiling goroutines..."
    
    curl http://localhost:6060/debug/pprof/goroutine \
        -o $PROFILE_OUTPUT/goroutine.prof
    
    echo "Profiling goroutines terminé"
}

# Analyse des profils
analyze_profiles() {
    echo "Analyse des profils..."
    
    # Analyse CPU
    if [ -f "$PROFILE_OUTPUT/cpu.prof" ]; then
        echo "=== Analyse CPU ==="
        go tool pprof -top $PROFILE_OUTPUT/cpu.prof
        echo ""
    fi
    
    # Analyse Mémoire
    if [ -f "$PROFILE_OUTPUT/heap.prof" ]; then
        echo "=== Analyse Mémoire ==="
        go tool pprof -top $PROFILE_OUTPUT/heap.prof
        echo ""
    fi
    
    # Analyse Goroutines
    if [ -f "$PROFILE_OUTPUT/goroutine.prof" ]; then
        echo "=== Analyse Goroutines ==="
        go tool pprof -top $PROFILE_OUTPUT/goroutine.prof
        echo ""
    fi
}

# Test de charge
load_test() {
    echo "Test de charge..."
    
    # Test avec wrk
    if command -v wrk &> /dev/null; then
        wrk -t4 -c100 -d30s $SERVICE_URL/health
    else
        echo "wrk non installé, utilisation d'ab"
        ab -n 1000 -c 10 $SERVICE_URL/health
    fi
}

# Menu principal
case "${1:-}" in
    cpu)
        cpu_profile
        ;;
    memory)
        memory_profile
        ;;
    goroutine)
        goroutine_profile
        ;;
    all)
        cpu_profile
        memory_profile
        goroutine_profile
        ;;
    analyze)
        analyze_profiles
        ;;
    load)
        load_test
        ;;
    *)
        echo "Usage: $0 {cpu|memory|goroutine|all|analyze|load}"
        exit 1
        ;;
esac
```

## Techniques d'Optimisation

### 1. Optimisation CPU

```go
// performance-profiling/optimization/cpu-optimization.go
package optimization

import (
    "runtime"
    "sync"
    "time"
)

// Optimisation du pool de workers
type WorkerPool struct {
    workers    int
    jobQueue   chan Job
    workerPool chan chan Job
    quit       chan bool
    wg         sync.WaitGroup
}

type Job struct {
    ID       int
    Data     interface{}
    Process  func(interface{}) error
}

func NewWorkerPool(workers int) *WorkerPool {
    return &WorkerPool{
        workers:    workers,
        jobQueue:   make(chan Job, 1000),
        workerPool: make(chan chan Job, workers),
        quit:       make(chan bool),
    }
}

func (wp *WorkerPool) Start() {
    for i := 0; i < wp.workers; i++ {
        wp.wg.Add(1)
        go wp.worker()
    }
    
    go wp.dispatcher()
}

func (wp *WorkerPool) worker() {
    defer wp.wg.Done()
    
    jobChannel := make(chan Job)
    wp.workerPool <- jobChannel
    
    for {
        select {
        case job := <-jobChannel:
            job.Process(job.Data)
            wp.workerPool <- jobChannel
        case <-wp.quit:
            return
        }
    }
}

func (wp *WorkerPool) dispatcher() {
    for {
        select {
        case job := <-wp.jobQueue:
            go func(job Job) {
                jobChannel := <-wp.workerPool
                jobChannel <- job
            }(job)
        case <-wp.quit:
            return
        }
    }
}

func (wp *WorkerPool) Submit(job Job) {
    wp.jobQueue <- job
}

func (wp *WorkerPool) Stop() {
    close(wp.quit)
    wp.wg.Wait()
}

// Optimisation du cache
type Cache struct {
    data map[string]interface{}
    mu   sync.RWMutex
    ttl  time.Duration
}

func NewCache(ttl time.Duration) *Cache {
    cache := &Cache{
        data: make(map[string]interface{}),
        ttl:  ttl,
    }
    
    // Nettoyage périodique
    go cache.cleanup()
    
    return cache
}

func (c *Cache) Set(key string, value interface{}) {
    c.mu.Lock()
    defer c.mu.Unlock()
    
    c.data[key] = value
}

func (c *Cache) Get(key string) (interface{}, bool) {
    c.mu.RLock()
    defer c.mu.RUnlock()
    
    value, exists := c.data[key]
    return value, exists
}

func (c *Cache) cleanup() {
    ticker := time.NewTicker(c.ttl)
    defer ticker.Stop()
    
    for range ticker.C {
        c.mu.Lock()
        // Nettoyage des entrées expirées
        for key := range c.data {
            // Logique de nettoyage basée sur TTL
            delete(c.data, key)
        }
        c.mu.Unlock()
    }
}
```

### 2. Optimisation Mémoire

```go
// performance-profiling/optimization/memory-optimization.go
package optimization

import (
    "sync"
    "time"
)

// Pool d'objets pour réduire les allocations
type ObjectPool struct {
    pool sync.Pool
}

func NewObjectPool(newFunc func() interface{}) *ObjectPool {
    return &ObjectPool{
        pool: sync.Pool{
            New: newFunc,
        },
    }
}

func (op *ObjectPool) Get() interface{} {
    return op.pool.Get()
}

func (op *ObjectPool) Put(obj interface{}) {
    op.pool.Put(obj)
}

// Optimisation des slices
type OptimizedSlice struct {
    data []interface{}
    mu   sync.RWMutex
}

func NewOptimizedSlice(capacity int) *OptimizedSlice {
    return &OptimizedSlice{
        data: make([]interface{}, 0, capacity),
    }
}

func (os *OptimizedSlice) Add(item interface{}) {
    os.mu.Lock()
    defer os.mu.Unlock()
    
    os.data = append(os.data, item)
}

func (os *OptimizedSlice) Get(index int) (interface{}, bool) {
    os.mu.RLock()
    defer os.mu.RUnlock()
    
    if index < len(os.data) {
        return os.data[index], true
    }
    return nil, false
}

// Optimisation des maps
type OptimizedMap struct {
    data map[string]interface{}
    mu   sync.RWMutex
}

func NewOptimizedMap() *OptimizedMap {
    return &OptimizedMap{
        data: make(map[string]interface{}),
    }
}

func (om *OptimizedMap) Set(key string, value interface{}) {
    om.mu.Lock()
    defer om.mu.Unlock()
    
    om.data[key] = value
}

func (om *OptimizedMap) Get(key string) (interface{}, bool) {
    om.mu.RLock()
    defer om.mu.RUnlock()
    
    value, exists := om.data[key]
    return value, exists
}
```

## Bonnes Pratiques

### 1. Règles de Performance

```yaml
# performance-profiling/best-practices/performance-rules.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: veza-performance-rules
  namespace: veza
data:
  # Règles CPU
  cpu_rules:
    - "Utiliser des goroutines pour la concurrence"
    - "Éviter les boucles infinies"
    - "Optimiser les algorithmes critiques"
    - "Utiliser des pools de workers"
    - "Profiler régulièrement"
  
  # Règles Mémoire
  memory_rules:
    - "Réutiliser les objets avec sync.Pool"
    - "Pré-allouer les slices et maps"
    - "Éviter les allocations dans les boucles"
    - "Utiliser des pointeurs pour les gros objets"
    - "Surveiller les fuites mémoire"
  
  # Règles Réseau
  network_rules:
    - "Utiliser des connexions persistantes"
    - "Implémenter du connection pooling"
    - "Optimiser les requêtes base de données"
    - "Utiliser la compression"
    - "Implémenter du caching"
  
  # Règles Base de données
  database_rules:
    - "Utiliser des index appropriés"
    - "Optimiser les requêtes"
    - "Utiliser des transactions"
    - "Implémenter du connection pooling"
    - "Surveiller les slow queries"
```

### 2. Monitoring de Performance

```go
// performance-profiling/monitoring/performance-monitor.go
package monitoring

import (
    "context"
    "runtime"
    "time"
)

// Moniteur de performance
type PerformanceMonitor struct {
    metrics map[string]float64
    mu      sync.RWMutex
}

func NewPerformanceMonitor() *PerformanceMonitor {
    pm := &PerformanceMonitor{
        metrics: make(map[string]float64),
    }
    
    // Démarrage du monitoring
    go pm.monitor()
    
    return pm
}

func (pm *PerformanceMonitor) monitor() {
    ticker := time.NewTicker(30 * time.Second)
    defer ticker.Stop()
    
    for range ticker.C {
        pm.collectMetrics()
    }
}

func (pm *PerformanceMonitor) collectMetrics() {
    pm.mu.Lock()
    defer pm.mu.Unlock()
    
    // Métriques mémoire
    var m runtime.MemStats
    runtime.ReadMemStats(&m)
    
    pm.metrics["memory_alloc"] = float64(m.Alloc)
    pm.metrics["memory_total_alloc"] = float64(m.TotalAlloc)
    pm.metrics["memory_sys"] = float64(m.Sys)
    pm.metrics["memory_num_gc"] = float64(m.NumGC)
    
    // Métriques goroutines
    pm.metrics["goroutines"] = float64(runtime.NumGoroutine())
    
    // Métriques CPU
    pm.metrics["cpu_cores"] = float64(runtime.NumCPU())
}

func (pm *PerformanceMonitor) GetMetrics() map[string]float64 {
    pm.mu.RLock()
    defer pm.mu.RUnlock()
    
    metrics := make(map[string]float64)
    for k, v := range pm.metrics {
        metrics[k] = v
    }
    
    return metrics
}
```

## Pièges à Éviter

### 1. Allocations Excessives

❌ **Mauvais** :
```go
// Allocations dans une boucle
for i := 0; i < 1000000; i++ {
    data := make([]byte, 1024)  // Allocation à chaque itération
    process(data)
}
```

✅ **Bon** :
```go
// Réutilisation d'objets
data := make([]byte, 1024)
for i := 0; i < 1000000; i++ {
    // Réutilisation du même slice
    process(data)
}
```

### 2. Goroutines Non Contrôlées

❌ **Mauvais** :
```go
// Goroutines non contrôlées
for i := 0; i < 1000000; i++ {
    go process(i)  // Création de trop de goroutines
}
```

✅ **Bon** :
```go
// Pool de workers
pool := NewWorkerPool(100)
pool.Start()

for i := 0; i < 1000000; i++ {
    pool.Submit(Job{ID: i, Process: process})
}
```

### 3. Pas de Monitoring

❌ **Mauvais** :
```go
// Pas de monitoring
func process() {
    // Logique sans monitoring
}
```

✅ **Bon** :
```go
// Avec monitoring
func process() {
    start := time.Now()
    defer func() {
        duration := time.Since(start)
        metrics.RecordDuration("process", duration)
    }()
    
    // Logique avec monitoring
}
```

## Ressources

### Documentation Interne

- [Guide de Monitoring](../monitoring/README.md)
- [Guide de Debugging](./debugging.md)
- [Guide de Tests](../testing/README.md)

### Outils Recommandés

- **pprof** : Profiling Go
- **perf** : Profiling système
- **Prometheus** : Métriques
- **Grafana** : Visualisation

### Commandes Utiles

```bash
# Profiling CPU
go tool pprof http://localhost:6060/debug/pprof/profile

# Profiling Mémoire
go tool pprof http://localhost:6060/debug/pprof/heap

# Analyse de performance
go test -bench=. -benchmem

# Monitoring en temps réel
htop
iotop
```

---

**Dernière mise à jour** : $(date)
**Version du guide** : 1.0.0
**Mainteneur** : Équipe Performance Veza 