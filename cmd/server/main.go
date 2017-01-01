package main

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/urfave/negroni"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/rafax/scari/handlers"
	"github.com/rafax/scari/services"
	negroniprometheus "github.com/zbindenren/negroni-prometheus"
)

func main() {
	js := services.NewJobService()
	n := negroni.New()
	n.Use(negroni.NewLogger())
	router := mux.NewRouter()
	p := negroniprometheus.NewMiddleware("scari")

	n.Use(negroni.NewRecovery())
	n.Use(p)
	router.Handle("/metrics", prometheus.Handler())

	handlers.New(js).Register(router)
	n.UseHandler(router)
	http.ListenAndServe(":3001", n)
}
