package block

import (
	"encoding/json"

	"time"

	errorpk "../errorpk" //  write an error on the json file
	globalPkg "../globalPkg"
	transaction "../transaction"

	"github.com/syndtr/goleveldb/leveldb"
)

//------------------------------------------------------------------------------------------------------------
//struct for object to be saved in db
//------------------------------------------------------------------------------------------------------------
type BlockStruct struct {
	BlockIndex        string
	BlockTransactions []transaction.Transaction

	BlockPreviousHash  string
	BlockHash          string
	BlockTimeStamp     time.Time
	ValidatorPublicKey string
	Deleted            bool // for delete file
}

//------------------------------------------------------------------------------------------------------------
//struct for object reated to TXInput object
//------------------------------------------------------------------------------------------------------------

// type TXInput struct {
// 	Txid      []byte
// 	Vout      int
// 	Signature []byte
// 	PubKey    []byte
// }

// //------------------------------------------------------------------------------------------------------------
// //struct for object reated to TXOutput object
// //------------------------------------------------------------------------------------------------------------
// type TXOutput struct {
// 	Value      int
// 	PubKeyHash []byte
// }

// //------------------------------------------------------------------------------------------------------------
// //struct for object reated to Transaction object
// //------------------------------------------------------------------------------------------------------------
// type Transaction struct {
// 	ID   []byte
// 	Vin  []TXInput
// 	Vout []TXOutput
// }

//var dbpath = "Database/BlockStruct"
var DB *leveldb.DB
var Open = false

//------------------------------------------------------------------------------------------------------------
// create or open DB if exist
//------------------------------------------------------------------------------------------------------------
func opendatabase() bool {
	if !Open {
		Open = true
		DBpath := "Database/BlockStruct"
		var err error
		DB, err = leveldb.OpenFile(DBpath, nil)
		if err != nil {
			errorpk.AddError("opendatabase BlockStruct package", "can't open the database", "critical error")
			return false
		}
		return true
	}
	return true

}

//------------------------------------------------------------------------------------------------------------
// close DB if exist
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

//-------------------------------------------------------------------------------------------------------------
// insert BlockStruct
//-------------------------------------------------------------------------------------------------------------
func blockCreate(data BlockStruct) bool {

	opendatabase()
	d, convert := globalPkg.ConvetToByte(data, "Block create block package")
	if !convert {
		closedatabase()
		return false
	}
	err := DB.Put([]byte(data.BlockIndex), d, nil)
	closedatabase()

	if err != nil {
		errorpk.AddError("BlockCreate  BlockStruct package", "can't create BlockStruct", "runtime error")
		return false
	}
	return true
}

//-------------------------------------------------------------------------------------------------------------
// select By key (BlockIndex) BlockStruct
//-------------------------------------------------------------------------------------------------------------
func findBlockByKey(key string) (BlockStructObj BlockStruct) {
	opendatabase()
	data, _ := DB.Get([]byte(key), nil)
	closedatabase()
	// if err != nil {
	// 	errorpk.AddError("BlockFinDByKey  BlockStruct package",   "can't get BlockStruct")
	// }

	json.Unmarshal(data, &BlockStructObj)

	return BlockStructObj
}

// func FindBlockByKey(key string) (BlockStructObj BlockStruct) {
// 	opendatabase()
// 	data, _ := DB.Get([]byte(key), nil)
// 	closedatabase()
// 	// if err != nil {
// 	// 	errorpk.AddError("BlockFinDByKey  BlockStruct package",   "can't get BlockStruct")
// 	// }

// 	json.Unmarshal(data, &BlockStructObj)

// 	return BlockStructObj
// }
//-------------------------------------------------------------------------------------------------------------
// get all AccountStruct
//-------------------------------------------------------------------------------------------------------------
func getAllBlocks() (values []BlockStruct) {

	opendatabase()
	iter := DB.NewIterator(nil, nil)
	for iter.Next() {

		value := iter.Value()
		var newdata BlockStruct
		json.Unmarshal(value, &newdata)
		values = append(values, newdata)
	}
	closedatabase()
	return values
}

//-------------------------------------------------------------------------------------------------------------
// get last key
//-------------------------------------------------------------------------------------------------------------
func getLastBlock() (result BlockStruct) {

	opendatabase()
	//var result BlockStruct
	iter := DB.NewIterator(nil, nil)
	for iter.Last() {
		value := iter.Value()
		json.Unmarshal(value, &result)
		break
	}
	closedatabase()
	return result
}

//-------------------------------------------------------------------------------------------------------------
// delete BlockStruct by key
//-------------------------------------------------------------------------------------------------------------
func deleteBlock(key string) (delete bool) {
	opendatabase()

	err := DB.Delete([]byte(key), nil)
	closedatabase()
	if err != nil {
		errorpk.AddError("BlockDelete  ErrorSBlockStruct package", "can't delete BlockStruct", "runtime error")
		return false
	}

	return true
}
