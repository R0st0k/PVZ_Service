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

func TestCloseReception_Success(t *testing.T) {
	_, pvzMock, handler := setupHandler(t)

	pvzID := uuid.New()
	expectedReception := &models.Reception{
		ID:       uuid.New(),
		DateTime: time.Now(),
		PVZID:    pvzID,
		Status:   models.ReceptionStatusClose,
	}

	pvzMock.On("CloseReception", mock.Anything, pvzID).Return(expectedReception, nil)

	req, rec := createRequest(http.MethodPost, "/pvz/"+pvzID.String()+"/receptions/close", nil)
	req = addURLParams(req, map[string]string{"pvzId": pvzID.String()})
	handler.CloseReception().ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp api.Reception
	err := json.NewDecoder(rec.Body).Decode(&resp)
	assert.NoError(t, err)
	assert.Equal(t, expectedReception.ID, *resp.Id)
	assert.Equal(t, api.Close, resp.Status)
}

func TestCloseReception_NoActiveReception(t *testing.T) {
	_, pvzMock, handler := setupHandler(t)

	pvzID := uuid.New()
	pvzMock.On("CloseReception", mock.Anything, pvzID).Return(
		(*models.Reception)(nil), e.ErrNoActiveReception(),
	)

	req, rec := createRequest(http.MethodPost, "/pvz/"+pvzID.String()+"/receptions/close", nil)
	req = addURLParams(req, map[string]string{"pvzId": pvzID.String()})
	handler.CloseReception().ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}
