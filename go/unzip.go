package p

import (
	"archive/zip"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	Unzip(os.Args[1], os.Args[2], false)
}

func Unzip(src string, dest string, assets bool) ([]string, error) {

	var filenames []string

	r, err := zip.OpenReader(src)
	if err != nil {
		log.Printf("File provided %s is not a valid zip file", src)
		return filenames, err
	}
	defer r.Close()

	for _, f := range r.File {

		// Store filename/path for returning and using later on
		fpath := filepath.Join(dest, f.Name)

		if assets {
			if strings.Contains(fpath, "Assets/") || strings.Contains(fpath, "Library/") {
				continue
			}

		} else {
			if strings.Contains(fpath, "Template") || strings.Contains(fpath, "index.html") {
				continue
			}

			if strings.Contains(fpath, "Build/") {
				fpath = strings.ReplaceAll(fpath, "Build/", "")
			}

		}

		// Check for ZipSlip. More Info: http://bit.ly/2MsjAWE
		if !strings.HasPrefix(fpath, filepath.Clean(dest)+string(os.PathSeparator)) {
			return filenames, fmt.Errorf("%s: illegal file path", fpath)
		}

		filenames = append(filenames, fpath)

		if f.FileInfo().IsDir() {
			// Make Folder
			os.MkdirAll(fpath, os.ModePerm)
			continue
		}

		// Make File
		if err = os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return filenames, err
		}

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return filenames, err
		}

		rc, err := f.Open()
		if err != nil {
			return filenames, err
		}

		_, err = io.Copy(outFile, rc)

		// Close the file without defer to close before next iteration of loop
		outFile.Close()
		rc.Close()

		if err != nil {
			return filenames, err
		}
	}
	return filenames, nil
}
