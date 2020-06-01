package account

import (
	"encoding/json"
	"net/http"

	"../responses"

	"../accountdb"
	"../admin"
	"../globalPkg"
	"../logpkg"
)

// GetAllAccountsAPI endpoint to get all accounts from the miner
func GetAllAccountsAPI(w http.ResponseWriter, req *http.Request) {
	//log
	now, userIP := globalPkg.GetIP(req)
	logobj := logpkg.LogStruct{"_", now, userIP, "macAdress", "GetAllAccount", "Account", "_", "_", "_", 0}
	if !admin.AdminAPIDecoderAndValidation(w, req.Body, logobj) {
		return
	}
	jsonObj, _ := json.Marshal(accountdb.GetAllAccounts())
	globalPkg.SendResponse(w, jsonObj)
	globalPkg.WriteLog(logobj, "get all accounts", "success")
	return
}

// GetAccountInfoByAccountPublicKeyAPI endpoint to get specific account using public key from the miner
func GetAccountInfoByAccountPublicKeyAPI(w http.ResponseWriter, req *http.Request) {
	//log
	now, userIP := globalPkg.GetIP(req)
	logobj := logpkg.LogStruct{"_", now, userIP, "macAdress", "GetAccountInfoByAccountPublicKeyAPI", "Account", "_", "_", "_", 0}

	var accountPk globalPkg.JSONString
	decoder := json.NewDecoder(req.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&accountPk)

	if err != nil {
		responseObj := responses.FindResponseByID("1")
		globalPkg.SendError(w, responseObj.EngResponse)
		globalPkg.WriteLog(logobj, responseObj.EngResponse, "failed")
		return
	}
	AccountObj := accountdb.FindAccountByAccountPublicKey(accountPk.Name)
	if AccountObj.AccountPublicKey == "" {
		responseObj := responses.FindResponseByID("10")
		globalPkg.SendError(w, responseObj.EngResponse)
		globalPkg.WriteLog(logobj, responseObj.EngResponse+accountPk.Name+"\n", "failed")
		return
	}
		jsonObj, _ := json.Marshal(accountdb.FindAccountByAccountPublicKey(accountPk.Name))
		globalPkg.SendResponse(w, jsonObj)
		globalPkg.WriteLog(logobj, "find object by  this publickey"+accountPk.Name+"\n", "success")
	
}

//EmailuserStruct email and name
type EmailuserStruct struct {
	Name  string
	Email string
}

//GetAllEmailsUsernameAPI get all emails and names
func GetAllEmailsUsernameAPI(w http.ResponseWriter, req *http.Request) {

	//log
	now, userIP := globalPkg.GetIP(req)
	logobj := logpkg.LogStruct{"_", now, userIP, "macAdress", "GetAllEmailsUsernameAPI", "Account", "_", "_", "_", 0}
	if !admin.AdminAPIDecoderAndValidation(w, req.Body, logobj) {
		return
	}
	// var sendJSON []byte
	arrEmailsUsername := []EmailuserStruct{}
	accountobj := accountdb.GetAllAccounts()
	for _, account := range accountobj {
		emailsUsername := EmailuserStruct{account.AccountName, account.AccountEmail}
		arrEmailsUsername = append(arrEmailsUsername, emailsUsername)
	}

	jsonObj, _ := json.Marshal(arrEmailsUsername)
	globalPkg.SendResponse(w, jsonObj)
	globalPkg.WriteLog(logobj, "success to get all emails and username", "success")
	return
}

//GetnumberAccountsAPI get number of accounts
func GetnumberAccountsAPI(w http.ResponseWriter, req *http.Request) {
	//log
	now, userIP := globalPkg.GetIP(req)
	logobj := logpkg.LogStruct{"_", now, userIP, "macAdress", "GetnumberAccountsAPI", "Account", "_", "_", "_", 0}
	if !admin.AdminAPIDecoderAndValidation(w, req.Body, logobj) {
		return
	}
	data := map[string]interface{}{
		"Number_Of_Accounts": len(accountdb.GetAllAccounts()),
	}
	jsonObj, _ := json.Marshal(data)
	globalPkg.SendResponse(w, jsonObj)
	logobj.OutputData = "success to get number of accounts"
	logobj.Process = "success"
	globalPkg.WriteLog(logobj, "success to get number of accounts", "success")
}

//GetPkandValidatorPkUsingAddress End point create GetPublickeyUsingAddress
func GetPkandValidatorPkUsingAddress(w http.ResponseWriter, req *http.Request) {
	now, userIP := globalPkg.GetIP(req)
	logStruct := logpkg.LogStruct{"", now, userIP, "macAdress", "GetPublickeyUsingAddress", "Account", "", "", "", 0}

	accountObj := accountdb.AccountStruct{}

	decoder := json.NewDecoder(req.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&accountObj)
	if err != nil {
		responseObj := responses.FindResponseByID("1")
		globalPkg.SendError(w, responseObj.EngResponse)
		globalPkg.WriteLog(logStruct, responseObj.EngResponse, "failed")
		return
	}

	accountobj := GetAccountByAccountPubicKey(accountObj.AccountPublicKey)
	if accountObj.AccountPublicKey == "" {
		responseObj := responses.FindResponseByID("10")
		globalPkg.SendError(w, responseObj.EngResponse)
		globalPkg.WriteLog(logStruct, responseObj.EngResponse, "failed")
		return
	}
	if accountObj.AccountPassword != accountobj.AccountPassword {
		responseObj := responses.FindResponseByID("11")
		globalPkg.SendError(w, responseObj.EngResponse)
		globalPkg.WriteLog(logStruct, responseObj.EngResponse, "failed")
		return
	}

	if accountObj.AccountName != accountobj.AccountName {
		responseObj := responses.FindResponseByID("11")
		globalPkg.SendError(w, responseObj.EngResponse)
		globalPkg.WriteLog(logStruct, responseObj.EngResponse, "failed")
		return
	}

	pkValidator := globalPkg.RSAPublic
	data := map[string]interface{}{
		"validatorPublicKey ": pkValidator,
	}
	jsonObj, _ := json.Marshal(data)
	globalPkg.SendResponse(w, jsonObj)
	return
}
