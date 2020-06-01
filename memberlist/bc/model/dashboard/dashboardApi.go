package dashboard

import (
	"encoding/json"

	"net/http"
	"strconv"

	//"time"

	"../admin"
	block "../block"
	globalPkg "../globalPkg"
	"../logpkg"
	"../serverworkload"
	"../transaction"
	"../validator"
)

//GetBlockDashboard type Blocks []Dashboard
func GetBlockDashboard(w http.ResponseWriter, req *http.Request) {
	//log
	now, userIP := globalPkg.SetLogObj(req)
	logobj := logpkg.LogStruct{"_", now, userIP, "macAdress", "GetBlockDashboard", "dashboard", "_", "_", "_", 0}

	Adminobj := admin.Admin{}

	decoder := json.NewDecoder(req.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&Adminobj)
	
	if err != nil {
		globalPkg.SendError(w, "please enter your correct request")
		globalPkg.WriteLog(logobj, "please enter your correct request", "failed")

		return
	}

	// if Adminobj.AdminUsername == globalPkg.AdminObj.AdminUsername && Adminobj.AdminPassword == globalPkg.AdminObj.AdminPassword {
	if admin.ValidationAdmin(Adminobj) {
		AllBlocks := block.GetBlockchain()

		var blockInfoObj blockInfo
		var dashboardObj dashboard
		var lastBlock block.BlockStruct // using this variable because it is panic when getting the last index of the block lst

		// for i := 0; i < len(AllBlocks); i++ {
		for _, blockObj := range AllBlocks {
			blockInfoObj = blockInfo{
				BlockId:           globalPkg.ConvertFixedLengthStringtoInt(blockObj.BlockIndex),
				Blockcreationdate: blockObj.BlockTimeStamp,
				NoOfTransaction:   len(blockObj.BlockTransactions),
				GnodeID:           blockObj.ValidatorPublicKey,
			}
			lastBlock = blockObj
			dashboardObj.Blockinfo = append(dashboardObj.Blockinfo, blockInfoObj)
		}
		dashboardObj.Blockheight = globalPkg.ConvertFixedLengthStringtoInt(lastBlock.BlockIndex)
		dashboardObj.PayloadBase = serverworkload.GetCPUWorkloadPrecentage()
		sendJSON, _ := json.Marshal(dashboardObj)
		globalPkg.SendResponse(w, sendJSON)
		globalPkg.WriteLog(logobj, "get dashboard success", "success")
	} else {
		globalPkg.SendError(w, "you are not admin")
		globalPkg.WriteLog(logobj, "you are not admin", "failed")

	}
}

//GetStatistics get number of transaction and block
func GetStatistics(w http.ResponseWriter, req *http.Request) {
	//log
	now, userIP := globalPkg.SetLogObj(req)
	logobj := logpkg.LogStruct{"_", now, userIP, "macAdress", "GetStatistics", "dashboard", "_", "_", "_", 0}

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
		statObj := statStruct{}
		statObj.NumberOfBlocks = globalPkg.ConvertFixedLengthStringtoInt(block.GetLastBlock().BlockIndex) + 1
		statObj.NumberOfTransactions = len(transaction.GetPendingTransactions())
		statObj.NumberOfValidator = len(validator.ValidatorsLstObj)
		for _, obj := range validator.ValidatorsLstObj {
			statObj.NumberOfStakeCoin = append(statObj.NumberOfStakeCoin, obj.ValidatorIP+"_"+strconv.FormatFloat(obj.ValidatorStakeCoins, 'f', 6, 64))
		}

		sendJSON, _ := json.Marshal(statObj)
		globalPkg.SendResponse(w, sendJSON)
		globalPkg.WriteLog(logobj, "get number of blocks and transaction success", "success")
	} else {
		globalPkg.SendError(w, "you are not admin")
		globalPkg.WriteLog(logobj, "you are not admin", "failed")
	}
}
