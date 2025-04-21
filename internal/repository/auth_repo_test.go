package repository

import (
	"context"
	"errors"
	"log/slog"
	"pvz-service/internal/config"
	"pvz-service/internal/models"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockAuthRepository implements AuthRepository for testing
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

// MockPostgresAuthGetter mocks the postgres repository getter
type MockPostgresAuthGetter struct {
	mock.Mock
}

func (m *MockPostgresAuthGetter) GetRepository(cfg *config.Config) (AuthRepository, error) {
	args := m.Called(cfg)
	return args.Get(0).(AuthRepository), args.Error(1)
}

func TestCreateAuthRepo(t *testing.T) {
	// Save original PostgresGetter and restore after test
	originalPostgresGetter := PostgresGetter
	defer func() { PostgresGetter = originalPostgresGetter }()

	tests := []struct {
		name          string
		cfg           *config.Config
		mockSetup     func(*MockPostgresAuthGetter)
		expectError   bool
		expectNilRepo bool
	}{
		{
			name: "Success with postgres",
			cfg: &config.Config{
				Database: config.Database{
					Protocol: "postgres",
				},
			},
			mockSetup: func(m *MockPostgresAuthGetter) {
				mockRepo := new(MockAuthRepository)
				m.On("GetRepository", mock.Anything).Return(mockRepo, nil)
			},
			expectError:   false,
			expectNilRepo: false,
		},
		{
			name: "Postgres error",
			cfg: &config.Config{
				Database: config.Database{
					Protocol: "postgres",
				},
			},
			mockSetup: func(m *MockPostgresAuthGetter) {
				m.On("GetRepository", mock.Anything).Return(&MockAuthRepository{}, errors.New("connection failed"))
			},
			expectError:   true,
			expectNilRepo: true,
		},
		{
			name: "Unknown protocol",
			cfg: &config.Config{
				Database: config.Database{
					Protocol: "unknown",
				},
			},
			mockSetup:     func(m *MockPostgresAuthGetter) {},
			expectError:   true,
			expectNilRepo: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockGetter := new(MockPostgresAuthGetter)
			tt.mockSetup(mockGetter)
			PostgresGetter = func(cfg *config.Config) (interface{}, error) {
				return mockGetter.GetRepository(cfg)
			}

			log := slog.Default()
			repo, err := CreateAuthRepo(tt.cfg, log)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			if tt.expectNilRepo {
				assert.Nil(t, repo)
			} else {
				assert.NotNil(t, repo)
			}

			mockGetter.AssertExpectations(t)
		})
	}
}
