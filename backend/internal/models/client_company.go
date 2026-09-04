package models

import "time"

type ClientCompany struct {
	ID         uint       `gorm:"primaryKey" json:"id"`
	Name       string     `gorm:"size:255;not null" json:"name"`
	Logo       *string    `gorm:"size:255" json:"logo,omitempty"`
	Address    *string    `json:"address,omitempty"`
	PostalCode *string    `gorm:"size:20" json:"postal_code,omitempty"`
	City       *string    `gorm:"size:255" json:"city,omitempty"`
	CountryID  *uint      `json:"-"`
	Country    *Country   `gorm:"foreignKey:CountryID" json:"country,omitempty"`
	CurrencyID *uint      `json:"-"`
	Currency   *Currency  `gorm:"foreignKey:CurrencyID" json:"currency,omitempty"`
	Email      *string    `gorm:"size:255;uniqueIndex" json:"email,omitempty"`
	Phone      *string    `gorm:"size:50" json:"phone,omitempty"`
	Web        *string    `gorm:"size:255" json:"web,omitempty"`
	Clients    []User     `gorm:"many2many:client_company_user;" json:"clients,omitempty"`
	ArchivedAt *time.Time `json:"archived_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}
