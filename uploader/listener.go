package uploader

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/iterum-provenance/sidecar/data"
	"github.com/iterum-provenance/sidecar/transmit"
	"github.com/prometheus/common/log"
)

// Listener is the structure that listens to RabbitMQ and redirects messages to a channel
type Listener struct {
	downloadChannel <-chan transmit.Serializable // data.RemoteFragmentDesc
	DaemonURL       string
}

// NewListener creates a new uploader listener
func NewListener(channel <-chan transmit.Serializable, daemonURL string) Listener {

	return Listener{channel, daemonURL}
}

// StartBlocking listens on the rabbitMQ messagequeue and redirects messages on the INPUT_QUEUE to a channel
func (listener Listener) StartBlocking() {
	for message := range listener.downloadChannel {
		fmt.Printf("Received on channel: %v\n", message)
		lfd := fragmentDesc{*message.(*data.LocalFragmentDesc)}

		filemap := make(map[string]string)

		for _, file := range lfd.Files {
			// fmt.Println(file.Name)
			filemap[file.Name] = file.LocalPath
		}
		fmt.Printf("Frag: %v\n", lfd.Files)
		fmt.Printf("Sending file to daemon..\n")
		fmt.Printf("DaemonUrl: %s..\n", listener.DaemonURL)
		response, err := PostMultipartForm(listener.DaemonURL, filemap)
		if err != nil {
			log.Errorf("Upload failed due to: '%v'", err)
		}
		switch response.StatusCode {
		case http.StatusOK:
			break
		default:
			err = fmt.Errorf("Error: POST multipart form failed, daemon responded with statuscode %v", response.StatusCode)
			return
		}

	}
}

// Start asychronously calls StartBlocking via Gorouting
func (listener Listener) Start(wg *sync.WaitGroup) {
	go func() {
		listener.StartBlocking()
	}()
}
