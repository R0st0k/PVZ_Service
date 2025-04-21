package service

import (
	"context"
	"errors"
	"log/slog"
	"pvz-service/internal/config"
	e "pvz-service/internal/errors"
	"pvz-service/internal/models"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockAuthRepository struct {
	mock.Mock
}

func (m *MockAuthRepository) CloseConnection() {
	m.Called()
}

func (m *MockAuthRepository) CreateUser(ctx context.Context, email, password string, role models.UserRole) (*models.User, error) {
	args := m.Called(ctx, email, password, role)
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockAuthRepository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	args := m.Called(ctx, email)
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockAuthRepository) GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockAuthRepository) VerifyPassword(ctx context.Context, email, password string) (bool, error) {
	args := m.Called(ctx, email, password)
	return args.Bool(0), args.Error(1)
}

func TestAuthService_Register(t *testing.T) {
	tests := []struct {
		name        string
		email       string
		password    string
		role        models.UserRole
		mockSetup   func(*MockAuthRepository)
		expectError error
	}{
		{
			name:     "Success",
			email:    "test@example.com",
			password: "password",
			role:     models.UserRoleModerator,
			mockSetup: func(m *MockAuthRepository) {
				m.On("CreateUser", mock.Anything, "test@example.com", "password", models.UserRoleModerator).
					Return(&models.User{Email: "test@example.com", Role: models.UserRoleModerator}, nil)
			},
			expectError: nil,
		},
		{
			name:     "Create user error",
			email:    "test@example.com",
			password: "password",
			role:     models.UserRoleModerator,
			mockSetup: func(m *MockAuthRepository) {
				m.On("CreateUser", mock.Anything, "test@example.com", "password", models.UserRoleModerator).
					Return(&models.User{}, e.ErrAlreadyExists())
			},
			expectError: e.ErrAlreadyExists(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockAuthRepository)
			tt.mockSetup(mockRepo)

			cfg := &config.Config{
				JWT: config.JWT{
					SecretKey: "test_secret",
					ExpiresIn: time.Hour,
				},
			}
			log := slog.Default()

			service := NewAuthService(mockRepo, cfg, log)
			token, err := service.Register(context.Background(), tt.email, tt.password, tt.role)

			if tt.expectError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectError, err)
				assert.Empty(t, token)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, token)

				// Verify token can be parsed and contains correct data
				parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
					return []byte(cfg.JWT.SecretKey), nil
				})
				assert.NoError(t, err)

				claims := parsedToken.Claims.(jwt.MapClaims)
				assert.Equal(t, tt.email, claims["email"])
				assert.Equal(t, string(tt.role), claims["role"])
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestAuthService_Login(t *testing.T) {
	tests := []struct {
		name        string
		email       string
		password    string
		mockSetup   func(*MockAuthRepository)
		expectError error
	}{
		{
			name:     "Success",
			email:    "test@example.com",
			password: "password",
			mockSetup: func(m *MockAuthRepository) {
				m.On("VerifyPassword", mock.Anything, "test@example.com", "password").
					Return(true, nil)
				m.On("GetUserByEmail", mock.Anything, "test@example.com").
					Return(&models.User{
						Email: "test@example.com",
						Role:  models.UserRoleModerator,
					}, nil)
			},
			expectError: nil,
		},
		{
			name:     "Wrong password",
			email:    "test@example.com",
			password: "wrong",
			mockSetup: func(m *MockAuthRepository) {
				m.On("VerifyPassword", mock.Anything, "test@example.com", "wrong").
					Return(false, nil)
			},
			expectError: e.ErrInvalidCredentials(),
		},
		{
			name:     "Verify error",
			email:    "test@example.com",
			password: "password",
			mockSetup: func(m *MockAuthRepository) {
				m.On("VerifyPassword", mock.Anything, "test@example.com", "password").
					Return(false, errors.New("verify error"))
			},
			expectError: errors.New("verify error"),
		},
		{
			name:     "Get user error",
			email:    "test@example.com",
			password: "password",
			mockSetup: func(m *MockAuthRepository) {
				m.On("VerifyPassword", mock.Anything, "test@example.com", "password").
					Return(true, nil)
				m.On("GetUserByEmail", mock.Anything, "test@example.com").
					Return(&models.User{}, errors.New("get user error"))
			},
			expectError: errors.New("get user error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockAuthRepository)
			tt.mockSetup(mockRepo)

			cfg := &config.Config{
				JWT: config.JWT{
					SecretKey: "test_secret",
					ExpiresIn: time.Hour,
				},
			}
			log := slog.Default()

			service := NewAuthService(mockRepo, cfg, log)
			token, err := service.Login(context.Background(), tt.email, tt.password)

			if tt.expectError != nil {
				assert.Error(t, err)
				if tt.expectError != e.ErrInvalidCredentials() {
					assert.Equal(t, tt.expectError, err)
				}
				assert.Empty(t, token)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, token)

				// Verify token can be parsed and contains correct data
				parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
					return []byte(cfg.JWT.SecretKey), nil
				})
				assert.NoError(t, err)

				claims := parsedToken.Claims.(jwt.MapClaims)
				assert.Equal(t, tt.email, claims["email"])
				assert.Equal(t, string(models.UserRoleModerator), claims["role"])
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestAuthService_DummyLogin(t *testing.T) {
	cfg := &config.Config{
		JWT: config.JWT{
			SecretKey: "test_secret",
			ExpiresIn: time.Hour,
		},
	}
	log := slog.Default()
	service := NewAuthService(nil, cfg, log)

	tests := []struct {
		name string
		role models.UserRole
	}{
		{
			name: "Employee role",
			role: models.UserRoleEmployee,
		},
		{
			name: "Moderator role",
			role: models.UserRoleModerator,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := service.DummyLogin(tt.role)
			assert.NoError(t, err)
			assert.NotEmpty(t, token)

			// Verify token can be parsed and contains correct role
			parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
				return []byte(cfg.JWT.SecretKey), nil
			})
			assert.NoError(t, err)

			claims := parsedToken.Claims.(jwt.MapClaims)
			assert.Contains(t, claims["email"].(string), "dummy_")
			assert.Equal(t, string(tt.role), claims["role"])
		})
	}
}

func TestAuthService_ParseToken(t *testing.T) {
	cfg := &config.Config{
		JWT: config.JWT{
			SecretKey: "test_secret",
			ExpiresIn: time.Hour,
		},
	}
	log := slog.Default()
	service := NewAuthService(nil, cfg, log)

	// Generate valid tokens
	validToken, err := service.generateToken("test@example.com", models.UserRoleModerator)
	assert.NoError(t, err)

	expiredToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"email": "test@example.com",
		"role":  models.UserRoleModerator,
		"exp":   time.Now().Add(-time.Hour).Unix(),
	}).SignedString([]byte(cfg.JWT.SecretKey))
	assert.NoError(t, err)

	tests := []struct {
		name        string
		token       string
		expectError bool
	}{
		{
			name:        "Valid token",
			token:       validToken,
			expectError: false,
		},
		{
			name:        "Invalid token",
			token:       "invalid.token.string",
			expectError: true,
		},
		{
			name:        "Expired token",
			token:       expiredToken,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := service.ParseToken(tt.token)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestAuthService_GetUserFromToken(t *testing.T) {
	cfg := &config.Config{
		JWT: config.JWT{
			SecretKey: "test_secret",
			ExpiresIn: time.Hour,
		},
	}
	log := slog.Default()
	service := NewAuthService(nil, cfg, log)

	// Generate valid tokens
	validToken, err := service.generateToken("test@example.com", models.UserRoleModerator)
	assert.NoError(t, err)

	invalidToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"email": "test@example.com",
		// missing role
	}).SignedString([]byte(cfg.JWT.SecretKey))
	assert.NoError(t, err)

	tests := []struct {
		name        string
		token       string
		expectEmail string
		expectRole  models.UserRole
		expectError bool
	}{
		{
			name:        "Valid token",
			token:       validToken,
			expectEmail: "test@example.com",
			expectRole:  models.UserRoleModerator,
			expectError: false,
		},
		{
			name:        "Invalid token - missing role",
			token:       invalidToken,
			expectError: true,
		},
		{
			name:        "Malformed token",
			token:       "malformed.token",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			email, role, err := service.GetUserFromToken(tt.token)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectEmail, email)
				assert.Equal(t, tt.expectRole, role)
			}
		})
	}
}
