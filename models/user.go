package models

import (
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// UserRole user role type
type UserRole string

const (
	RoleAdmin    UserRole = "admin"
	RoleCustomer UserRole = "customer"
)

// User user model
type User struct {
	BaseModel
	Phone    string   `gorm:"uniqueIndex;not null" json:"phone"`
	Password string   `gorm:"not null" json:"-"`
	Name     string   `gorm:"not null" json:"name"`               // First name
	LastName string   `gorm:"type:varchar(100)" json:"last_name"` // Last name (optional)
	Role     UserRole `gorm:"type:varchar(20);default:'customer'" json:"role"`

	// Relationships
	Reservations  []Reservation  `gorm:"foreignKey:UserID" json:"reservations,omitempty"`
	Notifications []Notification `gorm:"foreignKey:UserID" json:"notifications,omitempty"`
	Orders        []Order        `gorm:"foreignKey:UserID" json:"orders,omitempty"`
}

// BeforeCreate hash password before creating user
func (u *User) BeforeCreate(tx *gorm.DB) error {
	// If password is empty, set a temporary password
	if u.Password == "" {
		u.Password = "TEMP_PASSWORD_NO_LOGIN" // Temporary password, user must set password to login
	}

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hashedPassword)
	return nil
}

// CheckPassword checks if provided password matches user's password
func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	return err == nil
}

// IsAdmin checks if user is admin
func (u *User) IsAdmin() bool {
	return u.Role == RoleAdmin
}
