package services

import (
	scari "github.com/rafax/scari"
	uuid "github.com/satori/go.uuid"
)

type JobService interface {
	New(url string, output scari.OutputType) (*scari.Job, error)
	Get(id scari.JobID) (*scari.Job, error)
	GetAll() ([]scari.Job, error)
	LeaseOne() (*scari.Job, scari.LeaseID, error)
	Complete(lid scari.LeaseID, storageID string) (*scari.Job, error)
}

type jobService struct {
	store         scari.JobStore
	storageClient scari.StorageClient
}

func NewJobService(store scari.JobStore, storageClient scari.StorageClient) JobService {
	return &jobService{
		store:         store,
		storageClient: storageClient}
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
