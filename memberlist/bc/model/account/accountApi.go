package account

import (
	"../accountdb"
	"../admin"
	"../globalPkg"
	"../logpkg"
	"../validator"
	"encoding/json"
	"net/http"
	"strconv"
)

/*----------------- endpoint to get all accounts from the miner  -----------------*/
func GetAllAccountsAPI(w http.ResponseWriter, req *http.Request) {
	//log
	now, userIP := globalPkg.SetLogObj(req)
	logobj := logpkg.LogStruct{"_", now, userIP, "macAdress", "GetAllAccount", "Account", "_", "_", "_", 0}

	Adminobj := admin.Admin{}
	decoder := json.NewDecoder(req.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&Adminobj)

	if err != nil {
		globalPkg.SendError(w, "please enter your correct request ")
		globalPkg.WriteLog(logobj, "failed to decode admin object", "failed")
		return
	}
	// if Adminobj.AdminUsername == globalPkg.AdminObj.AdminUsername && Adminobj.AdminPassword == globalPkg.AdminObj.AdminPassword {
	if admin.ValidationAdmin(Adminobj) {
		jsonObj, _ := json.Marshal(accountdb.GetAllAccounts())
		globalPkg.SendResponse(w, jsonObj)
		globalPkg.WriteLog(logobj, "get all accounts", "success")
	} else {

		globalPkg.SendError(w, "you are not the admin ")
		globalPkg.WriteLog(logobj, "you are not the admin to get all accounts ", "failed")
	}
}

/*----------------- endpoint to get specific account using public key from the miner  -----------------*/
func GetAccountInfoByAccountPublicKeyAPI(w http.ResponseWriter, req *http.Request) {
	//log
	now, userIP := globalPkg.SetLogObj(req)
	logobj := logpkg.LogStruct{"_", now, userIP, "macAdress", "GetAccountInfoByAccountPublicKeyAPI", "Account", "_", "_", "_", 0}

	var AccountPublicKey string

	decoder := json.NewDecoder(req.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&AccountPublicKey)

	if err != nil {
		//http.Error(w, err.Error()+"  please enter your correct request", http.StatusBadRequest)
		globalPkg.SendError(w, "please enter your correct request ")
		globalPkg.WriteLog(logobj, "failed to decode admin object", "failed")
		return
	}
	AccountObj := accountdb.FindAccountByAccountPublicKey(AccountPublicKey)
	if AccountObj.AccountPublicKey == "" {
		// w.WriteHeader(http.StatusInternalServerError)
		// w.Write([]byte(errorpk.AddError("GetAccountInfoByAccountPublicKeyAPI", "Can't find the obj "+AccountPublicKey)))
		globalPkg.SendError(w, "Can't find the obj "+AccountPublicKey)
		globalPkg.WriteLog(logobj, "Can't find the obj by this publickey"+AccountPublicKey+"\n", "failed")
	} else {
		jsonObj, _ := json.Marshal(accountdb.FindAccountByAccountPublicKey(AccountPublicKey))
		globalPkg.SendResponse(w, jsonObj)
		globalPkg.WriteLog(logobj, "find object by  this publickey"+AccountPublicKey+"\n", "success")
	}
}

//EmailuserStruct email and name
type EmailuserStruct struct {
	Name  string
	Email string
}

//GetAllEmailsUsernameAPI get all emails and names
func GetAllEmailsUsernameAPI(w http.ResponseWriter, req *http.Request) {

	//log
	now, userIP := globalPkg.SetLogObj(req)
	logobj := logpkg.LogStruct{"_", now, userIP, "macAdress", "GetAllEmailsUsernameAPI", "Account", "_", "_", "_", 0}

	Adminobj := admin.Admin{}

	decoder := json.NewDecoder(req.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&Adminobj)

	if err != nil {
		globalPkg.SendError(w, "please enter your correct request ")
		globalPkg.WriteLog(logobj, "failed to decode admin object", "failed")
		return
	}

	// var sendJSON []byte
	arrEmailsUsername := []EmailuserStruct{}

	if admin.ValidationAdmin(Adminobj) {
		accountobj := accountdb.GetAllAccounts()
		for _, account := range accountobj {
			emailsUsername := EmailuserStruct{account.AccountName, account.AccountEmail}
			arrEmailsUsername = append(arrEmailsUsername, emailsUsername)
		}

		jsonObj, _ := json.Marshal(arrEmailsUsername)
		globalPkg.SendResponse(w, jsonObj)
		globalPkg.WriteLog(logobj, "success to get all emails and username", "success")
	} else {
		globalPkg.SendError(w, "you are not admin ")
		globalPkg.WriteLog(logobj, "you are not the admin to get all Emails and username ", "failed")
	}

}

//GetnumberAccountsAPI get number of accounts
func GetnumberAccountsAPI(w http.ResponseWriter, req *http.Request) {

	//log
	now, userIP := globalPkg.SetLogObj(req)
	logobj := logpkg.LogStruct{"_", now, userIP, "macAdress", "GetnumberAccountsAPI", "Account", "_", "_", "_", 0}

	Adminobj := admin.Admin{}

	decoder := json.NewDecoder(req.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&Adminobj)

	if err != nil {
		globalPkg.SendError(w, "please enter your correct request ")
		globalPkg.WriteLog(logobj, "failed to decode admin object", "failed")
		return
	}

	if admin.ValidationAdmin(Adminobj) {
		data := map[string]interface{}{
			"Number_Of_Accounts": len(accountdb.GetAllAccounts()),
		}
		jsonObj, _ := json.Marshal(data)
		globalPkg.SendResponse(w, jsonObj)
		logobj.OutputData = "success to get number of accounts"
		logobj.Process = "success"
		globalPkg.WriteLog(logobj, "success to get number of accounts", "success")
	} else {
		globalPkg.SendError(w, "you are not admin ")
		globalPkg.WriteLog(logobj, "you are not the admin to get all Emails and username ", "failed")

	}
}

//GetnumAccountsAPI get number of accounts
func GetnumAccountsAPI(w http.ResponseWriter, req *http.Request) {

	//log
	now, userIP := globalPkg.SetLogObj(req)
	logobj := logpkg.LogStruct{"_", now, userIP, "macAdress", "GetnumberAccountsAPI", "Account", "_", "_", "_", 0}

	// data := map[string]int{
	// 	"NumberOfAccounts": len(accountdb.GetAllAccounts()),
	// }
	// var responsedata globalPkg.StructData
	// for key, value := range data {
	// 	responsedata.Name = key
	// 	responsedata.Length = value
	// }
	// jsonObj, _ := json.Marshal(responsedata)

	globalPkg.SendResponseMessage(w, strconv.Itoa(len(accountdb.GetAllAccounts())))
	logobj.OutputData = "success to get number of accounts"
	logobj.Process = "success"
	globalPkg.WriteLog(logobj, "success to get number of accounts", "success")

}

//GetAddressbyNameAPI get address by name
func GetAddressbyNameAPI(w http.ResponseWriter, req *http.Request) {
	//log
	now, userIP := globalPkg.SetLogObj(req)
	logobj := logpkg.LogStruct{"_", now, userIP, "macAdress", "GetAddressbyNameAPI", "Account", "_", "_", "_", 0}

	var AccountName globalPkg.JSONString

	decoder := json.NewDecoder(req.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&AccountName)

	if err != nil {
		globalPkg.SendError(w, "please enter your correct request ")
		globalPkg.WriteLog(logobj, "failed to decode admin object", "failed")
		return
	}
	address := GetAccountByName(AccountName.Name).AccountPublicKey
	if address != "" {
		globalPkg.SendResponseMessage(w, address)
		logobj.OutputData = "get address by account name"
		logobj.Process = "success"
		globalPkg.WriteLog(logobj, "success to get address by account name", "success")
		return
	}
	globalPkg.SendError(w, "account Name not found")
	globalPkg.WriteLog(logobj, "account Name not found", "failed")
}

//GetPkandValidatorPkUsingAddress End point create GetPublickeyUsingAddress
func GetPkandValidatorPkUsingAddress(w http.ResponseWriter, req *http.Request) {
	now, userIP := globalPkg.SetLogObj(req)
	logStruct := logpkg.LogStruct{"", now, userIP, "macAdress", "GetPublickeyUsingAddress", "Account", "", "", "", 0}

	accountObj := accountdb.AccountStruct{}

	decoder := json.NewDecoder(req.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&accountObj)
	if err != nil {
		globalPkg.SendError(w, "please enter your correct request ")
		globalPkg.WriteLog(logStruct, "failed to decode Object", "failed")
		return
	}

	accountobj := GetAccountByAccountPubicKey(accountObj.AccountPublicKey)
	if accountObj.AccountPublicKey == "" {
		globalPkg.SendError(w, "invalid address ")
		globalPkg.WriteLog(logStruct, "invalid address", "failed")
		return
	}
	if accountObj.AccountPassword != accountobj.AccountPassword {
		globalPkg.SendError(w, "incorrect password ")
		globalPkg.WriteLog(logStruct, "incorrect password", "failed")
		return
	}

	if accountObj.AccountName != accountobj.AccountName {
		globalPkg.SendError(w, "incorrect name")
		globalPkg.WriteLog(logStruct, "incorrect name", "failed")
		return
	}

	//pk := FindpkByAddress(accountObj.AccountPublicKey)

	pkValidator := validator.CurrentValidator.ValidatorPublicKey

	data := map[string]interface{}{
		//"Public key for User":   pk.Publickey,
		"validatorPublicKey": pkValidator,
	}
	jsonObj, _ := json.Marshal(data)
	globalPkg.SendResponse(w, jsonObj)
	return
}
