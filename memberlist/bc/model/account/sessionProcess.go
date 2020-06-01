package account

import (
	"encoding/json"
	"fmt"

	"net/http"
	"time"

	"../errorpk"
	"../globalPkg"
	"../logpkg"
	"../accountdb"
)

type DelSession struct {
	AccounIndex string
	SessionID   string
	Passord     string
}

type SessionStruct struct {
	SessionId string
	Time      time.Time
	AccountID string
}

var sessionslst []SessionStruct

//AddSessioninTemp func
func AddSessioninTemp(sessionId accountdb.AccountSessionStruct) {
	accountdb.AddSessionIdStruct(sessionId)
	AddSession(sessionId.AccountIndex, sessionId.SessionId)
}

//AddSession fun
func AddSession(accountIndex string, sessionId string) {
	accountobj := GetAccountByIndex(accountIndex)
	accountobj.SessionID = sessionId
	UpdateAccount2(accountobj) ///new22
}

//Getaccountsessionid fun
func Getaccountsessionid(publickey string) string {
	accountobj := GetAccountByAccountPubicKey(publickey)
	return accountobj.SessionID
}

//UpdateUserTotemp for update and reset password ---->isra
//Deleted

//RemoveSessionFromtemp fun
func RemoveSessionFromtemp(sessionstruct accountdb.AccountSessionStruct) {
	var Delsession accountdb.AccountSessionStruct
	Delsession = accountdb.FindSessionByKey(sessionstruct.SessionId)
	accountdb.DeleteSession(sessionstruct.SessionId)
	deleteSessionId(Delsession)
}

//CheckIfsessionFound
func CheckIfsessionFound(sessionstruct accountdb.AccountSessionStruct) (bool, string) {
	acc := accountdb.FindSessionByKey(sessionstruct.SessionId)
	if acc.SessionId != "" && acc.AccountIndex != sessionstruct.AccountIndex {
		return true, acc.AccountIndex
	} else {
		return false, ""
	}
}

//deleteSessionId fun
func deleteSessionId(sessionobj accountdb.AccountSessionStruct) {
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

	// w.Header().Set("Content-Type", "application/json")
	var delObj DelSession
	// err := json.NewDecoder(req.Body).Decode(&delObj)

	decoder := json.NewDecoder(req.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&delObj)
	errStr := ""
	if err != nil {
		errStr = errorpk.AddError("DeleteSessionID AccountModuleAPI  "+req.Method, "can't convert body to Transaction obj", "runtime error")
		globalPkg.SendError(w, "please enter your correct request ")
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
		deleteSessionId(sessionobj)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
	} else {
		// w.WriteHeader(http.StatusInternalServerError)
		// w.Write([]byte(errStr))

		globalPkg.SendError(w, errStr)
		globalPkg.WriteLog(logobj, errStr, "failed")
	}

}
