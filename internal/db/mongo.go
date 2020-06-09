package db

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo/options"
	"slicerapi/internal/config"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
)

// Mongo is the main MongpDB client.
var Mongo *mongo.Client

// Connect creates a MongoDB session and assigns Mongo to it.
func Connect() (err error) {
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	Mongo, err = mongo.Connect(ctx, options.Client().ApplyURI(config.Config.DB.MongoDB.URI))

	return
}
