package models

// OrderStatus order status type
type OrderStatus string

const (
	OrderStatusPending   OrderStatus = "pending"   // Waiting for confirmation
	OrderStatusConfirmed OrderStatus = "confirmed" // Confirmed by admin
	OrderStatusPreparing OrderStatus = "preparing" // Being prepared
	OrderStatusReady     OrderStatus = "ready"     // Ready for delivery
	OrderStatusDelivered OrderStatus = "delivered" // Delivered to customer
	OrderStatusCancelled OrderStatus = "cancelled" // Cancelled
)

// Order order model
type Order struct {
	BaseModel
	UserID     uint        `gorm:"not null;index" json:"user_id"`
	Status     OrderStatus `gorm:"type:varchar(20);default:'pending'" json:"status"`
	TotalPrice float64     `gorm:"not null" json:"total_price"` // Total price of all items

	// Relationships
	User       User        `gorm:"foreignKey:UserID" json:"user,omitempty"`
	OrderItems []OrderItem `gorm:"foreignKey:OrderID" json:"order_items,omitempty"`
}

// OrderItem order item model (many-to-many relationship between Order and MenuItem)
type OrderItem struct {
	BaseModel
	OrderID    uint    `gorm:"not null;index" json:"order_id"`
	MenuItemID uint    `gorm:"not null;index" json:"menu_item_id"`
	Quantity   int     `gorm:"not null" json:"quantity"`
	Price      float64 `gorm:"not null" json:"price"` // Price at the time of order (snapshot)

	// Relationships
	Order    Order    `gorm:"foreignKey:OrderID" json:"order,omitempty"`
	MenuItem MenuItem `gorm:"foreignKey:MenuItemID" json:"menu_item,omitempty"`
}
