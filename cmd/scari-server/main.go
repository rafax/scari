package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/urfave/negroni"
	newrelic "github.com/yadvendar/negroni-newrelic-go-agent"

	"github.com/rafax/scari/handlers"
	"github.com/rafax/scari/mock"
	"github.com/rafax/scari/postgres"
	"github.com/rafax/scari/services"
)

func main() {
	pgConn := os.Getenv("DATABASE_URL")
	if pgConn == "" {
		pgConn = "postgresql://scari@localhost:5432/scari?sslmode=disable"
	}
	js := services.NewJobService(postgres.New(pgConn), mock.NewStorageClient())
	n := negroni.New()
	newRelicKey := os.Getenv("SCARI_NEW_RELIC_LICENCE_KEY")
	if newRelicKey != "" {
		config := newrelic.NewConfig("scari", newRelicKey)
		newRelicMiddleware, _ := newrelic.New(config)
		n.Use(newRelicMiddleware)
	}
	n.Use(negroni.NewLogger())
	router := mux.NewRouter()

	n.Use(negroni.NewRecovery())
	handlers.New(js).Register(router)
	n.UseHandler(router)

	port := os.Getenv("PORT")
	if port == "" {
		port = "3001"
	}
	log.Fatal(http.ListenAndServe(":"+port, n))
}
