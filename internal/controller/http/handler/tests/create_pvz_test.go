package tests

import (
	"encoding/json"
	"errors"
	"net/http"
	api "pvz-service/api/generated"
	"pvz-service/internal/models"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCreatePVZ_Success(t *testing.T) {
	_, pvzMock, handler := setupHandler(t)

	expectedID := uuid.New()
	expectedDate := time.Now()
	expectedPVZ := &models.PVZ{
		ID:               expectedID,
		RegistrationDate: expectedDate,
		CityName:         "Москва",
	}

	pvzMock.On("CreatePVZ", mock.Anything, mock.Anything).Return(expectedPVZ, nil)

	reqBody := api.PostPvzJSONRequestBody{
		City:             "Москва",
		Id:               &expectedID,
		RegistrationDate: &expectedDate,
	}

	req, rec := createRequest(http.MethodPost, "/pvz", reqBody)
	handler.CreatePVZ().ServeHTTP(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)

	var resp api.PVZ
	err := json.NewDecoder(rec.Body).Decode(&resp)
	assert.NoError(t, err)
	assert.Equal(t, expectedID, *resp.Id)
	assert.Equal(t, api.PVZCity("Москва"), resp.City)
}

func TestCreatePVZ_ValidationError(t *testing.T) {
	_, _, handler := setupHandler(t)

	reqBody := api.PostPvzJSONRequestBody{} // Empty city - invalid
	req, rec := createRequest(http.MethodPost, "/pvz", reqBody)
	handler.CreatePVZ().ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var resp api.Error
	err := json.NewDecoder(rec.Body).Decode(&resp)
	assert.NoError(t, err)
	assert.Contains(t, resp.Message, "field City is a required field")
}

func TestCreatePVZ_ServiceError(t *testing.T) {
	_, pvzMock, handler := setupHandler(t)

	pvzMock.On("CreatePVZ", mock.Anything, mock.Anything).Return((*models.PVZ)(nil), errors.New("service error"))

	reqBody := api.PostPvzJSONRequestBody{
		City: "Москва",
	}

	req, rec := createRequest(http.MethodPost, "/pvz", reqBody)
	handler.CreatePVZ().ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)

	var resp api.Error
	err := json.NewDecoder(rec.Body).Decode(&resp)
	assert.NoError(t, err)
	assert.Equal(t, "failed to create pvz", resp.Message)
}
