package models

import (
	"time"

	"github.com/google/uuid"
)

type ReceptionStatus string

const (
	ReceptionStatusInProgress ReceptionStatus = "in_progress"
	ReceptionStatusClose      ReceptionStatus = "close"
)

type Reception struct {
	ID       uuid.UUID       `db:"id" json:"id"`
	DateTime time.Time       `db:"date_time" json:"dateTime"`
	PVZID    uuid.UUID       `db:"pvz_id" json:"pvzId"`
	Status   ReceptionStatus `db:"status" json:"status"`
}
