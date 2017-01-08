package scari

type JobID string

type Job struct {
	ID        JobID
	Output    OutputType
	Source    string
	Status    JobStatus
	StorageID string
}

type OutputType string

const (
	AUDIO = "audio"
	VIDEO = "video"
)

type JobStatus int

const (
	Pending    = iota
	Processing = 1
	Completed  = 2
	Failed     = 3
)

type LeaseID string

type JobsResponse struct {
	Jobs []Job `json:"jobs"`
}

type JobResponse struct {
	Job Job `json:"job"`
}

type LeaseJobResponse struct {
	Job     Job     `json:"job"`
	LeaseID LeaseID `json:"leaseId"`
}

const (
	StorageBucketName = "scari-666.appspot.com"
)
