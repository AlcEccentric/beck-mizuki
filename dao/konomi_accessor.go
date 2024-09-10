package dao

import (
	model "github.com/alceccentric/beck-crawler/model"
	jet "github.com/go-jet/jet/v2/postgres"
)

type KonomiAccessor interface {
	GetRowCount(table jet.Table) (int, error)
	GetUser(uid string) (model.User, error)
	GetUserIdsPaginated(offset, limit int) ([]string, error)
	InsertUser(user model.User) error
	BatchInsertUser(user []model.User, size int) error
	DeleteUser(uid string) error
	InsertCollection(collection model.Collection) error
	BatchInsertCollection(collections []model.Collection, size int) error
	DeleteCollectionByUid(uid string) error
	Disconnect()
}
