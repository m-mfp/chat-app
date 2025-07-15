package main

import (
	"time"

	"gorm.io/gorm"
)

type Message struct {
	gorm.Model
	Username  string
	Content   string
	CreatedAt time.Time
}
