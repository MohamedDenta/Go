package errorPkgModule

import (
	"encoding/json"
	"net/http"

	"../admin"
	"../errorpk"
	"../globalPkg"
	"../logpkg"
)

type timee struct {
	From string
	To   string
}

type name struct {
	FunctionName string
}

// DeleteError delete error api
func DeleteError(w http.ResponseWriter, req *http.Request) {
	//log
	now, userIP := globalPkg.SetLogObj(req)
	logobj := logpkg.LogStruct{"_", now, userIP, "macAdress", "DeleteError", "error", "_", "_", "_", 0}
	
	var key timee
	decoder := json.NewDecoder(req.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&key)
	
	if err != nil {
		globalPkg.SendError(w, "please enter your correct request your req should contain  FunctionName : Funcname_Time")
		globalPkg.WriteLog(logobj, "please enter your correct request", "failed")
		return
	}
	if errorpk.ErrorDelete(key.From) {
		globalPkg.SendResponseMessage(w, "deleted successfully")
		globalPkg.WriteLog(logobj, "deleted successfully", "success")
	} else {
		globalPkg.SendError(w, "can't delete the errors")
		globalPkg.WriteLog(logobj, "can't delete the errors", "failed")
	}
}

//DeleteErrorsBetweenTwoTimes delete error between two times
func DeleteErrorsBetweenTwoTimes(w http.ResponseWriter, req *http.Request) {
	//log
	now, userIP := globalPkg.SetLogObj(req)
	logobj := logpkg.LogStruct{"_", now, userIP, "macAdress", "DeleteErrorsBetweenTwoTimes", "error", "_", "_", "_", 0}

	var key timee

	decoder := json.NewDecoder(req.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&key)

	if err != nil {
		globalPkg.SendError(w, "please enter your correct request your rq should contain 2 parameters and in order fo : From & To")
		globalPkg.WriteLog(logobj, "please enter your correct request", "failed")

		return
	}
	if errorpk.DeleteErrorsBetweenTowTimes(key.From, key.To) {
		globalPkg.SendResponseMessage(w, "deleted successfully")
		globalPkg.WriteLog(logobj, "deleted successfully", "success")
	} else {
		globalPkg.SendError(w, "can't delete the errors")
		globalPkg.WriteLog(logobj, "can't delete the errors", "failed")
	}
}

//GetAllErrorsAPI API to get all errors
func GetAllErrorsAPI(w http.ResponseWriter, req *http.Request) {
	//log
	now, userIP := globalPkg.SetLogObj(req)
	logobj := logpkg.LogStruct{"_", now, userIP, "macAdress", "GetAllErrorsAPI", "error", "_", "_", "_", 0}

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
		sendJSON, _ := json.Marshal(errorpk.GetAllErrors())
		globalPkg.SendResponse(w, sendJSON)
		globalPkg.WriteLog(logobj, "get all errors success", "success")
	} else {
		globalPkg.SendError(w, "you are not admin")
		globalPkg.WriteLog(logobj, "you are not admin", "failed")
	}
}

//GetFunctionErrors get all errors happend to a specific function
func GetFunctionErrors(w http.ResponseWriter, req *http.Request) {

	//log
	now, userIP := globalPkg.SetLogObj(req)
	logobj := logpkg.LogStruct{"_", now, userIP, "macAdress", "DeleteErrorsBetweenTwoTimes", "error", "_", "_", "_", 0}

	var funcName name

	decoder := json.NewDecoder(req.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&funcName)
	if err != nil {
		globalPkg.SendError(w, "please enter your correct request")
		globalPkg.WriteLog(logobj, "please enter your correct request", "failed")
		return
	}
	sendJSON, _ := json.Marshal(errorpk.GetErrorsByPrefix(funcName.FunctionName))
	globalPkg.SendResponse(w, sendJSON)
	globalPkg.WriteLog(logobj, "get all errors success", "success")
}
