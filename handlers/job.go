package handlers

import (
	"net/http"
	"net/url"
	"time"

	"github.com/gorilla/feeds"
	"github.com/gorilla/mux"
	"github.com/mholt/binding"
	"github.com/rafax/scari"
	"github.com/rafax/scari/services"
	"github.com/unrolled/render"
)

type Handlers interface {
	Register(*mux.Router)
}

type handlers struct {
	r  *render.Render
	js services.JobService
}

func New(js services.JobService) Handlers {
	return handlers{r: render.New(), js: js}
}

func (h handlers) Register(r *mux.Router) {
	r.HandleFunc("/jobs", h.createJob).Methods("POST")
	r.HandleFunc("/jobs", h.getAllJobs)
	r.HandleFunc("/feed", h.getFeed)
	r.HandleFunc("/jobs/lease", h.leaseJob).Methods("POST")
	r.HandleFunc("/jobs/{jobID}/complete", h.completeJob).Methods("POST")
}

func (h handlers) getFeed(w http.ResponseWriter, req *http.Request) {
	now := time.Now()
	feed := &feeds.Feed{
		Title:   "Scari downloads",
		Link:    &feeds.Link{Href: "http://scari.herokuapp.com/feed"},
		Author:  &feeds.Author{Name: "Rafal Gajdulewicz"},
		Created: now,
	}
	jobs, err := h.js.GetAll()
	if err != nil {
		h.r.XML(w, 500, map[string]string{"error": err.Error()})
	}
	items := []*feeds.Item{}
	for _, j := range jobs {
		if j.StorageID != "" {
			items = append(items, &feeds.Item{Link: &feeds.Link{Href: "http://scari-666.appspot.com/files/" + j.StorageID}})
		}
	}
	feed.Items = items
	atom, err := feed.ToAtom()
	if err != nil {
		h.r.XML(w, 500, map[string]string{"error": err.Error()})
	}
	w.WriteHeader(200)
	w.Header().Set("Content-Type", "application/atom+xml")
	w.Write([]byte(atom))
}

func (h handlers) createJob(w http.ResponseWriter, req *http.Request) {
	jr := new(JobRequest)
	errs := binding.Bind(req, jr)
	if errs.Handle(w) {
		return
	}
	j, err := h.js.New(jr.Source, scari.OutputType(jr.OutputType))
	if err != nil {
		h.r.JSON(w, 500, map[string]string{"error": err.Error()})
	}
	h.r.JSON(w, 200, scari.JobResponse{Job: *j})
}

func (h handlers) getAllJobs(w http.ResponseWriter, req *http.Request) {
	jobs, err := h.js.GetAll()
	if err != nil {
		h.r.JSON(w, 500, map[string]string{"error": err.Error()})
	}
	h.r.JSON(w, 200, scari.JobsResponse{Jobs: jobs})
}

func (h handlers) leaseJob(w http.ResponseWriter, req *http.Request) {
	job, lid, err := h.js.LeaseOne()
	if err != nil {
		h.r.JSON(w, 500, map[string]string{"error": err.Error()})
		return
	}
	if job == nil {
		h.r.JSON(w, 204, map[string]string{})
		return
	}
	h.r.JSON(w, 200, scari.LeaseJobResponse{Job: *job, LeaseID: lid})
}

func (h handlers) completeJob(w http.ResponseWriter, req *http.Request) {
	cjr := new(CompleteJobRequest)
	errs := binding.Bind(req, cjr)
	if errs.Handle(w) {
		return
	}
	j, err := h.js.Complete(cjr.LeaseID, cjr.FileName)
	if err != nil {
		h.r.JSON(w, 500, map[string]string{"error": err.Error()})
		return
	}
	h.r.JSON(w, 200, scari.JobResponse{Job: *j})
}

type JobRequest struct {
	Source     string `json:"source"`
	OutputType string `json:"outputType"`
}

type CompleteJobRequest struct {
	FileName string        `json:"storageUrl"`
	LeaseID  scari.LeaseID `json:"leaseId"`
}

func (jr *JobRequest) FieldMap(req *http.Request) binding.FieldMap {
	return binding.FieldMap{
		&jr.Source:     "source",
		&jr.OutputType: "outputType",
	}
}

func (jr *JobRequest) Validate(req *http.Request, errs binding.Errors) binding.Errors {
	url, err := url.Parse(jr.Source)
	if err != nil {
		errs = append(errs, binding.Error{
			FieldNames: []string{"source"},
			Message:    "source must be a valid url.URL :" + err.Error()})
	}
	if url.Scheme != "http" && url.Scheme != "https" {
		errs = append(errs, binding.Error{
			FieldNames: []string{"source"},
			Message:    "source must use a http(s) scheme"})
	}
	return errs
}

func (cjr *CompleteJobRequest) FieldMap(req *http.Request) binding.FieldMap {
	return binding.FieldMap{
		&cjr.LeaseID:  "leaseId",
		&cjr.FileName: "fileName",
	}
}
