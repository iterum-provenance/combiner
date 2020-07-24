package downloader

import (
	"fmt"
	"io"
	"sync"
	"time"

	desc "github.com/iterum-provenance/iterum-go/descriptors"
	"github.com/iterum-provenance/iterum-go/minio"
	"github.com/prometheus/common/log"

	"github.com/iterum-provenance/iterum-go/transmit"
)

// Downloader is the structure that consumes RemoteFragmentDesc structures and downloads them
type Downloader struct {
	ToDownload chan transmit.Serializable // desc.RemoteFragmentDesc
	Completed  chan transmit.Serializable // downloadedFragment
	Minio      minio.Config
	fragments  int
}

// NewDownloader creates a new Downloader and initiates a client of the Minio service
func NewDownloader(toDownload, completed chan transmit.Serializable) Downloader {
	minio := minio.NewMinioConfigFromEnv() // defaults to an upload setup
	minio.TargetBucket = "INVALID"         // adjust such that the target output is unusable
	if err := minio.Connect(); err != nil {
		log.Fatal(err)
	}
	return Downloader{toDownload, completed, minio, 0}
}

func download(miniocfg minio.Config, descriptor desc.RemoteFileDesc) (io.ReadCloser, error) {
	fhandle, err := miniocfg.GetFileAsReader(descriptor, true)
	if err != nil {
		return nil, fmt.Errorf("Download failed due to: '%v'\n %s", err, fmt.Sprintf("Bucket: '%v', Name: '%v'", descriptor.Bucket, descriptor.Name))
	}
	return fhandle, nil
}

// StartBlocking enters an endless loop consuming RemoteFragmentDescs and downloading the associated data
func (downloader Downloader) StartBlocking() {
	var wg sync.WaitGroup
	maxWorkers := 10

	// Spawn workers that consume from the ToDownload channel
	// they download the files as file handles and send these on to the uploader
	for i := 0; i < maxWorkers; i++ {
		wg.Add(1)
		go func(miniocfg minio.Config, toDownload, onComplete chan transmit.Serializable) {
			defer wg.Done()
			for msg := range toDownload {
				rfd := *msg.(*desc.RemoteFragmentDesc)
				fhandles := []io.ReadCloser{}
				// download all associated files
				for _, file := range rfd.Files {
					fhandle, err := download(miniocfg, file)
					if err != nil {
						log.Errorln(err)
						break // Fail this fragment
					}
					fhandles = append(fhandles, fhandle)
				}
				downloader.fragments++
				dload := downloadedFragment{fhandles, rfd.Metadata}
				onComplete <- &dload
			}
		}(downloader.Minio, downloader.ToDownload, downloader.Completed)
	}
	wg.Wait()
	downloader.Stop()
}

// Start asychronously calls StartBlocking via a Goroutine
func (downloader Downloader) Start(wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		startTime := time.Now()
		downloader.StartBlocking()
		log.Infof("downloader ran for %v", time.Now().Sub(startTime))
	}()
}

// Stop finishes up and notifies the user of its progress
func (downloader Downloader) Stop() {
	log.Infof("Downloader finishing up, (tried to) download(ed) %v fragments", downloader.fragments)
	close(downloader.Completed)
}
