package account

import (
	"encoding/json"
	"fmt"
	"../accountdb"
	"../broadcastTcp"
	"../globalPkg"
)

//BoardcastHandleAccount to handle account case
func BoardcastHandleAccount(tCPDataObj broadcastTcp.TCPData) {
	var accountObjc accountdb.AccountStruct
	json.Unmarshal(tCPDataObj.Obj, &accountObjc)

	if tCPDataObj.Method == "POST" {
		AddAccount(accountObjc)
		lst := GetUserObjLst()
		for index, data := range lst {
			if data.Account.AccountName == accountObjc.AccountName {
				RemoveUserFromtemp(index)
				break
			}
		}
	} else if tCPDataObj.Method == "PUT" {

		UpdateAccount(accountObjc)
		lst := GetUserObjLst()
		for index, data := range lst {
			if data.Account.AccountName == accountObjc.AccountName {
				RemoveUserFromtemp(index)
				break
			}
		}
	} else if tCPDataObj.Method == "set public key" {
		SetPublicKey(accountObjc)
		savepkobj := RemovePKFromTemp(accountObjc.AccountPublicKey) // Denta
		savepkobj.CurrentTime = globalPkg.UTCtime()
		savepkobj.Index = accountObjc.AccountIndex
		SavePKAddress(savepkobj)
	} else if tCPDataObj.Method == "Resetpass" {
		UpdateAccount2(accountObjc)
		lst2 := GetResetPasswordData()
		for index, data := range lst2 {

			if data.Email == accountObjc.AccountEmail {
				RemoveResetpassFromtemp(index)
				break
			}
		}

	} else if tCPDataObj.Method == "update2" { //change status
		UpdateAccount2(accountObjc)
	}

}

//BoardcastHandleAccountModule to handle account module case
func BoardcastHandleAccountModule(tCPDataObj broadcastTcp.TCPData) {
	var accmodObjec ResetPasswordData
	var accmodObjecuser User
	if tCPDataObj.Method == "addRestPassword" {
		json.Unmarshal(tCPDataObj.Obj, &accmodObjec)
		fmt.Println("account module")
		//to not repeat reset password codes
		lst := GetResetPasswordData()
		add := true
		if len(lst) != 0 {
			for index, data := range lst {
				if data.Email == accmodObjec.Email {
					UpdateResetpassObjInTemp(index, accmodObjec)
					add = false
					break
				}
			}
		}
		if add == true {
			AddResetpassObjInTemp(accmodObjec)
		}
		//end update

	} else if tCPDataObj.Method == "adduser" {
		json.Unmarshal(tCPDataObj.Obj, &accmodObjecuser)
		fmt.Println("your object : ", accmodObjecuser)
		AddUserIntemp(accmodObjecuser)
	} else if tCPDataObj.Method == "Update" {
		json.Unmarshal(tCPDataObj.Obj, &accmodObjecuser)
		UpdateconfirmAtribute(accmodObjecuser)
	}
}

//BoardcastHandleAddSession to handle add session case
func BoardcastHandleAddSession(tCPDataObj broadcastTcp.TCPData) {
	var sessionid accountdb.AccountSessionStruct
	json.Unmarshal(tCPDataObj.Obj, &sessionid)
	AddSessioninTemp(sessionid)
}

//BoardcastHandleDeleteSession to handle delete session case
func BoardcastHandleDeleteSession(tCPDataObj broadcastTcp.TCPData) {
	var sessionid accountdb.AccountSessionStruct
	json.Unmarshal(tCPDataObj.Obj, &sessionid)
	RemoveSessionFromtemp(sessionid)
}

//BoardcastHandleOwnership to handle ownership case
func BoardcastHandleOwnership(tCPDataObj broadcastTcp.TCPData) {

	var ownership accountdb.AccountOwnershipStruct
	json.Unmarshal(tCPDataObj.Obj, &ownership)
	accountdb.OwnershipCreate(ownership)
}

//BoardcastHandleSavePK to handle save pk case
func BoardcastHandleSavePK(tCPDataObj broadcastTcp.TCPData) {
	var savepkobj SavePKStruct
	json.Unmarshal(tCPDataObj.Obj, &savepkobj)
	// add in temp
	AddInTempPK(savepkobj)
}
