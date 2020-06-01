package account

import (
	"encoding/json"
	"fmt"

	"net/http"

	"../accountdb"
	"../errorpk"
	"../globalPkg"
	"../logpkg"
	"../responses"
)
// DelSession delete session
type DelSession struct {
	AccounIndex string
	SessionID   string
	Passord     string
}

//AddSessioninTemp func
func AddSessioninTemp(sessionID accountdb.AccountSessionStruct) {
	accountdb.AddSessionIdStruct(sessionID)
	AddSession(sessionID.AccountIndex, sessionID.SessionId)
}

//AddSession fun
func AddSession(accountIndex string, sessionID string) {
	accountobj := GetAccountByIndex(accountIndex)
	accountobj.AccountLastUpdatedTime = globalPkg.UTCtime()
	accountobj.SessionID = sessionID
	UpdateAccount(accountobj) ///new22
}
// UpdateSession update session
func UpdateSession(sessionobj accountdb.AccountSessionStruct) {
	accountobj := GetAccountByIndex(sessionobj.AccountIndex)
	accountobj.SessionID = sessionobj.SessionId
	accountobj.AccountLastUpdatedTime = globalPkg.UTCtime()
	UpdateAccount(accountobj) ///new22
}

//Getaccountsessionid fun
func Getaccountsessionid(publickey string) string {
	accountobj := GetAccountByAccountPubicKey(publickey)
	return accountobj.SessionID
}

//RemoveSessionFromtemp fun
func RemoveSessionFromtemp(sessionstruct accountdb.AccountSessionStruct) {
	var Delsession accountdb.AccountSessionStruct
	Delsession = accountdb.FindSessionByKey(sessionstruct.SessionId)
	accountdb.DeleteSession(sessionstruct.SessionId)
	deleteSessionID(Delsession)
}

//CheckIfsessionFound check if session found
func CheckIfsessionFound(sessionstruct accountdb.AccountSessionStruct) (bool, string) {
	acc := accountdb.FindSessionByKey(sessionstruct.SessionId)
	if acc.SessionId != "" && acc.AccountIndex != sessionstruct.AccountIndex {
		return true, acc.AccountIndex
	} 
		return false, ""	
}

//deleteSessionID fun delete session id 
func deleteSessionID(sessionobj accountdb.AccountSessionStruct) {
	accountobj := GetAccountByIndex(sessionobj.AccountIndex)
	fmt.Println("accountupdated1", accountobj)
	if accountobj.SessionID == sessionobj.SessionId {
		accountobj.SessionID = ""
		UpdateAccount2(accountobj)
		fmt.Println("accountupdated", accountobj)
	}
}

// DeleteSessionID endpoint to broadcast adding or deleting a transaction
func DeleteSessionID(w http.ResponseWriter, req *http.Request) {

	//log
	now, userIP := globalPkg.SetLogObj(req)
	logobj := logpkg.LogStruct{"_", now, userIP, "macAdress", "DeleteSessionID", "AccountModule", "_", "_", "_", 0}

	var delObj DelSession
	decoder := json.NewDecoder(req.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&delObj)
	errStr := ""
	if err != nil {
		errStr = errorpk.AddError("DeleteSessionID AccountModuleAPI  "+req.Method, "can't convert body to Transaction obj", "runtime error")
		responseObj := responses.FindResponseByID("1")
		globalPkg.SendError(w, responseObj.EngResponse)
		globalPkg.WriteLog(logobj, "please enter your correct request", "failed")
		return
	} else {
		account := GetAccountByIndex(delObj.AccounIndex)
		if account.AccountInitialPassword == delObj.Passord {
			//errStr still null and will call dellete seesion id in the next if condition
		} else {
			errStr = errorpk.AddError("DeleteSessionID AccountModuleAPI "+req.Method, "Wrong password and not authorized to delete this session!", "hack error")
		}
	}
	//var sessionobj SessionStruct
	var sessionobj accountdb.AccountSessionStruct
	if errStr == "" {
		sessionobj.SessionId = delObj.SessionID
		sessionobj.AccountIndex = delObj.AccounIndex
		deleteSessionID(sessionobj)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
	} else {
		// w.WriteHeader(http.StatusInternalServerError)
		// w.Write([]byte(errStr))

		globalPkg.SendError(w, errStr)
		globalPkg.WriteLog(logobj, errStr, "failed")
	}

}
