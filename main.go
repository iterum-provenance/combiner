package main

import (
	"os"
	"sync"

	"github.com/iterum-provenance/combiner/uploader"

	"github.com/iterum-provenance/combiner/util"
	"github.com/iterum-provenance/sidecar/messageq"
	"github.com/iterum-provenance/sidecar/store"
	"github.com/iterum-provenance/sidecar/transmit"
)

func main() {
	var wg sync.WaitGroup

	mqDownloaderBridgeBufferSize := 10
	mqDownloaderBridge := make(chan transmit.Serializable, mqDownloaderBridgeBufferSize)

	downloaderSocketBridgeBufferSize := 10
	downloaderSocketBridge := make(chan transmit.Serializable, downloaderSocketBridgeBufferSize)

	// Download manager setup
	downloadManager := store.NewDownloadManager(mqDownloaderBridge, downloaderSocketBridge)
	downloadManager.Start(&wg)

	brokerURL := os.Getenv("BROKER_URL")
	inputQueue := os.Getenv("INPUT_QUEUE")

	// MessageQueue setup
	mqListener, err := messageq.NewListener(mqDownloaderBridge, brokerURL, inputQueue)
	util.Ensure(err, "MessageQueue listener succesfully created and listening")
	mqListener.Start(&wg)

	daemonURL := os.Getenv("DAEMON_URL")
	pipelineHash := os.Getenv("PIPELINE_HASH")
	datasetName := os.Getenv("DATASET_NAME")

	uploaderListener := uploader.NewListener(downloaderSocketBridge, daemonURL+"/"+datasetName+"/pipeline_result/"+pipelineHash)
	uploaderListener.Start(&wg)

	wg.Wait()
}
