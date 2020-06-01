package accountdb

import (
	"encoding/json"
	"time"

	"../errorpk"
	"../globalPkg"
)

//------------------------------------------------------------------------------------------------------------
//struct for object reated to AcountStruct object
//------------------------------------------------------------------------------------------------------------
type AccountLastUpdatedTimestruct struct {
	AccountLastUpdatedTime time.Time
	AccountIndex           string
}

//-------------------------------------------------------------------------------------------------------------
// insert AccountLastUpdatedTimestruct
//-------------------------------------------------------------------------------------------------------------
func accountLastUpdatedTimeCreate(data AccountLastUpdatedTimestruct) bool {
	var err error
	open, dbobj := opendatabaseCandidate("Database/TempAccount/AccountLastUpdatedTimestruct")
	if !open {
		errorpk.AddError("opendatabase AccountLastUpdatedTimestruct package", "can't open the database", "critical error")
		return false
	}
	d, convert := globalPkg.ConvetToByte(data, "accountLastUpdateTimeCreate account package")
	if !convert {
		return false
	}
	err = dbobj.Put([]byte(data.AccountLastUpdatedTime.String()), d, nil)
	if err != nil {
		errorpk.AddError("AccountLastUpdatedTimestructCreate  AccountLastUpdatedTimestruct package", "can't create AccountLastUpdatedTimestruct", "runtime error")
		return false
	}
	dbobj.Close()
	return true
}

//-------------------------------------------------------------------------------------------------------------
// select By AccountLastUpdatedTime  AccountStruct
//-------------------------------------------------------------------------------------------------------------
func findAccountByAccountLastUpdatedTime(AccountLastUpdatedTime string) (AccountStructObj AccountStruct) {

	data, err := findAccountLastUpdatedTimeByKey(AccountLastUpdatedTime)
	if err {
		AccountStructObj = FindAccountByAccountKey(data.AccountIndex)
	}
	return AccountStructObj
}

//-------------------------------------------------------------------------------------------------------------
// select By key AccountLastUpdatedTimeStruct
//-------------------------------------------------------------------------------------------------------------
func findAccountLastUpdatedTimeByKey(key string) (AccountLastUpdatedTimestructObj AccountLastUpdatedTimestruct, er bool) {
	var err error
	open, dbobj := opendatabaseCandidate("Database/TempAccount/AccountLastUpdatedTimestruct")
	if !open {
		errorpk.AddError("opendatabase AccountLastUpdatedTimestruct package", "can't open the database", "critical error")
		return AccountLastUpdatedTimestruct{}, false
	}
	data, err := dbobj.Get([]byte(key), nil)
	if err != nil {
		errorpk.AddError("AccountLastUpdatedTimestructFindByKey  AccountLastUpdatedTimestruct package", "can't get AccountLastUpdatedTimestruct", "runtime error")
	}
	json.Unmarshal(data, &AccountLastUpdatedTimestructObj)
	dbobj.Close()
	return AccountLastUpdatedTimestructObj, true
}

//delete AccountLastUpdatedTimestruct
//-------------------------------------------------------------------------------------------------------------
func accountLastUpdatedTimeDelete(AccountLastUpdatedTime string) bool {
	var err error
	open, dbobj := opendatabaseCandidate("Database/TempAccount/AccountLastUpdatedTimestruct")
	if !open {
		errorpk.AddError("opendatabase AccountLastUpdatedTimestruct package", "can't open the database", "critical error")
		return false
	}
	err = dbobj.Delete([]byte(AccountLastUpdatedTime), nil)
	if err != nil {
		errorpk.AddError("AccountLastUpdatedTimestructDelete  AccountLastUpdatedTimestruct package", "can't delete AccountLastUpdatedTimestruct", "runtime error")
		return false
	}
	dbobj.Close()
	return true
}
