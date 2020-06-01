package transaction

import (
	"encoding/json"
	"net/http"
	"time"

	"../admin"
	"../logpkg"

	"../errorpk"
	"../globalPkg"

	"github.com/syndtr/goleveldb/leveldb"
)

//SignatureDB save signature in database
type SignatureDB struct {
	Signature     string
	TransactionID []map[string]time.Time
	// Timetransaction time.Time
}

var DBsign *leveldb.DB
var Opensign = false

//***************************************************************************************************************
// open database for TransactionDB with path Database/TransactionStruct
//***************************************************************************************************************

func opendatabasesign() bool {
	if !Opensign {
		Opensign = true
		DBpath := "Database/SignatureDB"
		var err error
		DBsign, err = leveldb.OpenFile(DBpath, nil)
		if err != nil {
			errorpk.AddError("opendatabase SignatureDB package", "can't open the database", "")
			return false
		}
		return true
	}
	return true

}

func closedatabaseSign() bool {
	// var err error
	// err = DB.Close()
	// if err != nil {
	// 	errorpk.AddError("closedatabase TransactionStruct package", "can't close the database")
	// 	return false
	// }
	return true
}

//SaveSignature insert SignatureDB
func SaveSignature(data SignatureDB) bool {

	opendatabasesign()
	d, convert := globalPkg.ConvetToByte(data, " save signature SignatureDB package")
	if !convert {
		closedatabaseSign()
		return false
	}
	err := DBsign.Put([]byte(data.Signature), d, nil)
	if err != nil {
		errorpk.AddError("save signature SignatureDB package", "can't create SignatureDB", "runtime error")
		return false
	}
	closedatabaseSign()
	return true
}

// FindsignatureBySender select By key
func FindsignatureBySender(id string) (SignObj SignatureDB) {
	opendatabasesign()
	data, err := DBsign.Get([]byte(id), nil)
	if err != nil {
		errorpk.AddError("FindsignatureByid save signature SignatureDB package", "can't get SignatureDB", "runtime error")
	}

	json.Unmarshal(data, &SignObj)
	closedatabaseSign()
	return SignObj
}

// FindsignatureBySender select By key
// func FindsignatureBySender(id string) (values []SignatureDB) {
// 	opendatabasesign()
// 	iter := DBsign.NewIterator(nil, nil)
// 	for iter.Next() {

// 		value := iter.Value()
// 		var newdata SignatureDB
// 		json.Unmarshal(value, &newdata)
// 		if id == newdata.Signature {
// 			values = append(values, newdata)
// 		}
// 	}
// 	closedatabaseSign()

// 	return values
// }

//GetAllSignaturesave get all save signture and sender address
func GetAllSignaturesave() (values []SignatureDB) {
	opendatabasesign()
	iter := DBsign.NewIterator(nil, nil)
	for iter.Next() {

		value := iter.Value()
		var newdata SignatureDB
		json.Unmarshal(value, &newdata)
		values = append(values, newdata)
	}
	closedatabaseSign()

	return values
}

//GetAllsignatureAPI get signture from database
func GetAllsignatureAPI(w http.ResponseWriter, req *http.Request) {
	//log
	now, userIP := globalPkg.SetLogObj(req)
	logobj := logpkg.LogStruct{"_", now, userIP, "macAdress", "GetAllsignatureAPI", "transaction", "_", "_", "_", 0}

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
		jsonObj, _ := json.Marshal(GetAllSignaturesave())
		globalPkg.SendResponse(w, jsonObj)
		globalPkg.WriteLog(logobj, "get all accounts", "success")
	} else {

		globalPkg.SendError(w, "you are not the admin ")
		globalPkg.WriteLog(logobj, "you are not the admin to get all accounts ", "failed")
	}
}
