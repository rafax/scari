package main

import (
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/urfave/negroni"

	"github.com/rafax/scari/handlers"
	"github.com/rafax/scari/services"
)

func main() {
	js := services.NewJobService()
	n := negroni.New()
	n.Use(negroni.NewLogger())
	router := mux.NewRouter()

	n.Use(negroni.NewRecovery())
	handlers.New(js).Register(router)
	n.UseHandler(router)

	port := os.Getenv("PORT")
	if port == "" {
		port = "3001"
	}
	http.ListenAndServe(":"+port, n)
}
