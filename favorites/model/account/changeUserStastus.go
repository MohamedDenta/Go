package account

import (
	"encoding/json"
	"net/http"

	"../accountdb"
	"../broadcastTcp"
	"../globalPkg"
	"../logpkg"
	"../responses"
)

// deactivateInfo reason to deactivate account
type deactivateInfo struct {
	PublicKey          string
	DeactivationReason string
	UserName           string
	Password           string
}

//ChangeStatus End Point to make user change his statusfrom active  to deactive OR from deactive to active
func ChangeStatus(w http.ResponseWriter, req *http.Request) {
	//log
	now, userIP := globalPkg.GetIP(req)
	logobj := logpkg.LogStruct{"_", now, userIP, "macAdress", "ChangeStatus", "AccountModule", "_", "_", "_", 0}

	DeactivationData := deactivateInfo{}
	decoder := json.NewDecoder(req.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&DeactivationData)

	if err != nil {
		responseObj := responses.FindResponseByID("1")
		globalPkg.SendError(w, responseObj.EngResponse)
		globalPkg.WriteLog(logobj, "please enter your correct request ", "failed")
		return
	}

	//approve username is lowercase and trim
	DeactivationData.UserName = convertStringTolowerCaseAndtrimspace(DeactivationData.UserName)
	accountObjByPK := accountdb.FindAccountByAccountPublicKey(DeactivationData.PublicKey)
	if accountObjByPK.AccountPublicKey == "" {
		responseObj := responses.FindResponseByID("10")
		globalPkg.SendError(w, responseObj.EngResponse)
		globalPkg.WriteLog(logobj, "invalid public key", "failed")
		return
	}

	if DeactivationData.UserName != accountObjByPK.AccountName || DeactivationData.Password != accountObjByPK.AccountPassword {
		responseObj := responses.FindResponseByID("11")
		globalPkg.SendNotFound(w, responseObj.EngResponse)
		globalPkg.WriteLog(logobj, "invalid UserName or Passsword ", "failed")
		return
	}

	accountObjByPK.AccountStatus = !accountObjByPK.AccountStatus
	accountObjByPK.AccountDeactivatedReason = DeactivationData.DeactivationReason
	accountObjByPK.AccountLastUpdatedTime = now
	broadcastTcp.BoardcastingTCP(accountObjByPK, "update2", "account")
	sendJSON, _ := json.Marshal(accountObjByPK)
	globalPkg.SendResponse(w, sendJSON)
	logobj.OutputData = "update status successful "
	logobj.Process = "success"
	globalPkg.WriteLog(logobj, "update status successful", "success")

}
