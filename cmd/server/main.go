package main

import (
	"slicerapi/internal/db"
	"slicerapi/internal/http"
	"slicerapi/internal/logger"
)

func main() {
	if err := db.Connect(); err != nil {
		logger.L.Fatalln(err)
	}

	http.Start()
}
