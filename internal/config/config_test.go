package config

import (
	"os"
	"testing"
)

func TestLoad(t *testing.T) {
	// Test loading config with environment variables
	os.Setenv("DIREITO_LUX_SERVER_PORT", "8080")
	os.Setenv("DIREITO_LUX_DATABASE_HOST", "localhost")
	
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}
	
	if cfg.Server.Port != "8080" {
		t.Errorf("Expected port 8080, got %s", cfg.Server.Port)
	}
	
	if cfg.Database.Host != "localhost" {
		t.Errorf("Expected database host localhost, got %s", cfg.Database.Host)
	}
	
	// Clean up
	os.Unsetenv("DIREITO_LUX_SERVER_PORT")
	os.Unsetenv("DIREITO_LUX_DATABASE_HOST")
}

func TestGetDSN(t *testing.T) {
	cfg := &Config{
		Database: DatabaseConfig{
			Host:     "localhost",
			Port:     "5432",
			User:     "test",
			Password: "test",
			DBName:   "testdb",
			SSLMode:  "disable",
		},
	}
	
	expected := "host=localhost port=5432 user=test password=test dbname=testdb sslmode=disable"
	dsn := cfg.GetDSN()
	
	if dsn != expected {
		t.Errorf("Expected DSN %s, got %s", expected, dsn)
	}
}

func TestGetRedisAddr(t *testing.T) {
	cfg := &Config{
		Redis: RedisConfig{
			Host: "localhost",
			Port: "6379",
		},
	}
	
	expected := "localhost:6379"
	addr := cfg.GetRedisAddr()
	
	if addr != expected {
		t.Errorf("Expected Redis address %s, got %s", expected, addr)
	}
}

func TestIsProduction(t *testing.T) {
	cfg := &Config{
		Server: ServerConfig{
			Mode: "release",
		},
	}
	
	if !cfg.IsProduction() {
		t.Error("Expected IsProduction to return true for release mode")
	}
	
	cfg.Server.Mode = "debug"
	if cfg.IsProduction() {
		t.Error("Expected IsProduction to return false for debug mode")
	}
}