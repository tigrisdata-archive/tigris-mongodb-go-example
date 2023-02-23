package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Podcast struct {
	ID     primitive.ObjectID `bson:"_id,omitempty"`
	Title  string             `bson:"title,omitempty"`
	Author string             `bson:"author,omitempty"`
	Tags   []string           `bson:"tags,omitempty"`
}

type Episode struct {
	ID          primitive.ObjectID `bson:"_id,omitempty"`
	Podcast     primitive.ObjectID `bson:"podcast,omitempty"`
	Title       string             `bson:"title,omitempty"`
	Description string             `bson:"description,omitempty"`
	Duration    int32              `bson:"duration,omitempty"`
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mongoUri := fetchMongoUri(ctx)

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoUri))
	if err != nil {
		panic(err)
	}
	defer client.Disconnect(ctx)

	fmt.Println("Connected.")

	database := client.Database(os.Getenv("TIGRIS_PROJECT"))
	podcastsCollection := database.Collection("podcasts")
	episodesCollection := database.Collection("episodes")

	// InsertOne
	podcastResult, err := podcastsCollection.InsertOne(ctx, Podcast{
		Title:  "The Polyglot Developer",
		Author: "Nic Raboy",
		Tags:   []string{"development", "programming", "coding"},
	})
	if err != nil {
		panic(err)
	}

	podcastID := podcastResult.InsertedID.(primitive.ObjectID)
	fmt.Printf("Inserted document into podcast collection: %v\n", podcastID.String())

	// InsertMany
	episodeResult, err := episodesCollection.InsertMany(ctx, []interface{}{
		Episode{
			Podcast:     podcastID,
			Title:       "GraphQL for API Development",
			Description: "Learn about GraphQL from the co-creator of GraphQL, Lee Byron.",
			Duration:    25,
		},
		Episode{
			Podcast:     podcastID,
			Title:       "Progressive Web Application Development",
			Description: "Learn about PWA development with Tara Manicsic.",
			Duration:    32,
		},
	})
	if err != nil {
		panic(err)
	}
	fmt.Printf("Inserted %v documents into episode collection!\n", len(episodeResult.InsertedIDs))

	// Find
	var episodes []Episode
	cursor, err := episodesCollection.Find(ctx, bson.M{"duration": bson.D{{"$gt", 25}}})
	if err != nil {
		panic(err)
	}
	if err = cursor.All(ctx, &episodes); err != nil {
		panic(err)
	}

	jsonData, err := json.MarshalIndent(episodes, "", "    ")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Found %v documents matching filter!\n", len(episodes))
	fmt.Printf("%s\n", jsonData)

	// Update
	updateResult, err := podcastsCollection.UpdateOne(
		ctx,
		bson.M{"_id": podcastID},
		bson.D{
			{"$set", bson.D{{"title", "The Polyglot Developer Podcast"}}},
		})
	if err != nil {
		panic(err)
	}
	fmt.Printf("Updated %v Documents!\n", updateResult.ModifiedCount)

	// Delete
	deleteResult, err := episodesCollection.DeleteOne(ctx, bson.D{{"title", "GraphQL for API Development"}})
	if err != nil {
		panic(err)
	}
	fmt.Printf("DeleteOne removed %v document(s)\n", deleteResult.DeletedCount)
}

func fetchMongoUri(ctx context.Context) string {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	uri := os.Getenv("TIGRIS_URI")
	clientId := os.Getenv("TIGRIS_CLIENT_ID")
	clientSecret := os.Getenv("TIGRIS_CLIENT_SECRET")
	if uri == "" || clientId == "" || clientSecret == "" {
		log.Fatal("You must set the 'TIGRIS_URI', 'TIGRIS_CLIENT_ID', and 'TIGRIS_CLIENT_SECRET' environment variables.")
	}

	return fmt.Sprintf("mongodb://%s:%s@%s/?authMechanism=PLAIN", clientId, clientSecret, uri)
}
