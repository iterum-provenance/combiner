package uploader

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"

	"github.com/iterum-provenance/iterum-go/util"
)

// constructMultiFileRequest creates a new file upload http request with optional extra otherParams
func constructMultiFileRequest(url string, otherParams map[string]string, nameFileMap map[string]string) (request *http.Request, err error) {
	defer util.ReturnErrOnPanic(&err)()

	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)

	for filename, path := range nameFileMap {
		file, err := os.Open(path)
		util.PanicIfErr(err, "")
		defer file.Close()

		part, err := writer.CreateFormFile(filepath.Base(path), filename)
		util.PanicIfErr(err, "")
		io.Copy(part, file)
	}

	for key, val := range otherParams {
		_ = writer.WriteField(key, val)
	}
	err = writer.Close()
	util.PanicIfErr(err, "")

	request, err = http.NewRequest("POST", url, body)
	util.PanicIfErr(err, "")

	request.Header.Add("Content-Type", writer.FormDataContentType())
	return
}

func postMultipartForm(url string, filemap map[string]string) (response *http.Response, err error) {
	defer util.ReturnErrOnPanic(&err)()
	request, err := constructMultiFileRequest(url, nil, filemap)
	util.PanicIfErr(err, "")
	client := &http.Client{}
	return client.Do(request)
}
