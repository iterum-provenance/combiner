package downloader

import (
	"encoding/json"
	"io"

	desc "github.com/iterum-provenance/iterum-go/descriptors"
	"github.com/iterum-provenance/iterum-go/transmit"
)

// downloadedFragment stores a fragment as file handles instead of referencing files on disk
type downloadedFragment struct {
	FileHandles []io.ReadCloser
	Metadata    desc.RemoteMetadata
}

// Serialize tries to transform `dfrag` into a json encoded bytearray. Errors on failure
func (dfrag *downloadedFragment) Serialize() (data []byte, err error) {
	data, err = json.Marshal(dfrag)
	if err != nil {
		err = transmit.ErrSerialization(err)
	}
	return
}

// Deserialize tries to decode a json encoded byte array into `dfrag`. Errors on failure
func (dfrag *downloadedFragment) Deserialize(data []byte) (err error) {
	err = json.Unmarshal(data, dfrag)
	if err != nil {
		err = transmit.ErrSerialization(err)
	}
	return
}
