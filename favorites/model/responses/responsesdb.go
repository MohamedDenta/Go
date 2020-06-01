package responses

import (
	"encoding/json"
	"../globalPkg"
	"../errorpk"
	"github.com/syndtr/goleveldb/leveldb"
)

// Response response structure
type Response struct {
	ID          string
	EngResponse string
	ArResponse  string
	Param       string
}

//DB name leveldb
var DB *leveldb.DB

//Open flag open db or not
var Open = false

// opendatabase create or open DB if exist
func opendatabase() bool {
	if !Open {
		Open = true
		DBpath := "Database/Responses"
		var err error
		DB, err = leveldb.OpenFile(DBpath, nil)
		if err != nil {
			errorpk.AddError("opendatabase Chunkdb package", "can't open the database", "DBError")
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

// AddResponse add response to db
func AddResponse(data Response) bool {
	opendatabase()
	d, convert := globalPkg.ConvetToByte(data, "AddResponse create Response package")
	if !convert {
		closedatabase()
		return false
	}
	err := DB.Put([]byte(data.ID), d, nil)
	if err != nil {
		errorpk.AddError("AddResponse  Responsedb package", "can't create Response", "DBError")
		return false
	}
	closedatabase()
	return true
}

// GetAllChunks get all Chunkdb
func GetAllResponses() (values []Response) {
	opendatabase()
	iter := DB.NewIterator(nil, nil)
	for iter.Next() {
		value := iter.Value()
		var newdata Response
		json.Unmarshal(value, &newdata)
		values = append(values, newdata)
	}
	closedatabase()
	return values
}

// FindResponseByID select By response id
func FindResponseByID(id string) (res Response) {
	opendatabase()
	data, err := DB.Get([]byte(id), nil)
	if err != nil {
		errorpk.AddError("FindChunkByid  Chunkdb package", "can't get Chunkdb", "DBError")
	}

	json.Unmarshal(data, &res)
	closedatabase()
	return res
}

// DeleteResponse delete chunk by chunk id
func DeleteResponse(key string) (delete bool) {
	opendatabase()
	err := DB.Delete([]byte(key), nil)
	closedatabase()
	if err != nil {
		errorpk.AddError("DeleteResponse Response package", "can't delete Responnse", "logic")
		return false
	}

	return true
}


