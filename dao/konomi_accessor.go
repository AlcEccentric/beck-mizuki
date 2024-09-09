package dao

import model "github.com/alceccentric/beck-crawler/model"

type KonomiAccessor interface {
	InsertUser(user model.User) error
	BatchInsertUser(user []model.User, size int) error
	InsertCollection(collection model.Collection) error
	BatchInsertCollection(collections []model.Collection, size int) error
	Disconnect()
}
