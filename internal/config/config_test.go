package config

import (
	"os"
	"testing"
)

func TestLoad(t *testing.T) {
	// Save original environment
	originalPort := os.Getenv("PORT")
	originalDB := os.Getenv("DATABASE_URL")
	originalEnv := os.Getenv("ENV")

	// Clean up after test
	defer func() {
		os.Setenv("PORT", originalPort)
		os.Setenv("DATABASE_URL", originalDB)
		os.Setenv("ENV", originalEnv)
	}()

	t.Run("loads defaults when no env vars set", func(t *testing.T) {
		// Clear environment variables
		os.Unsetenv("PORT")
		os.Unsetenv("DATABASE_URL")
		os.Unsetenv("ENV")

		cfg := Load()

		if cfg.Port != "8080" {
			t.Errorf("Expected default port 8080, got %s", cfg.Port)
		}

		expectedDB := "postgres://postgres:password@localhost:5433/pepo_dev?sslmode=disable"
		if cfg.DatabaseURL != expectedDB {
			t.Errorf("Expected default database URL, got %s", cfg.DatabaseURL)
		}

		if cfg.Environment != "development" {
			t.Errorf("Expected default environment 'development', got %s", cfg.Environment)
		}
	})

	t.Run("loads from environment variables", func(t *testing.T) {
		// Set environment variables
		os.Setenv("PORT", "3000")
		os.Setenv("DATABASE_URL", "postgres://test:test@localhost:5432/test_db")
		os.Setenv("ENV", "production")

		cfg := Load()

		if cfg.Port != "3000" {
			t.Errorf("Expected port 3000, got %s", cfg.Port)
		}

		if cfg.DatabaseURL != "postgres://test:test@localhost:5432/test_db" {
			t.Errorf("Expected custom database URL, got %s", cfg.DatabaseURL)
		}

		if cfg.Environment != "production" {
			t.Errorf("Expected environment 'production', got %s", cfg.Environment)
		}
	})
}

func TestIsDevelopment(t *testing.T) {
	tests := []struct {
		name        string
		environment string
		expected    bool
	}{
		{"development environment", "development", true},
		{"dev environment", "dev", false},
		{"production environment", "production", false},
		{"test environment", "test", false},
		{"empty environment", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{Environment: tt.environment}
			result := cfg.IsDevelopment()
			if result != tt.expected {
				t.Errorf("IsDevelopment() = %v, expected %v for environment %s",
					result, tt.expected, tt.environment)
			}
		})
	}
}

func TestIsProduction(t *testing.T) {
	tests := []struct {
		name        string
		environment string
		expected    bool
	}{
		{"production environment", "production", true},
		{"prod environment", "prod", false},
		{"development environment", "development", false},
		{"test environment", "test", false},
		{"empty environment", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{Environment: tt.environment}
			result := cfg.IsProduction()
			if result != tt.expected {
				t.Errorf("IsProduction() = %v, expected %v for environment %s",
					result, tt.expected, tt.environment)
			}
		})
	}
}

func TestGetEnv(t *testing.T) {
	// Save original value
	original := os.Getenv("TEST_VAR")
	defer func() {
		if original != "" {
			os.Setenv("TEST_VAR", original)
		} else {
			os.Unsetenv("TEST_VAR")
		}
	}()

	t.Run("returns environment value when set", func(t *testing.T) {
		os.Setenv("TEST_VAR", "test_value")
		result := getEnv("TEST_VAR", "default_value")
		if result != "test_value" {
			t.Errorf("Expected 'test_value', got '%s'", result)
		}
	})

	t.Run("returns fallback when env var not set", func(t *testing.T) {
		os.Unsetenv("TEST_VAR")
		result := getEnv("TEST_VAR", "default_value")
		if result != "default_value" {
			t.Errorf("Expected 'default_value', got '%s'", result)
		}
	})

	t.Run("returns fallback when env var is empty", func(t *testing.T) {
		os.Setenv("TEST_VAR", "")
		result := getEnv("TEST_VAR", "default_value")
		if result != "default_value" {
			t.Errorf("Expected 'default_value', got '%s'", result)
		}
	})
}

func TestConfigStructure(t *testing.T) {
	cfg := &Config{
		Port:        "8080",
		DatabaseURL: "postgres://localhost/test",
		Environment: "test",
	}

	// Test that all fields are accessible
	if cfg.Port == "" {
		t.Error("Port field should be accessible")
	}
	if cfg.DatabaseURL == "" {
		t.Error("DatabaseURL field should be accessible")
	}
	if cfg.Environment == "" {
		t.Error("Environment field should be accessible")
	}
}

// Benchmark the Load function to ensure it's performant
func BenchmarkLoad(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Load()
	}
}
