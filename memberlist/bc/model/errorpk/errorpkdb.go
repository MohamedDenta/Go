package errorpk

import (
	"encoding/json"

	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/util"
)

//ErrorStruct is the error structure
type ErrorStruct struct {
	ErrorTime         string
	ErrorFunctionName string
	ErrorNotes        string
	ErrorType         string
}

//DB ...Path = "Database/errorPk" //Data base path
var DB *leveldb.DB

//Open is the status of db
var Open = false

// create or open db if exist
func opendatabase() bool {
	if !Open {
		Open = true
		DB_Path := "Database/errorPk"
		var err error
		DB, err = leveldb.OpenFile(DB_Path, nil)
		if err != nil {
			return false
		}
		return true
	}
	return true

}

// close db for opendatabase()
func closedatabase() bool {
	// var err error
	// err = db.Close()
	// if err != nil {
	// 	return false
	// }
	return true
}

// for errorStruct
func marshaltobson(Data ErrorStruct) (value []byte, convert bool) {
	var err error
	value, err = json.Marshal(Data)
	if err != nil {
		return value, false
	}
	return value, true
}

//create error in database
func errorCreate(data ErrorStruct) bool {
	opendatabase()
	var err error
	d, convert := marshaltobson(data)
	if !convert {
		closedatabase()
		return false
	}
	//key containing the func name and the time happened to it
	key := data.ErrorFunctionName + "_" + data.ErrorTime
	err = DB.Put([]byte(key), d, nil)
	closedatabase()
	if err != nil {
		return false
	}
	return true
}

func findErrorByKey(key string) bool {
	opendatabase()
	data, _ := DB.Get([]byte(key), nil)
	closedatabase()
	if data == nil {
		return false
	}
	return true //found error
}

//ErrorDelete delets an error
func ErrorDelete(key string) (delete bool) {
	opendatabase()
	err := DB.Delete([]byte(key), nil)
	if err != nil {
		return false
	}
	closedatabase()
	return true
}

// get the last error
func errorGetlast() (errorStructlst ErrorStruct) {
	opendatabase()
	var result ErrorStruct
	iter := DB.NewIterator(nil, nil)
	for iter.Last() {
		value := iter.Value()
		json.Unmarshal(value, &result)
		break
	}
	closedatabase()
	return result
}

//GetAllErrors get all stored errors
func GetAllErrors() (values []ErrorStruct) {
	opendatabase()

	iter := DB.NewIterator(nil, nil)
	for iter.Next() {

		value := iter.Value()
		var newdata ErrorStruct
		json.Unmarshal(value, &newdata)
		values = append(values, newdata)
	}
	closedatabase()
	return values
}

//GetErrorsBetweenTimes get all errors happend between specific 2 times
func GetErrorsBetweenTimes(from string, to string) (errorStructlst []ErrorStruct) {
	opendatabase()
	var lst []ErrorStruct
	iter := DB.NewIterator(&util.Range{Start: []byte(from), Limit: []byte(to)}, nil)
	for iter.Next() {
		value := iter.Value()
		var result ErrorStruct
		json.Unmarshal(value, &result)
		lst = append(lst, result)
	}
	iter.Last()
	last := iter.Value()
	iter = DB.NewIterator(nil, nil)
	nxt := iter.Value()
	for iter.Next() {
		if string(iter.Value()) == string(last) {
			iter.Next()
			nxt = iter.Value()
			break
		}
	}
	var result ErrorStruct
	json.Unmarshal(nxt, &result)
	lst = append(lst, result)
	closedatabase()
	return lst
}

//GetErrorsByPrefix get all errors that happend to a specific function
func GetErrorsByPrefix(prefix string) (values []ErrorStruct) {
	opendatabase()
	iter := DB.NewIterator(util.BytesPrefix([]byte(prefix)), nil)
	for iter.Next() {
		value := iter.Value()
		var newdata ErrorStruct
		json.Unmarshal(value, &newdata)
		values = append(values, newdata)
	}
	closedatabase()
	return values
}
