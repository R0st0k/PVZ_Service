package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"pvz-service/internal/controller/http/handler"
	"pvz-service/internal/metrics"
	"pvz-service/internal/models"
	"sync"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

var (
	metricsOnce sync.Once
	testMetrics *metrics.Metrics
)

type MockAuthService struct {
	mock.Mock
}

func (m *MockAuthService) Register(ctx context.Context, email, password string, role models.UserRole) (string, error) {
	args := m.Called(ctx, email, password, role)
	return args.String(0), args.Error(1)
}

func (m *MockAuthService) Login(ctx context.Context, email, password string) (string, error) {
	args := m.Called(ctx, email, password)
	return args.String(0), args.Error(1)
}

func (m *MockAuthService) DummyLogin(role models.UserRole) (string, error) {
	args := m.Called(role)
	return args.String(0), args.Error(1)
}

func (m *MockAuthService) ParseToken(tokenString string) (*jwt.Token, error) {
	args := m.Called(tokenString)
	return args.Get(0).(*jwt.Token), args.Error(1)
}

func (m *MockAuthService) GetUserFromToken(tokenString string) (string, models.UserRole, error) {
	args := m.Called(tokenString)
	return args.String(0), args.Get(1).(models.UserRole), args.Error(2)
}

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

func setupHandler(t *testing.T) (*MockAuthService, *MockPVZService, *handler.Handler) {
	t.Helper()
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	metricsOnce.Do(func() {
		testMetrics = metrics.NewMetrics()
	})
	authMock := new(MockAuthService)
	pvzMock := new(MockPVZService)
	var handler = handler.NewHandler(log, testMetrics, authMock, pvzMock)
	return authMock, pvzMock, handler
}

func createRequest(method, url string, body interface{}) (*http.Request, *httptest.ResponseRecorder) {
	var buf bytes.Buffer
	if body != nil {
		json.NewEncoder(&buf).Encode(body)
	}
	req := httptest.NewRequest(method, url, &buf)
	rec := httptest.NewRecorder()
	return req, rec
}

func addURLParams(r *http.Request, params map[string]string) *http.Request {
	ctx := chi.NewRouteContext()
	for key, val := range params {
		ctx.URLParams.Add(key, val)
	}
	return r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, ctx))
}
