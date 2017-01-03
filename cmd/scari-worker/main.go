package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	log "github.com/Sirupsen/logrus"

	"os/exec"

	"github.com/rafax/scari"
)

var (
	c             = http.Client{}
	command       = "youtube-dl"
	commandParams = []string{"-o", "/out/%(title)s.%(ext)s"}
	audioSuffix   = []string{"-x", "--audio-format", "mp3"}
	videoSuffix   = []string{"--recode-video", "mp4"}
	audioParams   = append(commandParams, audioSuffix...)
	videoParams   = append(commandParams, videoSuffix...)
	seen          = map[scari.JobID]bool{}
)

func main() {
	apiserver := os.Getenv("SCARI_SERVER")
	if apiserver == "" {
		apiserver = "http://localhost:3001/"
	}
	doIt(apiserver)
	for {
		select {
		case <-time.Tick(60 * time.Second):
			doIt(apiserver)
		}
	}
}

type noPendingJobs error

func doIt(apiserver string) {
	j, err := fetchOne(apiserver)
	if err != nil {
		log.Error(err)
		return
	}
	if j == nil {
		log.Info("No pending jobs found")
		return
	}
	var params []string
	if j.Output == scari.AUDIO {
		params = audioParams
	} else {
		params = videoParams
	}
	c := exec.Command(command, append(params, j.Source)...)
	c.Dir = "/out"
	log.Infof("Starting %v", c)
	out, err := c.Output()
	if err != nil {
		log.Warnf("Error when starting(): %v", err)
	}
	log.Info(string(out))
}

func fetchOne(apiserver string) (*scari.Job, error) {
	r, err := c.Get(apiserver + "jobs")
	if err != nil {
		return nil, err
	}
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	var jr scari.JobsResponse
	err = json.Unmarshal(body, &jr)
	if err != nil {
		return nil, err
	}
	if len(jr.Jobs) == 0 {
		return nil, nil
	}
	for _, j := range jr.Jobs {
		if j.Status == scari.Pending && !seen[j.ID] {
			seen[j.ID] = true
			return &j, nil
		}
	}
	return nil, nil
}
