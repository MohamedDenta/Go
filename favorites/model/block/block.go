package block

import (
	"encoding/hex"
	"encoding/json"
	"fmt"

	"../globalPkg"
	"../transaction"

	"../account"
	errorpk "../errorpk" //  write an error on the json file

	//to use transactions in the block structure

	"../cryptogrpghy"
	validator "../validator"
)

//CalculateBlockHash calculate block hash
func (blockObj *BlockStruct) CalculateBlockHash() string {
	transactionsByte, _ := json.Marshal(blockObj.BlockTransactions)
	data := string(blockObj.BlockIndex) + blockObj.BlockTimeStamp.String() + string(transactionsByte) + blockObj.BlockPreviousHash
	hashed := cryptogrpghy.ClacHash([]byte(data))
	return hex.EncodeToString(hashed[:])
}

// function to add block on the account
func addAccountBlock(accountPublicKeyLst []string, publicKey string) []string {
	existsObj := false
	for _, accountPublicKeyObj := range accountPublicKeyLst {
		if accountPublicKeyObj == publicKey {
			existsObj = true
		}

	}
	if !existsObj {
		accountPublicKeyLst = append(accountPublicKeyLst, publicKey)
	}
	return accountPublicKeyLst
}

//AddBlock add block
func (blockObj *BlockStruct) AddBlock(postLedger bool) string {
	var accountPublicKeyLst []string
	var tokenidinput string
	var tokenidoutput []string
	var lengthtxInput int

	if (findBlockByKey(blockObj.BlockIndex)).BlockHash == "" && blockObj.validateBlock(postLedger) {
		if blockCreate(blockObj) {
			for _, transactionObj := range blockObj.BlockTransactions {
				if transactionObj.Filestruct.FileSize == 0 {
					// traverse over all transactions in block
					for _, transactionInputObj := range transactionObj.TransactionInput {
						accountPublicKeyLst = addAccountBlock(accountPublicKeyLst[:], transactionInputObj.SenderPublicKey)
						tokenidinput = transactionInputObj.TokenID
					}

					for _, transactionOutPutObj := range transactionObj.TransactionOutPut {
						accountPublicKeyLst = addAccountBlock(accountPublicKeyLst[:], transactionOutPutObj.RecieverPublicKey)
						tokenidoutput = append(tokenidoutput, transactionOutPutObj.TokenID)
					}
				} else {

					account.AddBlockFileToAccount((transactionObj.Filestruct), (blockObj.BlockIndex))

				}
				transactionObj.DeleteTransaction()

			}

			for _, accountObj := range accountPublicKeyLst {
				if lengthtxInput == 0 {
					account.AddBlockToAccount(accountObj, blockObj.BlockIndex, tokenidoutput[0])
				} else {
					account.AddBlockToAccount(accountObj, blockObj.BlockIndex, tokenidinput)
				}
			}

			return ""
		}
		// in case of failing in create block == invalid block
		for _, objValidator := range validator.ValidatorsLstObj {
			if objValidator.ValidatorIP == blockObj.ValidatorIP {
				objValidator.ValidatorStakeCoins = 0
				(&objValidator).UpdateValidator()
			}
		}
		return errorpk.AddError("AddBlock Block package", "Check your path or object to Add Block", "logical error")
	}

	return errorpk.AddError("AddBlock Block package", "The Block is already exists ", "hack error")

}

// validateBlock check block data
func (blockObj *BlockStruct) validateBlock(postLedger bool) bool {

	lstBlockObj := getLastBlock()

	trans := transaction.GetPendingTransactions()
	var existtrans bool
	var err string

	if len(blockObj.BlockTransactions) == 0 {
		return false
	}
	if !postLedger {
		for _, transactionObj := range blockObj.BlockTransactions {
			existtrans = false
			for _, objtrans := range trans {

				trans1 := fmt.Sprintf("%v", transactionObj)
				trans2 := fmt.Sprintf("%v", objtrans)

				if trans1 == trans2 {
					existtrans = true
					break
				}
			}
			if existtrans == false {
				err = errorpk.AddError("validateBlock Block package", "transaction id is not correct", "hack error")
				fmt.Println("-", err)
				return false
			}
		}
	}
	firstBlockIndex, _ := globalPkg.ConvertIntToFixedLengthString(0, globalPkg.GlobalObj.StringFixedLength)
	if blockObj.BlockIndex != firstBlockIndex {
		if blockObj.BlockPreviousHash != lstBlockObj.BlockHash {
			err = errorpk.AddError("validateBlock Block package", "previous Block Hash is not correct", "hack error")

			return false
		}

		blockIndex, _ := globalPkg.ConvertIntToFixedLengthString(
			globalPkg.ConvertFixedLengthStringtoInt(lstBlockObj.BlockIndex)+1, globalPkg.GlobalObj.StringFixedLength,
		)

		if blockIndex != blockObj.BlockIndex {
			err = errorpk.AddError("validateBlock Block package", "BlockIndex is not correct", "hack error")

			return false
		}
	}

	if blockObj.CalculateBlockHash() != blockObj.BlockHash {
		err = errorpk.AddError("validateBlock Block package", "Block Hash is not correct", "hack error")
		fmt.Println("*", err)
		return false
	}

	for _, objValidator := range validator.ValidatorsLstObj {
		if objValidator.ValidatorIP == blockObj.ValidatorIP {
			return true
		}
	}
	return false

}

// DeleteBlock delete block
func (blockObj *BlockStruct) DeleteBlock() string {
	if (findBlockByKey(blockObj.BlockIndex)).BlockHash != "" {
		if deleteBlock(blockObj.BlockIndex) {
			return ""
		}
		return errorpk.AddError("DeleteBlock Block package", "Check your path to Delete the Block", "logical error")
	}
	return errorpk.AddError("FindjsonFile Block package", "Can't find the Block obj ", "hack error")
}

// GetBlockchain get all blockchain
func GetBlockchain() []BlockStruct {
	return getAllBlocks()

}

// GetBlockInfoByID get block info by id
func GetBlockInfoByID(index string) BlockStruct {
	return findBlockByKey(index)
}

// GetLastBlock get block info
func GetLastBlock() BlockStruct {
	return getLastBlock()
}

// AddLedgerBlock call create block function
func AddLedgerBlock(data *BlockStruct) bool {

	return blockCreate(data)
}
