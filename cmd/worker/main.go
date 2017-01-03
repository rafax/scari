package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	log "github.com/Sirupsen/logrus"

	"os/exec"

	"errors"

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
func doIt(apiserver string) {
	j, err := fetchOne(apiserver)
	if err != nil {
		log.Error(err)
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
		log.Info("No jobs found")
		return nil, errors.New("No jobs found")
	}
	pending := 0
	var first *scari.Job
	for _, j := range jr.Jobs {
		if j.Status == scari.Pending {
			if first == nil {
				first = &j
			}
			pending++
		}
	}
	log.Infof("Found %v pending jobs, returning %v", pending, *first)
	return first, nil
}
