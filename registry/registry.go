package registry

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// Assignment represents a port assignment to a project
type Assignment struct {
	Port        int    `json:"port"`
	Description string `json:"description,omitempty"`
	Path        string `json:"path,omitempty"`
}

// BlockedPort represents a port or range of ports that should not be assigned
type BlockedPort struct {
	Ports       string `json:"ports"`
	Description string `json:"description,omitempty"`
}

// registryData represents the JSON structure of the registry file
type registryData struct {
	Assignments  []Assignment  `json:"assignments"`
	BlockedPorts []BlockedPort `json:"blockedPorts"`
}

// Registry manages port assignments and persistence
type Registry struct {
	path         string
	assignments  []Assignment
	blockedPorts []BlockedPort
}

// Custom errors
var (
	ErrPortAlreadyAssigned = errors.New("port is already assigned")
	ErrPortNotAssigned     = errors.New("port is not assigned")
	ErrPortBlocked         = errors.New("port is in blocked range")
	ErrNoPortsAvailable    = errors.New("no available ports found")
	ErrInvalidPortRange    = errors.New("invalid port range")
)

// New creates a new Registry instance, loading from file if it exists
func New(path string) (*Registry, error) {
	r := &Registry{
		path:         path,
		assignments:  []Assignment{},
		blockedPorts: []BlockedPort{},
	}

	// Load existing registry if file exists
	if _, err := os.Stat(path); err == nil {
		if err := r.load(); err != nil {
			return nil, fmt.Errorf("failed to load registry: %w", err)
		}
	}

	return r, nil
}

// Init initializes a new registry file with default blocked ports
func (r *Registry) Init() error {
	// Check if file already exists
	if _, err := os.Stat(r.path); err == nil {
		return fmt.Errorf("registry file already exists at %s", r.path)
	}

	// Set default blocked ports for common services
	r.blockedPorts = []BlockedPort{
		{Ports: "3306", Description: "MySQL default port"},
		{Ports: "5432", Description: "PostgreSQL default port"},
		{Ports: "6379", Description: "Redis default port"},
		{Ports: "8080", Description: "Common HTTP alternative port"},
		{Ports: "27017", Description: "MongoDB default port"},
	}

	r.assignments = []Assignment{}

	return r.Save()
}

// AssignPort assigns a specific port to a project
func (r *Registry) AssignPort(port int, description, path string) error {
	// Check if port is already assigned
	for _, a := range r.assignments {
		if a.Port == port {
			return fmt.Errorf("%w: port %d is already assigned to '%s'", ErrPortAlreadyAssigned, port, a.Description)
		}
	}

	// Check if port is blocked
	if r.isPortBlocked(port) {
		return fmt.Errorf("%w: port %d", ErrPortBlocked, port)
	}

	// Add assignment
	r.assignments = append(r.assignments, Assignment{
		Port:        port,
		Description: description,
		Path:        path,
	})

	return r.Save()
}

// AssignNextAvailable finds and assigns the next available port
func (r *Registry) AssignNextAvailable(description, path string) (int, error) {
	port := r.findNextAvailablePort()
	if port == -1 {
		return 0, ErrNoPortsAvailable
	}

	if err := r.AssignPort(port, description, path); err != nil {
		return 0, err
	}

	return port, nil
}

// UnassignPort releases a port assignment
func (r *Registry) UnassignPort(port int) error {
	found := false
	newAssignments := []Assignment{}

	for _, a := range r.assignments {
		if a.Port == port {
			found = true
		} else {
			newAssignments = append(newAssignments, a)
		}
	}

	if !found {
		return fmt.Errorf("%w: port %d", ErrPortNotAssigned, port)
	}

	r.assignments = newAssignments
	return r.Save()
}

// ListAssignments returns all current port assignments
func (r *Registry) ListAssignments() []Assignment {
	return r.assignments
}

// IsPortAvailable checks if a port can be assigned
func (r *Registry) IsPortAvailable(port int) bool {
	// Check assignments
	for _, a := range r.assignments {
		if a.Port == port {
			return false
		}
	}

	// Check blocked ports
	return !r.isPortBlocked(port)
}

// Save persists the registry to disk
func (r *Registry) Save() error {
	data := registryData{
		Assignments:  r.assignments,
		BlockedPorts: r.blockedPorts,
	}

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal registry: %w", err)
	}

	// Ensure directory exists
	dir := filepath.Dir(r.path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Write to temporary file first for atomic write
	tmpFile := r.path + ".tmp"
	if err := os.WriteFile(tmpFile, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write temporary file: %w", err)
	}

	// Rename temporary file to actual file (atomic on most systems)
	if err := os.Rename(tmpFile, r.path); err != nil {
		os.Remove(tmpFile) // Clean up on error
		return fmt.Errorf("failed to save registry: %w", err)
	}

	return nil
}

// load reads the registry from disk
func (r *Registry) load() error {
	data, err := os.ReadFile(r.path)
	if err != nil {
		return fmt.Errorf("failed to read registry file: %w", err)
	}

	var regData registryData
	if err := json.Unmarshal(data, &regData); err != nil {
		return fmt.Errorf("failed to unmarshal registry: %w", err)
	}

	r.assignments = regData.Assignments
	r.blockedPorts = regData.BlockedPorts

	return nil
}

// isPortBlocked checks if a port is in any blocked range
func (r *Registry) isPortBlocked(port int) bool {
	for _, bp := range r.blockedPorts {
		if isPortInRange(port, bp.Ports) {
			return true
		}
	}
	return false
}

// findNextAvailablePort finds the lowest available port starting from 3100
func (r *Registry) findNextAvailablePort() int {
	startPort := 3100
	maxPort := 65535

	for port := startPort; port <= maxPort; port++ {
		if r.IsPortAvailable(port) {
			return port
		}
	}

	return -1
}

// isPortInRange checks if a port is within a range specification
func isPortInRange(port int, rangeSpec string) bool {
	// Check if it's a range (contains hyphen)
	if strings.Contains(rangeSpec, "-") {
		parts := strings.Split(rangeSpec, "-")
		if len(parts) != 2 {
			return false
		}

		start, err1 := strconv.Atoi(strings.TrimSpace(parts[0]))
		end, err2 := strconv.Atoi(strings.TrimSpace(parts[1]))

		if err1 != nil || err2 != nil {
			return false
		}

		return port >= start && port <= end
	}

	// Single port
	singlePort, err := strconv.Atoi(strings.TrimSpace(rangeSpec))
	if err != nil {
		return false
	}

	return port == singlePort
}
