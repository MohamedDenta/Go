package filestoragemodule

import (
	"time"

	"../accountdb"
)

const maxUploadSize = 5 * 1024 * 1024 * 1024 // 5 GB
const uploadPath = "upload"

//Fileapi data from front , for front only
type Fileapi struct {
	Fileid   string
	Ownerpk  string
	FileName string
	FileType string
	FileHash string
	Timefile time.Time
	Signture string
}

// ExploreResponse explore response body
type ExploreResponse struct {
	OwnedFiles      []accountdb.FileList
	TotalSizeOwned  int64
	SharedFile      []accountdb.FileList
	TotalSizeShared int64
}

// ExploreBody expore request body
type ExploreBody struct {
	Publickey string
	Password  string
}

// RetrieveBody retrieve request body
type RetrieveBody struct {
	Publickey string
	Password  string
	FileID    string
	Time      string
	Signture  string
}

// ShareFiledata share file
type ShareFiledata struct {
	Publickey        string
	Password         string
	FileID           string
	PermissionPkList []string
	Signture         string
}

// FavouriteNodesData add/remove/explore favourite nodes
type FavouriteNodesData struct {
	Storage   int
	PublicKey string
	NodeIDs   []string
	Signture  string
}

// ExploreResp for explore favored nodes
type ExploreResp struct {
	Nodes []string
}
