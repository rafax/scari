package scari

type JobID string

type Job struct {
	ID     JobID
	Output OutputType
	Source string
	Status JobStatus
}

type OutputType string

const (
	AUDIO = "audio"
	VIDEO = "video"
)

type JobStatus int

const (
	Pending    = iota
	Processing = 2
	Completed  = 3
	Failed     = 4
)

type LeaseID string

type JobsResponse struct {
	Jobs []Job `json:"jobs"`
}

type JobResponse struct {
	Job Job `json:"job"`
}
