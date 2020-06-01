package privacyandterms

import (
	"encoding/json"
	errorpk "../errorpk" //  write an error on the json file
	globalPkg "../globalPkg"
	"strconv"

	"github.com/syndtr/goleveldb/leveldb"
)

type Privacyandterms struct {
	ID    int
	Items []string
}

// var dbpath = "Database/HeartBeatStruct"
var DB *leveldb.DB
var Open = false

//------------------------------------------------------------------------------------------------------------
// create or open db if exist
//------------------------------------------------------------------------------------------------------------
func opendatabase() bool {
	if !Open {
		Open = true
		dbpath := "Database/Privacyandterms"
		var err error
		DB, err = leveldb.OpenFile(dbpath, nil)
		if err != nil {
			errorpk.AddError("opendatabase Privacyandterms package", "can't open the database", "critical error")
			return false
		}
		return true

	}
	return true

}

//------------------------------------------------------------------------------------------------------------
// close db for opendatabase()
//------------------------------------------------------------------------------------------------------------
func closedatabase() bool {
	// var err error
	// err = db.Close()
	// if err != nil {
	// 	return false
	// }
	return true
}

// //-------------------------------------------------------------------------------------------------------------
// // insert
// //-------------------------------------------------------------------------------------------------------------
func CreateORUpdate(data Privacyandterms) bool {
	opendatabase()
	var err error
	d, convert := globalPkg.ConvetToByte(data, "PrivacyandtermsCreate  Privacyandterms package")
	if !convert {
		closedatabase()
		return false
	}
	err = DB.Put([]byte(strconv.Itoa(data.ID)), d, nil)
	closedatabase()
	if err != nil {
		errorpk.AddError("PrivacyandtermsCreate  Privacyandterms package", "can't create Privacyandterms", "runtime error")
		return false
	}
	return true
}

// //-------------------------------------------------------------------------------------------------------------
// // get all
// //-------------------------------------------------------------------------------------------------------------
func GetAll() (values []Privacyandterms) {
	opendatabase()
	iter := DB.NewIterator(nil, nil)
	for iter.Next() {

		value := iter.Value()
		var newdata Privacyandterms
		json.Unmarshal(value, &newdata)
		values = append(values, newdata)
	}
	closedatabase()
	return values
}

//-------------------------------------------------------------------------------------------------------------
// select By key
//-------------------------------------------------------------------------------------------------------------
func findByid(id string) (Obj Privacyandterms) {

	opendatabase()

	data, err := DB.Get([]byte(id), nil)

	if err != nil {
		errorpk.AddError("PrivacyandtermsFindByKey  Privacyandterms package", "can't get Privacyandterms", "runtime error")
	}
	json.Unmarshal(data, &Obj)
	closedatabase()
	return Obj
}
