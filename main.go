package main

import (
	"sync"

	"github.com/iterum-provenance/combiner/daemon"
	"github.com/iterum-provenance/iterum-go/env"
	"github.com/iterum-provenance/iterum-go/minio"
	"github.com/iterum-provenance/iterum-go/transmit"
	"github.com/iterum-provenance/iterum-go/util"
	"github.com/iterum-provenance/sidecar/manager"
	"github.com/iterum-provenance/sidecar/messageq"
	"github.com/iterum-provenance/sidecar/store"

	"github.com/iterum-provenance/combiner/uploader"
)

func main() {
	var wg sync.WaitGroup

	mqDownloaderBridgeBufferSize := 10
	mqDownloaderBridge := make(chan transmit.Serializable, mqDownloaderBridgeBufferSize)

	downloaderSocketBridgeBufferSize := 10
	downloaderSocketBridge := make(chan transmit.Serializable, downloaderSocketBridgeBufferSize)

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
	mqListener, err := messageq.NewListener(mqDownloaderBridge, env.MQBrokerURL, env.MQInputQueue)
	util.Ensure(err, "MessageQueue listener succesfully created and listening")
	mqListener.Start(&wg)

	uri := daemonConfig.DaemonURL + "/" + daemonConfig.Dataset + "/pipeline_result/" + env.PipelineHash
	uploaderListener := uploader.NewListener(downloaderSocketBridge, uri)
	uploaderListener.Start(&wg)

	usChecker := manager.NewUpstreamChecker(env.ManagerURL, env.PipelineHash, env.ProcessName, 5)
	usChecker.Start(&wg)
	usChecker.Register <- mqListener.CanExit

	wg.Wait()
}
