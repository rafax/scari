package services

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"path"
	"sync"

	scari "github.com/rafax/scari"
	uuid "github.com/satori/go.uuid"
)

type JobStore interface {
	Put(j scari.Job) error
	Get(id scari.JobID) (*scari.Job, error)
	GetAll() ([]scari.Job, error)
	LeaseOne(scari.LeaseID) (*scari.Job, error)
	Complete(lid scari.LeaseID, storageID string) (*scari.Job, error)
}

type JobService interface {
	New(url string, output scari.OutputType) (*scari.Job, error)
	Get(id scari.JobID) (*scari.Job, error)
	GetAll() ([]scari.Job, error)
	LeaseOne() (*scari.Job, scari.LeaseID, error)
	Complete(lid scari.LeaseID, storageID string) (*scari.Job, error)
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

type jobService struct {
	store         JobStore
	storageClient StorageClient
}

func NewJobService() JobService {
	key := scari.JobID(uuid.NewV4().String())
	return &jobService{
		store:         &mapJobStore{jobs: map[scari.JobID]scari.Job{key: scari.Job{ID: key, StorageID: "5644406560391168"}}, leasedJobs: map[scari.LeaseID]scari.JobID{}},
		storageClient: &mockStorageClient{client: &http.Client{}}}
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

func (js *jobService) LeaseOne() (*scari.Job, scari.LeaseID, error) {
	lid := scari.LeaseID(uuid.NewV4().String())
	j, err := js.store.LeaseOne(lid)
	if err != nil {
		return nil, "", err
	}
	return j, lid, nil
}

func (js *jobService) Complete(lid scari.LeaseID, fileName string) (*scari.Job, error) {
	sid, err := js.storageClient.Register(fileName)
	if err != nil {
		return nil, err
	}
	j, err := js.store.Complete(lid, sid)
	if err != nil {
		return nil, err
	}
	return j, err
}

type StorageClient interface {
	Register(string) (string, error)
}

type mockStorageClient struct {
	client *http.Client
}

func (msc *mockStorageClient) Register(fileName string) (string, error) {
	sfr := scari.StaticFileRequest{FileName: path.Base(fileName)}
	body, err := json.Marshal(sfr)
	if err != nil {
		return "", err
	}
	resp, err := msc.client.Post("http://scari-666.appspot.com/files", "application/json", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	rbody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	sfresp := new(scari.StaticFileResponse)
	if err = json.Unmarshal(rbody, sfresp); err != nil {
		return "", err
	}
	return sfresp.Id, nil
}
