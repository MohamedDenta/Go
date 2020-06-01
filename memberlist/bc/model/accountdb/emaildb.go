package accountdb

import (
	"encoding/json"

	"../errorpk"
	"../globalPkg"
)

//------------------------------------------------------------------------------------------------------------
//struct for object reated to AcountStruct object
//------------------------------------------------------------------------------------------------------------
type AccountEmailStruct struct {
	AccountEmail string
	AccountIndex string
}

//-------------------------------------------------------------------------------------------------------------
// insert AccountEmailStruct
//-------------------------------------------------------------------------------------------------------------
func accountEmailCreate(data AccountEmailStruct) bool {
	var err error
	open, dbobj := opendatabaseCandidate("Database/TempAccount/AccountEmailStruct")
	if !open {
		errorpk.AddError("opendatabase AccountEmailStruct package", "can't open the database", "critical error")
		return false
	}
	d, convert := globalPkg.ConvetToByte(data, "accountEmailCreate account package")
	if !convert {
		return false
	}
	err = dbobj.Put([]byte(data.AccountEmail), d, nil)
	if err != nil {
		errorpk.AddError("AccountEmailStructCreate  AccountEmailStruct package", "can't create AccountEmailStruct", "runtime error")
		return false
	}
	dbobj.Close()
	return true
}

//-------------------------------------------------------------------------------------------------------------
// select By AccountEmail  AccountStruct
//-------------------------------------------------------------------------------------------------------------
func FindAccountByAccountEmail(AccountEmail string) (AccountStructObj AccountStruct) {
	data, err := findAccountEmailByKey(AccountEmail)
	if err {
		AccountStructObj = FindAccountByAccountKey(data.AccountIndex)
	}
	return AccountStructObj

}

//-------------------------------------------------------------------------------------------------------------
// select By key AccountEmailStruct
//-------------------------------------------------------------------------------------------------------------
func findAccountEmailByKey(key string) (AccountEmailStructObj AccountEmailStruct, er bool) {
	var err error
	open, dbobj := opendatabaseCandidate("Database/TempAccount/AccountEmailStruct")
	if !open {
		errorpk.AddError("opendatabase AccountEmailStruct package", "can't open the database", "critical error")
		return AccountEmailStruct{}, false
	}
	data, err := dbobj.Get([]byte(key), nil)
	if err != nil {
		errorpk.AddError("AccountEmailStructFindByKey  AccountEmailStruct package", "can't get AccountEmailStruct", "runtime error")
	}
	json.Unmarshal(data, &AccountEmailStructObj)
	dbobj.Close()
	return AccountEmailStructObj, true
}

// delete AccountEmailStruct
//-------------------------------------------------------------------------------------------------------------
func accountEmailDelete(AccountEmail string) bool {
	var err error
	open, dbobj := opendatabaseCandidate("Database/TempAccount/AccountEmailStruct")
	if !open {
		errorpk.AddError("opendatabase AccountEmailStruct package", "can't open the database", "critical error")
		return false
	}
	err = dbobj.Delete([]byte(AccountEmail), nil)
	if err != nil {
		errorpk.AddError("AccountEmailStructDelete  AccountEmailStruct package", "can't delete AccountEmailStruct", "runtime error")
		return false
	}
	dbobj.Close()
	return true
}

// get all AccountStruct
//-------------------------------------------------------------------------------------------------------------
func GetAllEmails() (values []AccountEmailStruct) {
	//var err error
	_, dbobj := opendatabaseCandidate("Database/TempAccount/AccountEmailStruct")

	iter := dbobj.NewIterator(nil, nil)
	for iter.Next() {

		value := iter.Value()
		var newdata AccountEmailStruct
		json.Unmarshal(value, &newdata)
		values = append(values, newdata)
	}
	dbobj.Close()

	return values
}
