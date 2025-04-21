package repository

import (
	"pvz-service/internal/config"
	"pvz-service/internal/repository/postgres"
)

var PostgresGetter = func(cfg *config.Config) (interface{}, error) {
	return postgres.GetRepository(cfg)
}
