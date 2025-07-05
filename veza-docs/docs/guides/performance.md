# Guide de Performance - Veza Platform

## Vue d'ensemble

Ce guide détaille les stratégies d'optimisation de performance pour la plateforme Veza, couvrant l'optimisation CPU, mémoire, réseau et base de données.

## Table des matières

- [Métriques de Performance](#métriques-de-performance)
- [Optimisation CPU](#optimisation-cpu)
- [Optimisation Mémoire](#optimisation-mémoire)
- [Optimisation Réseau](#optimisation-réseau)
- [Optimisation Base de Données](#optimisation-base-de-données)
- [Bonnes Pratiques](#bonnes-pratiques)
- [Pièges à Éviter](#pièges-à-éviter)
- [Ressources](#ressources)

## Métriques de Performance

### 1. Métriques Clés

```yaml
# performance/metrics/key-metrics.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: veza-performance-metrics
  namespace: veza
data:
  # Métriques système
  system_metrics:
    cpu:
      - "cpu_usage_percent"
      - "cpu_load_average"
      - "cpu_context_switches"
      - "cpu_interrupts"
    
    memory:
      - "memory_usage_bytes"
      - "memory_available_bytes"
      - "memory_swap_usage"
      - "memory_page_faults"
    
    disk:
      - "disk_io_read_bytes"
      - "disk_io_write_bytes"
      - "disk_io_ops"
      - "disk_usage_percent"
    
    network:
      - "network_bytes_received"
      - "network_bytes_sent"
      - "network_packets_received"
      - "network_packets_sent"
  
  # Métriques application
  application_metrics:
    http:
      - "http_requests_total"
      - "http_request_duration_seconds"
      - "http_requests_in_flight"
      - "http_response_size_bytes"
    
    database:
      - "db_connections_active"
      - "db_connections_idle"
      - "db_query_duration_seconds"
      - "db_transactions_total"
    
    cache:
      - "cache_hits_total"
      - "cache_misses_total"
      - "cache_hit_ratio"
      - "cache_memory_usage_bytes"
  
  # Métriques business
  business_metrics:
    - "active_users"
    - "messages_per_second"
    - "streams_active"
    - "concurrent_connections"
    - "response_time_p95"
    - "error_rate"
```

### 2. Collecteur de Métriques

```go
// performance/metrics/collector.go
package metrics

import (
    "runtime"
    "time"
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

// Collecteur de métriques de performance
type PerformanceCollector struct {
    // Métriques système
    cpuUsage prometheus.Gauge
    memoryUsage prometheus.Gauge
    goroutinesCount prometheus.Gauge
    
    // Métriques application
    httpRequestsTotal prometheus.Counter
    httpRequestDuration prometheus.Histogram
    dbConnectionsActive prometheus.Gauge
    cacheHitRatio prometheus.Gauge
    
    // Métriques business
    activeUsers prometheus.Gauge
    messagesPerSecond prometheus.Counter
    concurrentConnections prometheus.Gauge
}

func NewPerformanceCollector() *PerformanceCollector {
    return &PerformanceCollector{
        // Métriques système
        cpuUsage: promauto.NewGauge(prometheus.GaugeOpts{
            Name: "cpu_usage_percent",
            Help: "CPU usage percentage",
        }),
        memoryUsage: promauto.NewGauge(prometheus.GaugeOpts{
            Name: "memory_usage_bytes",
            Help: "Memory usage in bytes",
        }),
        goroutinesCount: promauto.NewGauge(prometheus.GaugeOpts{
            Name: "goroutines_count",
            Help: "Number of active goroutines",
        }),
        
        // Métriques application
        httpRequestsTotal: promauto.NewCounter(prometheus.CounterOpts{
            Name: "http_requests_total",
            Help: "Total number of HTTP requests",
        }),
        httpRequestDuration: promauto.NewHistogram(prometheus.HistogramOpts{
            Name: "http_request_duration_seconds",
            Help: "HTTP request duration in seconds",
            Buckets: prometheus.DefBuckets,
        }),
        dbConnectionsActive: promauto.NewGauge(prometheus.GaugeOpts{
            Name: "db_connections_active",
            Help: "Number of active database connections",
        }),
        cacheHitRatio: promauto.NewGauge(prometheus.GaugeOpts{
            Name: "cache_hit_ratio",
            Help: "Cache hit ratio",
        }),
        
        // Métriques business
        activeUsers: promauto.NewGauge(prometheus.GaugeOpts{
            Name: "active_users",
            Help: "Number of active users",
        }),
        messagesPerSecond: promauto.NewCounter(prometheus.CounterOpts{
            Name: "messages_per_second",
            Help: "Messages processed per second",
        }),
        concurrentConnections: promauto.NewGauge(prometheus.GaugeOpts{
            Name: "concurrent_connections",
            Help: "Number of concurrent connections",
        }),
    }
}

func (pc *PerformanceCollector) Collect() {
    // Collecte des métriques système
    var m runtime.MemStats
    runtime.ReadMemStats(&m)
    
    pc.memoryUsage.Set(float64(m.Alloc))
    pc.goroutinesCount.Set(float64(runtime.NumGoroutine()))
    
    // Collecte des métriques CPU (simplifiée)
    pc.cpuUsage.Set(pc.getCPUUsage())
}

func (pc *PerformanceCollector) getCPUUsage() float64 {
    // Logique de calcul CPU (simplifiée)
    return 0.0
}

func (pc *PerformanceCollector) RecordHTTPRequest(duration time.Duration) {
    pc.httpRequestsTotal.Inc()
    pc.httpRequestDuration.Observe(duration.Seconds())
}

func (pc *PerformanceCollector) RecordDatabaseConnection(active int) {
    pc.dbConnectionsActive.Set(float64(active))
}

func (pc *PerformanceCollector) RecordCacheHitRatio(ratio float64) {
    pc.cacheHitRatio.Set(ratio)
}

func (pc *PerformanceCollector) RecordActiveUsers(count int) {
    pc.activeUsers.Set(float64(count))
}

func (pc *PerformanceCollector) RecordMessage() {
    pc.messagesPerSecond.Inc()
}

func (pc *PerformanceCollector) RecordConnection(count int) {
    pc.concurrentConnections.Set(float64(count))
}
```

## Optimisation CPU

### 1. Pool de Workers

```go
// performance/cpu/worker-pool.go
package cpu

import (
    "context"
    "sync"
    "time"
)

// Job représente une tâche à exécuter
type Job struct {
    ID       int
    Data     interface{}
    Process  func(interface{}) error
    Priority int
}

// WorkerPool gère un pool de workers
type WorkerPool struct {
    workers    int
    jobQueue   chan Job
    workerPool chan chan Job
    quit       chan bool
    wg         sync.WaitGroup
    metrics    *PerformanceCollector
}

func NewWorkerPool(workers int, metrics *PerformanceCollector) *WorkerPool {
    return &WorkerPool{
        workers:    workers,
        jobQueue:   make(chan Job, 1000),
        workerPool: make(chan chan Job, workers),
        quit:       make(chan bool),
        metrics:    metrics,
    }
}

func (wp *WorkerPool) Start() {
    for i := 0; i < wp.workers; i++ {
        wp.wg.Add(1)
        go wp.worker(i)
    }
    
    go wp.dispatcher()
}

func (wp *WorkerPool) worker(id int) {
    defer wp.wg.Done()
    
    jobChannel := make(chan Job)
    wp.workerPool <- jobChannel
    
    for {
        select {
        case job := <-jobChannel:
            start := time.Now()
            
            // Exécution de la tâche
            if err := job.Process(job.Data); err != nil {
                // Log de l'erreur
            }
            
            // Métriques de performance
            duration := time.Since(start)
            wp.metrics.RecordHTTPRequest(duration)
            
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
```

### 2. Optimisation des Algorithmes

```go
// performance/cpu/algorithm-optimization.go
package cpu

import (
    "sort"
    "sync"
)

// Optimisation de recherche
type OptimizedSearch struct {
    data []int
    mu   sync.RWMutex
}

func NewOptimizedSearch(data []int) *OptimizedSearch {
    // Tri des données pour recherche binaire
    sortedData := make([]int, len(data))
    copy(sortedData, data)
    sort.Ints(sortedData)
    
    return &OptimizedSearch{
        data: sortedData,
    }
}

// Recherche binaire optimisée
func (os *OptimizedSearch) BinarySearch(target int) int {
    os.mu.RLock()
    defer os.mu.RUnlock()
    
    left, right := 0, len(os.data)-1
    
    for left <= right {
        mid := (left + right) / 2
        
        if os.data[mid] == target {
            return mid
        } else if os.data[mid] < target {
            left = mid + 1
        } else {
            right = mid - 1
        }
    }
    
    return -1
}

// Optimisation de cache
type CacheOptimizer struct {
    cache map[string]interface{}
    mu    sync.RWMutex
}

func NewCacheOptimizer() *CacheOptimizer {
    return &CacheOptimizer{
        cache: make(map[string]interface{}),
    }
}

func (co *CacheOptimizer) Get(key string) (interface{}, bool) {
    co.mu.RLock()
    defer co.mu.RUnlock()
    
    value, exists := co.cache[key]
    return value, exists
}

func (co *CacheOptimizer) Set(key string, value interface{}) {
    co.mu.Lock()
    defer co.mu.Unlock()
    
    co.cache[key] = value
}
```

## Optimisation Mémoire

### 1. Pool d'Objets

```go
// performance/memory/object-pool.go
package memory

import (
    "sync"
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

### 2. Gestion de la Mémoire

```go
// performance/memory/memory-manager.go
package memory

import (
    "runtime"
    "time"
)

// Gestionnaire de mémoire
type MemoryManager struct {
    maxMemory uint64
    gcPercent int
}

func NewMemoryManager(maxMemory uint64) *MemoryManager {
    return &MemoryManager{
        maxMemory: maxMemory,
        gcPercent: runtime.GOGC,
    }
}

func (mm *MemoryManager) StartMonitoring() {
    go func() {
        ticker := time.NewTicker(30 * time.Second)
        defer ticker.Stop()
        
        for range ticker.C {
            mm.checkMemoryUsage()
        }
    }()
}

func (mm *MemoryManager) checkMemoryUsage() {
    var m runtime.MemStats
    runtime.ReadMemStats(&m)
    
    // Vérification de l'utilisation mémoire
    if m.Alloc > mm.maxMemory {
        // Déclenchement du GC
        runtime.GC()
    }
    
    // Ajustement du pourcentage GC
    if m.Alloc > mm.maxMemory*80/100 {
        runtime.GOGC = 50 // Plus agressif
    } else {
        runtime.GOGC = mm.gcPercent // Normal
    }
}

func (mm *MemoryManager) ForceGC() {
    runtime.GC()
}

func (mm *MemoryManager) GetMemoryStats() runtime.MemStats {
    var m runtime.MemStats
    runtime.ReadMemStats(&m)
    return m
}
```

## Optimisation Réseau

### 1. Connection Pooling

```go
// performance/network/connection-pool.go
package network

import (
    "context"
    "database/sql"
    "sync"
    "time"
)

// Pool de connexions
type ConnectionPool struct {
    db     *sql.DB
    mu     sync.RWMutex
    config *PoolConfig
}

type PoolConfig struct {
    MaxOpenConns    int
    MaxIdleConns    int
    ConnMaxLifetime time.Duration
    ConnMaxIdleTime time.Duration
}

func NewConnectionPool(dsn string, config *PoolConfig) (*ConnectionPool, error) {
    db, err := sql.Open("postgres", dsn)
    if err != nil {
        return nil, err
    }
    
    // Configuration du pool
    db.SetMaxOpenConns(config.MaxOpenConns)
    db.SetMaxIdleConns(config.MaxIdleConns)
    db.SetConnMaxLifetime(config.ConnMaxLifetime)
    db.SetConnMaxIdleTime(config.ConnMaxIdleTime)
    
    return &ConnectionPool{
        db:     db,
        config: config,
    }, nil
}

func (cp *ConnectionPool) GetConnection(ctx context.Context) (*sql.DB, error) {
    // Vérification de la santé de la connexion
    if err := cp.db.PingContext(ctx); err != nil {
        return nil, err
    }
    
    return cp.db, nil
}

func (cp *ConnectionPool) Close() error {
    return cp.db.Close()
}

// Optimisation des requêtes HTTP
type HTTPOptimizer struct {
    client *http.Client
    cache  *CacheOptimizer
}

func NewHTTPOptimizer() *HTTPOptimizer {
    return &HTTPOptimizer{
        client: &http.Client{
            Timeout: 30 * time.Second,
            Transport: &http.Transport{
                MaxIdleConns:        100,
                MaxIdleConnsPerHost: 10,
                IdleConnTimeout:     90 * time.Second,
            },
        },
        cache: NewCacheOptimizer(),
    }
}

func (ho *HTTPOptimizer) GetWithCache(url string) ([]byte, error) {
    // Vérification du cache
    if cached, exists := ho.cache.Get(url); exists {
        return cached.([]byte), nil
    }
    
    // Requête HTTP
    resp, err := ho.client.Get(url)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, err
    }
    
    // Mise en cache
    ho.cache.Set(url, body)
    
    return body, nil
}
```

## Optimisation Base de Données

### 1. Optimisation des Requêtes

```go
// performance/database/query-optimizer.go
package database

import (
    "context"
    "database/sql"
    "time"
)

// Optimiseur de requêtes
type QueryOptimizer struct {
    db *sql.DB
}

func NewQueryOptimizer(db *sql.DB) *QueryOptimizer {
    return &QueryOptimizer{
        db: db,
    }
}

// Requête optimisée avec index
func (qo *QueryOptimizer) GetUserByEmail(email string) (*User, error) {
    query := `
        SELECT id, email, name, created_at 
        FROM users 
        WHERE email = $1 
        LIMIT 1
    `
    
    var user User
    err := qo.db.QueryRow(query, email).Scan(
        &user.ID,
        &user.Email,
        &user.Name,
        &user.CreatedAt,
    )
    
    if err != nil {
        return nil, err
    }
    
    return &user, nil
}

// Requête avec pagination optimisée
func (qo *QueryOptimizer) GetUsersPaginated(offset, limit int) ([]User, error) {
    query := `
        SELECT id, email, name, created_at 
        FROM users 
        ORDER BY created_at DESC 
        LIMIT $1 OFFSET $2
    `
    
    rows, err := qo.db.Query(query, limit, offset)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    
    var users []User
    for rows.Next() {
        var user User
        err := rows.Scan(
            &user.ID,
            &user.Email,
            &user.Name,
            &user.CreatedAt,
        )
        if err != nil {
            return nil, err
        }
        users = append(users, user)
    }
    
    return users, nil
}

// Requête avec cache
func (qo *QueryOptimizer) GetUserWithCache(userID string, cache *CacheOptimizer) (*User, error) {
    cacheKey := "user:" + userID
    
    // Vérification du cache
    if cached, exists := cache.Get(cacheKey); exists {
        return cached.(*User), nil
    }
    
    // Requête base de données
    user, err := qo.GetUserByID(userID)
    if err != nil {
        return nil, err
    }
    
    // Mise en cache
    cache.Set(cacheKey, user)
    
    return user, nil
}

// Batch insert optimisé
func (qo *QueryOptimizer) BatchInsertUsers(users []User) error {
    query := `
        INSERT INTO users (email, name, created_at) 
        VALUES ($1, $2, $3)
    `
    
    tx, err := qo.db.Begin()
    if err != nil {
        return err
    }
    defer tx.Rollback()
    
    stmt, err := tx.Prepare(query)
    if err != nil {
        return err
    }
    defer stmt.Close()
    
    for _, user := range users {
        _, err := stmt.Exec(user.Email, user.Name, user.CreatedAt)
        if err != nil {
            return err
        }
    }
    
    return tx.Commit()
}
```

## Bonnes Pratiques

### 1. Règles de Performance

```yaml
# performance/best-practices/performance-rules.yaml
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
  
  # Règles mémoire
  memory_rules:
    - "Réutiliser les objets avec sync.Pool"
    - "Pré-allouer les slices et maps"
    - "Éviter les allocations dans les boucles"
    - "Utiliser des pointeurs pour les gros objets"
    - "Surveiller les fuites mémoire"
  
  # Règles réseau
  network_rules:
    - "Utiliser des connexions persistantes"
    - "Implémenter du connection pooling"
    - "Optimiser les requêtes base de données"
    - "Utiliser la compression"
    - "Implémenter du caching"
  
  # Règles base de données
  database_rules:
    - "Utiliser des index appropriés"
    - "Optimiser les requêtes"
    - "Utiliser des transactions"
    - "Implémenter du connection pooling"
    - "Surveiller les slow queries"
```

### 2. Monitoring de Performance

```go
// performance/monitoring/performance-monitor.go
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

### 3. Requêtes Non Optimisées

❌ **Mauvais** :
```go
// Requête non optimisée
query := "SELECT * FROM users WHERE email = '" + email + "'"
```

✅ **Bon** :
```go
// Requête optimisée avec index
query := "SELECT id, email, name FROM users WHERE email = $1"
rows, err := db.Query(query, email)
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