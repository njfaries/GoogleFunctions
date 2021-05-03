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
	Files Href `json:"files"`
}

type Links struct {
	Url       Href     `json:"api_self"`
	Artifacts Artifact `json:"artifacts"`
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

	if err := Download(GetAssetUrl(hook)); err != nil {
		log.Printf("error occured while downloading assets: %v", err)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
	}

	if err := Download(url); err != nil {
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

	opts := minio.PutObjectOptions{}
	for _, f := range files {
		log.Printf("File being uploaded: %s", f)
		trimmedFilePath := strings.ReplaceAll(f, "/tmp/build/Default WebGL/", "")
		_, err := client.FPutObject("deleptualspace", "final-verdict-cicd-test/"+trimmedFilePath, f, opts)
		if err != nil {
			log.Printf("error occured while uploading: %v", err)
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		}

	}

	for _, f := range assets {
		log.Printf("File being uploaded: %s", f)
		trimmedFilePath := strings.ReplaceAll(f, "/tmp/assets/ServerData/", "")
		_, err := client.FPutObject("monikerspace", "final-verdict-cicd-test/"+trimmedFilePath, f, opts)
		if err != nil {
			log.Printf("error occured while uploading assets: %v", err)
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		}
	}
}

//extract download_direct url
func ConstructUrl(request Hook) string {
	url := request.LinkList.Url.Url
	return RootUrl + url
}

func GetAssetUrl(request Hook) string {
	url := request.LinkList.Artifacts.Files.Url
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

func Download(url string) error {
	out, err := os.Create("/tmp/build.zip")
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

// Unzip will decompress a zip archive, moving all files and folders
// within the zip file (parameter 1) to an output directory (parameter 2).
// func Unzip(src string, dest string) ([]string, error) {

// 	var filenames []string

// 	r, err := zip.OpenReader(src)
// 	if err != nil {
// 		log.Printf("File provided %s is not a valid zip file", src)
// 		return filenames, err
// 	}
// 	defer r.Close()

// 	for _, f := range r.File {

// 		// Store filename/path for returning and using later on
// 		fpath := filepath.Join(dest, f.Name)

// 		// Check for ZipSlip. More Info: http://bit.ly/2MsjAWE
// 		if !strings.HasPrefix(fpath, filepath.Clean(dest)+string(os.PathSeparator)) {
// 			return filenames, fmt.Errorf("%s: illegal file path", fpath)
// 		}

// 		filenames = append(filenames, fpath)

// 		if f.FileInfo().IsDir() {
// 			// Make Folder
// 			os.MkdirAll(fpath, os.ModePerm)
// 			continue
// 		}

// 		// Make File
// 		if err = os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
// 			return filenames, err
// 		}

// 		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
// 		if err != nil {
// 			return filenames, err
// 		}

// 		rc, err := f.Open()
// 		if err != nil {
// 			return filenames, err
// 		}

// 		_, err = io.Copy(outFile, rc)

// 		// Close the file without defer to close before next iteration of loop
// 		outFile.Close()
// 		rc.Close()

// 		if err != nil {
// 			return filenames, err
// 		}
// 	}
// 	return filenames, nil
// }
