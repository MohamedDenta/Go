package block

import (
	"encoding/json"
	"net/http"

	//"time"

	"../admin"
	"../globalPkg"
	"../logpkg"
	"../responses"
	"../broadcastTcp"
)

// GetAllBlocksAPI endpoint to get all Blocks from the miner  -----------------*/
func GetAllBlocksAPI(w http.ResponseWriter, req *http.Request) {
	//log
	now, userIP := globalPkg.GetIP(req)
	logobj := logpkg.LogStruct{"_", now, userIP, "macAdress", "GetAllBlocksAPI", "Block", "_", "_", "_", 0}
	if !admin.AdminAPIDecoderAndValidation(w, req.Body, logobj) {
		return
	}
	sendJSON, _ := json.Marshal(GetBlockchain())
	globalPkg.SendResponse(w, sendJSON)
	globalPkg.WriteLog(logobj, "get all blocks success", "success")
}

//GetBlockByIDAPI endpoint to get specific Block using block id from the miner
func GetBlockByIDAPI(w http.ResponseWriter, req *http.Request) {
	//log
	now, userIP := globalPkg.GetIP(req)
	logobj := logpkg.LogStruct{"_", now, userIP, "macAdress", "GetBlockByIDAPI", "block", "_", "_", "_", 0}

	id := globalPkg.JSONString{}
	decoder := json.NewDecoder(req.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&id)

	if err != nil {
		responseObj := responses.FindResponseByID("1")
		globalPkg.SendError(w, responseObj.EngResponse)
		globalPkg.WriteLog(logobj, responseObj.EngResponse, "failed")
		return
	}

	blockObj := GetBlockInfoByID(id.Name)
	if blockObj.BlockHash == "" {
		responseObj := responses.FindResponseByID("158")
		globalPkg.SendNotFound(w, responseObj.EngResponse)
		globalPkg.WriteLog(logobj, responseObj.EngResponse, "failed")

	} else {
		sendJSON, _ := json.Marshal(GetBlockInfoByID(id.Name))
		globalPkg.SendResponse(w, sendJSON)
		globalPkg.WriteLog(logobj, "get block by pk success", "success")
	}
}

//BoardcatHandleBlock handle block case
func BoardcatHandleBlock(tCPDataObj broadcastTcp.TCPData) {
	var blockObjec BlockStruct
	json.Unmarshal(tCPDataObj.Obj, &blockObjec)

	var blockObjec2 BlockStruct
	blockObjec2 = blockObjec
	blockObjec2.BlockTransactions = nil

	for _, obj := range blockObjec.BlockTransactions {
		blockObjec2.BlockTransactions = append(blockObjec2.BlockTransactions, obj)
	}

	(&blockObjec2).AddBlock(false)
}