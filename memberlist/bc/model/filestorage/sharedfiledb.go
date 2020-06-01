package filestorage

import (
	"encoding/json"

	errorpk "../errorpk"
	"../globalPkg"
	"github.com/syndtr/goleveldb/leveldb"
)

//SharedFile share file
type SharedFile struct {
	AccountIndex   string
	OwnerSharefile []OwnersharedFile //take more file from users
}
type OwnersharedFile struct {
	OwnerPublicKey string
	Fileid         []string //share more than file
}

//DB name leveldb
var DBshare *leveldb.DB

//Open flag open db or not
var Openshare = false

// opendatabase create or open DB if exist
func opendatabaseshare() bool {

	if !Openshare {
		Openshare = true
		DBpathshare := "Database/SharedFile"
		var err error
		DBshare, err = leveldb.OpenFile(DBpathshare, nil)
		if err != nil {
			errorpk.AddError("opendatabase SharedFile package", "can't open the database", "DBError")
			return false
		}
		return true
	}
	return true

}

// close DB if exist
func closedatabaseshare() bool {
	return true
}

//AddSharedFile insert SharedFile
func AddSharedFile(data SharedFile) bool {
	opendatabaseshare()
	d, convert := globalPkg.ConvetToByte(data, "SharedFile create SharedFile package")
	if !convert {
		closedatabaseshare()
		return false
	}
	err := DBshare.Put([]byte(data.AccountIndex), d, nil)
	if err != nil {
		errorpk.AddError("AddSharedFile  SharedFile package", "can't create SharedFile", "DBError")
		return false
	}
	closedatabaseshare()
	return true
}

// GetAllSharedFile get all SharedFile
func GetAllSharedFile() (values []SharedFile) {
	opendatabaseshare()
	iter := DBshare.NewIterator(nil, nil)
	for iter.Next() {
		value := iter.Value()
		var newdata SharedFile
		json.Unmarshal(value, &newdata)
		values = append(values, newdata)
	}
	closedatabaseshare()
	return values
}

// FindSharedfileByAccountIndex select By account index
func FindSharedfileByAccountIndex(id string) (SharedfileObj SharedFile) {
	opendatabaseshare()
	data, err := DBshare.Get([]byte(id), nil)
	if err != nil {
		errorpk.AddError("FindSharedfileByAccountIndex  Sharedfile package", "can't Sharedfile Chunkdb", "DBError")
	}
	json.Unmarshal(data, &SharedfileObj)
	closedatabaseshare()
	return SharedfileObj
}

// FindSharedFileByownerpk get by owner pk
func FindSharedFileByownerpk(ownerpk string) (values []SharedFile) {
	opendatabaseshare()
	iter := DBshare.NewIterator(nil, nil)
	for iter.Next() {

		value := iter.Value()

		var newdata SharedFile
		json.Unmarshal(value, &newdata)

		for _, sharefileobj := range newdata.OwnerSharefile {
			if sharefileobj.OwnerPublicKey == ownerpk {
				values = append(values, newdata)
				break
			}
		}
	}
	closedatabaseshare()
	return values
}

//DeleteSharedFile delete shared file by SharedFile
func DeleteSharedFile(key string) (delete bool) {
	opendatabaseshare()

	err := DBshare.Delete([]byte(key), nil)
	closedatabaseshare()
	if err != nil {
		errorpk.AddError("DeleteSharedFile SharedFile package", "can't delete SharedFile", "logic")
		return false
	}

	return true
}

//Updatesharefile update share file by account index
func Updatesharefile(data SharedFile) bool {

	err := AddSharedFile(data)
	if err {
		return true
	}
	return false

}
