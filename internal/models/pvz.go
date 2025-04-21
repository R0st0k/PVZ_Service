package models

import (
	"time"

	"github.com/google/uuid"
)

type PVZ struct {
	ID               uuid.UUID `db:"id" json:"id"`
	RegistrationDate time.Time `db:"registration_date" json:"registrationDate"`
	CityID           int       `db:"city_id" json:"-"`
	CityName         string    `db:"city_name" json:"city"`
}
