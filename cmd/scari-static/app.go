package hello

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

	"cloud.google.com/go/storage"

	"github.com/gorilla/mux"
	"github.com/rafax/scari"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/file"
)

const (
	kind = "Key"
)

func init() {
	router := mux.NewRouter()
	router.HandleFunc("/", homeHandler)
	router.HandleFunc("/files/{fileid}", getHandler)
	router.HandleFunc("/files", postHandler).Methods("POST")
	http.Handle("/", router)
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "scari!")
}

func getHandler(w http.ResponseWriter, r *http.Request) {
	v := mux.Vars(r)
	id, err := strconv.ParseInt(v["fileid"], 10, 64)
	if err != nil {
		http.Error(w, format(err), 500)
		return
	}
	ctx := appengine.NewContext(r)
	file := new(staticFile)
	if err = datastore.Get(ctx, datastore.NewKey(ctx, kind, "", id, nil), file); err != nil {
		http.Error(w, format(err), 500)
		return
	}
	fmt.Fprintf(w, file.StorageURL)
}

func postHandler(w http.ResponseWriter, r *http.Request) {
	var sfr StaticFileRequest
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, format(err), 500)
		return
	}
	err = json.Unmarshal(body, &sfr)
	if err != nil {
		http.Error(w, format(err), 500)
		return
	}
	ctx := appengine.NewContext(r)
	client, err := storage.NewClient(ctx)
	if err != nil {
		http.Error(w, format(err), 500)
		return
	}
	bucketName, err := file.DefaultBucketName(ctx)
	if err != nil {
		http.Error(w, format(err), 500)
		return
	}
	bucket := client.Bucket(bucketName)
	_, err = bucket.Object(sfr.FileName).Attrs(ctx)
	if err != nil {
		http.Error(w, format(err), 500)
		return
	}
	key := datastore.NewIncompleteKey(ctx, kind, nil)
	file := staticFile{
		FileName:   sfr.FileName,
		StorageURL: fmt.Sprintf("https://storage.googleapis.com/%s/%s", scari.StorageBucketName, sfr.FileName),
	}
	k, err := datastore.Put(ctx, key, &file)
	if err != nil {
		http.Error(w, format(err), 500)
		return
	}
	id := strconv.FormatInt(k.IntID(), 10)
	resp, err := json.Marshal(StaticFileResponse{Id: id})
	if err != nil {
		http.Error(w, format(err), 500)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(resp)
}

type StaticFileRequest struct {
	FileName string `json:"fileName"`
}

type StaticFileResponse struct {
	Id string `json:"id"`
}

type staticFile struct {
	FileName   string
	StorageURL string
}

func format(err error) string {
	return fmt.Sprintf("{\"error\":\"%v\"}", err)
}
