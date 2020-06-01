package broadcastTcp

var TempData []TCPData

//TCPData struct contain data about object,package name ,method
type TCPData struct {
	Obj         []byte
	ValidatorIP string
	Method      string
	PackageName string // CurrentTime        []string
	Signature   string
}
type NetStruct struct {
	Encryptedkey  string
	Encrypteddata string
}

type TxBroadcastResponse struct {
	TxID  string
	Valid bool
}
type FileBroadcastResponse struct {
	ChunkData []byte
	Valid     bool
}

const BUFFERSIZE = 1024
