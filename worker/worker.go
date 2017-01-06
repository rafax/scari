package worker

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path"

	"cloud.google.com/go/storage"

	"github.com/rafax/scari"
	"golang.org/x/net/context"
)

type Worker interface {
	Process() ([]ProcessedJob, error)
}

const (
	bucketName = "scari-666.appspot.com"
)

var (
	command     = "youtube-dl"
	audioSuffix = []string{"-x", "--audio-format", "mp3"}
	videoSuffix = []string{"--recode-video", "mp4"}
)

type worker struct {
	apiserver   string
	outDir      string
	audioParams []string
	videoParams []string
	c           *http.Client
}

type youtubeDLOutput struct {
	FileName string `json:"_filename"`
}

type ProcessedJob struct {
	Job         scari.Job
	LeaseID     scari.LeaseID
	StorageID   string
	StoragePath string
	Existed     bool
}

func New(apiserver string, outDir string) Worker {
	commandParams := []string{"--print-json", "--restrict-filenames", "-o", outDir + "%(title)s.%(ext)s"}
	return worker{
		apiserver: apiserver, outDir: outDir, audioParams: append(commandParams, audioSuffix...),
		videoParams: append(commandParams, videoSuffix...), c: http.DefaultClient}
}

func (w worker) Process() ([]ProcessedJob, error) {
	j, lid, err := w.fetch()
	if err != nil {
		return nil, err
	}
	if j == nil {
		return nil, nil
	}
	out, err := w.convert(*j)
	if err != nil {
		return nil, err
	}
	url, existed, err := w.upload(out)
	if err != nil {
		return nil, err
	}
	pj := ProcessedJob{Job: *j, StoragePath: url, Existed: existed}
	w.complete(lid, pj)
	return []ProcessedJob{}, err
}

func (w worker) fetch() (*scari.Job, scari.LeaseID, error) {
	r, err := w.c.Post(w.apiserver+"jobs/lease", "application/json", nil)
	if err != nil {
		return nil, "", err
	}
	if r.StatusCode == 204 {
		return nil, "", nil
	}
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, "", err
	}
	var jr scari.LeaseJobResponse
	err = json.Unmarshal(body, &jr)
	if err != nil {
		return nil, "", err
	}
	return &jr.Job, jr.LeaseID, nil
}

func (w worker) convert(j scari.Job) (string, error) {
	var params []string
	if j.Output == scari.AUDIO {
		params = w.audioParams
	} else {
		params = w.videoParams
	}
	c := exec.Command(command, append(params, j.Source)...)
	c.Dir = w.outDir
	output, err := c.Output()
	if err != nil {
		return "", err
	}
	var out youtubeDLOutput
	err = json.Unmarshal(output, &out)
	if err != nil {
		return "", err
	}
	return out.FileName, nil
}

func (w worker) upload(filePath string) (string, bool, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return "", false, err
	}
	const publicURL = "https://storage.googleapis.com/%s/%s"
	name := path.Base(filePath)
	storageLocation := fmt.Sprintf(publicURL, bucketName, name)
	if err != nil {
		return "", false, err
	}
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		return "", false, err
	}
	bkt := client.Bucket(bucketName)
	_, err = bkt.Object(name).Attrs(ctx)
	if err == nil {
		// object exists
		return storageLocation, true, nil
	}
	if err != nil && err != storage.ErrObjectNotExist {
		return "", false, err
	}
	writer := bkt.Object(name).If(storage.Conditions{DoesNotExist: true}).NewWriter(ctx)
	writer.ACL = []storage.ACLRule{{Entity: storage.AllUsers, Role: storage.RoleReader}}
	writer.CacheControl = "public, max-age=86400"
	if _, err := io.Copy(writer, f); err != nil {
		return "", false, err
	}
	if err := writer.Close(); err != nil {
		return "", false, err
	}
	return storageLocation, false, nil
}

func (w worker) complete(lid scari.LeaseID, pj ProcessedJob) error {
	cjr := scari.CompleteJobRequest{LeaseID: lid, StorageURL: pj.StoragePath}
	body, err := json.Marshal(cjr)
	if err != nil {
		return err
	}
	_, err = w.c.Post(w.apiserver+"jobs/"+string(pj.Job.ID)+"/complete", "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	return nil
}
