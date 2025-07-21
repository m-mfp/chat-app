package main

import (
	"time"
)

type Message struct {
	UserID    string    `json:"userid" gorm:"type:varchar(100)"`
	MsgID     string    `json:"msgid" gorm:"type:varchar(100);primaryKey"`
	Text      string    `json:"text" gorm:"type:text"`
	CreatedAt time.Time `json:"created_at"`
}
