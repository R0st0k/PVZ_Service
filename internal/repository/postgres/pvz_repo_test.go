package postgres

import (
	"context"
	"testing"
	"time"

	"pvz-service/internal/models"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestInsertPVZ(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	repo := &Postgres{db: db}

	pvz := &models.PVZ{
		ID:               uuid.New(),
		RegistrationDate: time.Now(),
		CityID:           1,
	}

	t.Run("Success", func(t *testing.T) {
		mock.ExpectExec("INSERT INTO pvz \\(id, registration_date, city_id\\) VALUES \\(\\$1, \\$2, \\$3\\)").
			WithArgs(pvz.ID, pvz.RegistrationDate, pvz.CityID).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := repo.InsertPVZ(context.Background(), pvz)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestGetActiveReception(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	repo := &Postgres{db: db}

	pvzID := uuid.New()
	receptionID := uuid.New()
	now := time.Now()

	t.Run("Success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "date_time", "pvz_id", "status"}).
			AddRow(receptionID, now, pvzID, models.ReceptionStatusInProgress)
		mock.ExpectQuery("SELECT(.*)").
			WithArgs(pvzID).
			WillReturnRows(rows)

		reception, err := repo.GetActiveReception(context.Background(), pvzID)
		assert.NoError(t, err)
		assert.Equal(t, &models.Reception{
			ID:       receptionID,
			DateTime: now,
			PVZID:    pvzID,
			Status:   models.ReceptionStatusInProgress,
		}, reception)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestInsertReception(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	repo := &Postgres{db: db}

	reception := &models.Reception{
		ID:       uuid.New(),
		DateTime: time.Now(),
		PVZID:    uuid.New(),
		Status:   models.ReceptionStatusInProgress,
	}

	t.Run("Success", func(t *testing.T) {
		mock.ExpectExec("INSERT INTO receptions \\(id, date_time, pvz_id, status\\) VALUES \\(\\$1, \\$2, \\$3, \\$4\\)").
			WithArgs(reception.ID, reception.DateTime, reception.PVZID, reception.Status).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := repo.InsertReception(context.Background(), reception)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestInsertProduct(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	repo := &Postgres{db: db}

	product := &models.Product{
		ID:          uuid.New(),
		DateTime:    time.Now(),
		TypeID:      1,
		ReceptionID: uuid.New(),
	}

	t.Run("Success", func(t *testing.T) {
		mock.ExpectExec("INSERT INTO products \\(id, date_time, type_id, reception_id\\) VALUES \\(\\$1, \\$2, \\$3, \\$4\\)").
			WithArgs(product.ID, product.DateTime, product.TypeID, product.ReceptionID).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := repo.InsertProduct(context.Background(), product)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestGetPVZs(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	repo := &Postgres{db: db}

	now := time.Now()
	pvzID := uuid.New()
	cityName := "Москва"

	t.Run("Success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "registration_date", "city_id", "name"}).
			AddRow(pvzID, now, 1, cityName)
		mock.ExpectQuery("SELECT(.*)").
			WithArgs(now.Add(-24*time.Hour), now, 10, 0).
			WillReturnRows(rows)

		pvzs, err := repo.GetPVZs(context.Background(), now.Add(-24*time.Hour), now, 10, 0)
		assert.NoError(t, err)
		assert.Equal(t, []models.PVZ{
			{
				ID:               pvzID,
				RegistrationDate: now,
				CityID:           1,
				CityName:         cityName,
			},
		}, pvzs)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
