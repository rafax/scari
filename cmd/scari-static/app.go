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
	fmt.Fprintf(w, v["fileid"])
}

func postHandler(w http.ResponseWriter, r *http.Request) {
	var sfr StaticFileRequest
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(500)
		fmt.Fprintf(w, "{\"error\":\"%v\"}", err)
		return
	}
	err = json.Unmarshal(body, &sfr)
	if err != nil {
		w.WriteHeader(500)
		fmt.Fprintf(w, "{\"error\":\"%v\"}", err)
		return
	}
	ctx := appengine.NewContext(r)
	client, err := storage.NewClient(ctx)
	if err != nil {
		w.WriteHeader(500)
		fmt.Fprintf(w, "{\"error\":\"%v\"}", err)
		return
	}
	bucketName, err := file.DefaultBucketName(ctx)
	if err != nil {
		w.WriteHeader(500)
		fmt.Fprintf(w, "{\"error\":\"%v\"}", err)
		return
	}
	bucket := client.Bucket(bucketName)
	_, err = bucket.Object(sfr.FileName).Attrs(ctx)
	if err != nil {
		w.WriteHeader(500)
		fmt.Fprintf(w, "{\"error\":\"%v\"}", err)
		return
	}
	key := datastore.NewIncompleteKey(ctx, "File", nil)
	file := staticFile{
		FileName:   sfr.FileName,
		StorageURL: fmt.Sprintf("https://storage.googleapis.com/%s/%s", scari.StorageBucketName, sfr.FileName),
	}
	k, err := datastore.Put(ctx, key, &file)
	if err != nil {
		w.WriteHeader(500)
		fmt.Fprintf(w, "{\"error\":\"%v\"}", err)
		return
	}
	id := strconv.FormatInt(k.IntID(), 10)
	resp, err := json.Marshal(StaticFileResponse{Id: id})
	if err != nil {
		w.WriteHeader(500)
		fmt.Fprintf(w, "{\"error\":\"%v\"}", err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(resp)
}

func checkErr(err error, w http.ResponseWriter) {

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
