package model

import (
	"time"

	jetmodel "github.com/alceccentric/beck-crawler/model/gen/beck-konomi/public/model"
)

type User struct {
	ID             string    `bson:"_id" gorm:"primaryKey;column:id"`
	Nickname       string    `bson:"nickname,omitempty" gorm:"column:nickname"`
	AvatarURL      string    `bson:"avatar_url" gorm:"column:avatar_url"`
	LastActiveTime time.Time `bson:"last_active_time" gorm:"column:last_active_time"`
}

// convert to jet generated model
func (u *User) ToBgmUser() jetmodel.BgmUser {
	return jetmodel.BgmUser{
		ID:             u.ID,
		Nickname:       &u.Nickname,
		AvatarURL:      &u.AvatarURL,
		LastActiveTime: &u.LastActiveTime,
	}
}

func ToBgmUsers(users []User) []jetmodel.BgmUser {
	bgmUsers := make([]jetmodel.BgmUser, 0, len(users))
	for _, user := range users {
		bgmUsers = append(bgmUsers, user.ToBgmUser())
	}
	return bgmUsers
}

func FromBgmUser(bgmUser jetmodel.BgmUser) User {
	return User{
		ID:             bgmUser.ID,
		Nickname:       *bgmUser.Nickname,
		AvatarURL:      *bgmUser.AvatarURL,
		LastActiveTime: *bgmUser.LastActiveTime,
	}
}

func FromBgmUsers(bgmUsers []jetmodel.BgmUser) []User {
	users := make([]User, 0, len(bgmUsers))
	for _, bgmUser := range bgmUsers {
		users = append(users, FromBgmUser(bgmUser))
	}
	return users
}
