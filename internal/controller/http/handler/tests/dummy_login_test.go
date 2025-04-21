package tests

import (
	"encoding/json"
	"net/http"
	api "pvz-service/api/generated"
	"pvz-service/internal/models"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDummyLogin_Success(t *testing.T) {
	authMock, _, handler := setupHandler(t)

	role := models.UserRoleEmployee
	token := "dummy_token"

	authMock.On("DummyLogin", role).Return(token, nil)

	reqBody := api.PostDummyLoginJSONRequestBody{
		Role: api.PostDummyLoginJSONBodyRoleEmployee,
	}

	req, rec := createRequest(http.MethodPost, "/dummy-login", reqBody)
	handler.DummyLogin().ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp api.Token
	err := json.NewDecoder(rec.Body).Decode(&resp)
	assert.NoError(t, err)
	assert.Equal(t, token, string(resp))
}
