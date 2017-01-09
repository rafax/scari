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

	log "github.com/Sirupsen/logrus"
	"github.com/rafax/scari"
	"github.com/rafax/scari/handlers"
	"golang.org/x/net/context"
)

type Worker interface {
	Process() ([]ProcessedJob, error)
}

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
	Job      scari.Job
	LeaseID  scari.LeaseID
	FileName string
	Existed  bool
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
	name, existed, err := w.upload(out)
	if err != nil {
		return nil, err
	}
	pj := ProcessedJob{Job: *j, FileName: name, Existed: existed}
	err = w.complete(lid, pj)
	if err != nil {
		return nil, err
	}
	return []ProcessedJob{}, err
}

func (w worker) fetch() (*scari.Job, scari.LeaseID, error) {
	log.Debugf("Starting fetch")
	r, err := w.c.Post(w.apiserver+"jobs/lease", "application/json", nil)
	if err != nil {
		return nil, "", err
	}
	defer r.Body.Close()
	if r.StatusCode == 204 {
		log.Debugf("Got 204, no jobs available")
		io.Copy(ioutil.Discard, r.Body)
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
	log.Debugf("Received %v with leaseID %v", jr.Job, jr.LeaseID)
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
	log.Debugf("Will convert %v with %v", j.ID, c)
	output, err := c.Output()
	if err != nil {
		return "", err
	}
	var out youtubeDLOutput
	err = json.Unmarshal(output, &out)
	if err != nil {
		return "", err
	}
	log.Debugf("Converted %v to %v", j.ID, out.FileName)
	return out.FileName, nil
}

func (w worker) upload(filePath string) (string, bool, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return "", false, err
	}
	const publicURL = "https://storage.googleapis.com/%s/%s"
	name := path.Base(filePath)
	storageLocation := fmt.Sprintf(publicURL, scari.StorageBucketName, name)
	if err != nil {
		return "", false, err
	}
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		return "", false, err
	}
	bkt := client.Bucket(scari.StorageBucketName)
	_, err = bkt.Object(name).Attrs(ctx)
	if err == nil {
		// object exists
		return storageLocation, true, nil
	}
	if err != nil && err != storage.ErrObjectNotExist {
		return "", false, err
	}
	log.Debugf("File %v does not exist in storage", name)
	writer := bkt.Object(name).If(storage.Conditions{DoesNotExist: true}).NewWriter(ctx)
	writer.ACL = []storage.ACLRule{{Entity: storage.AllUsers, Role: storage.RoleReader}}
	writer.CacheControl = "public, max-age=86400"
	if _, err := io.Copy(writer, f); err != nil {
		return "", false, err
	}
	if err := writer.Close(); err != nil {
		return "", false, err
	}
	log.Debugf("File %v uploaded to %v", storageLocation)
	return storageLocation, false, nil
}

func (w worker) complete(lid scari.LeaseID, pj ProcessedJob) error {
	cjr := handlers.CompleteJobRequest{LeaseID: lid, FileName: pj.FileName}
	body, err := json.Marshal(cjr)
	if err != nil {
		return err
	}
	log.Debugf("Will mark %v as completed", cjr)
	r, err := w.c.Post(w.apiserver+"jobs/"+string(pj.Job.ID)+"/complete", "application/json", bytes.NewReader(body))
	if err != nil {
		log.Debugf("Failed completionwith %v", err)
		return err
	}
	if r.StatusCode/100 != 2 {
		rbody, err := ioutil.ReadAll(r.Body)
		if err != nil {
			return err
		}
		return fmt.Errorf("Failed to complete %v: got code %v reason %v", pj, r.StatusCode, string(rbody))
	}
	io.Copy(ioutil.Discard, r.Body)
	r.Body.Close()
	return nil
}
