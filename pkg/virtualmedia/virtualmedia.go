package virtualmedia

import "time"

type Source string

const (
	SourceHTTP    Source = "HTTP"
	SourceStorage Source = "Storage"
)

type Mode string

const (
	ModeCDROM Mode = "CDROM"
	ModeDisk  Mode = "Disk"
)

type State struct {
	Source   Source `json:"source"`
	Mode     Mode   `json:"mode"`
	Filename string `json:"filename,omitempty"`
	URL      string `json:"url,omitempty"`
	Size     int64  `json:"size"`
}

type StorageFile struct {
	Filename  string    `json:"filename"`
	Size      int64     `json:"size"`
	CreatedAt time.Time `json:"createdAt"`
}

type StorageSpace struct {
	BytesUsed int64 `json:"bytesUsed"`
	BytesFree int64 `json:"bytesFree"`
}

type UploadStart struct {
	AlreadyUploadedBytes int64  `json:"alreadyUploadedBytes"`
	DataChannel          string `json:"dataChannel"`
}

type UploadProgress struct {
	Sent      int64
	Total     int64
	BytesPerS float64
}
