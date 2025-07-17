package main

import (
	"time"

	"gorm.io/gorm"
)

type Message struct {
	gorm.Model
	UserID    string `json:"userid"`
	MsgID     string `json:"msgid"`
	Text      string `json:"text"`
	CreatedAt time.Time
}
