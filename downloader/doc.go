// Package downloader contains a goroutine that downloads files associated with MQ messages containing fragment descriptions.
// This package is currently unused, but targeted at replacing the iterum-provenance/sidecar dependency of this repo.
// This can be achieved by implementing an uploader that can handle a slice of io.ReadCloser (being the files), rather than
// requiring files to be stored on disk
package downloader
