package worker

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os/exec"

	"github.com/rafax/scari"
)

type Worker interface {
	Process() ([]scari.Job, error)
}

func New(apiserver string, outDir string) Worker {
	commandParams := []string{"--restrict-filenames", "-o", outDir + "%(title)s.%(ext)s"}
	return worker{
		apiserver: apiserver, outDir: outDir, audioParams: append(commandParams, audioSuffix...),
		videoParams: append(commandParams, videoSuffix...), c: http.DefaultClient}
}

type noPendingJobs error

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

func (w worker) Process() ([]scari.Job, error) {
	j, err := w.fetch()
	if err != nil {
		return nil, err
	}
	if j == nil {
		return nil, nil
	}

	return []scari.Job{*j}, err
}

func (w worker) fetch() (*scari.Job, error) {
	r, err := w.c.Post(w.apiserver+"jobs/lease", "application/json", nil)
	if err != nil {
		return nil, err
	}
	if r.StatusCode == 204 {
		return nil, nil
	}
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	var jr scari.LeaseJobResponse
	err = json.Unmarshal(body, &jr)
	if err != nil {
		return nil, err
	}
	return &jr.Job, nil
}

func (w worker) convert(j scari.Job) {
	var params []string
	if j.Output == scari.AUDIO {
		params = w.audioParams
	} else {
		params = w.videoParams
	}
	c := exec.Command(command, append(params, j.Source)...)
	c.Dir = w.outDir
	err := c.Run()
	if err != nil {
		return nil, err
	}
}
