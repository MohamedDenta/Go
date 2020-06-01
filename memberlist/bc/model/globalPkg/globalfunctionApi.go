package globalPkg

import (
	"encoding/json" //read and send json data through api
	"fmt"
	"net/http" // using API request

	//"time"

	errorpk "../errorpk" //  write an error on the json file
	"../logpkg"
)

//PostGlobalVariableAPI post global variable
func PostGlobalVariableAPI(w http.ResponseWriter, req *http.Request) {
	//log
	now, userIP := SetLogObj(req)
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

// RequestServiceAPI request a service
func RequestServiceAPI(data Voucher, publickey string) (ResponseCreateVoucher, bool) {
	//fmt.Println("putchasedataaa", data)
	is_fine := true
	cre := ServCredentials{"Administrator", "adeelakramrox2"}
	cred_json, _ := json.Marshal(cre)
	vou_json, _ := json.Marshal(data)
	r, c := CreateVoucher(cred_json, vou_json, GlobalObj.ServiceLogin, GlobalObj.ServiceCreateVoutcher, "POST")
	if c != 200 {
		is_fine = false
	} else {
		is_fine = true
	}
	return r, is_fine
}

// GetVoucherDataAPI get service data
func GetVoucherDataAPI(create_time int64) (ResponseCreateVoucher, bool, bool) {
	is_fine := true
	var v ResponseCreateVoucher
	v.Create_time = create_time
	cre := ServCredentials{"Administrator", "adeelakramrox2"}
	jsonObj, _ := json.Marshal(cre)
	fmt.Println(jsonObj)
	r := serviceLogin(GlobalObj.ServiceLogin, "POST", jsonObj) // login
	if r != 200 {
		// can not login
		fmt.Println(r)
		is_fine = false
		return v, is_fine, true
	}
	v, code := getVoucherData(v)
	ok := true
	fmt.Println(v, code)
	switch code {
	case 200:
		is_fine = true
		ok = true
		break
	case 5:
		is_fine = true
		ok = false
		break
	case 4:
		is_fine = false
		ok = true
		break
	default:
		is_fine = true
		ok = false
		break
	}
	return v, is_fine, ok
}
