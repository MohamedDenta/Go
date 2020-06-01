package validatorModule

import (
	"encoding/json" //read and send json data through api
	"net/http"      // using API request

	"../admin"
	"../broadcastTcp"
	"../errorpk"   //  write an error on the json file
	"../globalPkg" //to use send request function
	"../logpkg"
	"../responses"
	"../validator"
)

//BroadcastValidatorAPI endpoint to broadcasting adding, updating or deleting validator in the miner
func BroadcastValidatorAPI(w http.ResponseWriter, req *http.Request) {
	//log
	now, userIP := globalPkg.SetLogObj(req)
	logobj := logpkg.LogStruct{"", now, userIP, "macAdress", "BroadcastValidatorAPI", "Validator", "", "", "", 0}
	//read json body
	var parentObjec *MixedObjec
	err := parentObjec.Decode(req)
	if err != nil {
		responseObj := responses.FindResponseByID("1")
		globalPkg.SendError(w, responseObj.EngResponse)
		globalPkg.WriteLog(logobj, responseObj.EngResponse, "failed")
		return
	}

	// converting mixed obj to 2 object
	admin := parentObjec.Admn                     // admin object
	validator.NewValidatorObj = parentObjec.Vldtr // validator obj

	//check if authorized :
	if admin.UsernameAdmin != "inoadmin" && admin.PasswordAdmin != "a5601de47276914b0b2bc40e9555d826b382001897f9cf065cc147ab1a3b483b" {
		responseObj := responses.FindResponseByID("11")
		globalPkg.SendError(w, responseObj.EngResponse)
		globalPkg.WriteLog(logobj, responseObj.EngResponse, "failed")
		return
	}

	validator.NewValidatorObj.ValidatorRegisterTime = globalPkg.UTCtime()
	validator.NewValidatorObj.ValidatorLastHeartBeat = globalPkg.UTCtime()

	if req.Method == "PUT" {
		broadcastTcp.BoardcastingTCP(validator.NewValidatorObj, req.Method, "validator")
	} else {
		//create tempvalidator
		createvalidator(req)
	}

	responseObj := responses.FindResponseByID("108")
	globalPkg.SendResponseMessage(w, responseObj.EngResponse)
	globalPkg.WriteLog(logobj, responseObj.EngResponse, "success")
}
func createvalidator(req *http.Request) {
	now := globalPkg.UTCtime()
	confCode := validator.EncodeToString(4)
	tmpvalidator := validator.TempValidator{validator.NewValidatorObj, confCode, now}
	broadcastTcp.BoardcastingTCP(tmpvalidator, req.Method, "validator")
}

//ValidatorAPI endpoint to add, update or delete validator in the miner
func ValidatorAPI(w http.ResponseWriter, req *http.Request) {
	//log
	now, userIP := globalPkg.SetLogObj(req)
	logobj := logpkg.LogStruct{"_", now, userIP, "macAdress", "ValidatorAPI", "Validator", "_", "_", "_", 0}
	validatorObj := &validator.ValidatorStruct{}
	errorStr := ""
	err := validatorObj.DecodeBody(req)
	if err != nil {
		responseObj := responses.FindResponseByID("1")
		globalPkg.SendError(w, responseObj.EngResponse)
		globalPkg.WriteLog(logobj, responseObj.EngResponse, "failed")
		return
	}

	switch req.Method {
	case "POST":
		errorStr = (validatorObj).AddValidator()
	case "PUT":
		errorStr = (validatorObj).UpdateValidator()
	case "DELETE":
		errorStr = (validatorObj).DeleteValidator()
	default:
		errorStr = errorpk.AddError("Validator API validator package "+req.Method, "wrong method ", "logical error")

	}

	if errorStr == "" {
		sendJSON, _ := json.Marshal(validator.ValidatorsLstObj)
		globalPkg.SendResponse(w, sendJSON)
		globalPkg.WriteLog(logobj, "boardcast validator success to add or register validator", "success")

	} else {
		globalPkg.SendError(w, errorStr)
		globalPkg.WriteLog(logobj, errorStr, "failed")
	}
}

//GetAllValidatorAPI endpoint to get all validators from the miner
func GetAllValidatorAPI(w http.ResponseWriter, req *http.Request) {
	//log
	now, userIP := globalPkg.SetLogObj(req)
	logobj := logpkg.LogStruct{"_", now, userIP, "macAdress", "GetAllValidatorAPI", "ValidatorModule", "_", "_", "_", 0}

	Adminobj := admin.Admin{}

	decoder := json.NewDecoder(req.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&Adminobj)

	if err != nil {
		responseObj := responses.FindResponseByID("1")
		globalPkg.SendError(w, responseObj.EngResponse)
		globalPkg.WriteLog(logobj, responseObj.EngResponse, "failed")
		return
	}

	if admin.ValidationAdmin(Adminobj) {
		json.NewEncoder(w).Encode(validator.GetAllValidators())
		globalPkg.WriteLog(logobj, "get all validators success", "success")

	} else {
		responseObj := responses.FindResponseByID("2")
		globalPkg.SendError(w, responseObj.EngResponse)
		globalPkg.WriteLog(logobj, responseObj.EngResponse, "failed")

	}

}

// DeactiveNode admin can change or Update status of validator IP from validatorActive to disactive
func DeactiveNode(w http.ResponseWriter, req *http.Request) {

	now, userIP := globalPkg.SetLogObj(req)
	logobj := logpkg.LogStruct{"", now, userIP, "macAdress", "DeactiveNode", "adminModule", "", "", "", 0}

	AdminObj := admin.AdminStruct{}
	decoder := json.NewDecoder(req.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&AdminObj); err != nil {

		responseObj := responses.FindResponseByID("1")
		globalPkg.SendError(w, responseObj.EngResponse)
		globalPkg.WriteLog(logobj, responseObj.EngResponse, "failed")
		return
	}
	if AdminObj.AdminUsername == "" || AdminObj.AdminPassword == "" {
		responseObj := responses.FindResponseByID("11")
		globalPkg.SendError(w, responseObj.EngResponse)
		globalPkg.WriteLog(logobj, responseObj.EngResponse, "failed")
		return
	}

	Adminexist := admin.GetAdminsByUsername(AdminObj.AdminUsername)

	if AdminObj.AdminUsername == Adminexist.AdminUsername && AdminObj.AdminPassword == Adminexist.AdminPassword {
		listValidator := Adminexist.Validatorlst
		exist := false
		for _, validatorip := range listValidator {

			if validatorip == AdminObj.ValiatorIPtoDeactive {
				validatorObj := validator.FindValidatorByValidatorIP(validatorip)
				validatorObj.ValidatorActive = !validatorObj.ValidatorActive
				(&validatorObj).UpdateValidator()
				exist = true
			}
		}
		if exist == false {
			responseObj := responses.FindResponseByID("34")
			globalPkg.SendError(w, responseObj.EngResponse)
			globalPkg.WriteLog(logobj, responseObj.EngResponse, "failed")
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		globalPkg.WriteLog(logobj, "Update validator status", "success")
	} else {

		responseObj := responses.FindResponseByID("11")
		globalPkg.SendError(w, responseObj.EngResponse)
		globalPkg.WriteLog(logobj, responseObj.EngResponse, "failed")
	}
}

//ConfirmedValidatorAPI confirm validator
func ConfirmedValidatorAPI(w http.ResponseWriter, req *http.Request) {
	now, userIP := globalPkg.SetLogObj(req)
	logobj := logpkg.LogStruct{"_", now, userIP, "macAdress", "ConfirmedValidatorAPI", "validatorModule", "", "", "_", 0}

	var validValidator validator.TempValidator

	keys, ok := req.URL.Query()["confirmationcode"]

	if !ok || len(keys) == 0 {
		responseObj := responses.FindResponseByID("1")
		globalPkg.SendNotFound(w, responseObj.EngResponse)
		globalPkg.WriteLog(logobj, responseObj.EngResponse, "failed")
		return
	}

	//i is the index refer to location of the current confirmed validator in he tempvalidator array
	i := 0
	var flag bool
	for _, Valid := range validator.TempValidatorlst {
		if Valid.ConfirmationCode == keys[0] {
			validValidator = Valid
			flag = true
			break
		}
		i++
	}

	if flag != true {
		responseObj := responses.FindResponseByID("14")
		globalPkg.SendError(w, responseObj.EngResponse)
		globalPkg.WriteLog(logobj, responseObj.EngResponse, "failed")
		return
	}
	if now.Sub(validValidator.CurrentTime).Seconds() > globalPkg.GlobalObj.DeleteAccountTimeInseacond {
		responseObj := responses.FindResponseByID("8")
		globalPkg.SendError(w, responseObj.EngResponse)
		globalPkg.WriteLog(logobj, responseObj.EngResponse, "failed")
		return

	}
	broadcastTcp.SendObject(validValidator.ValidatorObjec, validator.CurrentValidator.ECCPublicKey, validator.CurrentValidator.ValidatorSoketIP, "Send public key back", validValidator.ValidatorObjec.ValidatorSoketIP)
	globalPkg.WriteLog(logobj, "sending success as response", "success")
	globalPkg.SendResponse(w, []byte("Validator addedd successfully"))
}
