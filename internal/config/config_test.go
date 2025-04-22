package config

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigDefaults(t *testing.T) {
	t.Parallel()

	cfg := Config{}

	assert.Equal(t, "", cfg.HTTP.Host)
	assert.Equal(t, "", cfg.HTTP.Port)
	assert.Equal(t, time.Duration(0), cfg.HTTP.Timeout)
	assert.Equal(t, time.Duration(0), cfg.HTTP.IdleTimeout)

	assert.Equal(t, false, cfg.GRPC.IsAble)
	assert.Equal(t, "", cfg.GRPC.Host)
	assert.Equal(t, "", cfg.GRPC.Port)

	assert.Equal(t, false, cfg.Prometheus.IsAble)
	assert.Equal(t, "", cfg.Prometheus.Host)
	assert.Equal(t, "", cfg.Prometheus.Port)
	assert.Equal(t, time.Duration(0), cfg.Prometheus.Timeout)
	assert.Equal(t, time.Duration(0), cfg.Prometheus.IdleTimeout)

	assert.Equal(t, "", cfg.Database.Protocol)
	assert.Equal(t, "", cfg.Database.Host)
	assert.Equal(t, "", cfg.Database.Port)
	assert.Equal(t, "", cfg.Database.User)
	assert.Equal(t, "", cfg.Database.Password)
	assert.Equal(t, "", cfg.Database.Name)
	assert.Equal(t, "", cfg.Database.SSLMode)

	assert.Equal(t, "", cfg.JWT.SecretKey)
	assert.Equal(t, time.Duration(0), cfg.JWT.ExpiresIn)
}

func TestLoad(t *testing.T) {
	tests := []struct {
		name          string
		prepare       func()
		expectedError string
		wantErr       bool
	}{
		{
			name: "successful load",
			prepare: func() {
				os.Setenv("CONFIG_PATH", "testdata/config.yaml")
			},
			wantErr: false,
		},
		{
			name: "missing config path",
			prepare: func() {
				os.Unsetenv("CONFIG_PATH")
			},
			expectedError: "CONFIG_PATH is not set",
			wantErr:       true,
		},
		{
			name: "config file not exists",
			prepare: func() {
				os.Setenv("CONFIG_PATH", "testdata/nonexistent.yaml")
			},
			expectedError: "config file does not exist",
			wantErr:       true,
		},
		{
			name: "invalid config file",
			prepare: func() {
				os.Setenv("CONFIG_PATH", "testdata/invalid_config.yaml")
			},
			expectedError: "config file does not exist",
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Backup and restore env vars
			oldConfigPath := os.Getenv("CONFIG_PATH")
			defer os.Setenv("CONFIG_PATH", oldConfigPath)

			tt.prepare()

			cfg, err := Load()

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, cfg)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, cfg)
				// Add assertions for loaded values
				assert.Equal(t, "0.0.0.0", cfg.HTTP.Host)
				assert.Equal(t, "8080", cfg.HTTP.Port)
				assert.Equal(t, 4*time.Second, cfg.HTTP.Timeout)
				assert.Equal(t, 30*time.Second, cfg.HTTP.IdleTimeout)
			}
		})
	}
}

func TestLoadWithEnvVars(t *testing.T) {
	// Backup and restore env vars
	oldConfigPath := os.Getenv("CONFIG_PATH")
	defer os.Setenv("CONFIG_PATH", oldConfigPath)

	// Set test env vars
	os.Setenv("CONFIG_PATH", "testdata/config.yaml")
	os.Setenv("HTTP_PORT", "9090")
	os.Setenv("DB_HOST", "localhost")
	os.Setenv("JWT_SECRET", "test_secret")

	cfg, err := Load()
	require.NoError(t, err)
	require.NotNil(t, cfg)

	// Check that env vars override config file values
	assert.Equal(t, "9090", cfg.HTTP.Port)
	assert.Equal(t, "localhost", cfg.Database.Host)
	assert.Equal(t, "test_secret", cfg.JWT.SecretKey)
}
