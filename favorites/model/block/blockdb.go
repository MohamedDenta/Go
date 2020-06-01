package block

import (
	"encoding/json"

	"../constants"

	"../cryptogrpghy"

	"fmt"
	"time"

	"../errorpk" //  write an error on the json file
	"../globalPkg"

	"github.com/syndtr/goleveldb/leveldb"
)

//opendatabase
func opendatabase() bool {
	if !Open {
		Open = true
		var err error
		DB, err = leveldb.OpenFile(constants.BLOCK_DB_PATH, nil)
		if err != nil {
			errorpk.AddError("opendatabase BlockStruct package", "can't open the database", "critical error")
			return false
		}
		return true
	}
	return true

}

//closedatabase close database
func closedatabase() bool {
	// var err error
	// err = DB.Close()
	// if err != nil {
	// 	errorpk.AddError("closedatabase AccountStruct package", "can't close the database")
	// 	return false
	// }
	return true
}

//blockCreate add block to data base
func blockCreate(data *BlockStruct) bool {
	opendatabase()
	encryptedblock := data.EncryptBlock()
	d, convert := globalPkg.ConvetToByte(encryptedblock, "Block create block package")
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

//findBlockByKey find block by index
func findBlockByKey(key string) (BlockStructObj BlockStruct) {
	opendatabase()
	data, _ := DB.Get([]byte(key), nil)
	closedatabase()
	var stringblock string
	if data != nil {
		err := json.Unmarshal(data, &stringblock)
		if err != nil {
			fmt.Println(err)
		}
		decblock := cryptogrpghy.KeyDecrypt(constants.ENCRYPT_KEY, stringblock)
		byteblock := []byte(decblock)
		tempblock := new(BlockStruct)
		BlockStructObj = tempblock.FromJSON(byteblock)

	}
	return BlockStructObj
}

//getAllBlocks get all blocks
func getAllBlocks() (values []BlockStruct) {
	opendatabase()
	iter := DB.NewIterator(nil, nil)
	for iter.Next() {
		value := iter.Value()
		var stringblock string
		err := json.Unmarshal(value, &stringblock)
		if err != nil {
			fmt.Println(err)
		}
		decblock := cryptogrpghy.KeyDecrypt(constants.ENCRYPT_KEY, stringblock)
		byteblock := []byte(decblock)
		var tempblock BlockStruct
		var newdata BlockStruct
		newdata = (&tempblock).FromJSON(byteblock[:])
		values = append(values, newdata)
	}
	closedatabase()
	return values
}

//getLastBlock get last block according to time stamp
func getLastBlock() (result BlockStruct) {

	opendatabase()
	iter := DB.NewIterator(nil, nil)
	for iter.Last() {
		value := iter.Value()

		var stringblock string
		err := json.Unmarshal(value, &stringblock)
		if err != nil {
			fmt.Println(err)
		}
		decblock := cryptogrpghy.KeyDecrypt(constants.ENCRYPT_KEY, stringblock)
		byteblock := []byte(decblock)
		var tempblock BlockStruct
		result = (&tempblock).FromJSON(byteblock[:])
		break
	}
	closedatabase()
	return result
}

//deleteBlock delete block
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

// GetBlocksByTimeRange get all blocks have been created in a period of time
func GetBlocksByTimeRange(dateTime time.Time) (blocks []BlockStruct) {
	opendatabase()
	iter := DB.NewIterator(nil, nil)
	for iter.Next() {
		value := iter.Value()
		var stringblock string
		err := json.Unmarshal(value, &stringblock)
		if err != nil {
			fmt.Println(err)
		}
		decblock := cryptogrpghy.KeyDecrypt(constants.ENCRYPT_KEY, stringblock)
		byteblock := []byte(decblock)
		var newdata BlockStruct
		var tempblock BlockStruct
		newdata = (&tempblock).FromJSON(byteblock[:])
		if newdata.BlockTimeStamp.After(dateTime) || dateTime.Equal(newdata.BlockTimeStamp) {
			blocks = append(blocks, newdata)
		}
	}
	iter.Release()
	closedatabase()
	return blocks
}
