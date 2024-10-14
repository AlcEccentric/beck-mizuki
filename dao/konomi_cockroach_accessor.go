package dao

import (
	"database/sql"
	"errors"
	"fmt"
	"os"

	. "github.com/go-jet/jet/v2/postgres"
	_ "github.com/lib/pq"
	"github.com/rs/zerolog/log"

	"github.com/AlcEccentric/beck-mizuki/model"
	jetmodel "github.com/AlcEccentric/beck-mizuki/model/gen/beck-konomi/public/model"
	. "github.com/AlcEccentric/beck-mizuki/model/gen/beck-konomi/public/table"
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

func (accessor *KonomiCRAccessor) GetRowCount(table Table) (int, error) {
	stmt := table.SELECT(COUNT(STAR)).
		FROM(table)

	var rows []struct {
		Count int
	}
	err := stmt.Query(accessor.db, &rows)

	if err != nil {
		return 0, err
	}

	return rows[0].Count, nil
}

func (accessor *KonomiCRAccessor) GetUserIdsPaginated(offset, limit int) ([]string, error) {
	stmt := BgmUser.SELECT(BgmUser.ID).
		FROM(BgmUser).
		LIMIT(int64(limit)).
		OFFSET(int64(offset))

	var rows []string
	err := stmt.Query(accessor.db, &rows)

	if err != nil {
		return nil, err
	}

	return rows, nil
}

func (accessor *KonomiCRAccessor) GetUser(uid string) (model.User, error) {
	stmt := BgmUser.SELECT(BgmUser.AllColumns).
		FROM(BgmUser).
		WHERE(BgmUser.ID.EQ(String(uid)))

	var rows []jetmodel.BgmUser
	err := stmt.Query(accessor.db, &rows)

	if err != nil {
		return model.User{}, err
	}

	if len(rows) == 0 {
		return model.User{}, errors.New("user not found")
	}

	return model.FromBgmUser(rows[0]), nil
}

func (accessor *KonomiCRAccessor) InsertUser(user model.User) error {
	stmt := BgmUser.INSERT(BgmUser.AllColumns).
		MODEL(user.ToBgmUser()).
		ON_CONFLICT(BgmUser.ID).
		DO_UPDATE(SET(
			BgmUser.Nickname.SET(String(user.Nickname)),
			BgmUser.AvatarURL.SET(String(user.AvatarURL)),
			BgmUser.LastActiveTime.SET(TimestampzT(user.LastActiveTime)),
		))

	_, err := stmt.Exec(accessor.db)

	if err != nil {
		return err
	}
	return nil
}

func (accessor *KonomiCRAccessor) BatchInsertUser(users []model.User, batchSize int) error {

	startIdx := 0
	errs := make([]error, 0)
	for startIdx < len(users) {
		endIdx := startIdx + batchSize
		if endIdx > len(users) {
			endIdx = len(users)
		}
		stmt := BgmUser.INSERT(BgmUser.AllColumns).
			MODELS(model.ToBgmUsers(users[startIdx:endIdx])).
			ON_CONFLICT(BgmUser.ID).DO_NOTHING()

		_, err := stmt.Exec(accessor.db)

		if err != nil {
			errs = append(errs, err)
		}
		startIdx = endIdx
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	return nil
}

func (accessor *KonomiCRAccessor) DeleteUser(uid string) error {
	stmt := BgmUser.DELETE().
		WHERE(BgmUser.ID.EQ(String(uid)))

	_, err := stmt.Exec(accessor.db)
	if err != nil {
		return err
	}
	return nil
}

func (accessor *KonomiCRAccessor) InsertCollection(collection model.Collection) error {
	stmt := BgmUserCollection.INSERT(BgmUserCollection.AllColumns).
		MODEL(collection.ToBgmUserCollection()).
		ON_CONFLICT(BgmUserCollection.UserID, BgmUserCollection.SubjectID).DO_NOTHING()

	_, err := stmt.Exec(accessor.db)

	if err != nil {
		return err
	}
	return nil
}

func (accessor *KonomiCRAccessor) BatchInsertCollection(collections []model.Collection, batchSize int) error {
	startIdx := 0
	errs := make([]error, 0)
	for startIdx < len(collections) {
		endIdx := startIdx + batchSize
		if endIdx > len(collections) {
			endIdx = len(collections)
		}
		stmt := BgmUserCollection.INSERT(BgmUserCollection.AllColumns).
			MODELS(model.ToBgmUserCollections(collections[startIdx:endIdx])).
			ON_CONFLICT(BgmUserCollection.UserID, BgmUserCollection.SubjectID).DO_NOTHING()

		_, err := stmt.Exec(accessor.db)

		if err != nil {
			errs = append(errs, err)
		}
		startIdx = endIdx
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	return nil
}

func (accessor *KonomiCRAccessor) DeleteCollectionByUid(uid string) error {
	stmt := BgmUserCollection.DELETE().
		WHERE(BgmUserCollection.UserID.EQ(String(uid)))

	_, err := stmt.Exec(accessor.db)
	if err != nil {
		return err
	}
	return nil
}
