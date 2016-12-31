package services

import (
	"errors"
	"time"

	scari "github.com/rafax/scari"
	uuid "github.com/satori/go.uuid"
)

type leasedJob struct {
	leaseID     string
	leasedUntil time.Time
	scari.Job
}

type JobService interface {
	New(url string, output scari.OutputType) (*scari.Job, error)
	Get(id scari.JobID) (*scari.Job, error)
	GetAll() ([]scari.Job, error)
	Lease(id scari.JobID, leaseID scari.LeaseID) (time.Time, error)
}

type jobService struct {
	jobs       map[scari.JobID]scari.Job
	leasedJobs map[scari.LeaseID]leasedJob
}

func NewJobService() JobService {
	return &jobService{jobs: map[scari.JobID]scari.Job{}, leasedJobs: map[scari.LeaseID]leasedJob{}}
}

func (js *jobService) New(source string, output scari.OutputType) (*scari.Job, error) {
	id := scari.JobID(uuid.NewV4().String())
	j := scari.Job{ID: id, Output: output, Source: source, Status: scari.Pending}
	js.jobs[id] = j
	return &j, nil
}

func (js *jobService) Get(id scari.JobID) (*scari.Job, error) {
	j, ok := js.jobs[id]
	if !ok {
		return nil, errors.New("Job not found")
	}
	return &j, nil
}

func (js *jobService) GetAll() ([]scari.Job, error) {
	res := make([]scari.Job, len(js.jobs))
	i := 0
	for _, j := range js.jobs {
		res[i] = j
		i++
	}
	return res, nil
}

func (js *jobService) Lease(id scari.JobID, leaseID scari.LeaseID) (time.Time, error) {
	return time.Now().Add(60 * time.Second), nil
}
