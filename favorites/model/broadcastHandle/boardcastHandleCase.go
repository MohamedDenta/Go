package broadcastHandle

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"../ledger"
	"../logfunc"
	"../logpkg"
	"../validatorModule"

	ecc "../ECC"
	"../account"
	"../accountdb"
	"../globalPkg"
	"../globalfinctiontransaction"
	"../service"
	"../token"
	"../tokenModule"
	"../transaction"
	"../transactionModule"
	"../validator"

	"../admin"
	"../broadcastTcp"
	file "../filestorage"
)

//BoardcastHandleAdmin to handle admin case
func BoardcastHandleAdmin(tCPDataObj broadcastTcp.TCPData) {

	var AdminObj admin.AdminStruct
	if tCPDataObj.Method == "addadmin" {
		json.Unmarshal(tCPDataObj.Obj, &AdminObj)
		admin.CreateAdmin(AdminObj)
	} else if tCPDataObj.Method == "updateadmin" {
		json.Unmarshal(tCPDataObj.Obj, &AdminObj)
		admin.UpdateAdmindb(AdminObj)
	}
}

//BoardcastHandleToken to handle  token case
func BoardcastHandleToken(tCPDataObj broadcastTcp.TCPData) {

	var TokenObj token.StructToken
	if tCPDataObj.Method == "addtoken" {
		json.Unmarshal(tCPDataObj.Obj, &TokenObj)
		tokenModule.AddToken(TokenObj)
	} else if tCPDataObj.Method == "updatetoken" {
		json.Unmarshal(tCPDataObj.Obj, &TokenObj)
		token.UpdateTokendb(TokenObj)
	}
}

//BoardcastHandleValidator to handle validator case
func BoardcastHandleValidator(tCPDataObj broadcastTcp.TCPData) {
	if tCPDataObj.Method == "POST" {
		var timpValidator validator.TempValidator
		json.Unmarshal(tCPDataObj.Obj, &timpValidator)

		(&timpValidator).AddValidatorTemporary()
	} else if tCPDataObj.Method == "PUT" {
		var validatorObj validator.ValidatorStruct
		json.Unmarshal(tCPDataObj.Obj, &validatorObj)
		(&validatorObj).UpdateValidator()
	}
}

//BoardcastHandleConfirmValidator to handle confirm validator case
func BoardcastHandleConfirmValidator(tCPDataObj broadcastTcp.TCPData) {
	var validatorObj validator.ValidatorStruct
	json.Unmarshal(tCPDataObj.Obj, &validatorObj) //add the validator in validators list after admin confirmation
	(&validatorObj).RemoveFromTemp()

	validatorObj.ECCPublicKey = ecc.UnmarshalECCPublicKey(tCPDataObj.Obj)
	(&validatorObj).AddValidator()
}

//BoardcastHandleService to handle service case
func BoardcastHandleService(tCPDataObj broadcastTcp.TCPData) {
	if tCPDataObj.Method == "Tmp" {
		serviceobj := new(service.ServiceStruct)
		json.Unmarshal(tCPDataObj.Obj, serviceobj)

		// service.AddserviceInTmp(serviceobj)
		serviceobj.AddserviceInTmp()
	}
	if tCPDataObj.Method == "DB" {
		serviceobj := new(service.ServiceStruct)
		json.Unmarshal(tCPDataObj.Obj, serviceobj)

		// service.AddAndUpdateServiceObj(serviceobj)
		serviceobj.AddAndUpdateServiceObj()
		servicetemp := service.GetAllservice()
		for index, obj := range servicetemp {
			if serviceobj.PublicKey == obj.PublicKey && serviceobj.ID == obj.ID {
				service.RemoveServicefromTmp(index)
				break
			}

		}
	}
}

//BoardcastHandleFile to handle file case
func BoardcastHandleFile(tCPDataObj broadcastTcp.TCPData, w http.ResponseWriter) {
	if tCPDataObj.Method == "addchunk" {
		var chunkobj file.Chunkdb
		json.Unmarshal(tCPDataObj.Obj, &chunkobj)
		var responseDataChunk broadcastTcp.FileBroadcastResponse
		if file.AddChunk(chunkobj) {
			responseDataChunk.Valid = true
		}
		responseByteData, _ := json.Marshal(responseDataChunk)
		globalPkg.SendResponse(w, responseByteData)
		return
	} else if tCPDataObj.Method == "addfile" {
		var fileobj transaction.Transaction
		json.Unmarshal(tCPDataObj.Obj, &fileobj.Filestruct)
		fileobj.AddTransaction()
	} else if tCPDataObj.Method == "deletefile" {
		var fileobj transaction.Transaction
		json.Unmarshal(tCPDataObj.Obj, &fileobj.Filestruct)
		fileobj.AddTransaction()
	} else if tCPDataObj.Method == "getchunkdata" {
		var chnkObj file.Chunkdb
		json.Unmarshal(tCPDataObj.Obj, &chnkObj)
		retrievedObj := file.FindChunkByid(chnkObj.Chunkid)
		var responseDataChunk broadcastTcp.FileBroadcastResponse
		if retrievedObj.Chunkhash == chnkObj.Chunkhash {
			if len(retrievedObj.Chunkdata) != 0 {
				responseDataChunk.Valid = true
				responseDataChunk.ChunkData = retrievedObj.Chunkdata
				responseByteData, _ := json.Marshal(responseDataChunk)
				globalPkg.SendResponse(w, responseByteData)
				return
			}
		}
	} else if tCPDataObj.Method == "sharefile" {
		var sharefileobj file.SharedFile
		json.Unmarshal(tCPDataObj.Obj, &sharefileobj)
		file.AddSharedFile(sharefileobj)
	} else if tCPDataObj.Method == "updatesharefile" {
		var sharefileobj file.SharedFile
		json.Unmarshal(tCPDataObj.Obj, &sharefileobj)
		file.Updatesharefile(sharefileobj)
	} else if tCPDataObj.Method == "deleteaccountindex" {
		var sharefileobj file.SharedFile
		json.Unmarshal(tCPDataObj.Obj, &sharefileobj)
		file.DeleteSharedFile(sharefileobj.AccountIndex)
	} else if tCPDataObj.Method == "updateaccountFilelist" {
		var accountObj accountdb.AccountStruct
		json.Unmarshal(tCPDataObj.Obj, &accountObj)
		account.UpdateAccount2(accountObj)
	} else if tCPDataObj.Method == "updateaccountFavoredNodes" {
		accountObj := new(accountdb.AccountStruct)
		json.Unmarshal(tCPDataObj.Obj, accountObj)
		account.UpdateAccount2(*accountObj)
	}

}

//BoardcastHandleTransaction to handle transaction case
func BoardcastHandleTransaction(tCPDataObj broadcastTcp.TCPData, w http.ResponseWriter) {
	if tCPDataObj.Method == "addTransaction" {
		var txMix transactionModule.MixedTxStruct
		json.Unmarshal(tCPDataObj.Obj, &txMix)

		responseData := broadcastTcp.TxBroadcastResponse{
			TxID: txMix.TxObj.TransactionID, Valid: true,
		}
		//alaa
		temMap := make(map[string]string)
		txvalidator := txMix.TxObj.Validator
		transactionInputlst := txMix.TxObj.TransactionID
		InputIDArray := strings.Split(transactionInputlst, "_")
		temMap[txvalidator] = InputIDArray[1]
		globalfinctiontransaction.SetTransactionIndexTemMap(temMap)
		if txValid := transactionModule.ValidateTx2(txMix.DigitalTxObj, txMix.TxObj); txValid == "true" {

			txMix.TxObj.AddTransaction()
			for _, validatorObj := range validator.ValidatorsLstObj {

				if validatorObj.ValidatorIP == tCPDataObj.ValidatorIP {
					validatorObj.ValidatorStakeCoins = validatorObj.ValidatorStakeCoins + globalPkg.GlobalObj.TransactionStakeCoins
					(&validatorObj).UpdateValidator()
					break
				}
			}

		} else {
			globalfinctiontransaction.MapTempRollBack(txvalidator, InputIDArray[1])
			responseData.Valid = false
		}

		responseByteData, _ := json.Marshal(responseData)
		globalPkg.SendResponse(w, responseByteData)
		return

	} else if tCPDataObj.Method == "addTokenTransaction" {
		var tokenTx transaction.Transaction

		json.Unmarshal(tCPDataObj.Obj, &tokenTx)
		temMap := make(map[string]string)
		txvalidator := tokenTx.Validator
		transactionInputlst := tokenTx.TransactionID
		InputIDArray := strings.Split(transactionInputlst, "_")
		temMap[txvalidator] = InputIDArray[1]
		globalfinctiontransaction.SetTransactionIndexTemMap(temMap)
		tokenTx.AddTransaction()

	} else if tCPDataObj.Method == "missed-transaction-db" {
		var txDB transaction.TransactionDB
		json.Unmarshal(tCPDataObj.Obj, &txDB)
		txDB.AddTransactiondb()
	} else if tCPDataObj.Method == "missed-transaction" {
		var tx transaction.Transaction
		json.Unmarshal(tCPDataObj.Obj, &tx)
		temMap := make(map[string]string)
		txvalidator := tx.Validator
		transactionInputlst := tx.TransactionID
		InputIDArray := strings.Split(transactionInputlst, "_")
		temMap[txvalidator] = InputIDArray[1]
		globalfinctiontransaction.SetTransactionIndexTemMap(temMap)
		tx.AddTransaction()
		fmt.Println("Missed-Transaction-Added-Successfully")
	}

}

//BoardcastHandleLog to handle log case
func BoardcastHandleLog(tCPDataObj broadcastTcp.TCPData) {
	var logobj logpkg.LogStruct
	json.Unmarshal(tCPDataObj.Obj, &logobj)
	logfunc.WriteAndUpdateLog(logobj)
}

//BoardcastHandleLedgerfornewNode to handle  ledger for new node case
func BoardcastHandleLedgerfornewNode(tCPDataObj broadcastTcp.TCPData) {
	var ledgObj ledger.Ledger
	json.Unmarshal(tCPDataObj.Obj, &ledgObj)
	ledger.PostLedger(ledgObj)
}

//BoardcastHandleSendPKback to handle send pk back case
func BoardcastHandleSendPKback(tCPDataObj broadcastTcp.TCPData) {
	//in the new node and request to get the new public key recieved here
	var validatorObj validator.ValidatorStruct
	json.Unmarshal(tCPDataObj.Obj, &validatorObj)
	validatorObj.ECCPublicKey = ecc.UnmarshalECCPublicKey(tCPDataObj.Obj)
	// fmt.Println("validatorObj is : ", validatorObj)
	validatorModule.GetValidatorPublicKey(validatorObj, tCPDataObj.Method) //tCPDataObj.Method = minerIP

}

//BoardcastHandleAttachedPK to handle  attached pk case
func BoardcastHandleAttachedPK(tCPDataObj broadcastTcp.TCPData) {
	validatorObj := new(validator.ValidatorStruct)
	json.Unmarshal(tCPDataObj.Obj, validatorObj)
	validatorObj.ECCPublicKey = ecc.UnmarshalECCPublicKey(tCPDataObj.Obj)
	// fmt.Println("validatorObj is : ", validatorObj)
	M := make(map[string]string)
	M[validatorObj.ValidatorIP] = "00000"
	globalfinctiontransaction.SetTransactionIndexTemMap(M)
	validatorModule.ActiveValidator(validatorObj) //tCPDataObj.Method = minerIP

}
