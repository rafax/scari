package main

import (
	"os"
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/rafax/scari/worker"
)

func main() {
	w := initWorker()
	for {
		processed, err := w.Process()
		if err != nil {
			log.Errorf("Error during processing: %v", err)
		}
		if processed == nil {
			log.Debugf("No jobs to process")
			time.Sleep(10 * time.Second)
		}
		for _, j := range processed {
			log.Infof("Processed %+v", j)
		}
	}
}

func initWorker() worker.Worker {
	log.SetLevel(log.DebugLevel)
	apiserver := os.Getenv("SCARI_SERVER")
	if apiserver == "" {
		apiserver = "http://localhost:3001/"
	}
	outDir := os.Getenv("SCARI_OUTDIR")
	if outDir == "" {
		outDir = "/tmp/out"
	}
	log.Infof("Starting scari-worker with apiserver: %v outDir: %v", apiserver, outDir)
	return worker.New(apiserver, outDir)
}
