package postgres

import (
	"database/sql"
	"fmt"
	"pvz-service/internal/config"

	"github.com/pressly/goose/v3"
)

type Postgres struct {
	db *sql.DB
}

var repo *Postgres

func newPostgresRepository(cfg *config.Config) (*Postgres, error) {
	const op = "repository.postgres.NewPostgresRepository"

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Name,
		cfg.Database.SSLMode,
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	psg := Postgres{db: db}

	err = psg.migrate()
	if err != nil {
		return nil, err
	}

	return &psg, nil
}

func (p *Postgres) migrate() error {
	const op = "repository.postgres.migrate"

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("%s: migration error: %w", op, err)
	}

	if err := goose.Up(p.db, "/migrations"); err != nil {
		return fmt.Errorf("%s: migration error: %w", op, err)
	}

	return nil
}

func GetRepository(cfg *config.Config) (*Postgres, error) {
	var err error = nil

	if repo == nil {
		repo, err = newPostgresRepository(cfg)
	}

	return repo, err
}

func (p *Postgres) SetRepository(db *sql.DB) {
	p.db = db
}

func (p *Postgres) CloseConnection() {
	if p.db != nil {
		p.db.Close()
	}
}
