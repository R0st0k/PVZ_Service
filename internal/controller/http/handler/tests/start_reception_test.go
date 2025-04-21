package tests

import (
	"encoding/json"
	"net/http"
	api "pvz-service/api/generated"
	e "pvz-service/internal/errors"
	"pvz-service/internal/models"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestStartReception_Success(t *testing.T) {
	_, pvzMock, handler := setupHandler(t)

	pvzID := uuid.New()
	expectedReception := &models.Reception{
		ID:       uuid.New(),
		DateTime: time.Now(),
		PVZID:    pvzID,
		Status:   models.ReceptionStatusInProgress,
	}

	pvzMock.On("StartReception", mock.Anything, pvzID).Return(expectedReception, nil)

	reqBody := api.PostReceptionsJSONRequestBody{
		PvzId: pvzID,
	}

	req, rec := createRequest(http.MethodPost, "/receptions", reqBody)
	handler.StartReception().ServeHTTP(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)

	var resp api.Reception
	err := json.NewDecoder(rec.Body).Decode(&resp)
	assert.NoError(t, err)
	assert.Equal(t, expectedReception.ID, *resp.Id)
	assert.Equal(t, pvzID, resp.PvzId)
}

func TestStartReception_ActiveReceptionExists(t *testing.T) {
	_, pvzMock, handler := setupHandler(t)

	pvzID := uuid.New()
	pvzMock.On("StartReception", mock.Anything, pvzID).Return(
		(*models.Reception)(nil), e.ErrActiveReceptionExists(),
	)

	reqBody := api.PostReceptionsJSONRequestBody{
		PvzId: pvzID,
	}

	req, rec := createRequest(http.MethodPost, "/receptions", reqBody)
	handler.StartReception().ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var resp api.Error
	err := json.NewDecoder(rec.Body).Decode(&resp)
	assert.NoError(t, err)
	assert.Equal(t, "active reception exists", resp.Message)
}
