// Package p contains an HTTP Cloud Function.
package p

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/minio/minio-go"
)

type Href struct {
	Url    string `json:"href"`
	Method string `json:"method"`
}

type Links struct {
	Url Href `json:"dashboard_download_direct"`
}

type Hook struct {
	LinkList Links `json:"links"`
}

func Decode(w http.ResponseWriter, r *http.Request) {
	accessKey := os.Getenv("accessKey")
	secretKey := os.Getenv("secret")
	endpoint := "nyc3.digitaloceanspaces.com"
	client, err := minio.New(endpoint, accessKey, secretKey, true)
	if err != nil {
		log.Printf("error creating client: %v", err)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
	}

	hook := Hook{}
	if err := json.NewDecoder(r.Body).Decode(&hook); err != nil {
		log.Printf("error occured: %v", err)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
	}

	url := ExtractUrl(hook)

	if err := Download(url); err != nil {
		log.Printf("error occured while downloading: %v", err)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
	}

	opts := minio.PutObjectOptions{}
	uploadInfo, err := client.FPutObject("deleptualspace", "final-verdict-cicd-test/build", "tmp/build.zip", opts)
	if err != nil {
		log.Printf("error occured while uploading: %v", err)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
	}

	log.Printf("Successfully uploaded object: %v", uploadInfo)
}

//extract download_direct url
func ExtractUrl(request Hook) string {
	return request.LinkList.Url.Url
}

func Download(url string) error {
	out, err := os.Create("tmp/build.zip")
	if err != nil {
		return err
	}
	defer out.Close()

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
}
