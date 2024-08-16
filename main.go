package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// import (
// 	"github.com/alceccentric/beck-crawler/dao/bgm"
// 	"github.com/alceccentric/beck-crawler/orch"
// )

// const (
// 	numOfSubjectProducers    = 3
// 	numOfUserProducers       = 3
// 	numOfCollectionProducers = 3
// )

type User struct {
	ID             string    `bson:"_id"`
	Username       string    `bson:"username,omitempty"`
	Nickname       string    `bson:"nickname,omitempty"`
	AvatarURL      string    `bson:"avatar_url,omitempty"`
	LastActiveTime time.Time `bson:"last_active_time"`
}

func main() {
	// bgmClient := bgm.NewBgmApiAccessor()
	// subjectOrch := orch.NewSubjectOrchestrator(&bgmClient, numOfCollectionProducers, numOfUserProducers, numOfSubjectProducers)
	// subjectOrch.Run()

	err := godotenv.Load("local_test_credentials.env")
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}
	userId := os.Getenv("BECK_MONGO_DB_USER")
	password := os.Getenv("BECK_MONGO_DB_PASSWORD")

	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	opts := options.Client().ApplyURI("mongodb+srv://" + userId + ":" + password + "@beck-konomi.lwpamdr.mongodb.net/?retryWrites=true&w=majority&appName=beck-konomi").SetServerAPIOptions(serverAPI)

	client, err := mongo.Connect(context.TODO(), opts)
	if err != nil {
		panic(err)
	}
	defer func() {
		if err = client.Disconnect(context.TODO()); err != nil {
			panic(err)
		}
	}()
	userTable := client.Database("beck-konomi").Collection("user")
	user := User{
		ID:             "3",
		Username:       "john_doe3",
		Nickname:       "John3",
		AvatarURL:      "http://example.com/avatar.jpg",
		LastActiveTime: time.Now(),
	}

	// Insert the document
	insertResult, err := userTable.InsertOne(context.TODO(), user)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Inserted document with ID:", insertResult.InsertedID)

}
