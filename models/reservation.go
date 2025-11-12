package models

import "time"

// ReservationStatus reservation status type
type ReservationStatus string

const (
	ReservationStatusPending   ReservationStatus = "pending"
	ReservationStatusConfirmed ReservationStatus = "confirmed"
	ReservationStatusCancelled ReservationStatus = "cancelled"
	ReservationStatusCompleted ReservationStatus = "completed"
)

// Reservation reservation model
type Reservation struct {
	BaseModel
	UserID   uint              `gorm:"not null;index" json:"user_id"`
	TableID  uint              `gorm:"not null;index" json:"table_id"`
	Date     time.Time         `gorm:"type:date;not null" json:"date"`
	Time     time.Time         `gorm:"type:time;not null" json:"time"`
	Status   ReservationStatus `gorm:"type:varchar(20);default:'pending'" json:"status"`
	
	// Relationships
	User  User  `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Table Table `gorm:"foreignKey:TableID" json:"table,omitempty"`
}

