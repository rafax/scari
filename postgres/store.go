package postgres

import (
	"database/sql"
	"time"

	"github.com/jackc/pgx"
	"github.com/jackc/pgx/stdlib"
	"github.com/pkg/errors"
	"github.com/rafax/scari"
)

type postgresStore struct {
	pool *pgx.ConnPool
}

func New(connURI string) scari.JobStore {
	conn, err := pgx.ParseURI(connURI)
	if err != nil {
		panic(err)
	}
	poolConfig := pgx.ConnPoolConfig{ConnConfig: conn, AcquireTimeout: 1 * time.Second, MaxConnections: 10}
	pool, err := pgx.NewConnPool(poolConfig)
	if err != nil {
		panic(err)
	}
	_, err = pool.Exec(schema)
	if err != nil {
		panic(err)
	}
	return &postgresStore{pool: pool}
}

func (ps *postgresStore) Put(j scari.Job) error {
	_, err := ps.pool.Exec("INSERT INTO Jobs (id, output, source, status) VALUES($1,$2,$3,$4)", j.ID, j.Output, j.Source, j.Status)
	return err
}
func (ps *postgresStore) Get(id scari.JobID) (*scari.Job, error) {
	row := ps.pool.QueryRow("SELECT id, output, source, status, storage_id, lease_id FROM Jobs WHERE id = $1", id)
	jm := JobModel{}
	err := row.Scan(&jm.ID, &jm.Output, &jm.Source, &jm.Status, &jm.StorageID, &jm.LeaseID)
	if err != nil {
		return nil, err
	}
	j := jm.toJob()
	return &j, nil
}
func (ps *postgresStore) GetAll() ([]scari.Job, error) {
	rows, err := ps.pool.Query(selectAll)
	if err != nil {
		return nil, errors.WithMessage(err, "When getting all jobs")
	}
	res := []scari.Job{}
	for rows.Next() {
		jm := JobModel{}
		err := rows.Scan(&jm.ID, &jm.Output, &jm.Source, &jm.Status, &jm.StorageID, &jm.LeaseID)
		if err != nil {
			return nil, errors.WithMessage(err, "When scanning")
		}
		res = append(res, jm.toJob())
	}
	return res, nil
}
func (ps *postgresStore) LeaseOne(lid scari.LeaseID) (*scari.Job, error) {
	row := ps.pool.QueryRow(leaseOne, scari.Pending, lid)
	jm := JobModel{}
	err := row.Scan(&jm.ID, &jm.Output, &jm.Source, &jm.Status, &jm.StorageID, &jm.LeaseID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	j := jm.toJob()
	return &j, nil
}
func (ps *postgresStore) Complete(lid scari.LeaseID, storageID string) (*scari.Job, error) {
	row := ps.pool.QueryRow(complete, scari.Completed, storageID, lid)
	jm := JobModel{}
	err := row.Scan(&jm.ID, &jm.Output, &jm.Source, &jm.Status, &jm.StorageID, &jm.LeaseID)
	if err != nil {
		return nil, err
	}
	j := jm.toJob()
	return &j, nil
}

type JobModel struct {
	ID        string         `sql:"id"`
	Output    string         `sql:"output"`
	Source    string         `sql:"source"`
	Status    string         `sql:"status"`
	StorageID sql.NullString `sql:"storage_id"`
	LeaseID   sql.NullString `sql:"lease_id"`
}

func (j *JobModel) toJob() scari.Job {
	return scari.Job{
		ID:        scari.JobID(j.ID),
		Output:    scari.OutputType(j.Output),
		Source:    j.Source,
		Status:    scari.JobStatus(j.Status),
		StorageID: j.StorageID.String,
	}
}

func (ps *postgresStore) Status() error {
	conn, err := stdlib.OpenFromConnPool(ps.pool)
	if err != nil {
		return err
	}
	return conn.Ping()
}

const schema = `
    CREATE  TABLE IF NOT EXISTS Jobs (
        id uuid PRIMARY KEY, 
        output text not null,
        source text not null,
        status text not null,
        storage_id text NULL,
        lease_id text NULL
    )`

const selectAll = `SELECT id, output, source, status, storage_id, lease_id FROM Jobs`
const leaseOne = `UPDATE jobs SET status = $1, lease_id = $2 
WHERE id =(SELECT id FROM jobs WHERE status = 'Pending' LIMIT 1 FOR UPDATE SKIP LOCKED) 
RETURNING id, output, source, status, storage_id, lease_id`
const complete = `UPDATE jobs SET status = $1, storage_id = $2 
WHERE id =(SELECT id FROM jobs WHERE lease_id = $3 LIMIT 1 FOR UPDATE SKIP LOCKED) 
RETURNING id, output, source, status, storage_id, lease_id`
