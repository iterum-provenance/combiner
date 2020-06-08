package main

import (
	"sync"

	"github.com/iterum-provenance/iterum-go/daemon"
	"github.com/iterum-provenance/iterum-go/env"
	"github.com/iterum-provenance/iterum-go/minio"
	"github.com/iterum-provenance/iterum-go/transmit"
	"github.com/iterum-provenance/iterum-go/util"
	"github.com/iterum-provenance/sidecar/manager"
	"github.com/iterum-provenance/sidecar/messageq"
	"github.com/iterum-provenance/sidecar/store"
	"github.com/prometheus/common/log"

	_ "github.com/iterum-provenance/combiner/env" // Run the init script checking the env variable values
	"github.com/iterum-provenance/combiner/uploader"
)

func main() {
	log.Base().SetLevel("DEBUG")
	var wg sync.WaitGroup

	mqDownloaderBridgeBufferSize := 10
	mqDownloaderBridge := make(chan transmit.Serializable, mqDownloaderBridgeBufferSize)

	downloaderSocketBridgeBufferSize := 10
	downloaderSocketBridge := make(chan transmit.Serializable, downloaderSocketBridgeBufferSize)

	uploaderAcknowledgerBridgeBufferSize := 10
	uploaderAcknowledgerBridge := make(chan transmit.Serializable, uploaderAcknowledgerBridgeBufferSize)

	// Download manager setup
	daemonConfig := daemon.NewDaemonConfigFromEnv()
	minioConf, err := minio.NewMinioConfigFromEnv() // defaults to an output setup
	util.PanicIfErr(err, "")
	minioConf.TargetBucket = "INVALID" // adjust such that the target output is unusable
	err = minioConf.Connect()
	util.PanicIfErr(err, "")
	downloadManager := store.NewDownloadManager(minioConf, mqDownloaderBridge, downloaderSocketBridge)
	downloadManager.Start(&wg)

	// MessageQueue setup
	mqListener, err := messageq.NewListener(mqDownloaderBridge, uploaderAcknowledgerBridge, env.MQBrokerURL, env.MQInputQueue, env.MQPrefetchCount)
	util.Ensure(err, "MessageQueue listener succesfully created and listening")
	mqListener.Start(&wg)

	uri := daemonConfig.DaemonURL + "/" + daemonConfig.Dataset + "/pipeline_result/" + env.PipelineHash
	uploaderListener := uploader.NewListener(downloaderSocketBridge, uploaderAcknowledgerBridge, uri)
	uploaderListener.Start(&wg)

	usChecker := manager.NewUpstreamChecker(env.ManagerURL, env.PipelineHash, env.ProcessName, 5)
	usChecker.Start(&wg)
	usChecker.Register <- mqListener.CanExit

	wg.Wait()
}
