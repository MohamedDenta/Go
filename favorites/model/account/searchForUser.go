package account

import (
	"encoding/json"
	"net/http"

	"../accountdb"
	"../globalPkg"
	"../logpkg"
	"../responses"
)

//GetSearchProperty user can search for  Any account pk using userName or Email Or Phone
//response have userName and Public key
func GetSearchProperty(w http.ResponseWriter, req *http.Request) {
	now, userIP := globalPkg.GetIP(req)
	logobj := logpkg.LogStruct{"_", now, userIP, "macAdress", "GetSearchProperty", "AccountModule", "", "", "_", 0}

	user := User{}
	decoder := json.NewDecoder(req.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&user)

	if err != nil {
		responseObj := responses.FindResponseByID("1")
		globalPkg.SendError(w, responseObj.EngResponse)
		globalPkg.WriteLog(logobj, "failed to decode Object", "failed")
		return
	}

	//approve name is lowercase
	user.Account.AccountEmail = convertStringTolowerCaseAndtrimspace(user.Account.AccountEmail)
	user.Account.AccountName = convertStringTolowerCaseAndtrimspace(user.Account.AccountName)
	user.TextSearch = convertStringTolowerCaseAndtrimspace(user.TextSearch)
	var accountObj accountdb.AccountStruct
	if user.Account.AccountName != "" {
		accountObj = GetAccountByName(user.Account.AccountName)
	}

	if accountObj.AccountName == "" || accountObj.AccountPassword != user.Account.AccountPassword {
		responseObj := responses.FindResponseByID("11")
		globalPkg.SendNotFound(w, responseObj.EngResponse)
		globalPkg.WriteLog(logobj, "ckeck user name and password", "failed")
		return
	}

	PublicKey := getPublicKeyUsingString(user.TextSearch)
	if PublicKey == "" {
		responseObj := responses.FindResponseByID("20")
		globalPkg.SendNotFound(w, responseObj.EngResponse)
		globalPkg.WriteLog(logobj, "can not find user using this property", "failed")
		return
	}
	accountBypk := accountdb.FindAccountByAccountPublicKey(PublicKey)
	if !accountBypk.AccountStatus {
		responseObj := responses.FindResponseByID("21")
		globalPkg.SendError(w, responseObj.EngResponse)
		globalPkg.WriteLog(logobj, "this user is not active", "failed")
		return
	}
	var SR searchResponse
	SR.PublicKey = PublicKey
	SR.UserName = accountBypk.AccountName

	sendJSON, _ := json.Marshal(SR)
	globalPkg.SendResponse(w, sendJSON)
	globalPkg.WriteLog(logobj, string(sendJSON), "success")

}
