package main

import (
	"log"
	"net/http"

	"github.com/iubondar/gophermart/internal/router"
	"go.uber.org/zap"
)

func init() {
	zap.ReplaceGlobals(zap.Must(zap.NewDevelopment()))
}

func main() {
	router, err := router.NewRouter()
	if err != nil {
		log.Fatal(err)
	}

	log.Fatal(
		http.ListenAndServe("localhost:8080", router),
	)
}
