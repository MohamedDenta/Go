package broadcastTcp

import (
	"encoding/json"

	"net/http"

	"../admin"

	"time"

	errorpk "../errorpk"
	"../globalPkg"

	"github.com/syndtr/goleveldb/leveldb"
)

//ManageSendObject struct for object to be saved in db
type ManageSendObject struct {
	ObjID              string
	ValidatorSocket    string
	CurrentvalidatorIP string
	DataObj            TCPData
	ObjTime            time.Time
	TransactionID      string //incase package name transaction
}

//DB name leveldb
var DB *leveldb.DB

//Open flag open db or not
var Open = false

// opendatabase create or open DB if exist
func opendatabase() bool {

	if !Open {
		Open = true
		DBpath := "Database/ManageSendObject"
		var err error
		DB, err = leveldb.OpenFile(DBpath, nil)
		if err != nil {
			errorpk.AddError("opendatabase ManageSendObject package", "can't open the database", "DBError")
			return false
		}
		return true
	}
	return true

}

// close DB if exist
func closedatabase() bool {
	return true
}

//ManageSendObjectCreate insert ManageSendObject
func ManageSendObjectCreate(data ManageSendObject) bool {

	opendatabase()
	d, convert := globalPkg.ConvetToByte(data, "ManageSendObject create TCP package")
	if !convert {
		closedatabase()
		return false
	}

	err := DB.Put([]byte(data.ObjID), d, nil)

	if err != nil {
		errorpk.AddError("ManageSendObject  Tcp package", "can't create ManageSendObject", "DBError")
		return false
	}
	closedatabase()
	return true
}

// GetAllsendObject get all ManageSendObject
func GetAllsendObject() (values []ManageSendObject) {

	opendatabase()
	iter := DB.NewIterator(nil, nil)
	for iter.Next() {

		value := iter.Value()
		var newdata ManageSendObject
		json.Unmarshal(value, &newdata)
		values = append(values, newdata)
	}
	closedatabase()
	return values
}

func GetLastSendObject() ManageSendObject {
	opendatabase()

	iter := DB.NewIterator(nil, nil)
	iter.Last()
	value := iter.Value()
	var newdata ManageSendObject
	json.Unmarshal(value, &newdata)

	closedatabase()

	return newdata
}

func getlastsendObjectindex() string {

	manageObj := GetLastSendObject()
	if manageObj.ObjID == "" {
		return "-1"
	}

	return manageObj.ObjID
}

//GetAllsendObjectAPI get unsend object from db  for Test
func GetAllsendObjectAPI(w http.ResponseWriter, req *http.Request) {

	Adminobj := admin.Admin{}

	decoder := json.NewDecoder(req.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&Adminobj)

	if err != nil {
		globalPkg.SendError(w, "please enter your correct request ")

		return
	}

	if admin.ValidationAdmin(Adminobj) {
		jsonObj, _ := json.Marshal(GetAllsendObject())
		globalPkg.SendResponse(w, jsonObj)

	} else {

		globalPkg.SendError(w, "you are not the admin ")

	}
}
