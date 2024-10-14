package dao

import (
	"context"
	"fmt"
	"os"

	model "github.com/AlcEccentric/beck-mizuki/model"
	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	mongoApp                 = "beck-konomi"
	mongoDatabase            = "beck-konomi"
	mongoUserTable           = "user"
	mongoUserCollectionTable = "user-collection"
)

type KonomiMongoAccessor struct {
	client *mongo.Client
}

func NewMongoKonomiAccessor() *KonomiMongoAccessor {
	client, err := getMongoClient()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to mongodb")
	}
	return &KonomiMongoAccessor{
		client: client,
	}
}

func getMongoClient() (*mongo.Client, error) {
	userId := os.Getenv("BECK_KONOMI_MONGO_DB_USER")
	password := os.Getenv("BECK_KONOMI_MONGO_DB_PASSWORD")

	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	opts := options.Client().ApplyURI("mongodb+srv://" + userId + ":" + password +
		"@beck-konomi.lwpamdr.mongodb.net/?retryWrites=true&w=majority&appName=" + mongoApp).SetServerAPIOptions(serverAPI)

	client, err := mongo.Connect(context.TODO(), opts)

	if err != nil {
		return nil, fmt.Errorf("error connecting to mongodb with error %v", err)
	}
	return client, nil
}

func (accessor *KonomiMongoAccessor) Disconnect() {
	err := accessor.client.Disconnect(context.TODO())
	if err != nil {
		log.Error().Msgf("Failed to disconnect from mongodb with error: %v", err)
	}
}

func (accessor *KonomiMongoAccessor) InsertUser(user model.User) error {
	userTable := accessor.client.Database("beck-konomi").Collection(mongoUserTable)
	filter := bson.M{"_id": user.ID}
	update := bson.M{"$set": user}

	_, err := userTable.UpdateOne(context.TODO(), filter, update, options.Update().SetUpsert(true))
	if err != nil {
		return err
	}

	return nil
}

func (accessor *KonomiMongoAccessor) InsertCollection(collection model.Collection) error {
	collectionTable := accessor.client.Database("beck-konomi").Collection(mongoUserCollectionTable)
	filter := bson.D{
		{Key: "user_id", Value: collection.UserID},
		{Key: "subject_id", Value: collection.SubjectID},
	}
	update := bson.M{"$set": collection}

	_, err := collectionTable.UpdateOne(context.TODO(), filter, update, options.Update().SetUpsert(true))
	if err != nil {
		return err
	}

	return nil
}
