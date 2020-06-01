package token

import (
	"encoding/json"
	"fmt"
	"time"

	errorpk "../errorpk"
	"../globalPkg"

	"github.com/syndtr/goleveldb/leveldb"
)

//StructToken struct for object to be saved in db
type StructToken struct {
	TokenID           string
	TokenName         string
	TokenSymbol       string
	// IconURL           string
	TokenIcon         string
	Description       string
	Password          string //password of account who create token
	InitiatorAddress  string // public key address who create token
	TokensTotalSupply float64
	TokenValue        float64
	Reissuability     bool
	Precision         int
	UsageType         string //security or utility
	TokenType         string //public or private
	ValueDynamic      bool
	ContractID        string
	UserPublicKey     []string //who users use this token
	Dynamicprice      float64  //get the value BiddingplatformAPIURL if value dynamic is true
	TokenTime         time.Time
}

//tokenTransactionStruct struct for transaction token
type tokenTransactionStruct struct {
	SenderPK   string
	TokenID    string
	ReceiverPK string
	Amount     float64
	Sendersign string
}

//DB name leveldb
var DB *leveldb.DB

//Open flag open db or not
var Open = false

// opendatabase create or open DB if exist
func opendatabase() bool {

	if !Open {
		Open = true
		DBpath := "Database/TokenStruct"
		var err error
		DB, err = leveldb.OpenFile(DBpath, nil)
		if err != nil {
			errorpk.AddError("opendatabase tokenStruct package", "can't open the database", "DBError")
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

//TokenCreate insert tokenStruct
func TokenCreate(data StructToken) bool {

	opendatabase()
	d, convert := globalPkg.ConvetToByte(data, "Token create Token package")
	if !convert {
		closedatabase()
		return false
	}
	data.Password = ""
	err := DB.Put([]byte(data.TokenID), d, nil)

	if err != nil {
		errorpk.AddError("TokenCreate  tokenStruct package", "can't create tokenStruct", "DBError")
		return false
	}
	closedatabase()
	return true
}

// GetAllTokens get all StructTokens
func GetAllTokens() (values []StructToken) {

	opendatabase()
	iter := DB.NewIterator(nil, nil)
	for iter.Next() {

		value := iter.Value()
		var newdata StructToken
		json.Unmarshal(value, &newdata)
		values = append(values, newdata)
	}
	closedatabase()
	return values
}

// FindTokenByid select By TokenID StructToken
func FindTokenByid(id string) (tokenStructObj StructToken) {
	opendatabase()
	data, err := DB.Get([]byte(id), nil)
	if err != nil {
		errorpk.AddError("FindtokenByid  StructToken package", "can't get StructToken", "DBError")
	}

	json.Unmarshal(data, &tokenStructObj)
	closedatabase()
	return tokenStructObj
}

// FindTokenByTokenName get by Token name
func FindTokenByTokenName(tokenname string) (values StructToken) {
	opendatabase()
	iter := DB.NewIterator(nil, nil)
	for iter.Next() {

		value := iter.Value()

		var newdata StructToken
		json.Unmarshal(value, &newdata)
		if newdata.TokenName == tokenname {
			values = newdata
		}

	}
	closedatabase()
	return values
}

//UpdateTokendb update StructToken by token id
func UpdateTokendb(data StructToken) bool {

	TokenStructObj := FindTokenByid(data.TokenID)
	if TokenStructObj.TokenID == "" {
		fmt.Println("Token ID  don't exist before ")
		return false
	}
	//update except the token name, symbol, id,InitiatorAddress public key ,logo fields are disabled.
	data.TokenID = TokenStructObj.TokenID
	data.TokenName = TokenStructObj.TokenName
	data.TokenSymbol = TokenStructObj.TokenSymbol
	data.InitiatorAddress = TokenStructObj.InitiatorAddress
	data.UsageType = TokenStructObj.UsageType

	err := TokenCreate(data)
	if err {
		return true
	}
	return false
}

func GetLastToken() StructToken {
	opendatabase()

	iter := DB.NewIterator(nil, nil)
	iter.Last()
	value := iter.Value()
	var newdata StructToken
	json.Unmarshal(value, &newdata)

	closedatabase()

	return newdata
}

func GetFirstToken() StructToken {
	opendatabase()

	iter := DB.NewIterator(nil, nil)
	iter.First()
	value := iter.Value()
	var newdata StructToken
	json.Unmarshal(value, &newdata)

	closedatabase()

	return newdata
}
