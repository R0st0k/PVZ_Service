package repository

import (
	"context"
	"fmt"
	"log/slog"
	"pvz-service/internal/config"
	"pvz-service/internal/models"

	"github.com/google/uuid"
)

type AuthRepository interface {
	CloseConnection()

	CreateUser(ctx context.Context, email, password string, role models.UserRole) (*models.User, error)
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error)
	VerifyPassword(ctx context.Context, email, password string) (bool, error)
}

func CreateAuthRepo(cfg *config.Config, log *slog.Logger) (AuthRepository, error) {
	const op = "repository.auth_repo.CreateAuthRepo"

	switch cfg.Database.Protocol {

	case "postgres":
		repo, err := PostgresGetter(cfg)
		if err != nil {
			return nil, err
		}
		log.Info("auth_repo is postgres")
		return repo.(AuthRepository), nil

	default:
		return nil, fmt.Errorf("%s: unknown database protocol (%s)", op, cfg.Database.Protocol)
	}
}
