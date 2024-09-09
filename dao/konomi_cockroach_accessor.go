package dao

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	_ "github.com/lib/pq"
	"github.com/rs/zerolog/log"

	"github.com/alceccentric/beck-crawler/model"
	. "github.com/alceccentric/beck-crawler/model/gen/beck-konomi/public/table"
)

const (
	crCluster  = "beck-konomi"
	crDatabase = "beck-konomi"
)

type KonomiCRAccessor struct {
	db *sql.DB
}

func NewCRKonomiAccessor() *KonomiCRAccessor {
	userId := os.Getenv("BECK_KONOMI_COCKROACH_DB_USER")
	password := os.Getenv("BECK_KONOMI_COCKROACH_DB_PASSWORD")

	dsn := fmt.Sprintf("postgresql://%s:%s@%s-11815.6wr.aws-us-west-2.cockroachlabs.cloud:26257/%s?sslmode=verify-full",
		userId, password, crCluster, crDatabase)
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to cockroachdb")
	}
	return &KonomiCRAccessor{
		db: db,
	}
}

func (accessor *KonomiCRAccessor) Disconnect() {
	accessor.db.Close()
}

func (accessor *KonomiCRAccessor) InsertUser(user model.User) error {
	stmt := BgmUser.INSERT(BgmUser.ID, BgmUser.Nickname, BgmUser.AvatarURL, BgmUser.LastActiveTime).
		MODEL(user.ToBgmUser()).
		ON_CONFLICT(BgmUser.ID).DO_NOTHING()

	start := time.Now()
	_, err := stmt.Exec(accessor.db)
	fmt.Printf("Insert user took %s\n", time.Since(start))

	if err != nil {
		return err
	}
	return nil
}

func (accessor *KonomiCRAccessor) InsertCollection(collection model.Collection) error {
	stmt := BgmUserCollection.INSERT(BgmUserCollection.UserID, BgmUserCollection.SubjectID,
		BgmUserCollection.SubjectType, BgmUserCollection.CollectionType,
		BgmUserCollection.CollectedTime, BgmUserCollection.Rating).
		MODEL(collection.ToBgmUserCollection()).
		ON_CONFLICT(BgmUserCollection.UserID, BgmUserCollection.SubjectID).DO_NOTHING()

	start := time.Now()
	_, err := stmt.Exec(accessor.db)
	fmt.Printf("Insert collection took %s\n", time.Since(start))

	if err != nil {
		return err
	}
	return nil
}
