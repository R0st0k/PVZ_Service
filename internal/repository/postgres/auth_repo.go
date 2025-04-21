package postgres

import (
	"context"
	"database/sql"
	"errors"
	e "pvz-service/internal/errors"
	"pvz-service/internal/models"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func (p *Postgres) CreateUser(ctx context.Context, email, password string, role models.UserRole) (*models.User, error) {
	var count int
	row := p.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM users WHERE email = $1", email)
	if err := row.Scan(&count); err != nil {
		return nil, err
	}
	if count > 0 {
		return nil, e.ErrAlreadyExists()
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &models.User{
		ID:           uuid.New(),
		Email:        email,
		PasswordHash: string(hashedPassword),
		Role:         role,
	}

	_, err = p.db.ExecContext(ctx,
		"INSERT INTO users (id, email, password_hash, role) VALUES ($1, $2, $3, $4)",
		user.ID, user.Email, user.PasswordHash, user.Role)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (p *Postgres) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	row := p.db.QueryRowContext(ctx,
		"SELECT id, email, password_hash, role FROM users WHERE email = $1", email)

	err := row.Scan(&user.ID, &user.Email, &user.PasswordHash, &user.Role)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, e.ErrNotFound()
		}
		return nil, err
	}
	return &user, nil
}

func (p *Postgres) GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	var user models.User
	row := p.db.QueryRowContext(ctx,
		"SELECT id, email, password_hash, role FROM users WHERE id = $1", id)

	err := row.Scan(&user.ID, &user.Email, &user.PasswordHash, &user.Role)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, e.ErrNotFound()
		}
		return nil, err
	}
	return &user, nil
}

func (p *Postgres) VerifyPassword(ctx context.Context, email, password string) (bool, error) {
	user, err := p.GetUserByEmail(ctx, email)
	if err != nil {
		return false, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	return err == nil, nil
}
