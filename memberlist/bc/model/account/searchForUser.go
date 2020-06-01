package account

import (
	"encoding/json"
	"net/http"

	"../globalPkg"
	"../logpkg"
	"../accountdb"
)

//GetSearchProperty user can search for  Any account pk using userName or Email Or Phone
//response have userName and Public key
func GetSearchProperty(w http.ResponseWriter, req *http.Request) {
	now, userIP := globalPkg.SetLogObj(req)
	logobj := logpkg.LogStruct{"_", now, userIP, "macAdress", "GetSearchProperty", "AccountModule", "", "", "_", 0}

	user := User{}
	decoder := json.NewDecoder(req.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&user)

	if err != nil {
		globalPkg.SendError(w, "please enter your correct request ")
		globalPkg.WriteLog(logobj, "failed to decode Object", "failed")
		return
	}

	//approve name is lowercase
	user.Account.AccountEmail = convertStringTolowerCaseAndtrimspace(user.Account.AccountEmail)
	user.Account.AccountName = convertStringTolowerCaseAndtrimspace(user.Account.AccountName)
	user.TextSearch = convertStringTolowerCaseAndtrimspace(user.TextSearch)
	var accountObj accountdb.AccountStruct
	if user.Account.AccountName != "" {
		accountObj = getAccountByName(user.Account.AccountName)
	}

	if accountObj.AccountName == "" || accountObj.AccountPassword != user.Account.AccountPassword {
		globalPkg.SendNotFound(w, "Please,Check your account user name and password")
		globalPkg.WriteLog(logobj, "ckeck user name and password", "failed")
		return
	}

	PublicKey := getPublicKeyUsingString(user.TextSearch)
	if PublicKey == "" {
		globalPkg.SendNotFound(w, "I canot find user using this property, enter anthor---*")
		globalPkg.WriteLog(logobj, "can not find user using this property", "failed")
		return
	}
	accountBypk := accountdb.FindAccountByAccountPublicKey(PublicKey)
	if !accountBypk.AccountStatus {
		globalPkg.SendError(w, "This user is not active")
		globalPkg.WriteLog(logobj, "this user is not active", "failed")
		return
	}
	var SR searchResponse
	SR.PublicKey = PublicKey

	SR.UserName = accountBypk.AccountName

	sendJson, _ := json.Marshal(SR)

	globalPkg.SendResponse(w, sendJson)
	globalPkg.WriteLog(logobj, string(sendJson), "success")
	// sendResponse(w,PublicKey)
}
