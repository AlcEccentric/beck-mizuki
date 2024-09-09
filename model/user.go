package model

import "time"

const (
	crUserTable = "bgm_user"
)

type User struct {
	ID             string    `bson:"_id" gorm:"primaryKey;column:id"`
	Nickname       string    `bson:"nickname,omitempty" gorm:"column:nickname"`
	AvatarURL      string    `bson:"avatar_url" gorm:"column:avatar_url"`
	LastActiveTime time.Time `bson:"last_active_time" gorm:"column:last_active_time"`
}

// It's for gorm to identify the target table
func (User) TableName() string {
	return crUserTable
}
