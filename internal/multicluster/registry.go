package multicluster

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"

	"gopkg.in/yaml.v3"
)

// Registry manages registered clusters.
type Registry struct {
	path     string
	clusters map[string]ClusterConfig
	mu       sync.RWMutex
}

// RegistryConfig is the on-disk format for the registry.
type RegistryConfig struct {
	Clusters []ClusterConfig `yaml:"clusters"`
}

// NewRegistry creates a new cluster registry.
func NewRegistry(path string) *Registry {
	return &Registry{
		path:     path,
		clusters: make(map[string]ClusterConfig),
	}
}

// Load reads the registry from disk.
func (r *Registry) Load() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	data, err := os.ReadFile(r.path)
	if os.IsNotExist(err) {
		// No clusters configured yet
		return nil
	}
	if err != nil {
		return fmt.Errorf("read registry: %w", err)
	}

	var config RegistryConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("parse registry: %w", err)
	}

	r.clusters = make(map[string]ClusterConfig)
	for _, c := range config.Clusters {
		r.clusters[c.Name] = c
	}

	return nil
}

// Save writes the registry to disk.
func (r *Registry) Save() error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.saveUnlocked()
}

func (r *Registry) saveUnlocked() error {
	// Ensure directory exists
	dir := filepath.Dir(r.path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create directory: %w", err)
	}

	// Build config
	var clusters []ClusterConfig
	for _, c := range r.clusters {
		clusters = append(clusters, c)
	}

	// Sort by name for consistent output
	sort.Slice(clusters, func(i, j int) bool {
		return clusters[i].Name < clusters[j].Name
	})

	config := RegistryConfig{Clusters: clusters}

	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("marshal registry: %w", err)
	}

	// Write with restricted permissions (contains API keys)
	if err := os.WriteFile(r.path, data, 0600); err != nil {
		return fmt.Errorf("write registry: %w", err)
	}

	return nil
}

// Add registers a new cluster.
func (r *Registry) Add(cluster ClusterConfig) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if cluster.Name == "" {
		return fmt.Errorf("cluster name is required")
	}

	if cluster.Endpoint == "" {
		return fmt.Errorf("cluster endpoint is required")
	}

	if _, exists := r.clusters[cluster.Name]; exists {
		return fmt.Errorf("cluster already exists: %s", cluster.Name)
	}

	r.clusters[cluster.Name] = cluster
	return r.saveUnlocked()
}

// Update modifies an existing cluster.
func (r *Registry) Update(cluster ClusterConfig) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.clusters[cluster.Name]; !exists {
		return fmt.Errorf("cluster not found: %s", cluster.Name)
	}

	r.clusters[cluster.Name] = cluster
	return r.saveUnlocked()
}

// Remove unregisters a cluster.
func (r *Registry) Remove(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.clusters[name]; !exists {
		return fmt.Errorf("cluster not found: %s", name)
	}

	delete(r.clusters, name)
	return r.saveUnlocked()
}

// Get retrieves a cluster by name.
func (r *Registry) Get(name string) (ClusterConfig, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	c, ok := r.clusters[name]
	return c, ok
}

// List returns all registered clusters.
func (r *Registry) List() []ClusterConfig {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var list []ClusterConfig
	for _, c := range r.clusters {
		list = append(list, c)
	}

	// Sort by name for consistent output
	sort.Slice(list, func(i, j int) bool {
		return list[i].Name < list[j].Name
	})

	return list
}

// Count returns the number of registered clusters.
func (r *Registry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return len(r.clusters)
}

// Exists checks if a cluster is registered.
func (r *Registry) Exists(name string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, exists := r.clusters[name]
	return exists
}

// GetByLabel returns clusters matching a label.
func (r *Registry) GetByLabel(key, value string) []ClusterConfig {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var matches []ClusterConfig
	for _, c := range r.clusters {
		if c.Labels != nil && c.Labels[key] == value {
			matches = append(matches, c)
		}
	}

	return matches
}

// SetLabel sets a label on a cluster.
func (r *Registry) SetLabel(name, key, value string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	cluster, exists := r.clusters[name]
	if !exists {
		return fmt.Errorf("cluster not found: %s", name)
	}

	if cluster.Labels == nil {
		cluster.Labels = make(map[string]string)
	}
	cluster.Labels[key] = value
	r.clusters[name] = cluster

	return r.saveUnlocked()
}

// RemoveLabel removes a label from a cluster.
func (r *Registry) RemoveLabel(name, key string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	cluster, exists := r.clusters[name]
	if !exists {
		return fmt.Errorf("cluster not found: %s", name)
	}

	if cluster.Labels != nil {
		delete(cluster.Labels, key)
		r.clusters[name] = cluster
	}

	return r.saveUnlocked()
}

// Rename changes the name of a cluster.
func (r *Registry) Rename(oldName, newName string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if newName == "" {
		return fmt.Errorf("new name is required")
	}

	cluster, exists := r.clusters[oldName]
	if !exists {
		return fmt.Errorf("cluster not found: %s", oldName)
	}

	if _, exists := r.clusters[newName]; exists {
		return fmt.Errorf("cluster already exists: %s", newName)
	}

	delete(r.clusters, oldName)
	cluster.Name = newName
	r.clusters[newName] = cluster

	return r.saveUnlocked()
}

// Clear removes all clusters.
func (r *Registry) Clear() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.clusters = make(map[string]ClusterConfig)
	return r.saveUnlocked()
}

// Import adds multiple clusters from a config.
func (r *Registry) Import(clusters []ClusterConfig) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, c := range clusters {
		if c.Name == "" || c.Endpoint == "" {
			continue
		}
		r.clusters[c.Name] = c
	}

	return r.saveUnlocked()
}

// Export returns all clusters as a config.
func (r *Registry) Export() RegistryConfig {
	return RegistryConfig{Clusters: r.List()}
}
