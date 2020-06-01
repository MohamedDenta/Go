package accountdb

import (
	"encoding/json"
	"fmt"
	"os"

	errorpk "../errorpk"
	"../globalPkg"
)

type AccountOwnershipStruct struct {
	Ownership        []string // current state of node or user
	HistoryOwnership []string // past state of node or user
	AccountIndex     string
	Owner            bool
}

func OwnershipCreate(data AccountOwnershipStruct) bool {
	fmt.Println("OwnershipCreate ", data)
	_, dbobj := opendatabaseCandidate("Database/TempAccount/Ownership")
	defer dbobj.Close()
	d, convert := globalPkg.ConvetToByte(data, "OwnershipCreate")
	if !convert {
		return false
	}
	err := dbobj.Put([]byte(data.AccountIndex), d, nil)
	if err != nil {
		errorpk.AddError("OwnershipCreate  AccountDB package", "can't create OwnershipCreate", "runtime error")
		return false
	}
	fmt.Println("data.AccountIndex", data.AccountIndex)
	return true
}

func FindOwnershipByKey(key string) (OwnershipStructObj AccountOwnershipStruct, isFine bool) {
	isFine = true
	_, dbobj := opendatabaseCandidate("Database/TempAccount/Ownership")
	data, err := dbobj.Get([]byte(key), nil)
	dbobj.Close()
	if err == os.ErrNotExist {
		isFine = false
	}
	json.Unmarshal(data, &OwnershipStructObj)
	if len(OwnershipStructObj.Ownership) == 0 {
		isFine = false
	}
	return OwnershipStructObj, isFine
}

// FindAllOwnerAccounts find all owner that have `Owner` = true
func FindAllOwnerAccounts() (values []AccountOwnershipStruct, isFine bool) {
	isFine, dbobj := opendatabaseCandidate("Database/TempAccount/Ownership")
	iter := dbobj.NewIterator(nil, nil)
	for iter.Next() {
		value := iter.Value()
		var newdata AccountOwnershipStruct
		json.Unmarshal(value, &newdata)
		fmt.Println("?_? ", newdata.AccountIndex)
		if newdata.Owner {
			values = append(values, newdata)
			isFine = true
		}

	}
	dbobj.Close()
	return values, isFine
}

// // delete Ownership by account index
// //-------------------------------------------------------------------------------------------------------
// func DeleteOwnership(key string) (delete bool) {
// 	_, dbobj := opendatabaseCandidate("Database/TempAccount/Ownership")
// 	err := dbobj.Delete([]byte(key), nil)
// 	dbobj.Close()
// 	if err != nil {
// 		errorpk.AddError("  Ownership package", "can't delete Ownership", "runtime error")
// 		return false
// 	}

// 	return true
// }

//-------------------------------------------------------------------------------------------------------------
// get all Ownerships
//-------------------------------------------------------------------------------------------------------------
func GetAllOwnerships() (values []AccountOwnershipStruct) {
	_, dbobj := opendatabaseCandidate("Database/TempAccount/Ownership")
	iter := dbobj.NewIterator(nil, nil)
	for iter.Next() {
		value := iter.Value()
		var newdata AccountOwnershipStruct
		json.Unmarshal(value, &newdata)
		values = append(values, newdata)
	}
	dbobj.Close()
	return values
}

// // AddOwnershipStruct add ownership struct
// func AddOwnershipStruct(accountownershipObj AccountOwnershipStruct) string {
// 	if OwnershipCreate(accountownershipObj) {
// 		return ""
// 	}
// 	return errorpk.AddError("AddOwnershipstruct account package", "Check your path or object to Add OwnershipStruct", "logical error")
// }
