package repository

import (
	"context"
	"fmt"
	"log/slog"
	"pvz-service/internal/config"
	"pvz-service/internal/models"
	"time"

	"github.com/google/uuid"
)

type PVZRepository interface {
	CloseConnection()

	// Basic PVZ operations
	InsertPVZ(ctx context.Context, pvz *models.PVZ) error
	CheckPVZ(ctx context.Context, pvzID uuid.UUID) (bool, error)
	GetCityID(ctx context.Context, cityName string) (int, error)
	GetPVZsWithNoFilter(ctx context.Context) ([]models.PVZ, error)

	// Reception operations
	InsertReception(ctx context.Context, reception *models.Reception) error
	GetActiveReception(ctx context.Context, pvzID uuid.UUID) (*models.Reception, error)
	UpdateReceptionStatus(ctx context.Context, receptionID uuid.UUID, status models.ReceptionStatus) error

	// Product operations
	InsertProduct(ctx context.Context, product *models.Product) error
	GetLastProduct(ctx context.Context, receptionID uuid.UUID) (*models.Product, error)
	DeleteProduct(ctx context.Context, productID uuid.UUID) error

	// Product type operations
	GetProductTypeID(ctx context.Context, productTypeName string) (int, error)

	// Query operations
	GetPVZs(ctx context.Context, from, to time.Time, limit, offset int) ([]models.PVZ, error)
	GetReceptionsForPVZs(ctx context.Context, pvzIDs []uuid.UUID, from, to time.Time) ([]models.Reception, error)
	GetProductsForReceptions(ctx context.Context, receptionIDs []uuid.UUID) ([]models.Product, error)
}

func CreatePVZRepo(cfg *config.Config, log *slog.Logger) (PVZRepository, error) {
	const op = "repository.pvz_repo.CreatePVZRepo"

	switch cfg.Database.Protocol {

	case "postgres":
		repo, err := PostgresGetter(cfg)
		if err != nil {
			return nil, err
		}
		log.Info("pvz_repo is postgres")
		return repo.(PVZRepository), nil

	default:
		return nil, fmt.Errorf("%s: unknown database protocol", op)
	}
}
