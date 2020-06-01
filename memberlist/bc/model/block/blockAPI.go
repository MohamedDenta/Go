package block

import (
	"encoding/json"
	"net/http"

	//"time"

	"../admin"
	"../globalPkg"
	"../logpkg"
)

/*----------------- -----------------------API------------------------------------------------*/
/*----------------- endpoint to broadcast adding or deleting the Block through the network  -----------------*/
// func BlockBroadCastAPI(w http.ResponseWriter, req *http.Request) {

// 	w.Header().Set("Content-Type", "application/json")
// 	blockObj := BlockStruct{}
// 	err := json.NewDecoder(req.Body).Decode(&blockObj)
// 	errStr := ""
// 	if err != nil {
// 		errStr = errorpk.AddError("Broadcast Block API Block package"+req.Method, "can't convert body to Block obj")

// 	} else {
// 		for _, validatorObj := range validator.ValidatorsLstObj {

// 			url := validatorObj.ValidatorIP
// 			switch req.Method {
// 			case "POST":
// 				url = url + "/RegisterBlock"
// 			case "DELETE":
// 				url = url + "/DeleteBlock"
// 			default:
// 				{
// 					errStr = errorpk.AddError("Broadcast Validator API validator package"+req.Method, "wrong method")
// 				}
// 			}

// 			jsonObj, _ := json.Marshal(blockObj)
// 			errStr = errStr + globalPkg.SendRequest(jsonObj, url, req.Method)
// 		}
// 	}

// 	if errStr == "" {
// 		sendJson, _ := json.Marshal(blockObj)
// 		w.Header().Set("Content-Type", "application/json")
// 		w.WriteHeader(http.StatusOK)
// 		w.Write(sendJson)
// 	} else {
// 		w.WriteHeader(http.StatusOK)
// 		w.Write([]byte(errStr))
// 	}

// }

// /*----------------- endpoint to add or delete Block in the miner  -----------------*/
// func BlockAPI(w http.ResponseWriter, req *http.Request) {

// 	w.Header().Set("Content-Type", "application/json")
// 	blockObj := BlockStruct{}
// 	errorStr := ""
// 	err := json.NewDecoder(req.Body).Decode(&blockObj)
// 	if err != nil {
// 		errorStr = errorpk.AddError("Block API Block package"+req.Method, "Can't convert Body to Block obj ")
// 	} else {
// 		defer req.Body.Close()

// 		switch req.Method {
// 		case "POST":
// 			errorStr = AddBlock(blockObj)
// 		case "DELETE":
// 			errorStr = DeleteBlock(blockObj)
// 		default:
// 			errorStr = errorpk.AddError("Block API Block package"+req.Method, "wrong method ")

// 		}

// 		if errorStr != "" {
// 			w.WriteHeader(http.StatusOK)
// 			w.Write([]byte(errorStr))
// 		} else {
// 			sendJson, _ := json.Marshal(blockObj)
// 			w.Header().Set("Content-Type", "application/json")
// 			w.WriteHeader(http.StatusOK)
// 			w.Write(sendJson)
// 		}
// 	}
// }

// GetAllBlocksAPI endpoint to get all Blocks from the miner  -----------------*/
func GetAllBlocksAPI(w http.ResponseWriter, req *http.Request) {
	//log
	now, userIP := globalPkg.SetLogObj(req)
	logobj := logpkg.LogStruct{"_", now, userIP, "macAdress", "GetAllBlocksAPI", "Block", "_", "_", "_", 0}

	Adminobj := admin.Admin{}
	decoder := json.NewDecoder(req.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&Adminobj)

	if err != nil {
		globalPkg.SendError(w, "please enter your correct request")
		globalPkg.WriteLog(logobj, "please enter your correct request", "failed")
		return
	}

	if admin.ValidationAdmin(Adminobj) {
		sendJSON, _ := json.Marshal(GetBlockchain())
		globalPkg.SendResponse(w, sendJSON)
		globalPkg.WriteLog(logobj, "get all blocks success", "success")
	} else {
		globalPkg.SendError(w, "you are not admin")
		globalPkg.WriteLog(logobj, "you are not admin", "failed")
	}
}

//GetBlockByIDAPI endpoint to get specific Block using block id from the miner  -----------------*/
func GetBlockByIDAPI(w http.ResponseWriter, req *http.Request) {
	//log
	now, userIP := globalPkg.SetLogObj(req)
	logobj := logpkg.LogStruct{"_", now, userIP, "macAdress", "GetBlockByIDAPI", "block", "_", "_", "_", 0}

	id := globalPkg.JSONString{}
	decoder := json.NewDecoder(req.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&id)

	if err != nil {
		globalPkg.SendError(w, "please enter your correct request")
		globalPkg.WriteLog(logobj, "please enter your correct request", "failed")
		return
	}

	blockObj := GetBlockInfoByID(id.Name)
	if blockObj.BlockHash == "" {
		globalPkg.SendNotFound(w, "Can't find the block obj")
		globalPkg.WriteLog(logobj, "can't find the block obj", "failed")
	} else {
		sendJSON, _ := json.Marshal(GetBlockInfoByID(id.Name))
		globalPkg.SendResponse(w, sendJSON)
		globalPkg.WriteLog(logobj, "get block by pk success", "success")
	}
}

// for test validation block
// func AddBlockAPI(w http.ResponseWriter, req *http.Request) {
//  	blockObj := BlockStruct{}
// 	decoder := json.NewDecoder(req.Body)
// 	decoder.DisallowUnknownFields()
// 	err := decoder.Decode(&blockObj)
// 	if err != nil {
// 		globalPkg.SendError(w, "  please enter your correct request")
// 		return
// 	}
// errs := AddBlock(blockObj)
// if errs == ""{
// 	sendJSON, _ := json.Marshal("Good")
// 	globalPkg.SendResponse(w, sendJSON)
// }else{
// 	globalPkg.SendError(w,errs)
// }

// }
