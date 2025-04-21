package tests

import (
	"net/http"
	e "pvz-service/internal/errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestDeleteLastProduct_Success(t *testing.T) {
	_, pvzMock, handler := setupHandler(t)

	pvzID := uuid.New()
	pvzMock.On("DeleteLastProduct", mock.Anything, pvzID).Return(nil)

	req, rec := createRequest(http.MethodDelete, "/pvz/"+pvzID.String()+"/products/last", nil)
	req = addURLParams(req, map[string]string{"pvzId": pvzID.String()})
	handler.DeleteLastProduct().ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, 0, rec.Body.Len()) // Check empty response
}

func TestDeleteLastProduct_NoActiveReception(t *testing.T) {
	_, pvzMock, handler := setupHandler(t)

	pvzID := uuid.New()
	pvzMock.On("DeleteLastProduct", mock.Anything, pvzID).Return(e.ErrNoActiveReception())

	req, rec := createRequest(http.MethodDelete, "/pvz/"+pvzID.String()+"/products/last", nil)
	req = addURLParams(req, map[string]string{"pvzId": pvzID.String()})
	handler.DeleteLastProduct().ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestDeleteLastProduct_NoProducts(t *testing.T) {
	_, pvzMock, handler := setupHandler(t)

	pvzID := uuid.New()
	pvzMock.On("DeleteLastProduct", mock.Anything, pvzID).Return(e.ErrNoProduct())

	req, rec := createRequest(http.MethodDelete, "/pvz/"+pvzID.String()+"/products/last", nil)
	req = addURLParams(req, map[string]string{"pvzId": pvzID.String()})
	handler.DeleteLastProduct().ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestDeleteLastProduct_InvalidUUID(t *testing.T) {
	_, _, handler := setupHandler(t)

	req, rec := createRequest(http.MethodDelete, "/pvz/invalid_uuid/products/last", nil)
	req = addURLParams(req, map[string]string{"pvzId": "invalid_uuid"})
	handler.DeleteLastProduct().ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}
