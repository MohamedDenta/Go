package privacyandterms

import (
	"encoding/json"
	"net/http"

	"../logpkg"

	"../account"
	"../admin"
	globalPkg "../globalPkg"
)

type getstr struct {
	ID       string
	UserPK   string
	UserPass string
}

//------------------------------------------------------------------------------------------------------------
// GetByIDAPI response with privacyandterms object
//------------------------------------------------------------------------------------------------------------
func GetByIDAPI(w http.ResponseWriter, req *http.Request) {

	id := getstr{}
	now, userIP := globalPkg.SetLogObj(req)
	logobj := logpkg.LogStruct{"_", now, userIP, "macAdress", "GetByIDAPI", "privacyandterms", "_", "_", "_", 0}

	decoder := json.NewDecoder(req.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&id)

	if err != nil {
		globalPkg.SendError(w, "  please enter your correct request")
		globalPkg.WriteLog(logobj, "Failed to decode object", "Failed")
		return
	}
	accdata := account.GetAccountByAccountPubicKey(id.UserPK)
	if accdata.AccountIndex != "" && accdata.AccountPassword == id.UserPass {
		PrivacyandtermsObj := findByid(id.ID)
		if PrivacyandtermsObj.ID == 0 && len(PrivacyandtermsObj.Items) == 0 {
			globalPkg.SendNotFound(w, "Can't find the PrivacyandtermsObj obj")
			globalPkg.WriteLog(logobj, "Can't find the PrivacyandtermsObj obj", "failed")
		} else {
			sendJson, _ := json.Marshal(PrivacyandtermsObj)
			globalPkg.SendResponse(w, sendJson)
			globalPkg.WriteLog(logobj, string(sendJson), "success")
		}
	} else {
		globalPkg.SendError(w, "you are not a user")
		globalPkg.WriteLog(logobj, "you are not a user", "failed")
	}

}

//------------------------------------------------------------------------------------------------------------
// GetAllAPI response with all privacyandterms
//------------------------------------------------------------------------------------------------------------
func GetAllAPI(w http.ResponseWriter, req *http.Request) {

	Adminobj := admin.Admin{}
	now, userIP := globalPkg.SetLogObj(req)
	logobj := logpkg.LogStruct{"_", now, userIP, "macAdress", "GetAllAPI", "privacyandterms", "_", "_", "_", 0}
	decoder := json.NewDecoder(req.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&Adminobj)

	if err != nil {
		globalPkg.SendError(w, "  please enter your correct request")
		globalPkg.WriteLog(logobj, "Failed to decode object", "Failed")
		return
	}
	if admin.ValidationAdmin(Adminobj) {
		sendJson, _ := json.Marshal(GetAll())
		globalPkg.SendResponse(w, sendJson)
	} else {
		globalPkg.SendError(w, "you are not admin")
	}
}

//------------------------------------------------------------------------------------------------------------
// GetAllAPI response with all privacyandterms
//------------------------------------------------------------------------------------------------------------
func AddAPI(w http.ResponseWriter, req *http.Request) {
	Privacyandtermsobj := Privacyandterms{}
	decoder := json.NewDecoder(req.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&Privacyandtermsobj)
	now, userIP := globalPkg.SetLogObj(req)
	logobj := logpkg.LogStruct{"_", now, userIP, "macAdress", "AddAPI", "privacyandterms", "_", "_", "_", 0}

	if err != nil {
		globalPkg.SendError(w, "  please enter your correct request")
		globalPkg.WriteLog(logobj, "Failed to decode object", "Failed")
		return
	}
	CreateORUpdate(Privacyandtermsobj)
	sendJson, _ := json.Marshal(Privacyandtermsobj)
	globalPkg.SendResponse(w, sendJson)
}

//------------------------------------------------------------------------------------------------------------
// GetAllAPI response with all privacyandterms
//------------------------------------------------------------------------------------------------------------
func UpdateAPI(w http.ResponseWriter, req *http.Request) {
	Privacyandtermsobj := Privacyandterms{}
	now, userIP := globalPkg.SetLogObj(req)
	logobj := logpkg.LogStruct{"_", now, userIP, "macAdress", "UpdateAPI", "privacyandterms", "_", "_", "_", 0}
	decoder := json.NewDecoder(req.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&Privacyandtermsobj)

	if err != nil {
		globalPkg.SendError(w, "  please enter your correct request")
		globalPkg.WriteLog(logobj, "Failed to decode object", "Failed")
		return
	}
	CreateORUpdate(Privacyandtermsobj)
	sendJson, _ := json.Marshal(Privacyandtermsobj)
	globalPkg.SendResponse(w, sendJson)
}
