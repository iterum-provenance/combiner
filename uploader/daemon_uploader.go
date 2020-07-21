package uploader

import (
	"net/http"
	"sync"

	desc "github.com/iterum-provenance/iterum-go/descriptors"
	"github.com/iterum-provenance/iterum-go/transmit"
	"github.com/prometheus/common/log"
)

// DaemonUploader is the structure that uploads data to the daemon
type DaemonUploader struct {
	downloadChannel <-chan transmit.Serializable // desc.RemoteFragmentDesc
	complete        chan<- transmit.Serializable // desc.FinishedFragmentMessage
	DaemonURL       string
}

// NewDaemonUploader creates a new uploader
func NewDaemonUploader(channel, complete chan transmit.Serializable, daemonURL string) DaemonUploader {
	return DaemonUploader{channel, complete, daemonURL}
}

// StartBlocking listens on the rabbitMQ messagequeue and redirects messages on the INPUT_QUEUE to a channel
func (uploader DaemonUploader) StartBlocking() {
	for message := range uploader.downloadChannel {
		log.Debugf("DaemonUploader received '%v'\n", message)

		// Convert LocalFragmentDescription to a filemap fit for posting
		lfd := fragmentDesc{*message.(*desc.LocalFragmentDesc)}
		filemap := make(map[string]string)
		for _, file := range lfd.Files {
			filemap[file.Name] = file.LocalPath
		}

		response, err := postMultipartForm(uploader.DaemonURL, filemap)
		if err != nil {
			log.Fatalf("DaemonUploader encountered an error: '%v'", err)
		}

		// Handle response
		switch response.StatusCode {
		case http.StatusOK:
			uploader.complete <- &desc.FinishedFragmentMessage{FragmentID: lfd.Metadata.FragmentID}
			break
		default:
			log.Fatalf("POST multipart form failed, daemon responded with statuscode %v", response.StatusCode)
			return
		}
	}

	log.Infoln("DaemonUploader is done, finishing up...")
	close(uploader.complete)
}

// Start asychronously calls StartBlocking via Gorouting
func (uploader DaemonUploader) Start(wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		uploader.StartBlocking()
	}()
}
