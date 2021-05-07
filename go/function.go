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

type File struct {
	Filename string `json:"filename"`
	Url      string `json:"href"`
}

type Artifact struct {
	Files []File `json:"files"`
}

type Links struct {
	Url       Href       `json:"api_self"`
	Artifacts []Artifact `json:"artifacts"`
	Download  Href       `json:"download_primary"`
}

type Hook struct {
	LinkList    Links  `json:"links"`
	ProjectName string `json:"projectName"`
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
	log.Printf("Dumping json body")
	log.Print(r.Body)

	hook := Hook{}
	if err := json.NewDecoder(r.Body).Decode(&hook); err != nil {
		log.Printf("error occured: %v", err)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
	}

	log.Printf("Dumping hook")
	log.Print(hook)

	formattedName := FormatName(hook.ProjectName)
	// url, err := GetDownloadUrl(ConstructUrl(hook))
	// if err != nil {
	// 	log.Printf("error occured while getting download url: %v", err)
	// 	http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
	// }

	if err := Download(GetDownloadUrl(hook), false); err != nil {
		log.Printf("error occured while downloading build data: %v", err)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
	}

	if err := Download(GetAssetUrl(hook), true); err != nil {
		log.Printf("error occured while downloading assets: %v", err)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
	}

	//Unzip build data
	files, err := Unzip("/tmp/build.zip", "/tmp/build", false)
	if err != nil {
		log.Printf("error occured while unzipping build data: %v", err)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
	}

	//Unzip asset data
	assets, err := Unzip("/tmp/assets.zip", "/tmp/assets", true)
	if err != nil {
		log.Printf("error occured while unzipping assets: %v", err)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
	}

	//Upload build data
	if err := Upload(files, "/tmp/build/webgl/", "dev/"+formattedName, "deleptualspace", client); err != nil {
		log.Printf("error occured while uploading build data: %v", err)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
	}

	//Upload assets
	if err := Upload(assets, "/tmp/assets/ServerData/", "dev/"+formattedName+"/WebGL", "deleptualspace", client); err != nil {
		log.Printf("error occured while uploading assets: %v", err)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
	}
}

//extract download_direct url
func ConstructUrl(request Hook) string {
	url := request.LinkList.Url.Url
	return RootUrl + url
}

func GetDownloadUrl(request Hook) string {
	url := request.LinkList.Artifacts[1].Files[0].Url
	log.Printf("Download URL with new method: %s", url)
	return url
}

func GetAssetUrl(request Hook) string {
	url := request.LinkList.Artifacts[0].Files[0].Url
	return url
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
		log.Printf("Uploading file %s to %s", f, dest)
		trimmedFilePath := strings.ReplaceAll(f, src, "")
		log.Printf("Trimmed path for this file: %s", trimmedFilePath)
		_, err := client.FPutObject(space, dest+trimmedFilePath, f, opts)
		if err != nil {
			return err
		}
	}
	return nil
}

func FormatName(name string) string {
	name = strings.ReplaceAll(name, " ", "-")
	name = strings.ToLower(name)
	return name
}
