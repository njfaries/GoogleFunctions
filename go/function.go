// Package p contains an HTTP Cloud Function.
package p

import (
	"encoding/json"
	"log"
	"net/http"
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
