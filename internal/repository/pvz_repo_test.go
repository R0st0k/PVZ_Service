package repository

import (
	"context"
	"errors"
	"log/slog"
	"pvz-service/internal/config"
	"pvz-service/internal/models"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockPVZRepository implements PVZRepository for testing
type MockPVZRepository struct {
	mock.Mock
}

func (m *MockPVZRepository) CloseConnection() {
	m.Called()
}

func (m *MockPVZRepository) InsertPVZ(ctx context.Context, pvz *models.PVZ) error {
	args := m.Called(ctx, pvz)
	return args.Error(0)
}

func (m *MockPVZRepository) CheckPVZ(ctx context.Context, pvzID uuid.UUID) (bool, error) {
	args := m.Called(ctx, pvzID)
	return args.Bool(0), args.Error(1)
}

func (m *MockPVZRepository) GetCityID(ctx context.Context, cityName string) (int, error) {
	args := m.Called(ctx, cityName)
	return args.Int(0), args.Error(1)
}

func (m *MockPVZRepository) InsertReception(ctx context.Context, reception *models.Reception) error {
	args := m.Called(ctx, reception)
	return args.Error(0)
}

func (m *MockPVZRepository) GetActiveReception(ctx context.Context, pvzID uuid.UUID) (*models.Reception, error) {
	args := m.Called(ctx, pvzID)
	return args.Get(0).(*models.Reception), args.Error(1)
}

func (m *MockPVZRepository) UpdateReceptionStatus(ctx context.Context, receptionID uuid.UUID, status models.ReceptionStatus) error {
	args := m.Called(ctx, receptionID, status)
	return args.Error(0)
}

func (m *MockPVZRepository) InsertProduct(ctx context.Context, product *models.Product) error {
	args := m.Called(ctx, product)
	return args.Error(0)
}

func (m *MockPVZRepository) GetLastProduct(ctx context.Context, receptionID uuid.UUID) (*models.Product, error) {
	args := m.Called(ctx, receptionID)
	return args.Get(0).(*models.Product), args.Error(1)
}

func (m *MockPVZRepository) DeleteProduct(ctx context.Context, productID uuid.UUID) error {
	args := m.Called(ctx, productID)
	return args.Error(0)
}

func (m *MockPVZRepository) GetProductTypeID(ctx context.Context, productTypeName string) (int, error) {
	args := m.Called(ctx, productTypeName)
	return args.Int(0), args.Error(1)
}

func (m *MockPVZRepository) GetPVZs(ctx context.Context, from, to time.Time, limit, offset int) ([]models.PVZ, error) {
	args := m.Called(ctx, from, to, limit, offset)
	return args.Get(0).([]models.PVZ), args.Error(1)
}

func (m *MockPVZRepository) GetPVZsWithNoFilter(ctx context.Context) ([]models.PVZ, error) {
	args := m.Called(ctx)
	return args.Get(0).([]models.PVZ), args.Error(1)
}

func (m *MockPVZRepository) GetReceptionsForPVZs(ctx context.Context, pvzIDs []uuid.UUID, from, to time.Time) ([]models.Reception, error) {
	args := m.Called(ctx, pvzIDs, from, to)
	return args.Get(0).([]models.Reception), args.Error(1)
}

func (m *MockPVZRepository) GetProductsForReceptions(ctx context.Context, receptionIDs []uuid.UUID) ([]models.Product, error) {
	args := m.Called(ctx, receptionIDs)
	return args.Get(0).([]models.Product), args.Error(1)
}

// MockPostgresGetter mocks the postgres repository getter
type MockPostgresPVZGetter struct {
	mock.Mock
}

func (m *MockPostgresPVZGetter) GetRepository(cfg *config.Config) (PVZRepository, error) {
	args := m.Called(cfg)
	return args.Get(0).(PVZRepository), args.Error(1)
}

func TestCreatePVZRepo(t *testing.T) {
	// Save original PostgresGetter and restore after test
	originalPostgresGetter := PostgresGetter
	defer func() { PostgresGetter = originalPostgresGetter }()

	tests := []struct {
		name          string
		cfg           *config.Config
		mockSetup     func(*MockPostgresPVZGetter)
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
			mockSetup: func(m *MockPostgresPVZGetter) {
				mockRepo := new(MockPVZRepository)
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
			mockSetup: func(m *MockPostgresPVZGetter) {
				m.On("GetRepository", mock.Anything).Return(&MockPVZRepository{}, errors.New("connection failed"))
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
			mockSetup:     func(m *MockPostgresPVZGetter) {},
			expectError:   true,
			expectNilRepo: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockGetter := new(MockPostgresPVZGetter)
			tt.mockSetup(mockGetter)
			PostgresGetter = func(cfg *config.Config) (interface{}, error) {
				return mockGetter.GetRepository(cfg)
			}

			log := slog.Default()
			repo, err := CreatePVZRepo(tt.cfg, log)

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

func TestPVZRepositoryInterface(t *testing.T) {
	mockRepo := new(MockPVZRepository)
	ctx := context.Background()
	testUUID := uuid.New()
	now := time.Now()
	testPVZ := &models.PVZ{ID: testUUID}
	testReception := &models.Reception{ID: testUUID}
	testProduct := &models.Product{ID: testUUID}
	testPVZs := []models.PVZ{*testPVZ}
	testReceptions := []models.Reception{*testReception}
	testProducts := []models.Product{*testProduct}

	// Test CloseConnection
	mockRepo.On("CloseConnection").Once()
	mockRepo.CloseConnection()
	mockRepo.AssertCalled(t, "CloseConnection")

	// Test PVZ operations
	mockRepo.On("InsertPVZ", ctx, testPVZ).Return(nil).Once()
	assert.NoError(t, mockRepo.InsertPVZ(ctx, testPVZ))

	mockRepo.On("CheckPVZ", ctx, testUUID).Return(true, nil).Once()
	exists, err := mockRepo.CheckPVZ(ctx, testUUID)
	assert.NoError(t, err)
	assert.True(t, exists)

	mockRepo.On("GetCityID", ctx, "Москва").Return(1, nil).Once()
	cityID, err := mockRepo.GetCityID(ctx, "Москва")
	assert.NoError(t, err)
	assert.Equal(t, 1, cityID)

	mockRepo.On("GetPVZsWithNoFilter", ctx).Return(testPVZs, nil).Once()
	pvzsNoFilter, err := mockRepo.GetPVZsWithNoFilter(ctx)
	assert.NoError(t, err)
	assert.Equal(t, testPVZs, pvzsNoFilter)

	// Test Reception operations
	mockRepo.On("InsertReception", ctx, testReception).Return(nil).Once()
	assert.NoError(t, mockRepo.InsertReception(ctx, testReception))

	mockRepo.On("GetActiveReception", ctx, testUUID).Return(testReception, nil).Once()
	reception, err := mockRepo.GetActiveReception(ctx, testUUID)
	assert.NoError(t, err)
	assert.Equal(t, testReception, reception)

	mockRepo.On("UpdateReceptionStatus", ctx, testUUID, models.ReceptionStatusClose).Return(nil).Once()
	assert.NoError(t, mockRepo.UpdateReceptionStatus(ctx, testUUID, models.ReceptionStatusClose))

	// Test Product operations
	mockRepo.On("InsertProduct", ctx, testProduct).Return(nil).Once()
	assert.NoError(t, mockRepo.InsertProduct(ctx, testProduct))

	mockRepo.On("GetLastProduct", ctx, testUUID).Return(testProduct, nil).Once()
	product, err := mockRepo.GetLastProduct(ctx, testUUID)
	assert.NoError(t, err)
	assert.Equal(t, testProduct, product)

	mockRepo.On("DeleteProduct", ctx, testUUID).Return(nil).Once()
	assert.NoError(t, mockRepo.DeleteProduct(ctx, testUUID))

	// Test Product type operations
	mockRepo.On("GetProductTypeID", ctx, "Electronics").Return(2, nil).Once()
	productTypeID, err := mockRepo.GetProductTypeID(ctx, "Electronics")
	assert.NoError(t, err)
	assert.Equal(t, 2, productTypeID)

	// Test Query operations
	mockRepo.On("GetPVZs", ctx, now, now.Add(24*time.Hour), 10, 0).Return(testPVZs, nil).Once()
	pvzs, err := mockRepo.GetPVZs(ctx, now, now.Add(24*time.Hour), 10, 0)
	assert.NoError(t, err)
	assert.Equal(t, testPVZs, pvzs)

	mockRepo.On("GetReceptionsForPVZs", ctx, []uuid.UUID{testUUID}, now, now.Add(24*time.Hour)).Return(testReceptions, nil).Once()
	receptions, err := mockRepo.GetReceptionsForPVZs(ctx, []uuid.UUID{testUUID}, now, now.Add(24*time.Hour))
	assert.NoError(t, err)
	assert.Equal(t, testReceptions, receptions)

	mockRepo.On("GetProductsForReceptions", ctx, []uuid.UUID{testUUID}).Return(testProducts, nil).Once()
	products, err := mockRepo.GetProductsForReceptions(ctx, []uuid.UUID{testUUID})
	assert.NoError(t, err)
	assert.Equal(t, testProducts, products)

	mockRepo.AssertExpectations(t)
}
