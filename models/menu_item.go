package models

// MenuCategory menu item category type
type MenuCategory string

const (
	CategoryAppetizer MenuCategory = "appetizer"
	CategoryMain      MenuCategory = "main"
	CategoryDessert   MenuCategory = "dessert"
	CategoryDrink     MenuCategory = "drink"
)

// MenuItem menu item model
type MenuItem struct {
	BaseModel
	Name        string      `gorm:"not null" json:"name"`
	Description string      `gorm:"type:text" json:"description"`
	Price       float64     `gorm:"not null" json:"price"`
	ImageURL    string      `gorm:"type:varchar(500)" json:"image_url"`
	Category    MenuCategory `gorm:"type:varchar(50);not null" json:"category"`
}

