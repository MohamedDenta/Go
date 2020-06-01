package filestorage

import (
	"encoding/json"

	errorpk "../errorpk"
	"../globalPkg"
	"github.com/syndtr/goleveldb/leveldb"
)

//Chunkdb chunk db
type Chunkdb struct {
	Chunkid     string
	Fileid      string
	Chunkhash   string
	ChunkNumber int
	Chunkdata   []byte
}

//DB name leveldb
var DB *leveldb.DB

//Open flag open db or not
var Open = false

// opendatabase create or open DB if exist
func opendatabase() bool {

	if !Open {
		Open = true
		DBpath := "Database/Chunkdb"
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

//AddChunk insert Chunkdb
func AddChunk(data Chunkdb) bool {
	opendatabase()
	d, convert := globalPkg.ConvetToByte(data, "AddChunk create Chunk package")
	if !convert {
		closedatabase()
		return false
	}
	err := DB.Put([]byte(data.Chunkid), d, nil)
	if err != nil {
		errorpk.AddError("AddChunk  Chunkdb package", "can't create Chunkdb", "DBError")
		return false
	}
	closedatabase()
	return true
}

// GetAllChunks get all Chunkdb
func GetAllChunks() (values []Chunkdb) {
	opendatabase()
	iter := DB.NewIterator(nil, nil)
	for iter.Next() {
		value := iter.Value()
		var newdata Chunkdb
		json.Unmarshal(value, &newdata)
		newdata.Chunkdata = nil
		values = append(values, newdata)
	}
	closedatabase()
	return values
}

// FindChunkByid select By ChunkID Chunkdb
func FindChunkByid(id string) (ChunkObj Chunkdb) {
	opendatabase()
	data, err := DB.Get([]byte(id), nil)
	if err != nil {
		errorpk.AddError("FindChunkByid  Chunkdb package", "can't get Chunkdb", "DBError")
	}

	json.Unmarshal(data, &ChunkObj)
	closedatabase()
	return ChunkObj
}

// FindChanksByFileId get by Token name
func FindChanksByFileId(fileID string) (values []Chunkdb) {
	opendatabase()
	iter := DB.NewIterator(nil, nil)
	for iter.Next() {

		value := iter.Value()

		var newdata Chunkdb
		json.Unmarshal(value, &newdata)
		if newdata.Fileid == fileID {
			values = append(values, newdata)
		}

	}
	closedatabase()
	return values
}

//DeleteChunk delete chunk by chunk id
func DeleteChunk(key string) (delete bool) {
	opendatabase()

	err := DB.Delete([]byte(key), nil)
	closedatabase()
	if err != nil {
		errorpk.AddError("ChunkDeleted Chunk package", "can't delete Chunk", "logic")
		return false
	}

	return true
}
