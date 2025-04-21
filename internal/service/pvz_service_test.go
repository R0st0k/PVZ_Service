package service

import (
	"context"
	"errors"
	"log/slog"
	e "pvz-service/internal/errors"
	"pvz-service/internal/models"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

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
	recp := args.Get(0)
	if recp == nil {
		return nil, args.Error(1)
	}
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
	prd := args.Get(0)
	if prd == nil {
		return nil, args.Error(1)
	}
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
	list := args.Get(0)
	if list == nil {
		return nil, args.Error(1)
	}
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

func TestPVZService_CreatePVZ(t *testing.T) {
	tests := []struct {
		name        string
		pvz         *models.PVZ
		mockSetup   func(*MockPVZRepository)
		expectError error
	}{
		{
			name: "Success",
			pvz: &models.PVZ{
				CityName: "Москва",
			},
			mockSetup: func(m *MockPVZRepository) {
				m.On("GetCityID", mock.Anything, "Москва").Return(1, nil)
				m.On("InsertPVZ", mock.Anything, mock.MatchedBy(func(p *models.PVZ) bool {
					return p.CityName == "Москва" && p.CityID == 1
				})).Return(nil)
			},
			expectError: nil,
		},
		{
			name: "City not found",
			pvz: &models.PVZ{
				CityName: "Unknown",
			},
			mockSetup: func(m *MockPVZRepository) {
				m.On("GetCityID", mock.Anything, "Unknown").Return(0, e.ErrCityNotAllowed())
			},
			expectError: e.ErrCityNotAllowed(),
		},
		{
			name: "Insert error",
			pvz: &models.PVZ{
				CityName: "Москва",
			},
			mockSetup: func(m *MockPVZRepository) {
				m.On("GetCityID", mock.Anything, "Москва").Return(1, nil)
				m.On("InsertPVZ", mock.Anything, mock.Anything).Return(errors.New("insert error"))
			},
			expectError: errors.New("failed to create PVZ: insert error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockPVZRepository)
			tt.mockSetup(mockRepo)

			service := NewPVZService(mockRepo, slog.Default())
			result, err := service.CreatePVZ(context.Background(), tt.pvz)

			if tt.expectError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectError.Error(), err.Error())
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, 1, result.CityID)
				assert.Equal(t, "Москва", result.CityName)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestPVZService_StartReception(t *testing.T) {
	testPVZID := uuid.New()
	testReception := &models.Reception{
		ID:       uuid.New(),
		PVZID:    testPVZID,
		DateTime: time.Now(),
		Status:   models.ReceptionStatusInProgress,
	}

	tests := []struct {
		name        string
		pvzID       uuid.UUID
		mockSetup   func(*MockPVZRepository)
		expectError error
	}{
		{
			name:  "Success",
			pvzID: testPVZID,
			mockSetup: func(m *MockPVZRepository) {
				m.On("CheckPVZ", mock.Anything, testPVZID).Return(true, nil)
				m.On("GetActiveReception", mock.Anything, testPVZID).Return(nil, e.ErrNoActiveReception())
				m.On("InsertReception", mock.Anything, mock.MatchedBy(func(r *models.Reception) bool {
					return r.PVZID == testPVZID && r.Status == models.ReceptionStatusInProgress
				})).Return(nil)
			},
			expectError: nil,
		},
		{
			name:  "PVZ not found",
			pvzID: testPVZID,
			mockSetup: func(m *MockPVZRepository) {
				m.On("CheckPVZ", mock.Anything, testPVZID).Return(false, e.ErrNotFound())
			},
			expectError: e.ErrCityNotAllowed(),
		},
		{
			name:  "Active reception exists",
			pvzID: testPVZID,
			mockSetup: func(m *MockPVZRepository) {
				m.On("CheckPVZ", mock.Anything, testPVZID).Return(true, nil)
				m.On("GetActiveReception", mock.Anything, testPVZID).Return(testReception, nil)
			},
			expectError: e.ErrActiveReceptionExists(),
		},
		{
			name:  "Insert error",
			pvzID: testPVZID,
			mockSetup: func(m *MockPVZRepository) {
				m.On("CheckPVZ", mock.Anything, testPVZID).Return(true, nil)
				m.On("GetActiveReception", mock.Anything, testPVZID).Return(nil, e.ErrNoActiveReception())
				m.On("InsertReception", mock.Anything, mock.Anything).Return(errors.New("insert error"))
			},
			expectError: errors.New("failed to create reception: insert error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockPVZRepository)
			tt.mockSetup(mockRepo)

			service := NewPVZService(mockRepo, slog.Default())
			result, err := service.StartReception(context.Background(), tt.pvzID)

			if tt.expectError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectError.Error(), err.Error())
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, testPVZID, result.PVZID)
				assert.Equal(t, models.ReceptionStatusInProgress, result.Status)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestPVZService_AddProduct(t *testing.T) {
	testPVZID := uuid.New()
	testReception := &models.Reception{
		ID:       uuid.New(),
		PVZID:    testPVZID,
		DateTime: time.Now(),
		Status:   models.ReceptionStatusInProgress,
	}
	testProductType := "electronics"

	tests := []struct {
		name          string
		pvzID         uuid.UUID
		productType   string
		mockSetup     func(*MockPVZRepository)
		expectError   error
		expectProduct bool
	}{
		{
			name:        "Success",
			pvzID:       testPVZID,
			productType: testProductType,
			mockSetup: func(m *MockPVZRepository) {
				m.On("GetActiveReception", mock.Anything, testPVZID).Return(testReception, nil)
				m.On("GetProductTypeID", mock.Anything, testProductType).Return(1, nil)
				m.On("InsertProduct", mock.Anything, mock.MatchedBy(func(p *models.Product) bool {
					return p.ReceptionID == testReception.ID && p.TypeName == testProductType && p.TypeID == 1
				})).Return(nil)
			},
			expectError:   nil,
			expectProduct: true,
		},
		{
			name:        "No active reception",
			pvzID:       testPVZID,
			productType: testProductType,
			mockSetup: func(m *MockPVZRepository) {
				m.On("GetActiveReception", mock.Anything, testPVZID).Return(nil, e.ErrNoActiveReception())
			},
			expectError:   e.ErrNoActiveReception(),
			expectProduct: false,
		},
		{
			name:        "Product type not allowed",
			pvzID:       testPVZID,
			productType: "forbidden",
			mockSetup: func(m *MockPVZRepository) {
				m.On("GetActiveReception", mock.Anything, testPVZID).Return(testReception, nil)
				m.On("GetProductTypeID", mock.Anything, "forbidden").Return(0, e.ErrProductTypeNotAllowed())
			},
			expectError:   e.ErrProductTypeNotAllowed(),
			expectProduct: false,
		},
		{
			name:        "Insert error",
			pvzID:       testPVZID,
			productType: testProductType,
			mockSetup: func(m *MockPVZRepository) {
				m.On("GetActiveReception", mock.Anything, testPVZID).Return(testReception, nil)
				m.On("GetProductTypeID", mock.Anything, testProductType).Return(1, nil)
				m.On("InsertProduct", mock.Anything, mock.Anything).Return(errors.New("insert error"))
			},
			expectError:   errors.New("failed to add product: insert error"),
			expectProduct: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockPVZRepository)
			tt.mockSetup(mockRepo)

			service := NewPVZService(mockRepo, slog.Default())
			product, err := service.AddProduct(context.Background(), tt.pvzID, tt.productType)

			if tt.expectError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectError.Error(), err.Error())
				assert.Nil(t, product)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, product)
				assert.Equal(t, testProductType, product.TypeName)
				assert.Equal(t, testReception.ID, product.ReceptionID)
				assert.Equal(t, 1, product.TypeID)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestPVZService_DeleteLastProduct(t *testing.T) {
	testPVZID := uuid.New()
	testReception := &models.Reception{
		ID:       uuid.New(),
		PVZID:    testPVZID,
		DateTime: time.Now(),
		Status:   models.ReceptionStatusInProgress,
	}
	testProduct := &models.Product{
		ID:          uuid.New(),
		ReceptionID: testReception.ID,
	}

	tests := []struct {
		name        string
		pvzID       uuid.UUID
		mockSetup   func(*MockPVZRepository)
		expectError error
	}{
		{
			name:  "Success",
			pvzID: testPVZID,
			mockSetup: func(m *MockPVZRepository) {
				m.On("GetActiveReception", mock.Anything, testPVZID).Return(testReception, nil)
				m.On("GetLastProduct", mock.Anything, testReception.ID).Return(testProduct, nil)
				m.On("DeleteProduct", mock.Anything, testProduct.ID).Return(nil)
			},
			expectError: nil,
		},
		{
			name:  "No active reception",
			pvzID: testPVZID,
			mockSetup: func(m *MockPVZRepository) {
				m.On("GetActiveReception", mock.Anything, testPVZID).Return(nil, e.ErrNoActiveReception())
			},
			expectError: e.ErrNoActiveReception(),
		},
		{
			name:  "No products",
			pvzID: testPVZID,
			mockSetup: func(m *MockPVZRepository) {
				m.On("GetActiveReception", mock.Anything, testPVZID).Return(testReception, nil)
				m.On("GetLastProduct", mock.Anything, testReception.ID).Return(nil, e.ErrNotFound())
			},
			expectError: e.ErrNoProduct(),
		},
		{
			name:  "Delete error",
			pvzID: testPVZID,
			mockSetup: func(m *MockPVZRepository) {
				m.On("GetActiveReception", mock.Anything, testPVZID).Return(testReception, nil)
				m.On("GetLastProduct", mock.Anything, testReception.ID).Return(testProduct, nil)
				m.On("DeleteProduct", mock.Anything, testProduct.ID).Return(errors.New("delete error"))
			},
			expectError: errors.New("failed to delete product: delete error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockPVZRepository)
			tt.mockSetup(mockRepo)

			service := NewPVZService(mockRepo, slog.Default())
			err := service.DeleteLastProduct(context.Background(), tt.pvzID)

			if tt.expectError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestPVZService_CloseReception(t *testing.T) {
	testPVZID := uuid.New()
	testReception := &models.Reception{
		ID:       uuid.New(),
		PVZID:    testPVZID,
		DateTime: time.Now(),
		Status:   models.ReceptionStatusInProgress,
	}

	tests := []struct {
		name        string
		pvzID       uuid.UUID
		mockSetup   func(*MockPVZRepository)
		expectError error
	}{
		{
			name:  "Success",
			pvzID: testPVZID,
			mockSetup: func(m *MockPVZRepository) {
				m.On("GetActiveReception", mock.Anything, testPVZID).Return(testReception, nil)
				m.On("UpdateReceptionStatus", mock.Anything, testReception.ID, models.ReceptionStatusClose).Return(nil)
			},
			expectError: nil,
		},
		{
			name:  "No active reception",
			pvzID: testPVZID,
			mockSetup: func(m *MockPVZRepository) {
				m.On("GetActiveReception", mock.Anything, testPVZID).Return(nil, e.ErrNoActiveReception())
			},
			expectError: e.ErrNoActiveReception(),
		},
		{
			name:  "Update error",
			pvzID: testPVZID,
			mockSetup: func(m *MockPVZRepository) {
				m.On("GetActiveReception", mock.Anything, testPVZID).Return(testReception, nil)
				m.On("UpdateReceptionStatus", mock.Anything, testReception.ID, models.ReceptionStatusClose).Return(errors.New("update error"))
			},
			expectError: errors.New("failed to close reception: update error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockPVZRepository)
			tt.mockSetup(mockRepo)

			service := NewPVZService(mockRepo, slog.Default())
			result, err := service.CloseReception(context.Background(), tt.pvzID)

			if tt.expectError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectError.Error(), err.Error())
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, models.ReceptionStatusClose, result.Status)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestPVZService_GetPVZsWithReceptions(t *testing.T) {
	now := time.Now()
	testPVZ := models.PVZ{
		ID:               uuid.New(),
		RegistrationDate: now,
		CityName:         "Москва",
		CityID:           1,
	}
	testReception := models.Reception{
		ID:       uuid.New(),
		PVZID:    testPVZ.ID,
		DateTime: now,
		Status:   models.ReceptionStatusClose,
	}
	testProduct := models.Product{
		ID:          uuid.New(),
		ReceptionID: testReception.ID,
		TypeName:    "electronics",
		DateTime:    now,
		TypeID:      1,
	}

	tests := []struct {
		name          string
		from          time.Time
		to            time.Time
		page          int
		limit         int
		mockSetup     func(*MockPVZRepository)
		expectError   error
		expectResults int
	}{
		{
			name:  "Success with full hierarchy",
			from:  now.Add(-24 * time.Hour),
			to:    now.Add(24 * time.Hour),
			page:  1,
			limit: 10,
			mockSetup: func(m *MockPVZRepository) {
				m.On("GetPVZs", mock.Anything, mock.Anything, mock.Anything, 10, 0).Return([]models.PVZ{testPVZ}, nil)
				m.On("GetReceptionsForPVZs", mock.Anything, []uuid.UUID{testPVZ.ID}, mock.Anything, mock.Anything).Return([]models.Reception{testReception}, nil)
				m.On("GetProductsForReceptions", mock.Anything, []uuid.UUID{testReception.ID}).Return([]models.Product{testProduct}, nil)
			},
			expectError:   nil,
			expectResults: 1,
		},
		{
			name:  "No data",
			from:  now.Add(-24 * time.Hour),
			to:    now.Add(24 * time.Hour),
			page:  1,
			limit: 10,
			mockSetup: func(m *MockPVZRepository) {
				m.On("GetPVZs", mock.Anything, mock.Anything, mock.Anything, 10, 0).Return([]models.PVZ{}, nil)
			},
			expectError:   nil,
			expectResults: 0,
		},
		{
			name:  "Get PVZs error",
			from:  now.Add(-24 * time.Hour),
			to:    now.Add(24 * time.Hour),
			page:  1,
			limit: 10,
			mockSetup: func(m *MockPVZRepository) {
				m.On("GetPVZs", mock.Anything, mock.Anything, mock.Anything, 10, 0).Return(nil, errors.New("get error"))
			},
			expectError:   errors.New("failed to get PVZs: get error"),
			expectResults: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockPVZRepository)
			tt.mockSetup(mockRepo)

			service := NewPVZService(mockRepo, slog.Default())
			result, err := service.GetPVZsWithReceptions(context.Background(), tt.from, tt.to, tt.page, tt.limit)

			if tt.expectError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectError.Error(), err.Error())
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.expectResults, len(result))

				if tt.expectResults > 0 {
					// Check PVZInfo
					pvzInfo := result[0]
					assert.Equal(t, testPVZ.ID, pvzInfo.PVZ.ID)
					assert.Equal(t, testPVZ.CityName, pvzInfo.PVZ.CityName)

					// Check ReceptionInfo
					assert.Equal(t, 1, len(pvzInfo.Receptions))
					receptionInfo := pvzInfo.Receptions[0]
					assert.Equal(t, testReception.ID, receptionInfo.Reception.ID)
					assert.Equal(t, testReception.Status, receptionInfo.Reception.Status)

					// Check Products
					assert.Equal(t, 1, len(receptionInfo.Products))
					product := receptionInfo.Products[0]
					assert.Equal(t, testProduct.ID, product.ID)
					assert.Equal(t, testProduct.TypeName, product.TypeName)
				}
			}

			mockRepo.AssertExpectations(t)
		})
	}
}
