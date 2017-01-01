package services

import (
	"errors"
	"sync"
	"time"

	scari "github.com/rafax/scari"
	uuid "github.com/satori/go.uuid"
)

type leasedJob struct {
	leaseID     string
	leasedUntil time.Time
	scari.Job
}

type JobStore interface {
	Put(j scari.Job) error
	Get(id scari.JobID) (*scari.Job, error)
	GetAll() ([]scari.Job, error)
}

type JobService interface {
	New(url string, output scari.OutputType) (*scari.Job, error)
	Get(id scari.JobID) (*scari.Job, error)
	GetAll() ([]scari.Job, error)
	Lease(id scari.JobID, leaseID scari.LeaseID) (time.Time, error)
}

type mapJobStore struct {
	jobsLock   sync.RWMutex
	jobs       map[scari.JobID]scari.Job
	leasedJobs map[scari.LeaseID]leasedJob
}

func (m *mapJobStore) Get(id scari.JobID) (*scari.Job, error) {
	m.jobsLock.RLock()
	defer m.jobsLock.RUnlock()
	j, ok := m.jobs[id]
	if !ok {
		return nil, errors.New("Job not found")
	}
	return &j, nil
}

func (m *mapJobStore) GetAll() ([]scari.Job, error) {
	m.jobsLock.RLock()
	defer m.jobsLock.RUnlock()
	res := make([]scari.Job, len(m.jobs))
	i := 0
	for _, j := range m.jobs {
		res[i] = j
		i++
	}
	return res, nil
}

func (m *mapJobStore) Put(j scari.Job) error {
	m.jobsLock.Lock()
	defer m.jobsLock.Unlock()
	m.jobs[j.ID] = j
	return nil
}

type jobService struct {
	store JobStore
}

func NewJobService() JobService {
	return &jobService{store: &mapJobStore{jobs: map[scari.JobID]scari.Job{}, leasedJobs: map[scari.LeaseID]leasedJob{}}}
}

func (js *jobService) New(source string, output scari.OutputType) (*scari.Job, error) {
	id := scari.JobID(uuid.NewV4().String())
	j := scari.Job{ID: id, Output: output, Source: source, Status: scari.Pending}
	js.store.Put(j)
	return &j, nil
}

func (js *jobService) Get(id scari.JobID) (*scari.Job, error) {
	return js.store.Get(id)
}

func (js *jobService) GetAll() ([]scari.Job, error) {
	return js.store.GetAll()
}

func (js *jobService) Lease(id scari.JobID, leaseID scari.LeaseID) (time.Time, error) {
	return time.Now().Add(60 * time.Second), nil
}
