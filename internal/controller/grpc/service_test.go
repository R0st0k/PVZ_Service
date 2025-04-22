package grpc

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"

	pvz_v1 "pvz-service/api/proto_v1"
	"pvz-service/internal/models"
)

// MockPVZService is a mock implementation of PVZServiceInterface
type MockPVZService struct {
	mock.Mock
}

func (m *MockPVZService) CreatePVZ(ctx context.Context, pvz *models.PVZ) (*models.PVZ, error) {
	args := m.Called(ctx, pvz)
	return args.Get(0).(*models.PVZ), args.Error(1)
}

func (m *MockPVZService) StartReception(ctx context.Context, pvzID uuid.UUID) (*models.Reception, error) {
	args := m.Called(ctx, pvzID)
	return args.Get(0).(*models.Reception), args.Error(1)
}

func (m *MockPVZService) AddProduct(ctx context.Context, pvzID uuid.UUID, productTypeName string) (*models.Product, error) {
	args := m.Called(ctx, pvzID, productTypeName)
	return args.Get(0).(*models.Product), args.Error(1)
}

func (m *MockPVZService) DeleteLastProduct(ctx context.Context, pvzID uuid.UUID) error {
	args := m.Called(ctx, pvzID)
	return args.Error(0)
}

func (m *MockPVZService) CloseReception(ctx context.Context, pvzID uuid.UUID) (*models.Reception, error) {
	args := m.Called(ctx, pvzID)
	return args.Get(0).(*models.Reception), args.Error(1)
}

func (m *MockPVZService) GetPVZsWithReceptions(ctx context.Context, from, to time.Time, page, limit int) ([]models.PVZInfo, error) {
	args := m.Called(ctx, from, to, page, limit)
	return args.Get(0).([]models.PVZInfo), args.Error(1)
}

func (m *MockPVZService) GetPVZs(ctx context.Context) ([]models.PVZ, error) {
	args := m.Called(ctx)
	return args.Get(0).([]models.PVZ), args.Error(1)
}

func TestNewPVZServer(t *testing.T) {
	mockService := new(MockPVZService)
	server := NewPVZServer(mockService)

	assert.NotNil(t, server)
	assert.Equal(t, mockService, server.service)
}

func TestRegister(t *testing.T) {
	mockService := new(MockPVZService)
	server := NewPVZServer(mockService)

	grpcServer := grpc.NewServer()
	server.Register(grpcServer)

	// Verify the service is registered by checking the service info
	services := grpcServer.GetServiceInfo()
	_, exists := services["pvz.v1.PVZService"]
	assert.True(t, exists)
}

func TestGetPVZList(t *testing.T) {
	now := time.Now()
	testUUID := uuid.New()

	tests := []struct {
		name          string
		mockReturn    []models.PVZ
		mockError     error
		expected      *pvz_v1.GetPVZListResponse
		expectedError error
	}{
		{
			name: "successful response with multiple PVZs",
			mockReturn: []models.PVZ{
				{
					ID:               testUUID,
					RegistrationDate: now,
					CityName:         "Москва",
				},
				{
					ID:               testUUID,
					RegistrationDate: now.Add(-24 * time.Hour),
					CityName:         "Санкт-Петербург",
				},
			},
			mockError: nil,
			expected: &pvz_v1.GetPVZListResponse{
				Pvzs: []*pvz_v1.PVZ{
					{
						Id:               testUUID.String(),
						RegistrationDate: timestamppb.New(now),
						City:             "Москва",
					},
					{
						Id:               testUUID.String(),
						RegistrationDate: timestamppb.New(now.Add(-24 * time.Hour)),
						City:             "Санкт-Петербург",
					},
				},
			},
			expectedError: nil,
		},
		{
			name:          "empty response",
			mockReturn:    []models.PVZ{},
			mockError:     nil,
			expected:      &pvz_v1.GetPVZListResponse{Pvzs: []*pvz_v1.PVZ{}},
			expectedError: nil,
		},
		{
			name:          "service error",
			mockReturn:    nil,
			mockError:     assert.AnError,
			expected:      nil,
			expectedError: assert.AnError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockPVZService)
			server := NewPVZServer(mockService)

			mockService.On("GetPVZs", mock.Anything).Return(tt.mockReturn, tt.mockError)

			resp, err := server.GetPVZList(context.Background(), &pvz_v1.GetPVZListRequest{})

			if tt.expectedError != nil {
				assert.Equal(t, tt.expectedError, err)
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, len(tt.expected.Pvzs), len(resp.Pvzs))

				for i, expectedPVZ := range tt.expected.Pvzs {
					assert.Equal(t, expectedPVZ.Id, resp.Pvzs[i].Id)
					assert.Equal(t, expectedPVZ.City, resp.Pvzs[i].City)
					assert.True(t, expectedPVZ.RegistrationDate.AsTime().Equal(resp.Pvzs[i].RegistrationDate.AsTime()))
				}
			}

			mockService.AssertExpectations(t)
		})
	}
}
