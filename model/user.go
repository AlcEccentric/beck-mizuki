package model

import "time"

type User struct {
	ID             string    `bson:"_id"`
	Nickname       string    `bson:"nickname,omitempty"`
	AvatarURL      string    `bson:"avatar_url"`
	LastActiveTime time.Time `bson:"last_active_time"`
}
