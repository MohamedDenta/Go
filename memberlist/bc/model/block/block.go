package block

import (
	"encoding/hex"
	"encoding/json"
	"fmt"

	"../account"
	errorpk "../errorpk"         //  write an error on the json file
	transaction "../transaction" //to use transactions in the block structure

	"../cryptogrpghy"
	globalPkg "../globalPkg"
	validator "../validator"
)

//CalculateBlockHash calculate block hash
func CalculateBlockHash(blockObj *BlockStruct) string {
	transactionsByte, _ := json.Marshal(blockObj.BlockTransactions)
	data := string(blockObj.BlockIndex) + blockObj.BlockTimeStamp.String() + string(transactionsByte) + blockObj.BlockPreviousHash
	hashed := cryptogrpghy.ClacHash([]byte(data))
	return hex.EncodeToString(hashed[:])
	// 	blockObj.BlockIndex =""
	//    strblockObj := fmt.Sprintf("%v", blockObj)
	// 	hashed := globalPkg.CreateHash(blockObj.BlockTimeStamp,strblockObj, 5)
	// 	return hashed
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

/*----------------- function to save a Block on a json file -----------------*/
// func AddBlock(blockObj BlockStruct, postLedger bool) string {
// 	var accountPublicKeyLst []string
// 	// fmt.Println("--Israa --  ", validateBlock(blockObj, postLedger))
// 	if (findBlockByKey(blockObj.BlockIndex)).BlockHash == "" && validateBlock(blockObj, postLedger) {
// 		if blockCreate(blockObj) {
// 			for _, transactionObj := range blockObj.BlockTransactions {
// 				for _, transactionInputObj := range transactionObj.TransactionInput {
// 					accountPublicKeyLst = addAccountBlock(accountPublicKeyLst, transactionInputObj.SenderPublicKey)
// 				}

// 				for _, transactionOutPutObj := range transactionObj.TransactionOutPut {
// 					accountPublicKeyLst = addAccountBlock(accountPublicKeyLst, transactionOutPutObj.RecieverPublicKey)
// 				}

// 				transaction.DeleteTransaction(transactionObj)

// 			}

// 			for _, accountObj := range accountPublicKeyLst {
// 				account.AddBlockToAccount(accountObj, blockObj.BlockIndex)
// 			}

// 			return ""
// 		} else {
// 			for _, objValidator := range validator.ValidatorsLstObj {
// 				if objValidator.ValidatorPublicKey == blockObj.ValidatorPublicKey {
// 					objValidator.ValidatorStakeCoins = 0
// 					validator.UpdateValidator(objValidator)
// 				}
// 			}
// 			return errorpk.AddError("AddBlock Block package", "Check your path or object to Add Block", "logical error")
// 		}
// 	}
// 	// + blockObj.BlockIndex

// 	return errorpk.AddError("AddBlock Block package", "The Block is already exists ", "hack error")

// }
func AddBlock(blockObj BlockStruct, postLedger bool) string {
	var accountPublicKeyLst []string
	var tokenidinput string
	var tokenidoutput []string
	var lengthtxInput int
	if blockObj.BlockIndex == "000000000000000000000000000000" || ((findBlockByKey(blockObj.BlockIndex)).BlockHash == "" && validateBlock(blockObj, postLedger)) {
		if blockCreate(blockObj) {

			for _, transactionObj := range blockObj.BlockTransactions {
				if transactionObj.Filestruct.FileSize == 0 {

					for _, transactionInputObj := range transactionObj.TransactionInput {
						accountPublicKeyLst = addAccountBlock(accountPublicKeyLst, transactionInputObj.SenderPublicKey)
						tokenidinput = transactionInputObj.TokenID
					}

					for _, transactionOutPutObj := range transactionObj.TransactionOutPut {
						accountPublicKeyLst = addAccountBlock(accountPublicKeyLst, transactionOutPutObj.RecieverPublicKey)
						tokenidoutput = append(tokenidoutput, transactionOutPutObj.TokenID)
					}
				} else {

					account.AddBlockFileToAccount(transactionObj.Filestruct, blockObj.BlockIndex)

				}
				//else add transaction file to the account list

				transaction.DeleteTransaction(transactionObj)

			}

			for _, accountObj := range accountPublicKeyLst {
				if lengthtxInput == 0 {
					account.AddBlockToAccount(accountObj, blockObj.BlockIndex, tokenidoutput[0])
				} else {
					account.AddBlockToAccount(accountObj, blockObj.BlockIndex, tokenidinput)
				}
			}

			//else
			//

			return ""
		} else {
			for _, objValidator := range validator.ValidatorsLstObj {
				if objValidator.ValidatorPublicKey == blockObj.ValidatorPublicKey {
					objValidator.ValidatorStakeCoins = 0
					validator.UpdateValidator(objValidator)
				}
			}
			return errorpk.AddError("AddBlock Block package", "Check your path or object to Add Block", "logical error")
		}
	}
	// + blockObj.BlockIndex

	return errorpk.AddError("AddBlock Block package", "The Block is already exists ", "hack error")

}

/*----------------- function to delete a Block on a json file -----------------*/
func DeleteBlock(blockObj BlockStruct) string {
	if (findBlockByKey(blockObj.BlockIndex)).BlockHash != "" {
		if deleteBlock(blockObj.BlockIndex) {
			return ""
		} else {
			return errorpk.AddError("DeleteBlock Block package", "Check your path to Delete the Block", "logical error")
		}

	}
	// +blockObj.BlockIndex
	return errorpk.AddError("FindjsonFile Block package", "Can't find the Block obj ", "hack error")

}

// func GetBalanceByPublickey(publicKey string) int {

// 	var balance int
// 	balance = 0
// 	accountObj := account.GetAccountByAccountPubicKey(publicKey)
// 	for _, blockObj := range accountObj.BlocksLst {
// 		block := findBlockByKey(blockObj)
// 		for _, transactionObj := range block.BlockTransactions {
// 			for _, vinObj := range transactionObj.TransactionInput {
// 				if string(vinObj.SenderPublicKey) == publicKey {
// 					balance = balance - vinObj.InputValue
// 				}
// 			}

// 			for _, VoutObj := range transactionObj.TransactionOutPut {
// 				if string(VoutObj.RecieverPublicKey) == publicKey {
// 					balance = balance - VoutObj.OutPutValue
// 				}
// 			}

// 		}
// 	}
// 	return balance

// }

/*----------------- function to get Blockchain -----------------*/
func GetBlockchain() []BlockStruct {
	return getAllBlocks()

}

/*----------------- function to get Block information using block id -----------------*/
func GetBlockInfoByID(index string) BlockStruct {
	return findBlockByKey(index)
}

/*----------------- function to validate the Block before adding  -----------------*/
// func validateBlock(blockObj BlockStruct) bool {

// 	return true
// }
/*----------------- function to validate the Block before adding  -----------------*/
func validateBlock(blockObj BlockStruct, postLedger bool) bool {

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
				// fmt.Println("   _________________    ", trans1)
				// fmt.Println("   _________________ **   ", trans2)
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

			fmt.Println("*", err)
			return false
		}
		blockIndex, _ := globalPkg.ConvertIntToFixedLengthString(
			globalPkg.ConvertFixedLengthStringtoInt(lstBlockObj.BlockIndex)+1, globalPkg.GlobalObj.StringFixedLength,
		)
		// blockIndex =""
		if blockIndex != blockObj.BlockIndex {
			err = errorpk.AddError("validateBlock Block package", "BlockIndex is not correct", "hack error")

			fmt.Println("  +++ ", err)
			return false
		}
	}

	if CalculateBlockHash(&blockObj) != blockObj.BlockHash {
		err = errorpk.AddError("validateBlock Block package", "Block Hash is not correct", "hack error")

		fmt.Println("*", err)
		return false
	}

	for _, objValidator := range validator.ValidatorsLstObj {
		fmt.Println("Validator pk :==============  ", objValidator.ValidatorPublicKey)
		if objValidator.ValidatorPublicKey == blockObj.ValidatorPublicKey {
			return true
		}
	}
	return false
}

/*------------------function to get the last block-------------*/
func GetLastBlock() BlockStruct {
	return getLastBlock()
}
