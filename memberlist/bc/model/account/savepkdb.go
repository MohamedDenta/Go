package account

import (
	"encoding/json"
	"net/http"

	"../admin"
	"../errorpk"
	"../globalPkg"
	"../logpkg"
	"github.com/syndtr/goleveldb/leveldb"
)

//SavePKStruct save pk and address
type SavePKStruct struct {
	Publickey string
	Address   string
}

//DB name leveldb
var DBsave *leveldb.DB

//Open flag open db or not
var Opensave = false

// opendatabase create or open DB if exist
func opendatabasesave() bool {
	if !Opensave {
		Opensave = true
		DBpath := "Database/SavePKStruct"
		var err error
		DBsave, err = leveldb.OpenFile(DBpath, nil)
		if err != nil {
			errorpk.AddError("opendatabase SavePKStruct package", "can't open the database", "critical error")
			return false
		}
		return true
	}
	return true

}

// close DB if exist
func closedatabasesave() bool {
	return true
}

//SavePKAddress insert SavePKStruct
func SavePKAddress(data SavePKStruct) bool {

	opendatabasesave()
	d, convert := globalPkg.ConvetToByte(data, " save pk account SavePKAddress package")
	if !convert {
		closedatabasesave()
		return false
	}
	err := DBsave.Put([]byte(data.Address), d, nil)
	if err != nil {
		errorpk.AddError("SavePKAddress  SavePKStruct package", "can't create SavePKStruct", "runtime error")
		return false
	}
	closedatabasesave()
	return true
}

// FindpkByAddress select By key
func FindpkByAddress(id string) (SavePKStructObj SavePKStruct) {
	opendatabasesave()
	data, err := DBsave.Get([]byte(id), nil)
	if err != nil {
		errorpk.AddError("FindAdminByid  SavePKStruct package", "can't get SavePKStruct", "runtime error")
	}

	json.Unmarshal(data, &SavePKStructObj)
	closedatabasesave()
	return SavePKStructObj
}

//GetAllsavepk get all save pk and address
func GetAllsavepksave() (values []SavePKStruct) {
	opendatabasesave()
	iter := DBsave.NewIterator(nil, nil)
	for iter.Next() {

		value := iter.Value()
		var newdata SavePKStruct
		json.Unmarshal(value, &newdata)
		values = append(values, newdata)
	}
	closedatabasesave()

	return values
}

func GetAllpkAPI(w http.ResponseWriter, req *http.Request) {
	//log
	now, userIP := globalPkg.SetLogObj(req)
	logobj := logpkg.LogStruct{"_", now, userIP, "macAdress", "GetAllAccount", "Account", "_", "_", "_", 0}

	Adminobj := admin.Admin{}

	decoder := json.NewDecoder(req.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&Adminobj)

	if err != nil {
		globalPkg.SendError(w, "please enter your correct request ")
		globalPkg.WriteLog(logobj, "failed to decode admin object", "failed")
		return
	}
	// if Adminobj.AdminUsername == globalPkg.AdminObj.AdminUsername && Adminobj.AdminPassword == globalPkg.AdminObj.AdminPassword {
	if admin.ValidationAdmin(Adminobj) {
		jsonObj, _ := json.Marshal(GetAllsavepksave())
		globalPkg.SendResponse(w, jsonObj)
		globalPkg.WriteLog(logobj, "get all accounts", "success")
	} else {

		globalPkg.SendError(w, "you are not the admin ")
		globalPkg.WriteLog(logobj, "you are not the admin to get all accounts ", "failed")
	}
}
