package transaction

import (
	"encoding/json"
	"fmt"
	"../accountdb"

	"../errorpk"
	"../globalPkg"

	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/util"
)

var DB *leveldb.DB
var Open = false

//***************************************************************************************************************
// open database for TransactionDB with path Database/TransactionStruct
//***************************************************************************************************************

func opendatabase() bool {
	if !Open {
		Open = true
		DBpath := "Database/TransactionStruct"
		var err error
		DB, err = leveldb.OpenFile(DBpath, nil)
		if err != nil {
			errorpk.AddError("opendatabase TransactionStruct package", "can't open the database", "")
			return false
		}
		return true
	}
	return true

}

func closedatabase() bool {
	// var err error
	// err = DB.Close()
	// if err != nil {
	// 	errorpk.AddError("closedatabase TransactionStruct package", "can't close the database")
	// 	return false
	// }
	return true
}

func getLastTransactionByID(id string) TransactionDB {
	opendatabase()
	var result TransactionDB
	iter := DB.NewIterator(util.BytesPrefix([]byte(id)), nil)
	for iter.Last() {
		value := iter.Value()
		json.Unmarshal(value, &result)
		break
	}
	closedatabase()
	return result
}

//***************************************************************************************************************
// add TransactionDB to database with path Database/TransactionStruct and the key is TransactionDB.TransactionID
//***************************************************************************************************************
func AddTransactiondb(TransactionObj TransactionDB) bool {

	firstAcc := accountdb.GetFirstAccount()
	haveSender := false
	fmt.Println("++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++ TransactionInput ", len(TransactionObj.TransactionInput))
	if len(TransactionObj.TransactionInput) >= 1 {
		TransactionObj.TransactionKey = TransactionObj.TransactionInput[0].SenderPublicKey + "_" + TransactionObj.TransactionTime.String() + "_" + TransactionObj.TransactionID
		haveSender = true
	} else if len(TransactionObj.TransactionOutPut) == 1 {
		if firstAcc.AccountPublicKey == TransactionObj.TransactionOutPut[0].RecieverPublicKey {
			TransactionObj.TransactionKey = "0000" + "_" + TransactionObj.TransactionTime.String() + "_" + TransactionObj.TransactionID
		}
	} else {
		errorpk.AddError("addTransaction transactionModule package", "TransactionObj have no sender or first account public key", "")
		return false
	}

	var lastTx TransactionDB
	if haveSender {
		lastTx = getLastTransactionByID(TransactionObj.TransactionInput[0].SenderPublicKey)
	} else {
		lastTx = getLastTransactionByID("0000")
	}

	diff := TransactionObj.TransactionTime.Sub(lastTx.TransactionTime)
	if diff.Seconds() < 5 {
		errorpk.AddError("addTransaction transactionModule package", "TransactionObj have the same time of the last transaction made for this public key", "")
		return false
	}

	opendatabase()

	iter := DB.NewIterator(nil, nil)
	for iter.Next() {

		value := iter.Value()
		var tx TransactionDB
		json.Unmarshal(value, &tx)
		if tx.TransactionKey == TransactionObj.TransactionKey {
			errorpk.AddError("addTransaction transactionModule package", "TransactionObj key exist in db", "")
			closedatabase()
			return false
		}
	}

	d, convert := globalPkg.ConvetToByte(TransactionObj, "addTransaction transactionModule package")
	if !convert {
		closedatabase()
		return false
	}
	fmt.Println("++++++++++++++++++++++++++++++++++++++++++TransactionObj.TransactionKey ", TransactionObj.TransactionKey)
	err := DB.Put([]byte(TransactionObj.TransactionKey), d, nil)
	closedatabase()

	if err != nil {
		errorpk.AddError("addTransaction transactionModule package", "can't insert transaction to db", "")
		closedatabase()
		return false
	}
	closedatabase()
	return true
}

//***************************************************************************************************************
// get TransactionDB from database with path Database/TransactionStruct and the key is TransactionDB.TransactionID
//***************************************************************************************************************

func GetTransactionByKey(key string) (transaction TransactionDB) {
	opendatabase()
	data, _ := DB.Get([]byte(key), nil)
	closedatabase()
	// if err != nil {
	// 	errorpk.AddError("getTransactionByKey transactionModule package",   "can't get TransactionDB")
	// }

	json.Unmarshal(data, &transaction)

	return transaction
}

//***************************************************************************************************************
// get all TransactionDB from database with path Database/TransactionStruct
//***************************************************************************************************************

func GetAllTransaction() (result []TransactionDB) {
	opendatabase()

	iter := DB.NewIterator(nil, nil)
	for iter.Next() {

		value := iter.Value()
		var tx TransactionDB
		json.Unmarshal(value, &tx)
		result = append(result, tx)
	}
	closedatabase()
	return result
}
func GetAllTransactionForPK(publicKey string) (result []TransactionDB) {
	opendatabase()

	iter := DB.NewIterator(nil, nil)
	for iter.Next() {

		value := iter.Value()
		var tx TransactionDB
		json.Unmarshal(value, &tx)
		if len(tx.TransactionInput) > 0 && tx.TransactionInput[0].SenderPublicKey == publicKey {
			result = append(result, tx)
		}

	}
	closedatabase()
	return result
}
