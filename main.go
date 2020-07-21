package main

import (
	"path"
	"sync"

	"github.com/iterum-provenance/iterum-go/daemon"
	"github.com/iterum-provenance/iterum-go/manager"
	"github.com/iterum-provenance/iterum-go/process"
	"github.com/iterum-provenance/iterum-go/transmit"
	"github.com/iterum-provenance/iterum-go/util"
	mq "github.com/iterum-provenance/sidecar/messageq"
	"github.com/iterum-provenance/sidecar/store"

	"github.com/iterum-provenance/combiner/uploader"
)

func main() {
	// log.Base().SetLevel("DEBUG")
	var wg sync.WaitGroup

	// Incoming messages from the queue are passed on for downloading
	mqDownloaderBridgeBufferSize := 10
	mqDownloaderBridge := make(chan transmit.Serializable, mqDownloaderBridgeBufferSize)

	// Connect minio downloader to daemon uploader
	downloaderUploaderBridgeBufferSize := 10
	downloaderUploaderBridge := make(chan transmit.Serializable, downloaderUploaderBridgeBufferSize)

	// Connect uploaded messages to the acknowledger of the mqListener
	uploaderAcknowledgerBridgeBufferSize := 10
	uploaderAcknowledgerBridge := make(chan transmit.Serializable, uploaderAcknowledgerBridgeBufferSize)

	// Consume messages from the queue, passing it on to the download manager
	mqListener, err := mq.NewListener(mqDownloaderBridge, uploaderAcknowledgerBridge, mq.BrokerURL, mq.InputQueue, mq.PrefetchCount)
	util.Ensure(err, "MessageQueue listener succesfully created and listening")
	mqListener.Start(&wg)

	// Pass the consumed messages to the downloader
	downloadManager := store.NewDownloadManager(process.DataVolumePath, mqDownloaderBridge, downloaderUploaderBridge)
	downloadManager.Start(&wg)

	// After downloading, upload thhe data to the daemon
	uri := path.Join(daemon.URL, daemon.Dataset, "pipeline_result", process.PipelineHash)
	daemonUploader := uploader.NewDaemonUploader(downloaderUploaderBridge, uploaderAcknowledgerBridge, uri)
	daemonUploader.Start(&wg)

	// Periodically check until previous steps are done, meaning that if we finish the queue now, it will be the last messages
	usChecker := manager.NewUpstreamChecker(manager.URL, process.PipelineHash, process.Name, 5)
	usChecker.Start(&wg)
	usChecker.Register <- mqListener.CanExit

	wg.Wait()
}
