package accountdb

import (
	"encoding/json"

	"../errorpk"
	"../globalPkg"
)

//------------------------------------------------------------------------------------------------------------
//struct for object reated to AcountStruct object
//------------------------------------------------------------------------------------------------------------
type AccountPhoneNumberStruct struct {
	AccountPhoneNumber string
	AccountIndex       string
}

//-------------------------------------------------------------------------------------------------------------
// insert AccountPhoneNumberStruct
//-------------------------------------------------------------------------------------------------------------
func accountPhoneNumberCreate(data AccountPhoneNumberStruct) bool {
	var err error
	open, dbobj := opendatabaseCandidate("Database/TempAccount/AccountPhoneNumberStruct")
	if !open {
		errorpk.AddError("opendatabase AccountPhoneNumberStruct package", "can't open the database", "critical error")
		return false
	}
	d, convert := globalPkg.ConvetToByte(data, "accountPhoneNumberCreate account package")
	if !convert {
		return false
	}
	err = dbobj.Put([]byte(data.AccountPhoneNumber), d, nil)
	if err != nil {
		errorpk.AddError("accountPhoneNumberCreate  AccountPhoneNumberStruct package", "can't create accountPhoneNumberStruct", "runtime error")
		return false
	}
	dbobj.Close()
	return true
}

//-------------------------------------------------------------------------------------------------------------
// select By key AccountPhoneNumberStruct
//-------------------------------------------------------------------------------------------------------------

func findAccountPhoneNumberByKey(key string) (AccountPhoneNumberStructObj AccountPhoneNumberStruct, er bool) {
	var err error
	open, dbobj := opendatabaseCandidate("Database/TempAccount/AccountPhoneNumberStruct")
	if !open {
		errorpk.AddError("opendatabase AccountPhoneNumberStruct package", "can't open the database", "critical error")
		return AccountPhoneNumberStruct{}, false
	}
	data, err := dbobj.Get([]byte(key), nil)
	if err != nil {
		errorpk.AddError("AccountPhoneNumberStructFindByKey  AccountPhoneNumberStruct package", "can't get AccountPhoneNumberStruct", "runtime error")
	}
	json.Unmarshal(data, &AccountPhoneNumberStructObj)
	dbobj.Close()
	return AccountPhoneNumberStructObj, true
}

//-------------------------------------------------------------------------------------------------------------
// select By key AccountPhoneNumberStruct
//-------------------------------------------------------------------------------------------------------------

//-------------------------------------------------------------------------------------------------------------
// select By PhoneNumber  AccountStruct
//-------------------------------------------------------------------------------------------------------------
func FindAccountByAccountPhoneNumber(AccountPhoneNumber string) (AccountStructObj AccountStruct) {

	data, err := findAccountPhoneNumberByKey(AccountPhoneNumber)
	if err {
		AccountStructObj = FindAccountByAccountKey(data.AccountIndex)
	}
	return AccountStructObj
}

// delete AccountPhoneNumberStruct
//-------------------------------------------------------------------------------------------------------------
func accountPhoneNumberDelete(AccountPhoneNumber string) bool {
	var err error
	open, dbobj := opendatabaseCandidate("Database/TempAccount/AccountPhoneNumberStruct")
	if !open {
		errorpk.AddError("opendatabase AccountPhoneNumberStruct package", "can't open the database", "critical error")
		return false
	}
	err = dbobj.Delete([]byte(AccountPhoneNumber), nil)
	if err != nil {
		errorpk.AddError("accountPhoneNumberDelete  AccountPhoneNumberStruct package", "can't delete accountPhoneNumberStruct", "runtime error")
		return false
	}
	dbobj.Close()
	return true
}

/////////////////////////////////////////////////
func GetAllPhones() (values []AccountPhoneNumberStruct) {
	//var err error
	_, dbobj := opendatabaseCandidate("Database/TempAccount/AccountPhoneNumberStruct")

	iter := dbobj.NewIterator(nil, nil)
	for iter.Next() {

		value := iter.Value()
		var newdata AccountPhoneNumberStruct
		json.Unmarshal(value, &newdata)
		values = append(values, newdata)
	}
	dbobj.Close()

	return values
}
