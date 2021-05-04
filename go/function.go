// Package p contains an HTTP Cloud Function.
package p

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/minio/minio-go"
)

type Href struct {
	Url    string `json:"href"`
	Method string `json:"method"`
}

type Artifact struct {
	Files Href `json:"files[0]"`
}

type Links struct {
	Url       Href     `json:"api_self"`
	Artifacts Artifact `json:"artifacts[0]"`
	Download  Href     `json:"download_primary"`
}

type Hook struct {
	LinkList Links `json:"links"`
}

type Request struct {
}

const RootUrl = "https://build-api.cloud.unity3d.com"

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

	url, err := GetDownloadUrl(ConstructUrl(hook))
	if err != nil {
		log.Printf("error occured while getting download url: %v", err)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
	}

	if err := Download(GetAssetUrl(hook), true); err != nil {
		log.Printf("error occured while downloading assets: %v", err)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
	}

	if err := Download(url, false); err != nil {
		log.Printf("error occured while downloading: %v", err)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
	}

	//Unzip build data
	files, err := Unzip("/tmp/build.zip", "/tmp/build", false)
	if err != nil {
		log.Printf("error occured while unzipping: %v", err)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
	}

	//Unzip asset data
	assets, err := Unzip("/tmp/assets.zip", "/tmp/assets", true)
	if err != nil {
		log.Printf("error occured while unzipping assets: %v", err)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
	}

	//Upload build data
	if err := Upload(files, "/tmp/build/Default WebGL/", "final-verdict-cicd-test/", "deleptualspace", client); err != nil {
		log.Printf("error occured while uploading build data: %v", err)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
	}

	//Upload assets
	if err := Upload(assets, "/tmp/assets/ServerData/", "final-verdict-cicd-test/", "monikerspace", client); err != nil {
		log.Printf("error occured while uploading assets: %v", err)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
	}
}

//extract download_direct url
func ConstructUrl(request Hook) string {
	url := request.LinkList.Url.Url
	return RootUrl + url
}

func GetAssetUrl(request Hook) string {
	url := request.LinkList.Artifacts.Files.Url
	log.Printf("request.LinkList.Artifacts: %v", request.LinkList.Artifacts)
	log.Printf("request.LinkList.Artifacts.Files: %v", request.LinkList.Artifacts.Files)
	return url
}

func GetDownloadUrl(url string) (string, error) {
	unityApiKey := os.Getenv("unityApiKey")
	// reader := strings.NewReader("{Content-Type: application/json, Authentication: Basic " + unityApiKey + "}")
	request, err := http.NewRequest("GET", url, nil)
	request.Header.Add("Content-Type", "application/json")
	request.Header.Add("Authorization", "Basic "+unityApiKey)
	if err != nil {
		return "", err
	}
	log.Printf("Request: %v", request)

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return "", err
	}

	log.Printf("Response: %v", response)

	downloadHook := Hook{}
	if err := json.NewDecoder(response.Body).Decode(&downloadHook); err != nil {
		log.Printf("error occured: %v", err)
		return "", err
	}

	log.Printf("Download link: %v", downloadHook.LinkList.Download.Url)

	return downloadHook.LinkList.Download.Url, nil
}

func Download(url string, isAssets bool) error {
	var out *os.File
	var err error
	if isAssets {
		out, err = os.Create("/tmp/assets.zip")
	} else {
		out, err = os.Create("/tmp/build.zip")
	}
	log.Printf("isAssets is %v", isAssets)
	log.Printf("Getting from url: %s", url)

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

func Upload(files []string, src string, dest string, space string, client *minio.Client) error {
	opts := minio.PutObjectOptions{}
	for _, f := range files {
		log.Printf("File being uploaded: %s", f)
		trimmedFilePath := strings.ReplaceAll(f, src, "")
		_, err := client.FPutObject(space, dest+trimmedFilePath, f, opts)
		if err != nil {
			return err
		}
	}
	return nil
}
