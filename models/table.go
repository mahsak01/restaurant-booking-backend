package models

// TableStatus table status type
type TableStatus string

const (
	TableStatusAvailable TableStatus = "available"
	TableStatusReserved  TableStatus = "reserved"
	TableStatusOccupied  TableStatus = "occupied"
	TableStatusMaintenance TableStatus = "maintenance"
)

// Table table model
type Table struct {
	BaseModel
	Number   int         `gorm:"not null;uniqueIndex" json:"number"`
	Capacity int         `gorm:"not null" json:"capacity"`
	Location string      `gorm:"not null" json:"location"`
	Status   TableStatus `gorm:"type:varchar(20);default:'available'" json:"status"`
	
	// Relationships
	Reservations []Reservation `gorm:"foreignKey:TableID" json:"reservations,omitempty"`
}

