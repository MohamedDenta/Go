package accountdb

import (
	"encoding/json"

	errorpk "../errorpk"
	"../globalPkg"
)

//------------------------------------------------------------------------------------------------------------

//------------------------------------------------------------------------------------------------------------
//struct for object reated to AcountStruct object
//------------------------------------------------------------------------------------------------------------
type AccountSessionStruct struct {
	SessionId    string
	AccountIndex string
}

//-------------------------------------------------------------------------------------------------------------
// insert SessionStruct
//-------------------------------------------------------------------------------------------------------------
func SessionCreate(data AccountSessionStruct) bool {

	_, dbobj := opendatabaseCandidate("Database/TempAccount/Session")
	d, convert := globalPkg.ConvetToByte(data, "session create block package")
	if !convert {
		dbobj.Close()
		return false
	}
	err := dbobj.Put([]byte(data.SessionId), d, nil)
	dbobj.Close()

	if err != nil {
		errorpk.AddError("SessionCreate  BlockStruct package", "can't create SessionStruct", "runtime error")
		return false
	}
	return true
}

//-------------------------------------------------------------------------------------------------------------
// select By key (BlockIndex) BlockStruct
//-------------------------------------------------------------------------------------------------------------
func FindSessionByKey(key string) (SessionStructObj AccountSessionStruct) {
	_, dbobj := opendatabaseCandidate("Database/TempAccount/Session")
	data, _ := dbobj.Get([]byte(key), nil)
	dbobj.Close()
	// if err != nil {
	// 	errorpk.AddError("BlockFinDByKey  BlockStruct package",   "can't get BlockStruct")
	// }

	json.Unmarshal(data, &SessionStructObj)

	return SessionStructObj
}

//-------------------------------------------------------------------------------------------------------------

//-------------------------------------------------------------------------------------------------------------
// delete BlockStruct by key
//-------------------------------------------------------------------------------------------------------------
func DeleteSession(key string) (delete bool) {
	_, dbobj := opendatabaseCandidate("Database/TempAccount/Session")

	err := dbobj.Delete([]byte(key), nil)
	dbobj.Close()
	if err != nil {
		errorpk.AddError("  ErrorSessionStruct package", "can't delete SessionStruct", "runtime error")
		return false
	}

	return true
}

//-------------------------------------------------------------------------------------------------------------
// get all AccountStruct
//-------------------------------------------------------------------------------------------------------------
func GetAllSessions() (values []AccountSessionStruct) {
	_, dbobj := opendatabaseCandidate("Database/TempAccount/Session")
	iter := dbobj.NewIterator(nil, nil)
	for iter.Next() {

		value := iter.Value()
		var newdata AccountSessionStruct
		json.Unmarshal(value, &newdata)
		values = append(values, newdata)
	}
	dbobj.Close()
	return values
}
func AddSessionIdStruct(accountsessionObj AccountSessionStruct) string {
	if SessionCreate(accountsessionObj) {

		return ""
	} else {
		return errorpk.AddError("AddAccount account package", "Check your path or object to Add AccountStruct", "logical error")
	}

}
