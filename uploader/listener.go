package uploader

import (
	"net/http"
	"sync"

	desc "github.com/iterum-provenance/iterum-go/descriptors"
	"github.com/iterum-provenance/iterum-go/transmit"
	"github.com/prometheus/common/log"
)

// Listener is the structure that listens to RabbitMQ and redirects messages to a channel
type Listener struct {
	downloadChannel <-chan transmit.Serializable // desc.RemoteFragmentDesc
	DaemonURL       string
}

// NewListener creates a new uploader listener
func NewListener(channel <-chan transmit.Serializable, daemonURL string) Listener {

	return Listener{channel, daemonURL}
}

// StartBlocking listens on the rabbitMQ messagequeue and redirects messages on the INPUT_QUEUE to a channel
func (listener Listener) StartBlocking() {
	for message := range listener.downloadChannel {
		log.Debugf("Received on channel: %v\n", message)
		lfd := fragmentDesc{*message.(*desc.LocalFragmentDesc)}

		filemap := make(map[string]string)

		for _, file := range lfd.Files {
			filemap[file.Name] = file.LocalPath
		}
		log.Debugf("Frag: %v\n", lfd.Files)
		log.Debugf("Sending files to daemon..\n")
		response, err := postMultipartForm(listener.DaemonURL, filemap)
		if err != nil {
			log.Fatalf("Upload failed due to: '%v'", err)
		}
		switch response.StatusCode {
		case http.StatusOK:
			break
		default:
			log.Fatalf("Error: POST multipart form failed, daemon responded with statuscode %v", response.StatusCode)
			return
		}

	}
}

// Start asychronously calls StartBlocking via Gorouting
func (listener Listener) Start(wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		listener.StartBlocking()
	}()
}
