package postgres

import (
	"database/sql"

	_ "github.com/lib/pq"

	"github.com/rafax/scari"
)

type postgresJobStore struct {
	db *sql.DB
}

func (s *postgresJobStore) Put(j scari.Job) error {
	_, err := s.db.Exec("INSERT INTO Jobs (id, output, source, status) VALUES($1,$2,$3,$4)", j.ID, j.Output, j.Source, j.Status)
	return err
}
func (s *postgresJobStore) Get(id scari.JobID) (*scari.Job, error) {
	return nil, nil
}
func (s *postgresJobStore) GetAll() ([]scari.Job, error) {
	jobs := []scari.Job{}
	r, err := s.db.Query("SELECT id::text,output,source,status,storageID FROM Jobs")
	if err != nil {
		return nil, err
	}
	defer r.Close()
	for r.Next() {
		c, err := r.Columns()
		if err != nil {
			return nil, err
		}
		jobs = append(jobs, scari.Job{ID: scari.JobID(c[0]), Output: scari.OutputType(c[1]), Source: c[2], Status: scari.JobStatus(c[3]), StorageID: c[4]})
	}
	return jobs, nil
}
func (s *postgresJobStore) LeaseOne(scari.LeaseID) (*scari.Job, error) {
	return nil, nil
}
func (s *postgresJobStore) Complete(lid scari.LeaseID, storageID string) (*scari.Job, error) {
	return nil, nil
}

func New(db *sql.DB) scari.JobStore {
	db.QueryRow(schema)
	return &postgresJobStore{db: db}
}

const schema = `
-- initial migration
CREATE TABLE IF NOT EXISTS Jobs (
    id uuid PRIMARY KEY,
    output text NOT NULL,
    source text NOT NULL,
    status text NOT NULL,
    storageId text ,
    leaseId text UNIQUE
)
-- DROP TABLE Jobs
`
