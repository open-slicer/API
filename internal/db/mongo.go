package db

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"slicerapi/internal/config"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
)

// Mongo is the main MongpDB client.
var Mongo *mongo.Client

// Connect creates a MongoDB session and assigns Mongo to it.
func Connect() error {
	var err error

	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	Mongo, err = mongo.Connect(ctx, options.Client().ApplyURI(config.C.MongoDB.URI))
	if err != nil {
		return err
	}

	ctx, _ = context.WithTimeout(context.Background(), 2*time.Second)
	return Mongo.Ping(ctx, readpref.Primary())
}
