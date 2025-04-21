package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"pvz-service/internal/config"
	e "pvz-service/internal/errors"
	"pvz-service/internal/logger/sl"
	"pvz-service/internal/models"
	"pvz-service/internal/repository"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type AuthService struct {
	repo         repository.AuthRepository
	log          *slog.Logger
	jwtSecret    string
	tokenExpires time.Duration
}

func NewAuthService(repo repository.AuthRepository, cfg *config.Config, log *slog.Logger) *AuthService {
	return &AuthService{
		repo:         repo,
		log:          log,
		jwtSecret:    cfg.JWT.SecretKey,
		tokenExpires: cfg.JWT.ExpiresIn,
	}
}

type AuthServiceInterface interface {
	Register(ctx context.Context, email, password string, role models.UserRole) (string, error)
	Login(ctx context.Context, email, password string) (string, error)
	DummyLogin(role models.UserRole) (string, error)
	ParseToken(tokenString string) (*jwt.Token, error)
	GetUserFromToken(tokenString string) (email string, role models.UserRole, err error)
}

func (s *AuthService) Register(ctx context.Context, email, password string, role models.UserRole) (string, error) {
	const op = "service.auth_service.Register"

	_, err := s.repo.CreateUser(ctx, email, password, role)
	if err != nil {
		s.log.Error(fmt.Sprintf("%s: create user erro", op), sl.Err(err))
		return "", err
	}

	return s.generateToken(email, role)
}

func (s *AuthService) Login(ctx context.Context, email, password string) (string, error) {
	const op = "service.auth_service.Login"

	valid, err := s.repo.VerifyPassword(ctx, email, password)
	if err != nil {
		s.log.Error(fmt.Sprintf("%s: verify password error", op), sl.Err(err))
		return "", err
	}
	if !valid {

		s.log.Info(fmt.Sprintf("%s: wrong password", op), "user", email)
		return "", e.ErrInvalidCredentials()
	}

	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		return "", err
	}

	return s.generateToken(user.Email, user.Role)
}

func (s *AuthService) DummyLogin(role models.UserRole) (string, error) {
	dummyEmail := "dummy_" + uuid.New().String() + "@example.com"
	return s.generateToken(dummyEmail, role)
}

func (s *AuthService) generateToken(email string, role models.UserRole) (string, error) {
	claims := jwt.MapClaims{
		"email": email,
		"role":  role,
		"exp":   time.Now().Add(s.tokenExpires).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}

func (s *AuthService) ParseToken(tokenString string) (*jwt.Token, error) {
	return jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, e.ErrWrongSigningMethod()
		}
		return []byte(s.jwtSecret), nil
	})
}

func (s *AuthService) GetUserFromToken(tokenString string) (email string, role models.UserRole, err error) {
	token, err := s.ParseToken(tokenString)
	if err != nil {
		return
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		err = errors.New("invalid token claims")
		return
	}

	email, ok = claims["email"].(string)
	if !ok {
		err = errors.New("email not found in token")
		return
	}

	roleStr, ok := claims["role"].(string)
	if !ok {
		err = errors.New("role not found in token")
		return
	}

	role = models.UserRole(roleStr)
	return
}
