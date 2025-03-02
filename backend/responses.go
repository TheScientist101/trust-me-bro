package main

import "github.com/google/uuid"

type TransactionResponse struct {
	UUID        uuid.UUID   `gorm:"primaryKey;type:uuid"`
	User        User        `gorm:"ForeignKey:UUID"`
	Trust       bool        `gorm:"not null"`
	Transaction Transaction `gorm:"ForeignKey:UUID"`
}
