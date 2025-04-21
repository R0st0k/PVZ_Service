package tests

import (
	"encoding/json"
	"net/http"
	api "pvz-service/api/generated"
	"pvz-service/internal/models"
	"testing"

	openapi_types "github.com/oapi-codegen/runtime/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRegister_Success(t *testing.T) {
	authMock, _, handler := setupHandler(t)

	email := "test@example.com"
	password := "password"
	role := api.PostRegisterJSONBodyRole(api.UserRoleEmployee)

	authMock.On("Register", mock.Anything, email, password, models.UserRole(role)).Return("token", nil)

	reqBody := api.PostRegisterJSONRequestBody{
		Email:    openapi_types.Email(email),
		Password: password,
		Role:     role,
	}

	req, rec := createRequest(http.MethodPost, "/register", reqBody)
	handler.Register().ServeHTTP(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)

	var resp api.User
	err := json.NewDecoder(rec.Body).Decode(&resp)
	assert.NoError(t, err)
	assert.Equal(t, email, string(resp.Email))
	assert.Equal(t, api.UserRole(role), resp.Role)
}

func TestRegister_ValidationError(t *testing.T) {
	_, _, handler := setupHandler(t)

	reqBody := api.PostRegisterJSONRequestBody{
		Email:    "invalid-email",
		Password: "",
		Role:     "invalid-role",
	}

	req, rec := createRequest(http.MethodPost, "/register", reqBody)
	handler.Register().ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}
