package postgres

import (
	"context"
	"testing"

	"pvz-service/internal/models"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

func TestCreateUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	repo := &Postgres{db: db}

	tests := []struct {
		name        string
		email       string
		password    string
		role        models.UserRole
		mock        func()
		expected    *models.User
		expectedErr error
	}{
		{
			name:     "Success Employee",
			email:    "employee@example.com",
			password: "password123",
			role:     models.UserRoleEmployee,
			mock: func() {
				mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM users WHERE email = \\$1").
					WithArgs("employee@example.com").
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

				mock.ExpectExec("INSERT INTO users \\(id, email, password_hash, role\\) VALUES \\(\\$1, \\$2, \\$3, \\$4\\)").
					WithArgs(sqlmock.AnyArg(), "employee@example.com", sqlmock.AnyArg(), models.UserRoleEmployee).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			expected: &models.User{
				Email: "employee@example.com",
				Role:  models.UserRoleEmployee,
			},
		},
		{
			name:     "Success Moderator",
			email:    "moderator@example.com",
			password: "password123",
			role:     models.UserRoleModerator,
			mock: func() {
				mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM users WHERE email = \\$1").
					WithArgs("moderator@example.com").
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

				mock.ExpectExec("INSERT INTO users \\(id, email, password_hash, role\\) VALUES \\(\\$1, \\$2, \\$3, \\$4\\)").
					WithArgs(sqlmock.AnyArg(), "moderator@example.com", sqlmock.AnyArg(), models.UserRoleModerator).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			expected: &models.User{
				Email: "moderator@example.com",
				Role:  models.UserRoleModerator,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock()

			user, err := repo.CreateUser(context.Background(), tt.email, tt.password, tt.role)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.expected.Email, user.Email)
			assert.Equal(t, tt.expected.Role, user.Role)
			assert.NotEmpty(t, user.ID)
			assert.NotEmpty(t, user.PasswordHash)

			err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(tt.password))
			assert.NoError(t, err)

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestGetUserByEmail(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	repo := &Postgres{db: db}

	testID := uuid.New()
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)

	tests := []struct {
		name        string
		email       string
		mock        func()
		expected    *models.User
		expectedErr error
	}{
		{
			name:  "Success Employee",
			email: "employee@example.com",
			mock: func() {
				rows := sqlmock.NewRows([]string{"id", "email", "password_hash", "role"}).
					AddRow(testID, "employee@example.com", string(hashedPassword), models.UserRoleEmployee)
				mock.ExpectQuery("SELECT id, email, password_hash, role FROM users WHERE email = \\$1").
					WithArgs("employee@example.com").
					WillReturnRows(rows)
			},
			expected: &models.User{
				ID:           testID,
				Email:        "employee@example.com",
				PasswordHash: string(hashedPassword),
				Role:         models.UserRoleEmployee,
			},
		},
		{
			name:  "Success Moderator",
			email: "moderator@example.com",
			mock: func() {
				rows := sqlmock.NewRows([]string{"id", "email", "password_hash", "role"}).
					AddRow(testID, "moderator@example.com", string(hashedPassword), models.UserRoleModerator)
				mock.ExpectQuery("SELECT id, email, password_hash, role FROM users WHERE email = \\$1").
					WithArgs("moderator@example.com").
					WillReturnRows(rows)
			},
			expected: &models.User{
				ID:           testID,
				Email:        "moderator@example.com",
				PasswordHash: string(hashedPassword),
				Role:         models.UserRoleModerator,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock()

			user, err := repo.GetUserByEmail(context.Background(), tt.email)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.expected, user)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
