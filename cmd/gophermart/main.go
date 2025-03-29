package main

import (
	"log"
	"net/http"

	"github.com/iubondar/gophermart/internal/router"
	"github.com/iubondar/gophermart/internal/storage"
	"go.uber.org/zap"
)

func init() {
	zap.ReplaceGlobals(zap.Must(zap.NewDevelopment()))
}

func main() {
	storage, err := storage.NewStorage("")
	if err != nil {
		log.Fatal(err)
	}

	router, err := router.NewRouter(storage)
	if err != nil {
		log.Fatal(err)
	}

	log.Fatal(
		http.ListenAndServe("localhost:8080", router),
	)
}
