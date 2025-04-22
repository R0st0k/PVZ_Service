// pvz_service.go
package service

import (
	"context"
	"fmt"
	"log/slog"
	e "pvz-service/internal/errors"
	"pvz-service/internal/logger/sl"
	"pvz-service/internal/models"
	"pvz-service/internal/repository"
	"time"

	"github.com/google/uuid"
)

type PVZService struct {
	repo repository.PVZRepository
	log  *slog.Logger
}

func NewPVZService(repo repository.PVZRepository, log *slog.Logger) *PVZService {
	return &PVZService{repo: repo, log: log}
}

type PVZServiceInterface interface {
	CreatePVZ(ctx context.Context, pvz *models.PVZ) (*models.PVZ, error)
	StartReception(ctx context.Context, pvzID uuid.UUID) (*models.Reception, error)
	AddProduct(ctx context.Context, pvzID uuid.UUID, productTypeName string) (*models.Product, error)
	DeleteLastProduct(ctx context.Context, pvzID uuid.UUID) error
	CloseReception(ctx context.Context, pvzID uuid.UUID) (*models.Reception, error)
	GetPVZsWithReceptions(ctx context.Context, from, to time.Time, page, limit int) ([]models.PVZInfo, error)
	GetPVZs(ctx context.Context) ([]models.PVZ, error)
}

func (s *PVZService) CreatePVZ(ctx context.Context, pvz *models.PVZ) (*models.PVZ, error) {
	const op = "service.pvz_service.CreatePVZ"

	cityID, err := s.repo.GetCityID(ctx, pvz.CityName)
	if err == e.ErrCityNotAllowed() {
		s.log.Error(fmt.Sprintf("%s: city not allowed", op), sl.Err(err))
		return nil, e.ErrCityNotAllowed()
	}
	if err != nil {
		s.log.Error(fmt.Sprintf("%s: failed to get city ID", op), sl.Err(err))
		return nil, fmt.Errorf("failed to get city ID: %w", err)
	}
	pvz.CityID = cityID

	if err := s.repo.InsertPVZ(ctx, pvz); err != nil {
		s.log.Error(fmt.Sprintf("%s: failed to create PVZ", op), sl.Err(err))
		return nil, fmt.Errorf("failed to create PVZ: %w", err)
	}

	return pvz, nil
}

func (s *PVZService) StartReception(ctx context.Context, pvzID uuid.UUID) (*models.Reception, error) {
	const op = "service.pvz_service.StartReception"

	_, err := s.repo.CheckPVZ(ctx, pvzID)
	if err == e.ErrNotFound() {
		s.log.Info(fmt.Sprintf("%s: city not found", op), sl.Err(err))
		return nil, e.ErrCityNotAllowed()
	}
	if err != nil {
		s.log.Error(fmt.Sprintf("%s: failed to get city ID", op), sl.Err(err))
		return nil, fmt.Errorf("failed to get city ID: %w", err)
	}

	// Check for existing active reception
	activeReception, err := s.repo.GetActiveReception(ctx, pvzID)
	if err != nil && err != e.ErrNoActiveReception() {
		s.log.Error(fmt.Sprintf("%s: failed to check active receptions", op), sl.Err(err))
		return nil, fmt.Errorf("failed to check active receptions: %w", err)
	}
	if activeReception != nil {
		s.log.Info(fmt.Sprintf("%s: active reception exists", op), "pvzID", pvzID)
		return nil, e.ErrActiveReceptionExists()
	}

	reception := &models.Reception{
		ID:       uuid.New(),
		DateTime: time.Now(),
		PVZID:    pvzID,
		Status:   models.ReceptionStatusInProgress,
	}

	// Simple repository call
	if err := s.repo.InsertReception(ctx, reception); err != nil {
		s.log.Error(fmt.Sprintf("%s: failed to create reception", op), sl.Err(err))
		return nil, fmt.Errorf("failed to create reception: %w", err)
	}

	return reception, nil
}

func (s *PVZService) AddProduct(ctx context.Context, pvzID uuid.UUID, productTypeName string) (*models.Product, error) {
	const op = "service.pvz_service.AddProduct"

	// Check for active reception
	reception, err := s.repo.GetActiveReception(ctx, pvzID)
	if err != nil {
		if err == e.ErrNoActiveReception() {
			s.log.Info(fmt.Sprintf("%s: no active reception", op), "pvzID", pvzID)
			return nil, e.ErrNoActiveReception()
		}
		s.log.Error(fmt.Sprintf("%s: failed to get active reception", op), sl.Err(err))
		return nil, fmt.Errorf("failed to get active reception: %w", err)
	}

	// Check product type
	productTypeID, err := s.repo.GetProductTypeID(ctx, productTypeName)
	if err != nil {
		if err == e.ErrProductTypeNotAllowed() {
			s.log.Info(fmt.Sprintf("%s: product type not allowed", op), "type", productTypeName)
			return nil, e.ErrProductTypeNotAllowed()
		}
		s.log.Error(fmt.Sprintf("%s: failed to get product type", op), sl.Err(err))
		return nil, fmt.Errorf("failed to get product type: %w", err)
	}

	product := &models.Product{
		ID:          uuid.New(),
		DateTime:    time.Now(),
		TypeID:      productTypeID,
		TypeName:    productTypeName,
		ReceptionID: reception.ID,
	}

	// Simple repository call
	if err := s.repo.InsertProduct(ctx, product); err != nil {
		s.log.Error(fmt.Sprintf("%s: failed to add product", op), sl.Err(err))
		return nil, fmt.Errorf("failed to add product: %w", err)
	}

	return product, nil
}

func (s *PVZService) DeleteLastProduct(ctx context.Context, pvzID uuid.UUID) error {
	const op = "service.pvz_service.DeleteLastProduct"

	// Check for active reception
	reception, err := s.repo.GetActiveReception(ctx, pvzID)
	if err != nil {
		if err == e.ErrNoActiveReception() {
			s.log.Info(fmt.Sprintf("%s: no active reception", op), "pvzID", pvzID)
			return e.ErrNoActiveReception()
		}
		s.log.Error(fmt.Sprintf("%s: failed to get active reception", op), sl.Err(err))
		return fmt.Errorf("failed to get active reception: %w", err)
	}

	// Get last product
	product, err := s.repo.GetLastProduct(ctx, reception.ID)
	if err != nil {
		if err == e.ErrNotFound() {
			s.log.Info(fmt.Sprintf("%s: no products to delete", op), "receptionID", reception.ID)
			return e.ErrNoProduct()
		}
		s.log.Error(fmt.Sprintf("%s: failed to get last product", op), sl.Err(err))
		return fmt.Errorf("failed to get last product: %w", err)
	}

	// Simple repository call
	if err := s.repo.DeleteProduct(ctx, product.ID); err != nil {
		s.log.Error(fmt.Sprintf("%s: failed to delete product", op), sl.Err(err))
		return fmt.Errorf("failed to delete product: %w", err)
	}

	return nil
}

func (s *PVZService) CloseReception(ctx context.Context, pvzID uuid.UUID) (*models.Reception, error) {
	const op = "service.pvz_service.CloseReception"

	// Check for active reception
	reception, err := s.repo.GetActiveReception(ctx, pvzID)
	if err != nil {
		if err == e.ErrNoActiveReception() {
			s.log.Info(fmt.Sprintf("%s: no active reception", op), "pvzID", pvzID)
			return nil, e.ErrNoActiveReception()
		}
		s.log.Error(fmt.Sprintf("%s: failed to get active reception", op), sl.Err(err))
		return nil, fmt.Errorf("failed to get active reception: %w", err)
	}

	// Simple repository call
	if err := s.repo.UpdateReceptionStatus(ctx, reception.ID, models.ReceptionStatusClose); err != nil {
		s.log.Error(fmt.Sprintf("%s: failed to close reception", op), sl.Err(err))
		return nil, fmt.Errorf("failed to close reception: %w", err)
	}

	reception.Status = models.ReceptionStatusClose
	return reception, nil
}

func (s *PVZService) GetPVZsWithReceptions(ctx context.Context, from, to time.Time, page, limit int) ([]models.PVZInfo, error) {
	const op = "service.pvz_service.GetPVZsWithReceptions"

	offset := (page - 1) * limit

	// Simple repository calls
	pvzs, err := s.repo.GetPVZs(ctx, from, to, limit, offset)
	if err != nil {
		s.log.Error(fmt.Sprintf("%s: failed to get PVZs", op), sl.Err(err))
		return nil, fmt.Errorf("failed to get PVZs: %w", err)
	}

	if len(pvzs) == 0 {
		return []models.PVZInfo{}, nil
	}

	pvzIDs := make([]uuid.UUID, len(pvzs))
	for i, pvz := range pvzs {
		pvzIDs[i] = pvz.ID
	}

	receptions, err := s.repo.GetReceptionsForPVZs(ctx, pvzIDs, from, to)
	if err != nil {
		s.log.Error(fmt.Sprintf("%s: failed to get receptions", op), sl.Err(err))
		return nil, fmt.Errorf("failed to get receptions: %w", err)
	}

	receptionIDs := make([]uuid.UUID, len(receptions))
	for i, rec := range receptions {
		receptionIDs[i] = rec.ID
	}

	products, err := s.repo.GetProductsForReceptions(ctx, receptionIDs)
	if err != nil {
		s.log.Error(fmt.Sprintf("%s: failed to get products", op), sl.Err(err))
		return nil, fmt.Errorf("failed to get products: %w", err)
	}

	// Build hierarchical response
	return s.buildPVZResponse(pvzs, receptions, products), nil
}

// Helper function: Build hierarchical response
func (s *PVZService) buildPVZResponse(pvzs []models.PVZ, receptions []models.Reception, products []models.Product) []models.PVZInfo {
	// Create maps for quick lookup
	receptionsByPVZ := make(map[uuid.UUID][]models.Reception)
	for _, rec := range receptions {
		receptionsByPVZ[rec.PVZID] = append(receptionsByPVZ[rec.PVZID], rec)
	}

	productsByReception := make(map[uuid.UUID][]models.Product)
	for _, prod := range products {
		productsByReception[prod.ReceptionID] = append(productsByReception[prod.ReceptionID], prod)
	}

	// Build response
	var response []models.PVZInfo
	for _, pvz := range pvzs {
		pvzResp := models.PVZInfo{
			PVZ: models.PVZ{
				ID:               pvz.ID,
				RegistrationDate: pvz.RegistrationDate,
				CityName:         pvz.CityName,
			},
			Receptions: []models.ReceptionInfo{},
		}

		for _, rec := range receptionsByPVZ[pvz.ID] {
			recProducts := productsByReception[rec.ID]
			productDTOs := make([]models.Product, len(recProducts))
			for i, prod := range recProducts {
				productDTOs[i] = models.Product{
					ID:          prod.ID,
					DateTime:    prod.DateTime,
					TypeName:    prod.TypeName,
					ReceptionID: prod.ReceptionID,
				}
			}

			pvzResp.Receptions = append(pvzResp.Receptions, models.ReceptionInfo{
				Reception: models.Reception{
					ID:       rec.ID,
					DateTime: rec.DateTime,
					PVZID:    rec.PVZID,
					Status:   rec.Status,
				},
				Products: productDTOs,
			})
		}

		response = append(response, pvzResp)
	}

	return response
}

func (s *PVZService) GetPVZs(ctx context.Context) ([]models.PVZ, error) {
	const op = "service.pvz_service.GetPVZs"

	// Simple repository calls
	pvzs, err := s.repo.GetPVZsWithNoFilter(ctx)
	if err != nil {
		s.log.Error(fmt.Sprintf("%s: failed to get PVZs", op), sl.Err(err))
		return nil, fmt.Errorf("failed to get PVZs: %w", err)
	}

	return pvzs, nil
}
