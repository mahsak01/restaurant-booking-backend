package models

// Category menu category model
type Category struct {
	BaseModel
	Name        string `gorm:"uniqueIndex;not null" json:"name"` // Category name (e.g., "appetizer", "main")
	DisplayName string `gorm:"not null" json:"display_name"`     // Display name (e.g., "Appetizer", "Main Course")
	Description string `gorm:"type:text" json:"description"`     // Optional description
	IsActive    bool   `gorm:"default:true" json:"is_active"`    // Category status
	SortOrder   int    `gorm:"default:0" json:"sort_order"`      // Sort order for display
}
