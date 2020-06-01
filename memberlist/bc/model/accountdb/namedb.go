package accountdb

import (
	"encoding/json"

	"../errorpk"
	"../globalPkg"
)

//------------------------------------------------------------------------------------------------------------
//struct for object reated to AcountStruct object
//------------------------------------------------------------------------------------------------------------
type AccountNameStruct struct {
	AccountName  string
	AccountIndex string
}

//-------------------------------------------------------------------------------------------------------------
// insert AccountNameStruct
//-------------------------------------------------------------------------------------------------------------

func accountNameCreate(data AccountNameStruct) bool {
	var err error
	open, dbobj := opendatabaseCandidate("Database/TempAccount/AccountNameStruct")
	if !open {
		errorpk.AddError("opendatabase AccountNameStruct package", "can't open the database", "critical error")
		return false
	}
	d, convert := globalPkg.ConvetToByte(data, "accountNameCreate account package")
	if !convert {
		return false
	}
	err = dbobj.Put([]byte(data.AccountName), d, nil)
	if err != nil {
		errorpk.AddError("AccountNameStructCreate  AccountNameStruct package", "can't create AccountNameStruct", "runtime error")
		return false
	}
	dbobj.Close()
	return true
}

//
// delete AccountNameStruct
//-------------------------------------------------------------------------------------------------------------
func accountNameDelete(AccountName string) bool {
	var err error
	open, dbobj := opendatabaseCandidate("Database/TempAccount/AccountNameStruct")
	if !open {
		errorpk.AddError("opendatabase AccountNameStruct package", "can't open the database", "critical error")
		return false
	}
	err = dbobj.Delete([]byte(AccountName), nil)
	if err != nil {
		errorpk.AddError("AccountNameDelete  AccountNameStruct package", "can't delete AccountNameStruct", "runtime error")
		return false
	}
	dbobj.Close()
	return true
}

///-----------------------------------------------------------------------------

//-------------------------------------------------------------------------------------------------------------
// select By AccountName  AccountStruct
//-------------------------------------------------------------------------------------------------------------
func FindAccountByAccountName(AccountName string) (AccountStructObj AccountStruct) {
	data, err := findAccountNameByKey(AccountName)
	if err {
		AccountStructObj = FindAccountByAccountKey(data.AccountIndex)
	}
	return AccountStructObj

}

//-------------------------------------------------------------------------------------------------------------
// select By key AccountNameStruct
//-------------------------------------------------------------------------------------------------------------
func findAccountNameByKey(key string) (AccountNameStructObj AccountNameStruct, er bool) {
	var err error
	open, dbobj := opendatabaseCandidate("Database/TempAccount/AccountNameStruct")
	if !open {
		errorpk.AddError("opendatabase AccountNameStruct package", "can't open the database", "critical error")
		return AccountNameStruct{}, false
	}
	data, err := dbobj.Get([]byte(key), nil)
	if err != nil {
		errorpk.AddError("AccountNameStructFindByKey  AccountNameStruct package", "can't get AccountNameStruct", "runtime error")
	}
	json.Unmarshal(data, &AccountNameStructObj)
	dbobj.Close()
	return AccountNameStructObj, true
}

//-------------------------------------------------------------------------------------------------------------
// select By AccountName  AccountStruct
//-------------------------------------------------------------------------------------------------------------
func findAccountByAccountName(AccountName string) (AccountStructObj AccountStruct) {

	data, err := findAccountNameByKey(AccountName)
	if err {
		AccountStructObj = FindAccountByAccountKey(data.AccountIndex)
	}
	return AccountStructObj
}

/////////////////////////
func GetAllNames() (values []AccountNameStruct) {
	//var err error
	_, dbobj := opendatabaseCandidate("Database/TempAccount/AccountNameStruct")

	iter := dbobj.NewIterator(nil, nil)
	for iter.Next() {

		value := iter.Value()
		var newdata AccountNameStruct
		json.Unmarshal(value, &newdata)
		values = append(values, newdata)
	}
	dbobj.Close()

	return values
}
