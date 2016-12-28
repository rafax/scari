package main

import (
	"encoding/json"
	"net/http"

	"github.com/codegangsta/negroni"
	"github.com/gorilla/mux"

	services "github.com/rafax/scari/services"
)

func main() {
	js := services.New()
	n := negroni.Classic()
	r := mux.NewRouter()
	r.HandleFunc("/jobs", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		jobs, err := js.GetAll()
		json.NewEncoder(w).Encode(jobs)
	})
	r.HandleFunc("/articles", ArticlesHandler)
	http.Handle("/", r)
}

type JobsResponse struct {
	jobs []Job
}
