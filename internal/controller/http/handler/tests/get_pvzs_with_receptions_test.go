package tests

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"pvz-service/internal/models"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetPVZsWithReceptions_Success(t *testing.T) {
	_, pvzMock, handler := setupHandler(t)

	from := time.Now().Truncate(time.Second).Add(-24 * time.Hour)
	to := time.Now().Truncate(time.Second).Add(24 * time.Hour)
	page := 1
	limit := 10

	expectedPVZs := []models.PVZInfo{
		{
			PVZ: models.PVZ{
				ID:               uuid.New(),
				CityName:         "Москва",
				RegistrationDate: time.Now(),
			},
			Receptions: []models.ReceptionInfo{
				{
					Reception: models.Reception{
						ID:       uuid.New(),
						DateTime: time.Now(),
					},
				},
			},
		},
	}

	pvzMock.On("GetPVZsWithReceptions", mock.Anything, from, to, page, limit).Return(expectedPVZs, nil)

	fmt.Println(from.Format(time.RFC3339))

	req, rec := createRequest(http.MethodGet, fmt.Sprintf("/pvz?startDate=%s&endDate=%s&page=%d&limit=%d",
		url.QueryEscape(from.Format(time.RFC3339)),
		url.QueryEscape(to.Format(time.RFC3339)),
		page,
		limit,
	), nil)
	handler.GetPVZsWithReceptions().ServeHTTP(rec, req)

	fmt.Println(rec.Body)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp []models.PVZInfo
	err := json.NewDecoder(rec.Body).Decode(&resp)
	assert.NoError(t, err)
	assert.Len(t, resp, 1)
}

func TestGetPVZsWithReceptions_DefaultParams(t *testing.T) {
	_, pvzMock, handler := setupHandler(t)

	pvzMock.On("GetPVZsWithReceptions", mock.Anything, time.Time{}, mock.AnythingOfType("time.Time"), 1, 10).
		Return([]models.PVZInfo{}, nil)

	req, rec := createRequest(http.MethodGet, "/pvz", nil)
	handler.GetPVZsWithReceptions().ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestGetPVZsWithReceptions_InvalidDate(t *testing.T) {
	_, _, handler := setupHandler(t)

	req, rec := createRequest(http.MethodGet, "/pvz?startDate=invalid", nil)
	handler.GetPVZsWithReceptions().ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}
