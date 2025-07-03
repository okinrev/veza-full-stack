/// Module Benchmarks pour mesures de performance précises

use std::time::{Duration, Instant};
use tracing::{info};
use serde::{Serialize, Deserialize};

/// Runner de benchmarks
#[derive(Debug)]
pub struct BenchmarkRunner {
    warmup_iterations: u32,
    measurement_iterations: u32,
}

/// Résultats de benchmark
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct BenchmarkResults {
    pub test_name: String,
    pub mean_duration: Duration,
    pub min_duration: Duration,
    pub max_duration: Duration,
    pub std_deviation: f64,
    pub iterations: u32,
    pub throughput_ops_per_sec: f64,
}

impl BenchmarkRunner {
    pub fn new() -> Self {
        Self {
            warmup_iterations: 100,
            measurement_iterations: 1000,
        }
    }

    pub async fn benchmark<F, Fut>(&self, test_name: &str, operation: F) -> BenchmarkResults
    where
        F: Fn() -> Fut + Send + Sync,
        Fut: std::future::Future<Output = ()> + Send,
    {
        info!("🏃 Benchmark: {}", test_name);
        
        // Warmup
        for _ in 0..self.warmup_iterations {
            operation().await;
        }
        
        // Mesures
        let mut durations = Vec::new();
        let start_benchmark = Instant::now();
        
        for _ in 0..self.measurement_iterations {
            let start = Instant::now();
            operation().await;
            let duration = start.elapsed();
            durations.push(duration);
        }
        
        let total_benchmark_time = start_benchmark.elapsed();
        
        // Calculs statistiques
        let mean = Duration::from_nanos(
            (durations.iter().map(|d| d.as_nanos()).sum::<u128>() / durations.len() as u128) as u64
        );
        
        let min = *durations.iter().min().unwrap();
        let max = *durations.iter().max().unwrap();
        
        let variance = durations.iter()
            .map(|d| {
                let diff = d.as_nanos() as f64 - mean.as_nanos() as f64;
                diff * diff
            })
            .sum::<f64>() / durations.len() as f64;
        
        let std_deviation = variance.sqrt();
        
        let throughput = self.measurement_iterations as f64 / total_benchmark_time.as_secs_f64();
        
        BenchmarkResults {
            test_name: test_name.to_string(),
            mean_duration: mean,
            min_duration: min,
            max_duration: max,
            std_deviation,
            iterations: self.measurement_iterations,
            throughput_ops_per_sec: throughput,
        }
    }
}

impl Default for BenchmarkRunner {
    fn default() -> Self {
        Self::new()
    }
}
