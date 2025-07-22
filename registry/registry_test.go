package registry

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	t.Run("creates new registry with non-existent file", func(t *testing.T) {
		tempFile := filepath.Join(t.TempDir(), "test.json")
		
		reg, err := New(tempFile)
		require.NoError(t, err)
		assert.NotNil(t, reg)
		assert.Equal(t, tempFile, reg.path)
		assert.Empty(t, reg.assignments)
		assert.Empty(t, reg.blockedPorts)
	})

	t.Run("loads existing registry file", func(t *testing.T) {
		tempFile := filepath.Join(t.TempDir(), "test.json")
		
		// Create a registry and save it
		reg1, err := New(tempFile)
		require.NoError(t, err)
		reg1.assignments = []Assignment{{Port: 8000, Description: "test"}}
		reg1.blockedPorts = []BlockedPort{{Ports: "9000-9010"}}
		require.NoError(t, reg1.Save())

		// Load it again
		reg2, err := New(tempFile)
		require.NoError(t, err)
		assert.Len(t, reg2.assignments, 1)
		assert.Equal(t, 8000, reg2.assignments[0].Port)
		assert.Len(t, reg2.blockedPorts, 1)
	})
}

func TestInit(t *testing.T) {
	t.Run("initializes new registry with defaults", func(t *testing.T) {
		tempFile := filepath.Join(t.TempDir(), "test.json")
		
		reg, err := New(tempFile)
		require.NoError(t, err)
		
		err = reg.Init()
		require.NoError(t, err)
		
		// Check file exists
		_, err = os.Stat(tempFile)
		require.NoError(t, err)
		
		// Check default blocked ports
		assert.NotEmpty(t, reg.blockedPorts)
		assert.Empty(t, reg.assignments)
		
		// Verify some expected blocked ports
		blockedDescriptions := make(map[string]bool)
		for _, bp := range reg.blockedPorts {
			blockedDescriptions[bp.Description] = true
		}
		assert.True(t, blockedDescriptions["MySQL default port"])
		assert.True(t, blockedDescriptions["PostgreSQL default port"])
		assert.True(t, blockedDescriptions["Common Ruby on Rails ports"])
	})

	t.Run("fails if file already exists", func(t *testing.T) {
		tempFile := filepath.Join(t.TempDir(), "test.json")
		
		reg, err := New(tempFile)
		require.NoError(t, err)
		
		// Initialize once
		err = reg.Init()
		require.NoError(t, err)
		
		// Try to initialize again
		err = reg.Init()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already exists")
	})
}

func TestAssignPort(t *testing.T) {
	t.Run("assigns available port", func(t *testing.T) {
		reg := createTestRegistry(t)
		
		err := reg.AssignPort(8000, "test project", "/path/to/project")
		require.NoError(t, err)
		
		assert.Len(t, reg.assignments, 1)
		assert.Equal(t, 8000, reg.assignments[0].Port)
		assert.Equal(t, "test project", reg.assignments[0].Description)
		assert.Equal(t, "/path/to/project", reg.assignments[0].Path)
	})

	t.Run("fails on already assigned port", func(t *testing.T) {
		reg := createTestRegistry(t)
		
		err := reg.AssignPort(8000, "project1", "")
		require.NoError(t, err)
		
		err = reg.AssignPort(8000, "project2", "")
		assert.ErrorIs(t, err, ErrPortAlreadyAssigned)
		assert.Contains(t, err.Error(), "project1")
	})

	t.Run("fails on blocked port", func(t *testing.T) {
		reg := createTestRegistry(t)
		reg.blockedPorts = []BlockedPort{{Ports: "3000-3010"}}
		
		err := reg.AssignPort(3005, "project", "")
		assert.ErrorIs(t, err, ErrPortBlocked)
	})
}

func TestAssignNextAvailable(t *testing.T) {
	t.Run("assigns first available port from 3100", func(t *testing.T) {
		reg := createTestRegistry(t)
		
		port, err := reg.AssignNextAvailable("test", "")
		require.NoError(t, err)
		assert.Equal(t, 3100, port)
		assert.Len(t, reg.assignments, 1)
	})

	t.Run("skips assigned and blocked ports", func(t *testing.T) {
		reg := createTestRegistry(t)
		reg.assignments = []Assignment{{Port: 3100}, {Port: 3101}}
		reg.blockedPorts = []BlockedPort{{Ports: "3102-3105"}}
		
		port, err := reg.AssignNextAvailable("test", "")
		require.NoError(t, err)
		assert.Equal(t, 3106, port)
	})
}

func TestUnassignPort(t *testing.T) {
	t.Run("unassigns existing port", func(t *testing.T) {
		reg := createTestRegistry(t)
		reg.assignments = []Assignment{
			{Port: 8000, Description: "project1"},
			{Port: 8001, Description: "project2"},
		}
		
		err := reg.UnassignPort(8000)
		require.NoError(t, err)
		
		assert.Len(t, reg.assignments, 1)
		assert.Equal(t, 8001, reg.assignments[0].Port)
	})

	t.Run("fails on non-assigned port", func(t *testing.T) {
		reg := createTestRegistry(t)
		
		err := reg.UnassignPort(8000)
		assert.ErrorIs(t, err, ErrPortNotAssigned)
	})
}

func TestIsPortAvailable(t *testing.T) {
	reg := createTestRegistry(t)
	reg.assignments = []Assignment{{Port: 8000}}
	reg.blockedPorts = []BlockedPort{{Ports: "9000-9010"}}
	
	tests := []struct {
		port      int
		available bool
		desc      string
	}{
		{7999, true, "unassigned and unblocked port"},
		{8000, false, "assigned port"},
		{9005, false, "blocked port in range"},
		{9011, true, "port outside blocked range"},
	}
	
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			assert.Equal(t, tt.available, reg.IsPortAvailable(tt.port))
		})
	}
}

func TestPortRangeParsing(t *testing.T) {
	tests := []struct {
		port      int
		rangeSpec string
		inRange   bool
		desc      string
	}{
		{3005, "3000-3010", true, "port in range"},
		{3000, "3000-3010", true, "start of range"},
		{3010, "3000-3010", true, "end of range"},
		{2999, "3000-3010", false, "before range"},
		{3011, "3000-3010", false, "after range"},
		{8080, "8080", true, "single port match"},
		{8081, "8080", false, "single port no match"},
		{5000, "invalid-range", false, "invalid range format"},
		{5000, "abc-def", false, "non-numeric range"},
	}
	
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			assert.Equal(t, tt.inRange, isPortInRange(tt.port, tt.rangeSpec))
		})
	}
}

func TestSaveAndLoad(t *testing.T) {
	t.Run("saves and loads registry data", func(t *testing.T) {
		tempFile := filepath.Join(t.TempDir(), "test.json")
		
		// Create and populate registry
		reg1, err := New(tempFile)
		require.NoError(t, err)
		
		reg1.assignments = []Assignment{
			{Port: 8000, Description: "project1", Path: "/path1"},
			{Port: 8001, Description: "project2"},
		}
		reg1.blockedPorts = []BlockedPort{
			{Ports: "3000-3010", Description: "Rails ports"},
			{Ports: "3306", Description: "MySQL"},
		}
		
		err = reg1.Save()
		require.NoError(t, err)
		
		// Load into new registry
		reg2, err := New(tempFile)
		require.NoError(t, err)
		
		assert.Equal(t, reg1.assignments, reg2.assignments)
		assert.Equal(t, reg1.blockedPorts, reg2.blockedPorts)
	})

	t.Run("handles missing directory", func(t *testing.T) {
		tempDir := t.TempDir()
		tempFile := filepath.Join(tempDir, "subdir", "test.json")
		
		reg, err := New(tempFile)
		require.NoError(t, err)
		
		reg.assignments = []Assignment{{Port: 8000}}
		err = reg.Save()
		require.NoError(t, err)
		
		// Verify directory was created
		_, err = os.Stat(filepath.Dir(tempFile))
		require.NoError(t, err)
	})
}

func createTestRegistry(t *testing.T) *Registry {
	tempFile := filepath.Join(t.TempDir(), "test.json")
	reg, err := New(tempFile)
	require.NoError(t, err)
	return reg
}