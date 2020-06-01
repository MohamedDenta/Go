package account

import (
	"encoding/json"
	"net/http"

	"../broadcastTcp"
	"../globalPkg"
	"../logpkg"
	"../accountdb"
)

type deactivateInfo struct {
	PublicKey          string
	DeactivationReason string
	UserName           string
	Password           string
}

//ChangeStatus End Point to make user change his statusfrom active  to deactive OR from deactive to active
func ChangeStatus(w http.ResponseWriter, req *http.Request) {
	//log
	now, userIP := globalPkg.SetLogObj(req)
	logobj := logpkg.LogStruct{"_", now, userIP, "macAdress", "ChangeStatus", "AccountModule", "_", "_", "_", 0}

	DeactivationData := deactivateInfo{}
	decoder := json.NewDecoder(req.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&DeactivationData)

	if err != nil {
		globalPkg.SendError(w, "please enter your correct request ")
		globalPkg.WriteLog(logobj, "please enter your correct request ", "failed")
		return
	}

	//approve username is lowercase and trim
	DeactivationData.UserName = convertStringTolowerCaseAndtrimspace(DeactivationData.UserName)
	accountObjByPK := accountdb.FindAccountByAccountPublicKey(DeactivationData.PublicKey)
	if accountObjByPK.AccountPublicKey == "" {
		globalPkg.SendError(w, "Invalid PublicKey")
		globalPkg.WriteLog(logobj, "invalid public key", "failed")
		return
	}

	if DeactivationData.UserName != accountObjByPK.AccountName || DeactivationData.Password != accountObjByPK.AccountPassword {
		globalPkg.SendNotFound(w, "invalid UserName or Passsword ")
		globalPkg.WriteLog(logobj, "invalid UserName or Passsword ", "failed")
		return
	}

	accountObjByPK.AccountStatus = !accountObjByPK.AccountStatus
	accountObjByPK.AccountDeactivatedReason = DeactivationData.DeactivationReason

	broadcastTcp.BoardcastingTCP(accountObjByPK, "update2", "account")
	sendJson, _ := json.Marshal(accountObjByPK)
	globalPkg.SendResponse(w, sendJson)
	logobj.OutputData = "update status successful "
	logobj.Process = "success"
	globalPkg.WriteLog(logobj, "update status successful", "success")

}
