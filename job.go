package scari

type JobID string

type Job struct {
	ID        JobID      `json:"id"`
	Output    OutputType `json:"output"`
	Source    string     `json:"source"`
	Status    JobStatus  `json:"status"`
	StorageID string     `json:"storageId,omitempty"`
}

type OutputType string

const (
	AUDIO = "audio"
	VIDEO = "video"
)

type JobStatus string

const (
	Pending    = "Pending"
	Processing = "Processing"
	Completed  = "Completed"
	Failed     = "Failed"
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

type StaticFileRequest struct {
	FileName string `json:"fileName"`
}

type StaticFileResponse struct {
	Id string `json:"id"`
}
