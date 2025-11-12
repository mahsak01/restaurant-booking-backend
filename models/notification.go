package models

// NotificationType notification type
type NotificationType string

const (
	NotificationTypeReservation NotificationType = "reservation"
	NotificationTypeSystem      NotificationType = "system"
	NotificationTypePromotion   NotificationType = "promotion"
)

// Notification notification model
type Notification struct {
	BaseModel
	UserID  uint            `gorm:"not null;index" json:"user_id"`
	Message string          `gorm:"type:text;not null" json:"message"`
	Type    NotificationType `gorm:"type:varchar(50);not null" json:"type"`
	IsRead  bool            `gorm:"default:false;index" json:"is_read"`
	
	// Relationships
	User User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

