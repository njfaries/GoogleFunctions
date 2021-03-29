// Package p contains an HTTP Cloud Function.
package p

import (
	"encoding/json"
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

	spaces, err := client.ListBuckets()
	if err != nil {
		log.Printf("error fetching buckets: %v", err)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
	}

	for _, space := range spaces {
		log.Printf("Space: %v", space.Name)
	}
	hook := Hook{}
	if err := json.NewDecoder(r.Body).Decode(&hook); err != nil {
		log.Printf("error occured: %v", err)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
	}

	url := ExtractUrl(hook)

	log.Print(url)
}

//extract download_direct url
func ExtractUrl(request Hook) string {
	return request.LinkList.Url.Url
}
