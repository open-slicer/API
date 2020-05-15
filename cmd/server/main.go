package main

import (
	"slicerapi/internal/db"
	"slicerapi/internal/http"
)

func main() {
	db.Connect()
	http.Start()
}
