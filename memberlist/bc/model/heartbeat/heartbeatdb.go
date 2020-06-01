package heartbeat

import (
	"encoding/json"

	errorpk "../errorpk" //  write an error on the json file
	globalPkg "../globalPkg"

	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/util"
)

type HeartBeatStruct struct {
	HeartBeatIp_Time  string
	HeartBeatStatus   bool
	HeartBeatworkLoad string
}

// var dbpath = "Database/HeartBeatStruct"
var DB *leveldb.DB
var Open = false

//------------------------------------------------------------------------------------------------------------
// create or open db if exist
//------------------------------------------------------------------------------------------------------------
func Opendatabase() bool {
	if !Open {
		Open = true
		dbpath := "Database/HeartBeatStruct"
		var err error
		DB, err = leveldb.OpenFile(dbpath, nil)
		if err != nil {
			errorpk.AddError("opendatabase HeartBeatStruct package", "can't open the database", "critical error")
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
func heartBeatStructCreate(data HeartBeatStruct) bool {
	Opendatabase()
	var err error
	d, convert := globalPkg.ConvetToByte(data, "heartBeatStructCreate  heartbeat package")
	if !convert {
		closedatabase()
		return false
	}
	err = DB.Put([]byte(data.HeartBeatIp_Time), d, nil)
	closedatabase()
	if err != nil {
		errorpk.AddError("HeartBeatStructCreate  HeartBeatStruct package", "can't create HeartBeatStruct", "runtime error")
		return false
	}
	return true
}

// //-------------------------------------------------------------------------------------------------------------
// // get all
// //-------------------------------------------------------------------------------------------------------------
func heartBeatStructGetAll() (values []HeartBeatStruct) {
	Opendatabase()
	iter := DB.NewIterator(nil, nil)
	for iter.Next() {

		value := iter.Value()
		var newdata HeartBeatStruct
		json.Unmarshal(value, &newdata)
		values = append(values, newdata)
	}
	closedatabase()
	return values
}

//-------------------------------------------------------------------------------------------------------------
// get last prefix key
//-------------------------------------------------------------------------------------------------------------
func heartBeatStructGetlastPrefix(prefix string) HeartBeatStruct {
	Opendatabase()
	var result HeartBeatStruct
	iter := DB.NewIterator(util.BytesPrefix([]byte(prefix)), nil)
	for iter.Last() {
		value := iter.Value()
		json.Unmarshal(value, &result)
		break
	}
	closedatabase()
	return result
}
