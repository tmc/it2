package plugins

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"
)

// PluginMetrics tracks execution statistics for a single plugin
type PluginMetrics struct {
	Name           string        `json:"name"`
	Type           string        `json:"type,omitempty"`
	ExecutionCount int           `json:"execution_count"`
	TotalDuration  time.Duration `json:"total_duration"`
	Durations      []int64       `json:"durations"` // Microseconds for compact storage
	LastExecuted   time.Time     `json:"last_executed"`
}

// MetricsStore manages plugin execution metrics
type MetricsStore struct {
	mu      sync.RWMutex
	metrics map[string]*PluginMetrics // plugin name -> metrics
	path    string
}

var (
	globalMetrics *MetricsStore
	metricsOnce   sync.Once
)

// GetMetricsStore returns the global metrics store singleton
func GetMetricsStore() *MetricsStore {
	metricsOnce.Do(func() {
		home, err := os.UserHomeDir()
		if err != nil {
			// Fall back to in-memory only
			globalMetrics = &MetricsStore{
				metrics: make(map[string]*PluginMetrics),
			}
			return
		}

		metricsPath := filepath.Join(home, ".it2", "plugin-metrics.json")
		globalMetrics = &MetricsStore{
			metrics: make(map[string]*PluginMetrics),
			path:    metricsPath,
		}
		globalMetrics.load()
	})
	return globalMetrics
}

// RecordExecution records a plugin execution with its duration
func (s *MetricsStore) RecordExecution(pluginName, pluginType string, duration time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := metricsKey(pluginName, pluginType)

	m, exists := s.metrics[key]
	if !exists {
		m = &PluginMetrics{
			Name:      pluginName,
			Type:      pluginType,
			Durations: make([]int64, 0, 100),
		}
		s.metrics[key] = m
	}

	if m.Type == "" {
		m.Type = pluginType
	}
	m.ExecutionCount++
	m.TotalDuration += duration
	m.LastExecuted = time.Now()

	// Store duration in microseconds
	durationUs := duration.Microseconds()
	m.Durations = append(m.Durations, durationUs)

	// Keep only last 1000 executions to prevent unbounded growth
	if len(m.Durations) > 1000 {
		m.Durations = m.Durations[len(m.Durations)-1000:]
	}
}

// GetAllMetrics returns metrics for all plugins
func (s *MetricsStore) GetAllMetrics() map[string]*PluginMetrics {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make(map[string]*PluginMetrics, len(s.metrics))
	for key, m := range s.metrics {
		copy := *m
		copy.Durations = make([]int64, len(m.Durations))
		copySlice(copy.Durations, m.Durations)
		result[key] = &copy
	}
	return result
}

// Save persists metrics to disk
func (s *MetricsStore) Save() error {
	if s.path == "" {
		return nil // In-memory only
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	// Ensure directory exists
	dir := filepath.Dir(s.path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(s.metrics, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(s.path, data, 0644)
}

// load reads metrics from disk
func (s *MetricsStore) load() error {
	if s.path == "" {
		return nil // In-memory only
	}

	data, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // No metrics yet, that's fine
		}
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	return json.Unmarshal(data, &s.metrics)
}

// CalculateStats computes statistics from recorded durations
func (m *PluginMetrics) CalculateStats() (avg, p50, p95, p99 time.Duration) {
	if len(m.Durations) == 0 {
		return 0, 0, 0, 0
	}

	// Sort a copy for percentile calculation
	sorted := make([]int64, len(m.Durations))
	copySlice(sorted, m.Durations)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i] < sorted[j] })

	// Average
	var sum int64
	for _, d := range sorted {
		sum += d
	}
	avg = time.Duration(sum/int64(len(sorted))) * time.Microsecond

	// Percentiles
	p50 = time.Duration(sorted[len(sorted)*50/100]) * time.Microsecond
	p95 = time.Duration(sorted[len(sorted)*95/100]) * time.Microsecond
	p99 = time.Duration(sorted[len(sorted)*99/100]) * time.Microsecond

	return avg, p50, p95, p99
}

// copySlice is a helper to copy int64 slices
func copySlice(dst, src []int64) {
	for i, v := range src {
		dst[i] = v
	}
}

func metricsKey(name, pluginType string) string {
	if pluginType == "" || pluginType == string(PluginTypeUnknown) {
		return name
	}
	return pluginType + "/" + name
}

// MetricsLookupKey returns the key used to store metrics for a plugin.
func MetricsLookupKey(name string, pluginType PluginType) string {
	return metricsKey(name, string(pluginType))
}
