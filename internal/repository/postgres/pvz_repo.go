package postgres

import (
	"context"
	"database/sql"
	"fmt"
	e "pvz-service/internal/errors"
	"pvz-service/internal/models"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

func (p *Postgres) InsertPVZ(ctx context.Context, pvz *models.PVZ) error {
	_, err := p.db.ExecContext(ctx,
		"INSERT INTO pvz (id, registration_date, city_id) VALUES ($1, $2, $3)",
		pvz.ID, pvz.RegistrationDate, pvz.CityID)
	return err
}

func (p *Postgres) GetCityID(ctx context.Context, city string) (int, error) {
	var cityID int

	err := p.db.QueryRowContext(ctx, "SELECT id FROM cities WHERE name = $1", city).Scan(&cityID)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, e.ErrCityNotAllowed()
		}
		return 0, err
	}

	return cityID, nil
}

func (p *Postgres) CheckPVZ(ctx context.Context, pvzID uuid.UUID) (bool, error) {
	var exists bool
	err := p.db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM pvz WHERE id = $1)", pvzID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check PVZ existence: %w", err)
	}
	if !exists {
		return false, e.ErrNotFound()
	}

	return true, nil
}

func (p *Postgres) GetActiveReception(ctx context.Context, pvzID uuid.UUID) (*models.Reception, error) {
	var reception models.Reception
	err := p.db.QueryRowContext(ctx,
		`SELECT id, date_time, pvz_id, status 
		 FROM receptions 
		 WHERE pvz_id = $1 AND status = 'in_progress' 
		 ORDER BY date_time DESC 
		 LIMIT 1`,
		pvzID).Scan(&reception.ID, &reception.DateTime, &reception.PVZID, &reception.Status)
	if err == sql.ErrNoRows {
		return nil, e.ErrNoActiveReception()
	}
	if err != nil {
		return nil, err
	}
	return &reception, nil
}

func (p *Postgres) InsertReception(ctx context.Context, reception *models.Reception) error {
	_, err := p.db.ExecContext(ctx,
		"INSERT INTO receptions (id, date_time, pvz_id, status) VALUES ($1, $2, $3, $4)",
		reception.ID, reception.DateTime, reception.PVZID, reception.Status)
	return err
}

func (p *Postgres) GetProductTypeID(ctx context.Context, productTypeName string) (int, error) {
	var productTypeID int
	err := p.db.QueryRowContext(ctx, "SELECT id FROM product_types WHERE name = $1", productTypeName).Scan(&productTypeID)
	if err != nil {
		if err == sql.ErrNoRows {
			return productTypeID, e.ErrProductTypeNotAllowed()
		}
		return productTypeID, err
	}
	return productTypeID, nil
}

func (p *Postgres) InsertProduct(ctx context.Context, product *models.Product) error {
	_, err := p.db.ExecContext(ctx,
		"INSERT INTO products (id, date_time, type_id, reception_id) VALUES ($1, $2, $3, $4)",
		product.ID, product.DateTime, product.TypeID, product.ReceptionID)
	return err
}

func (p *Postgres) GetLastProduct(ctx context.Context, receptionID uuid.UUID) (*models.Product, error) {
	var product models.Product
	err := p.db.QueryRowContext(ctx,
		`SELECT id, date_time, type_id, reception_id 
		 FROM products 
		 WHERE reception_id = $1 
		 ORDER BY date_time DESC 
		 LIMIT 1`,
		receptionID).Scan(&product.ID, &product.DateTime, &product.TypeID, &product.ReceptionID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, e.ErrNotFound()
		}
		return nil, err
	}
	return &product, nil
}

func (p *Postgres) DeleteProduct(ctx context.Context, productID uuid.UUID) error {
	_, err := p.db.ExecContext(ctx, "DELETE FROM products WHERE id = $1", productID)
	return err
}

func (p *Postgres) UpdateReceptionStatus(ctx context.Context, receptionID uuid.UUID, status models.ReceptionStatus) error {
	_, err := p.db.ExecContext(ctx,
		"UPDATE receptions SET status = $1 WHERE id = $2",
		status, receptionID)
	return err
}

func (p *Postgres) GetPVZs(ctx context.Context, from, to time.Time, limit, offset int) ([]models.PVZ, error) {
	rows, err := p.db.QueryContext(ctx,
		`SELECT p.id, p.registration_date, p.city_id, c.name
		 FROM pvz p
		 JOIN cities c ON p.city_id = c.id
		 WHERE EXISTS (
			 SELECT 1 FROM receptions r 
			 WHERE r.pvz_id = p.id 
			 AND r.date_time BETWEEN $1 AND $2
		 )
		 ORDER BY p.registration_date DESC
		 LIMIT $3 OFFSET $4`,
		from, to, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var pvzs []models.PVZ
	for rows.Next() {
		var pvz models.PVZ
		if err := rows.Scan(&pvz.ID, &pvz.RegistrationDate, &pvz.CityID, &pvz.CityName); err != nil {
			return nil, err
		}
		pvzs = append(pvzs, pvz)
	}
	return pvzs, rows.Err()
}

func (p *Postgres) GetReceptionsForPVZs(ctx context.Context, pvzIDs []uuid.UUID, from, to time.Time) ([]models.Reception, error) {
	rows, err := p.db.QueryContext(ctx,
		`SELECT id, date_time, pvz_id, status
         FROM receptions 
         WHERE pvz_id = ANY($1) AND date_time BETWEEN $2 AND $3
         ORDER BY date_time DESC`,
		pq.Array(pvzIDs), from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var receptions []models.Reception
	for rows.Next() {
		var rec models.Reception
		if err := rows.Scan(&rec.ID, &rec.DateTime, &rec.PVZID, &rec.Status); err != nil {
			return nil, err
		}
		receptions = append(receptions, rec)
	}
	return receptions, rows.Err()
}

func (p *Postgres) GetProductsForReceptions(ctx context.Context, receptionIDs []uuid.UUID) ([]models.Product, error) {
	if len(receptionIDs) == 0 {
		return []models.Product{}, nil
	}

	rows, err := p.db.QueryContext(ctx,
		`SELECT p.id, p.date_time, p.type_id, pt.name, p.reception_id
         FROM products p
         JOIN product_types pt ON p.type_id = pt.id
         WHERE p.reception_id = ANY($1)
         ORDER BY p.date_time DESC`,
		pq.Array(receptionIDs))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []models.Product
	for rows.Next() {
		var prod models.Product
		if err := rows.Scan(&prod.ID, &prod.DateTime, &prod.TypeID, &prod.TypeName, &prod.ReceptionID); err != nil {
			return nil, err
		}
		products = append(products, prod)
	}
	return products, rows.Err()
}
