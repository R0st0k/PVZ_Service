package tests

import (
	"database/sql"
	"log/slog"
	"net/http"
	"os"
	api "pvz-service/api/generated"
	"pvz-service/internal/config"
	"pvz-service/internal/controller/http/handler"
	"pvz-service/internal/metrics"
	"pvz-service/internal/models"
	"pvz-service/internal/repository"
	"pvz-service/internal/repository/postgres"
	"pvz-service/internal/service"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFullPVZWorkflow(t *testing.T) {
	// Init db
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	// Init logger
	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	// Mock config
	cfg := &config.Config{
		Database: config.Database{
			Protocol: "postgres",
		},
		JWT: config.JWT{
			SecretKey: "test_secret",
			ExpiresIn: 24 * time.Hour,
		},
	}

	// Init repo
	originalPostgresGetter := repository.PostgresGetter
	repository.PostgresGetter = func(cfg *config.Config) (interface{}, error) {
		p := &postgres.Postgres{}
		p.SetRepository(db)
		return p, nil
	}
	defer func() { repository.PostgresGetter = originalPostgresGetter }()

	authRepo, err := repository.CreateAuthRepo(cfg, log)
	require.NoError(t, err)
	defer authRepo.CloseConnection()

	pvzRepo, err := repository.CreatePVZRepo(cfg, log)
	require.NoError(t, err)
	defer pvzRepo.CloseConnection()

	// Create services
	metricsOnce.Do(func() {
		testMetrics = metrics.NewMetrics()
	})
	authService := service.NewAuthService(authRepo, cfg, log)
	pvzService := service.NewPVZService(pvzRepo, log)

	// Create handler
	h := handler.NewHandler(log, testMetrics, authService, pvzService)

	// Test data
	pvzID := uuid.New()
	receptionID := uuid.New()
	cityID := 1
	productTypeID := 1
	regDate := time.Now()

	// Test: Create PVZ
	t.Run("Create PVZ", func(t *testing.T) {
		mock.ExpectQuery("SELECT id FROM cities WHERE name = \\$1").
			WithArgs("Москва").
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(cityID))

		mock.ExpectExec("INSERT INTO pvz \\(id, registration_date, city_id\\) VALUES \\(\\$1, \\$2, \\$3\\)").
			WithArgs(pvzID, sqlmock.AnyArg(), cityID).
			WillReturnResult(sqlmock.NewResult(1, 1))

		reqBody := api.PostPvzJSONRequestBody{
			City:             "Москва",
			Id:               &pvzID,
			RegistrationDate: &regDate,
		}

		req, rec := createRequest(http.MethodPost, "/pvz", reqBody)
		h.CreatePVZ().ServeHTTP(rec, req)

		assert.Equal(t, http.StatusCreated, rec.Code)
	})

	// Test: Create reception
	t.Run("Start Reception", func(t *testing.T) {
		mock.ExpectQuery("SELECT EXISTS\\(SELECT 1 FROM pvz WHERE id = \\$1\\)").
			WithArgs(pvzID).
			WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

		mock.ExpectQuery("SELECT id, date_time, pvz_id, status FROM receptions WHERE pvz_id = \\$1 AND status = 'in_progress' ORDER BY date_time DESC LIMIT 1").
			WithArgs(pvzID).
			WillReturnError(sql.ErrNoRows)

		mock.ExpectExec("INSERT INTO receptions \\(id, date_time, pvz_id, status\\) VALUES \\(\\$1, \\$2, \\$3, \\$4\\)").
			WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), pvzID, models.ReceptionStatusInProgress).
			WillReturnResult(sqlmock.NewResult(1, 1))

		reqBody := api.PostReceptionsJSONRequestBody{
			PvzId: pvzID,
		}

		req, rec := createRequest(http.MethodPost, "/receptions", reqBody)
		h.StartReception().ServeHTTP(rec, req)

		assert.Equal(t, http.StatusCreated, rec.Code)
	})

	// Test: Add products
	t.Run("Add 50 Products", func(t *testing.T) {
		for i := 0; i < 50; i++ {

			mock.ExpectQuery("SELECT id, date_time, pvz_id, status FROM receptions WHERE pvz_id = \\$1 AND status = 'in_progress' ORDER BY date_time DESC LIMIT 1").
				WithArgs(pvzID).
				WillReturnRows(
					sqlmock.NewRows([]string{"id", "date_time", "pvz_id", "status"}).
						AddRow(receptionID, time.Now(), pvzID, models.ReceptionStatusInProgress))

			mock.ExpectQuery("SELECT id FROM product_types WHERE name = \\$1").
				WithArgs("одежда").
				WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(productTypeID))

			mock.ExpectExec("INSERT INTO products \\(id, date_time, type_id, reception_id\\) VALUES \\(\\$1, \\$2, \\$3, \\$4\\)").
				WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), productTypeID, receptionID).
				WillReturnResult(sqlmock.NewResult(1, 1))
		}

		for i := 0; i < 50; i++ {
			reqBody := api.PostProductsJSONRequestBody{
				PvzId: pvzID,
				Type:  "одежда",
			}

			req, rec := createRequest(http.MethodPost, "/products", reqBody)
			h.AddProduct().ServeHTTP(rec, req)

			assert.Equal(t, http.StatusCreated, rec.Code)
		}
	})

	// Test: Close reception
	t.Run("Close Reception", func(t *testing.T) {
		mock.ExpectQuery("SELECT id, date_time, pvz_id, status FROM receptions WHERE pvz_id = \\$1 AND status = 'in_progress' ORDER BY date_time DESC LIMIT 1").
			WithArgs(pvzID).
			WillReturnRows(
				sqlmock.NewRows([]string{"id", "date_time", "pvz_id", "status"}).
					AddRow(receptionID, time.Now(), pvzID, models.ReceptionStatusInProgress))

		mock.ExpectExec("UPDATE receptions SET status = \\$1 WHERE id = \\$2").
			WithArgs(models.ReceptionStatusClose, receptionID).
			WillReturnResult(sqlmock.NewResult(1, 1))

		req, rec := createRequest(http.MethodPost, "/pvz/"+pvzID.String()+"/receptions/close", nil)
		req = addURLParams(req, map[string]string{"pvzId": pvzID.String()})
		h.CloseReception().ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
	})

	assert.NoError(t, mock.ExpectationsWereMet())
}
