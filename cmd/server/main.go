package main

import (
	"context"
	"slicerapi/internal/db"
	"slicerapi/internal/http"
	"slicerapi/internal/logger"
	"slicerapi/internal/util"
	"time"
)

func main() {
	if err := db.Connect(); err != nil {
		logger.L.Fatalln(err)
	}
	ctx, _ := context.WithTimeout(context.Background(), 3 * time.Second)
	defer util.Chk(db.Mongo.Disconnect(ctx), false)

	http.Start()
}
