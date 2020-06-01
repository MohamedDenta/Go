package block

import (
	"encoding/json"
	"time"

	"../cryptogrpghy"

	"../constants"
	"../globalPkg"
	"../transaction"
	"github.com/mitchellh/mapstructure"
	"github.com/syndtr/goleveldb/leveldb"
)

// AbstractBlock abstract interface for block
type AbstractBlock interface {
	EncryptBlock() string
}

// BlockStruct data block struct
type BlockStruct struct {
	BlockIndex        string
	BlockTransactions []transaction.Transaction
	BlockPreviousHash string
	BlockHash         string
	BlockTimeStamp    time.Time
	ValidatorIP       string
}

// EncryptBlock encrypt block data using AES
func (b *BlockStruct) EncryptBlock() string {
	byteblock, _ := globalPkg.ConvetToByte(*b, "Block create block package")
	stringblock := string(byteblock)
	encryptedblock := cryptogrpghy.AESEncrypt(constants.ENCRYPT_KEY, stringblock)
	return encryptedblock
}

// FromJSON decode block from json format
func (b *BlockStruct) FromJSON(byteblock []byte) (BlockStructObj BlockStruct) {
	json.Unmarshal(byteblock, b)
	mapstructure.Decode(b, &BlockStructObj)
	return BlockStructObj
}

// DB database instance
var DB *leveldb.DB

//Open open database flag
var Open = false
