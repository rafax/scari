package mock

import (
	"errors"
	"sync"

	"github.com/rafax/scari"
)

func NewStore() scari.JobStore {
	return &mapJobStore{
		jobs:       map[scari.JobID]scari.Job{},
		leasedJobs: map[scari.LeaseID]scari.JobID{},
	}
}

type mapJobStore struct {
	jobsLock   sync.RWMutex
	jobs       map[scari.JobID]scari.Job
	leasedJobs map[scari.LeaseID]scari.JobID
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

func (m *mapJobStore) Complete(leaseID scari.LeaseID, storageID string) (*scari.Job, error) {
	m.jobsLock.Lock()
	defer m.jobsLock.Unlock()
	jid, ok := m.leasedJobs[leaseID]
	if !ok {
		return nil, errors.New("Lease not found")
	}
	j, ok := m.jobs[jid]
	if !ok {
		return nil, errors.New("Job not found")
	}
	delete(m.leasedJobs, leaseID)
	j.Status = scari.Completed
	j.StorageID = storageID
	m.jobs[j.ID] = j
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

func (m *mapJobStore) Status() error {
	return nil
}

func (m *mapJobStore) LeaseOne(lid scari.LeaseID) (*scari.Job, error) {
	m.jobsLock.Lock()
	defer m.jobsLock.Unlock()
	for _, j := range m.jobs {
		if j.Status == scari.Pending {
			j.Status = scari.Processing
			m.jobs[j.ID] = j
			m.leasedJobs[lid] = j.ID
			// TODO expire locks
			return &j, nil
		}
	}
	return nil, nil
}
