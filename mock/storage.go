package mock

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"path"

	"github.com/rafax/scari"
)

func NewStorageClient() scari.StorageClient {
	return &mockStorageClient{}
}

type mockStorageClient struct {
	client *http.Client
}

func (msc *mockStorageClient) Register(fileName string) (string, error) {
	sfr := scari.StaticFileRequest{FileName: path.Base(fileName)}
	body, err := json.Marshal(sfr)
	if err != nil {
		return "", err
	}
	resp, err := msc.client.Post("http://scari-666.appspot.com/files", "application/json", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	rbody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	sfresp := new(scari.StaticFileResponse)
	if err = json.Unmarshal(rbody, sfresp); err != nil {
		return "", err
	}
	return sfresp.Id, nil
}
