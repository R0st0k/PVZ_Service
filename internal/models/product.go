package models

import (
	"time"

	"github.com/google/uuid"
)

type Product struct {
	ID          uuid.UUID `db:"id" json:"id"`
	DateTime    time.Time `db:"date_time" json:"dateTime"`
	TypeID      int       `db:"type_id" json:"-"`
	TypeName    string    `db:"type_name" json:"type"`
	ReceptionID uuid.UUID `db:"reception_id" json:"receptionId"`
}
