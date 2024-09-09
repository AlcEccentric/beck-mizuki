package dao

import model "github.com/alceccentric/beck-crawler/model"

type KonomiAccessor interface {
	InsertUser(user model.User) error
	InsertCollection(collection model.Collection) error
	Disconnect()
}
