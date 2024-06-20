package database

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

type MongoInstance struct {
	Client *mongo.Client
	DB     *mongo.Database
}

const dbName = "fiber-mongodb"

var Mg MongoInstance

func ConnectDB(mongoURI string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clientOptions := options.Client().ApplyURI(mongoURI)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return err
	}

	if err := client.Ping(ctx, nil); err != nil {
		return err
	}

	Mg = MongoInstance{
		Client: client,
		DB:     client.Database(dbName),
	}

	fmt.Println("Successfully Connected to Database")
	return nil
}

func LoadDB() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	username := os.Getenv("USERNAME")
	password := os.Getenv("PASSWORD")
	mongoURI := fmt.Sprintf("mongodb+srv://%s:%s@cluster0.4nbswvd.mongodb.net/?retryWrites=true&w=majority&appName=Cluster0", username, password)
	if err := ConnectDB(mongoURI); err != nil {
		log.Fatalf("Could not connect to the database: %v", err)
	}
}
