package globalPkg

import (
	"encoding/json" //read and send json data through api
	"net/http"      // using API request

	//"time"

	errorpk "../errorpk" //  write an error on the json file
	"../logpkg"
)

//PostGlobalVariableAPI post global variable
func PostGlobalVariableAPI(w http.ResponseWriter, req *http.Request) {
	//log
	now, userIP :=GetIP(req)
	logobj := logpkg.LogStruct{"_", now, userIP, "macAdress", "PostGlobalVariableAPI", "globalPkg", "_", "_", "_", 0}

	globalObj := GlobalVariables{}
	errorStr := ""
	decoder := json.NewDecoder(req.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&globalObj)

	if err != nil {
		SendError(w, "please enter your correct request")
		WriteLog(logobj, "please enter your correct request", "failed")
		return
	}

	if Validation(globalObj) {
		GlobalObj = globalObj
	} else {
		errorStr = errorpk.AddError("PostGlobalVariable API globalfunction package "+req.Method, "Object is not valid", "hack error")
	}

	if errorStr == "" {
		sendJSON, _ := json.Marshal(globalObj)
		SendResponse(w, sendJSON)
		logobj.OutputData = "post global variables success"
		logobj.Process = "success"
		WriteLog(logobj, "post global variables success", "success")
	} else {
		SendError(w, errorStr)
		WriteLog(logobj, errorStr, "failed")
	}

}
