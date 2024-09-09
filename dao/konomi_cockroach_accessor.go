package dao

import (
	"fmt"
	"os"

	model "github.com/alceccentric/beck-crawler/model"
	"github.com/rs/zerolog/log"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

const (
	crCluster  = "beck-konomi"
	crDatabase = "beck-konomi"
)

type KonomiCRAccessor struct {
	db *gorm.DB
}

func NewCRKonomiAccessor() *KonomiCRAccessor {
	userId := os.Getenv("BECK_KONOMI_COCKROACH_DB_USER")
	password := os.Getenv("BECK_KONOMI_COCKROACH_DB_PASSWORD")

	dsn := fmt.Sprintf("postgresql://%s:%s@%s-11815.6wr.aws-us-west-2.cockroachlabs.cloud:26257/%s?sslmode=verify-full",
		userId, password, crCluster, crDatabase)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to cockroachdb")
	}
	return &KonomiCRAccessor{
		db: db,
	}
}

func (accessor *KonomiCRAccessor) Disconnect() {
	// not needed for gorm connection
}

func (accessor *KonomiCRAccessor) InsertUser(user model.User) error {
	err := accessor.db.Create(&user)
	if err.Error != nil {
		return err.Error
	}
	return nil
}

func (accessor *KonomiCRAccessor) InsertCollection(collection model.Collection) error {
	err := accessor.db.Create(&collection)
	if err.Error != nil {
		return err.Error
	}
	return nil
}
