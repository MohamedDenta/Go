package accountdb

import (
	"encoding/json"

	"../errorpk"
	"../globalPkg"
)

//------------------------------------------------------------------------------------------------------------
//struct for object reated to AcountStruct object
//------------------------------------------------------------------------------------------------------------
type AccountPublicKeyStruct struct {
	AccountPublicKey string
	AccountIndex     string
}

//-------------------------------------------------------------------------------------------------------------
// insert AccountPublicKeystruct
//-------------------------------------------------------------------------------------------------------------
func accountPublicKeyCreate(data AccountPublicKeyStruct) bool {
	var err error
	open, dbobj := opendatabaseCandidate("Database/TempAccount/AccountPublicKeyStruct")
	if !open {
		errorpk.AddError("opendatabase AccountPublicKeystruct package", "can't open the database", "critical error")
		return false
	}
	d, convert := globalPkg.ConvetToByte(data, "accountPublicKeyCreate account package")
	if !convert {
		return false
	}

	err = dbobj.Put([]byte(data.AccountPublicKey), d, nil)
	if err != nil {
		errorpk.AddError("AccountPublicKeyCreate  AccountPublicKeystruct package", "can't create AccountPublicKeystruct", "runtime error")
		return false
	}
	dbobj.Close()
	return true
}

// delete AccountPublicKeystruct
//-------------------------------------------------------------------------------------------------------------
func accountPublicKeyDelete(AccountPublicKey string) bool {
	var err error
	open, dbobj := opendatabaseCandidate("Database/TempAccount/AccountPublicKeyStruct")
	if !open {
		errorpk.AddError("opendatabase AccountPublicKeystruct package", "can't open the database", "critical error")
		return false
	}
	err = dbobj.Delete([]byte(AccountPublicKey), nil)
	if err != nil {
		errorpk.AddError("AccountPublicKeyDelete  AccountPublicKeystruct package", "can't delete AccountPublicKeystruct", "runtime error")
		return false
	}
	dbobj.Close()
	return true
}

//-------------------------------------------------------------------------------------------------------------
// select By PhoneNumber  AccountStruct
//-------------------------------------------------------------------------------------------------------------
func FindAccountByAccountPublicKey(AccountPublicKey string) (AccountStructObj AccountStruct) {

	data, err := findAccountPublicKeyByKey(AccountPublicKey)

	if err {
		AccountStructObj = FindAccountByAccountKey(data.AccountIndex)
	}
	return AccountStructObj
}

//-------------------------------------------------------------------------------------------------------------
// select By key AccountPublicKeyStruct
//-------------------------------------------------------------------------------------------------------------
func findAccountPublicKeyByKey(key string) (AccountPublicKeyStructObj AccountPublicKeyStruct, er bool) {
	var err error
	open, dbobj := opendatabaseCandidate("Database/TempAccount/AccountPublicKeyStruct")

	if !open {
		errorpk.AddError("opendatabase AccountPublicKeyStruct package", "can't open the database", "critical error")
		return AccountPublicKeyStruct{}, false
	}

	data, err := dbobj.Get([]byte(key), nil)
	if err != nil {
		// fmt.Println("errrrr from AccountPubKey struct??", err)
		errorpk.AddError("AccountPublicKeyStructFindByKey  AccountPublicKeyStruct package", "can't get AccountPublicKeyStruct", "runtime error")
	}
	json.Unmarshal(data, &AccountPublicKeyStructObj)
	dbobj.Close()
	return AccountPublicKeyStructObj, true

}
