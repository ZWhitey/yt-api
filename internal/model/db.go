package model

import (
	"context"
	"log"
	"os"
	"time"

	_ "github.com/joho/godotenv/autoload"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/x/mongo/driver/connstring"
)

var mgoClient *mongo.Client
var Db *mongo.Database

func init() {

	connectionString := os.Getenv("MONGO_CONNECTION_STRING")
	serverAPIOptions := options.ServerAPI(options.ServerAPIVersion1)
	clientOptions := options.Client().
		ApplyURI(connectionString).
		SetServerAPIOptions(serverAPIOptions)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	mgoClient, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	err = mgoClient.Ping(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}

	cs, err := connstring.ParseAndValidate(connectionString)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Connected to MongoDB:", cs.Hosts, cs.Database)

	Db = mgoClient.Database(cs.Database)
}
