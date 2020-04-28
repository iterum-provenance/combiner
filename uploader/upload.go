package uploader

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"

	"github.com/iterum-provenance/sidecar/util"
	"github.com/prometheus/common/log"
)

// constructMultiFileRequest creates a new file upload http request with optional extra otherParams
func constructMultiFileRequest(url string, otherParams map[string]string, nameFileMap map[string]string) (request *http.Request, err error) {
	defer util.ReturnErrOnPanic(&err)()

	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)

	for filename, path := range nameFileMap {
		file, err := os.Open(path)
		if err != nil {
			log.Errorf("Upload failed due to: '%v'", err)
		}
		defer file.Close()

		part, err := writer.CreateFormFile(filepath.Base(path), filename)
		if err != nil {
			log.Errorf("Upload failed due to: '%v'", err)
		}
		io.Copy(part, file)
	}

	for key, val := range otherParams {
		_ = writer.WriteField(key, val)
	}
	err = writer.Close()
	if err != nil {
		log.Errorf("Upload failed due to: '%v'", err)
	}
	request, err = http.NewRequest("POST", url, body)
	if err != nil {
		log.Errorf("Upload failed due to: '%v'", err)
	}
	request.Header.Add("Content-Type", writer.FormDataContentType())

	return
}

func PostMultipartForm(url string, filemap map[string]string) (response *http.Response, err error) {
	defer util.ReturnErrOnPanic(&err)()
	request, err := constructMultiFileRequest(url, nil, filemap)
	if err != nil {
		log.Errorf("Upload failed due to: '%v'", err)
	}

	client := &http.Client{}
	return client.Do(request)
}
