package service

import (
	"encoding/json"

	"time"

	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/util"

	"../errorpk" //  write an error on the json file
	"../globalPkg"
)

type ServiceStruct struct {
	ID          string
	Mbytes      bool
	Duration    int
	Day         bool
	Bandwidth   int
	PublicKey   string
	Password    string
	VoutcherId  string
	Calculation float64
	Time        time.Time
	CreateTime  int64
	SeviceType  string
	Amount      int
	M           int
}

//var DBpath = "Database/AccountStruct"
var DB *leveldb.DB
var Open = false

//------------------------------------------------------------------------------------------------------------
// create or open db if exist
//------------------------------------------------------------------------------------------------------------
func Opendatabase() bool {
	if !Open {
		Open = true
		dbpath := "Database/Service"
		var err error
		DB, err = leveldb.OpenFile(dbpath, nil)
		if err != nil {

			errorpk.AddError("opendatabase ServiceStruct package", "can't open the database", "critical error")
			return false
		}
		return true

	}
	return true

}

//------------------------------------------------------------------------------------------------------------
// close db
//------------------------------------------------------------------------------------------------------------

func closedatabase() bool {
	// var err error
	// err = DB.Close()
	// if err != nil {
	// 	errorpk.AddError("closedatabase AccountStruct package", "can't close the database")
	// 	return false
	// }
	return true
}

func ServiceCreateOUpdate(data ServiceStruct) bool {
	Opendatabase()
	d, convert := globalPkg.ConvetToByte(data, "serviceCreate ServiceStruct package")
	if !convert {
		return false
	}
	err := DB.Put([]byte(data.ID), d, nil)
	if err != nil {
		errorpk.AddError("ServiceStructCreate  ServiceStruct package", "can't create ServiceStruct", "runtime error")
		return false
	}
	closedatabase()
	return true
}

func FindServiceById(key string) ServiceStruct {
	Opendatabase()
	var Obj ServiceStruct
	data, err := DB.Get([]byte(key), nil)

	if err != nil {
		errorpk.AddError("FindServiceById  ServiceStruct package", "can't get ServiceStruct", "runtime error")
	}
	json.Unmarshal(data, &Obj)
	closedatabase()
	return Obj
}

//-------------------------------------------------------------------------------------------------------------
// get last prefix key
//-------------------------------------------------------------------------------------------------------------
func ServiceStructGetlastPrefix(prefix string) ServiceStruct {
	Opendatabase()
	var result ServiceStruct
	iter := DB.NewIterator(util.BytesPrefix([]byte(prefix)), nil)
	for iter.Last() {
		value := iter.Value()
		json.Unmarshal(value, &result)
		break
	}
	closedatabase()
	return result
}

//-------------------------------------------------------------------------------------------------------------
// get prefix key
//-------------------------------------------------------------------------------------------------------------
func ServiceStructGetByPrefix(prefix string) (values []ServiceStruct) {
	Opendatabase()
	iter := DB.NewIterator(util.BytesPrefix([]byte(prefix)), nil)
	for iter.Next() {
		value := iter.Value()
		var newdata ServiceStruct
		json.Unmarshal(value, &newdata)
		values = append(values, newdata)
	}
	closedatabase()
	return values
}

// //-------------------------------------------------------------------------------------------------------------
// // get all
// //-------------------------------------------------------------------------------------------------------------
func ServiceStructGetAll() (values []ServiceStruct) {
	Opendatabase()
	iter := DB.NewIterator(nil, nil)
	for iter.Next() {

		value := iter.Value()
		var newdata ServiceStruct
		json.Unmarshal(value, &newdata)
		values = append(values, newdata)
	}
	closedatabase()
	return values
}

//--------------------------------------------------------------------------------------------------------------

func GetCreateTime(voutcherId string) int64 {
	Opendatabase()
	iter := DB.NewIterator(nil, nil)
	for iter.Next() {

		value := iter.Value()
		var newdata ServiceStruct
		json.Unmarshal(value, &newdata)
		if newdata.VoutcherId == voutcherId {
			return newdata.CreateTime
		}
	}
	return -1
}
