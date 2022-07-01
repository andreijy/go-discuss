package main

import (
	"log"
	"net/http"

	"github.com/andreijy/go-discuss/postgres"
	"github.com/andreijy/go-discuss/web"
)

func main() {
	dsn := "postgres://postgres:secret@localhost/postgres?sslmode=disable"

	store, err := postgres.NewStore(dsn)
	if err != nil {
		log.Fatal(err)
	}

	sessions, err := web.NewSessionManager(dsn)
	if err != nil {
		log.Fatal(err)
	}

	h := web.NewHandler(store, sessions)
	http.ListenAndServe(":3000", h)
}
