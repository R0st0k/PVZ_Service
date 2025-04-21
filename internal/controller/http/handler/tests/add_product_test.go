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

func TestAddProduct_Success(t *testing.T) {
	_, pvzMock, handler := setupHandler(t)

	pvzID := uuid.New()
	productType := "одежда"
	expectedProduct := &models.Product{
		ID:          uuid.New(),
		DateTime:    time.Now(),
		TypeName:    productType,
		ReceptionID: uuid.New(),
	}

	pvzMock.On("AddProduct", mock.Anything, pvzID, productType).Return(expectedProduct, nil)

	reqBody := api.PostProductsJSONRequestBody{
		PvzId: pvzID,
		Type:  "одежда",
	}

	req, rec := createRequest(http.MethodPost, "/products", reqBody)
	handler.AddProduct().ServeHTTP(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)

	var resp api.Product
	err := json.NewDecoder(rec.Body).Decode(&resp)
	assert.NoError(t, err)
	assert.Equal(t, expectedProduct.ID, *resp.Id)
	assert.Equal(t, productType, string(resp.Type))
}

func TestAddProduct_NoActiveReception(t *testing.T) {
	_, pvzMock, handler := setupHandler(t)

	pvzID := uuid.New()
	pvzMock.On("AddProduct", mock.Anything, pvzID, mock.Anything).Return(
		(*models.Product)(nil), e.ErrNoActiveReception(),
	)

	reqBody := api.PostProductsJSONRequestBody{
		PvzId: pvzID,
		Type:  "одежда",
	}

	req, rec := createRequest(http.MethodPost, "/products", reqBody)
	handler.AddProduct().ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var resp api.Error
	err := json.NewDecoder(rec.Body).Decode(&resp)
	assert.NoError(t, err)
	assert.Equal(t, "no active reception", resp.Message)
}
