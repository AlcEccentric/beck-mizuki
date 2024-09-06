package dao

import (
	"context"
	"fmt"
	"log"
	"os"

	model "github.com/alceccentric/beck-crawler/model"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	mongoAppName            = "beck-konomi"
	databaseName            = "beck-konomi"
	userTableName           = "user"
	userCollectionTableName = "user-collection"
)

type KonomiAccessor struct {
	client *mongo.Client
}

func NewKonomiAccessor() *KonomiAccessor {
	client, err := getMongoClient()
	if err != nil {
		log.Fatal(err)
	}
	return &KonomiAccessor{
		client: client,
	}
}

func getMongoClient() (*mongo.Client, error) {
	err := godotenv.Load("local_test_credentials.env")
	if err != nil {
		return nil, fmt.Errorf("error loading .env file with error %v", err)
	}
	userId := os.Getenv("BECK_MONGO_DB_USER")
	password := os.Getenv("BECK_MONGO_DB_PASSWORD")

	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	opts := options.Client().ApplyURI("mongodb+srv://" + userId + ":" + password +
		"@beck-konomi.lwpamdr.mongodb.net/?retryWrites=true&w=majority&appName=" + mongoAppName).SetServerAPIOptions(serverAPI)

	client, err := mongo.Connect(context.TODO(), opts)

	if err != nil {
		return nil, fmt.Errorf("error connecting to mongodb with error %v", err)
	}
	return client, nil
}

func (mongoAccessor *KonomiAccessor) Disconnect() {
	err := mongoAccessor.client.Disconnect(context.TODO())
	if err != nil {
		log.Fatal(err)
	}
}

func (mongoAccessor *KonomiAccessor) InsertUser(user model.User) {
	userTable := mongoAccessor.client.Database("beck-konomi").Collection(userTableName)
	filter := bson.M{"_id": user.ID}
	update := bson.M{"$set": user}

	_, err := userTable.UpdateOne(context.TODO(), filter, update, options.Update().SetUpsert(true))
	if err != nil {
		log.Fatalf("Failed to insert user with id: %s with error: %v", user.ID, err)
	}
}

func (mongoAccessor *KonomiAccessor) InsertCollection(collection model.Collection) {
	collectionTable := mongoAccessor.client.Database("beck-konomi").Collection(userCollectionTableName)
	filter := bson.D{
		{Key: "user_id", Value: collection.UserID},
		{Key: "subject_id", Value: collection.SubjectID},
	}
	update := bson.M{"$set": collection}

	_, err := collectionTable.UpdateOne(context.TODO(), filter, update, options.Update().SetUpsert(true))
	if err != nil {
		log.Fatalf("Failed to insert user collection with user id: %s, subject id: %s with error: %v", collection.UserID, collection.SubjectID, err)
	}
}
