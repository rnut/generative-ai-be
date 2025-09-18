package auth

import "time"

// User represents a system user.
type User struct {
	ID           uint       `json:"id" gorm:"primaryKey"`
	Email        string     `json:"email" gorm:"uniqueIndex;size:255;not null"`
	PasswordHash string     `json:"-" gorm:"size:255;not null"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	LastLoginAt  *time.Time `json:"last_login_at"`
	IsActive     bool       `json:"-" gorm:"default:true"`
	// Profile fields
	FirstName       *string    `json:"first_name" gorm:"size:100"`
	LastName        *string    `json:"last_name" gorm:"size:100"`
	Phone           *string    `json:"phone" gorm:"size:20;index"`
	MembershipLevel string     `json:"membership_level" gorm:"size:20;default:Bronze"`
	MembershipCode  *string    `json:"membership_code" gorm:"size:50;uniqueIndex"`
	Points          int        `json:"points" gorm:"default:0"`
	JoinedAt        *time.Time `json:"joined_at"`
}
