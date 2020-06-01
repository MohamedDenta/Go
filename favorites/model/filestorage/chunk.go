package filestorage

import (
	"time"
)

//FileStruct file store file meta data
type FileStruct struct {
	Fileid        string
	Ownerpk       string
	Mapping       map[string][]Chunkdata // map for chunk id and []server id , chunkhash
	FileName      string
	FileType      string
	FileSize      int64  // file size in mb
	FileHash      string //hash of all file
	Timefile      time.Time
	Timefiletx    time.Time
	Deleted       bool // check is file list deleted or not
	Transactionid string
	FileHash2     string
}

//Chunkdata  cont
type Chunkdata struct {
	ValidatorIP string
	Chunkhash   string
}
