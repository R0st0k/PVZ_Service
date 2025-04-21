package tests

import (
	"encoding/json"
	"net/http"
	api "pvz-service/api/generated"
	e "pvz-service/internal/errors"
	"testing"

	openapi_types "github.com/oapi-codegen/runtime/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestLogin_Success(t *testing.T) {
	authMock, _, handler := setupHandler(t)

	email := "test@example.com"
	password := "password"
	token := "test_token"

	authMock.On("Login", mock.Anything, email, password).Return(token, nil)

	reqBody := api.PostLoginJSONRequestBody{
		Email:    openapi_types.Email(email),
		Password: password,
	}

	req, rec := createRequest(http.MethodPost, "/login", reqBody)
	handler.Login().ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp api.Token
	err := json.NewDecoder(rec.Body).Decode(&resp)
	assert.NoError(t, err)
	assert.Equal(t, token, string(resp))
}

func TestLogin_InvalidCredentials(t *testing.T) {
	authMock, _, handler := setupHandler(t)

	email := "test@example.com"
	password := "wrong_password"

	authMock.On("Login", mock.Anything, email, password).Return("", e.ErrInvalidCredentials())

	reqBody := api.PostLoginJSONRequestBody{
		Email:    openapi_types.Email(email),
		Password: password,
	}

	req, rec := createRequest(http.MethodPost, "/login", reqBody)
	handler.Login().ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}
