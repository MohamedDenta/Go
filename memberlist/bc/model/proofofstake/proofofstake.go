package proofofstake

import (
	"fmt"
	"math/rand"

	"../block" //use block
	"../broadcastTcp"
	"../cryptogrpghy"
	"../errorpk"     // save the error
	"../globalPkg"   // set the waiting time
	"../transaction" //use transaction
	"../validator"   // use validator and validator structer

	//"encoding/hex"
	"encoding/hex"
	"encoding/json" //read and send json data through api

	// using API request
	"time" // to calculate the sleeping time
)

//winningValidatorStruct structure
type WinningValidatorStruct struct {
	Index           string
	WinnerValidator validator.ValidatorStruct
	NextBlockIndex  string
	TimeStamp       time.Time
	CurrentNode     validator.ValidatorStruct
}

//////////////////////////////////////////////////////////

//winningValidatorStruct structure
// type  struct {
// 	Index           string
// 	IP  string
// Count  int
// }

/* variable to count the validator who choose the winner */
var WinnerCount []string
var NextBlockIndex = 0

/*----------------The path in where the Block files are going to be stored----------------- */

var globalObj globalPkg.GlobalVariables

//function to pic the winner according to stake coin
func pickWinner() validator.ValidatorStruct {
	var winnerValidator validator.ValidatorStruct
	winnerValidator = validator.CurrentValidator
	for _, validatorObj := range validator.ValidatorsLstObj {
		if ((winnerValidator.ValidatorStakeCoins < validatorObj.ValidatorStakeCoins) || (winnerValidator.ValidatorStakeCoins == validatorObj.ValidatorStakeCoins && winnerValidator.ValidatorRegisterTime.After(validatorObj.ValidatorRegisterTime))) && validatorObj.ValidatorActive && !validatorObj.ValidatorRemove {
			winnerValidator = validatorObj
		}

	}
	return winnerValidator
}

//function to choose the winner and make sure to announce the winner
func Mining() {
	for {
		time.Sleep(time.Second * time.Duration(rand.Int31n(globalPkg.GlobalObj.ElectionTimeInSeconds)))

		// fmt.Println("Readyyyyyyyyyyyy", transaction.CheckReadyTransaction())
		if len(transaction.GetPendingTransactions()) != 0 && transaction.CheckReadyTransaction() {

			var winnerObj WinningValidatorStruct
			lastBlock := block.GetLastBlock()
			if lastBlock.BlockHash == "" {
				lastBlock.BlockIndex, _ = globalPkg.ConvertIntToFixedLengthString(0, globalPkg.GlobalObj.StringFixedLength)
				winnerObj.NextBlockIndex, _ = globalPkg.ConvertIntToFixedLengthString(
					globalPkg.ConvertFixedLengthStringtoInt(lastBlock.BlockIndex), globalPkg.GlobalObj.StringFixedLength,
				)
			} else {
				winnerObj.NextBlockIndex, _ = globalPkg.ConvertIntToFixedLengthString(
					globalPkg.ConvertFixedLengthStringtoInt(lastBlock.BlockIndex)+1, globalPkg.GlobalObj.StringFixedLength,
				)
			}

			winnerObj.CurrentNode = validator.CurrentValidator
			winnerObj.WinnerValidator = pickWinner()
			winnerObj.TimeStamp, _ = time.Parse("2006-01-02 03:04:05 PM -0000", time.Now().UTC().Format("2006-01-02 03:04:05 PM -0000"))
			winnerObj.Index = validator.CurrentValidator.ValidatorPublicKey + "_" + winnerObj.NextBlockIndex + "_" + errorpk.ConvertTimeToString(winnerObj.TimeStamp)
			//announcementWinningValidator(winnerObj)

			// var lst []string
			// lst = append(lst, winnerObj.TimeStamp.String())

			if winnerObj.WinnerValidator.ValidatorIP == validator.CurrentValidator.ValidatorIP {

				broadcastTcp.SendObject(winnerObj, winnerObj.WinnerValidator.ValidatorPublicKey, "", "proofOfStake", validator.CurrentValidator.ValidatorSoketIP)
			} else {
				broadcastTcp.SendObject(winnerObj, winnerObj.WinnerValidator.ValidatorPublicKey, "", "proofOfStake", winnerObj.WinnerValidator.ValidatorSoketIP)
			}

		}
	}

}

//function to announce the validator winner to all validators
// func announcementWinningValidator(winnerValidator WinningValidatorStruct) {
// 	// errStr := ""
// 	// for _, validatorObj := range validator.ValidatorsLstObj {
// 	// 	jsonObj, _ := json.Marshal(winnerValidator)
// 	// 	errStr = errStr + globalPkg.SendRequest(jsonObj, validatorObj.ValidatorIP+"/LotteryWinner", "POST")
// 	// }
// 	// if errStr != "" {
// 	// 	errorpk.AddError("announcementWinningValidator proofofstake package LotteryWinner", errStr)
// 	// }
// 	broadcastTcp.SendObject(winnerValidator, nil, nil, validator.CurrentValidator.ValidatorPublicKey, "", "proofOfStake", winnerValidator.WinnerValidator.ValidatorSoketIP)

// }

//function to select transactions according to priority
func selectTransactions(blockObj *block.BlockStruct) {
	pendingTransactions := transaction.GetPendingTransactions()
	var transactions []transaction.Transaction

	if (globalPkg.GlobalObj.NoOfTransactionsPerBlock >= len(pendingTransactions)) || (globalPkg.GlobalObj.NoOfTransactionsPerBlock == 0) {
		blockObj.BlockTransactions = pendingTransactions
		return

	}
	index := 0
	t := globalPkg.UTCtime()

	for _, transactionObj := range pendingTransactions {

		Subtime := (t.Sub(transactionObj.TransactionTime)).Seconds()
		if Subtime > 10 {
			transactions = append(transactions, transactionObj)
		} else {
			continue
		}

		if index == globalPkg.GlobalObj.NoOfTransactionsPerBlock {
			blockObj.BlockTransactions = transactions
			break
		}

		index = index + 1
	}

}

//calculateBlockHash returns the hash of all block information
func calculateBlockHash(blockObj *block.BlockStruct) {
	transactionsByte, _ := json.Marshal(blockObj.BlockTransactions)
	data := string(blockObj.BlockIndex) + blockObj.BlockTimeStamp.String() + string(transactionsByte) + blockObj.BlockPreviousHash
	hashed := cryptogrpghy.ClacHash([]byte(data))
	blockObj.BlockHash = hex.EncodeToString(hashed[:])
}

/*function to generate the block*/
/*func generateBlock(currentNode validator.ValidatorStruct) block.BlockStruct {

	var blockObj, prevBlock block.BlockStruct

	prevBlock = block.GetLastBlock()

	if prevBlock.BlockHash == "" {
		blockObj.BlockIndex, _ = globalPkg.ConvertIntToFixedLengthString(0, globalPkg.GlobalObj.StringFixedLength)
		blockObj.BlockPreviousHash = "0"

	} else {

		blockObj.BlockIndex, _ = globalPkg.ConvertIntToFixedLengthString(
			globalPkg.ConvertFixedLengthStringtoInt(prevBlock.BlockIndex)+1, globalPkg.GlobalObj.StringFixedLength,
		)

		blockObj.BlockPreviousHash = prevBlock.BlockHash
	}
	selectTransactions(&blockObj)
	blockObj.BlockTimeStamp, _ = time.Parse("2006-01-02 03:04:05 PM -0000", time.Now().UTC().Format("2006-01-02 03:04:05 PM -0000"))
	blockObj.ValidatorPublicKey = currentNode.ValidatorPublicKey
	calculateBlockHash(&blockObj)

	return blockObj

}
*/
func calcActiveValidatorsNum() int { //-----return num of Active validator
	var TotalNum = 0
	for _, validatorObj := range validator.ValidatorsLstObj {
		if validatorObj.ValidatorActive == true {
			TotalNum++
		}

	}
	return TotalNum
}

/*function to forge the block*/
func ForgeTheBlock(WinningValidatorObj WinningValidatorStruct) {
	if len(transaction.GetPendingTransactions()) != 0 {
		totalNumberOfValidators := calcActiveValidatorsNum()
		fmt.Println("--------------------------\n", WinningValidatorObj.NextBlockIndex)
		fmt.Println("***************************\n", totalNumberOfValidators)
		if NextBlockIndex != globalPkg.ConvertFixedLengthStringtoInt(WinningValidatorObj.NextBlockIndex) {
			NextBlockIndex = globalPkg.ConvertFixedLengthStringtoInt(WinningValidatorObj.NextBlockIndex)
			WinnerCount = append(WinnerCount, WinningValidatorObj.CurrentNode.ValidatorPublicKey)
		} else {
			WinnerCount = append(WinnerCount, WinningValidatorObj.CurrentNode.ValidatorPublicKey)
		}

		if len(WinnerCount) >= totalNumberOfValidators { ////should be vaild lenght of
			var blockObj, prevBlock block.BlockStruct

			prevBlock = block.GetLastBlock()

			if prevBlock.BlockHash == "" {
				fmt.Println("iam wincount")
				blockObj.BlockIndex, _ = globalPkg.ConvertIntToFixedLengthString(0, globalPkg.GlobalObj.StringFixedLength)
				blockObj.BlockPreviousHash = "0"

			} else {

				blockObj.BlockIndex, _ = globalPkg.ConvertIntToFixedLengthString(globalPkg.ConvertFixedLengthStringtoInt(prevBlock.BlockIndex)+1, globalPkg.GlobalObj.StringFixedLength)

				blockObj.BlockPreviousHash = prevBlock.BlockHash
			}
			selectTransactions(&blockObj)
			blockObj.BlockTimeStamp = globalPkg.UTCtime()
			blockObj.ValidatorPublicKey = validator.CurrentValidator.ValidatorPublicKey
			calculateBlockHash(&blockObj)
			///
			selectTransactions(&blockObj)
			blockObj.BlockTimeStamp = globalPkg.UTCtime()
			blockObj.ValidatorPublicKey = validator.CurrentValidator.ValidatorPublicKey
			blockObj.BlockHash = block.CalculateBlockHash(&blockObj)

			fmt.Println(" broabeforedcast", blockObj)
			broadcastTcp.BoardcastingTCP(blockObj, "", "block")

		}
	}

}

/*----------------- -----------------------API------------------------------------------------*/
/*----------------- endpoint to save the winning validator  -----------------*/
// func LotteryWinner(w http.ResponseWriter, req *http.Request) {
// 	w.Header().Set("Content-Type", "application/json")
// 	WinningValidatorObj := winningValidatorStruct{}
// 	errStr := ""
// 	err := json.NewDecoder(req.Body).Decode(&WinningValidatorObj)

// 	if err != nil {
// 		errStr = errorpk.AddError("lotteryWinner API validator package "+WinningValidatorObj.WinnerValidator.ValidatorIP+"/lotteryWlotteryWinnerinner", "can't convert body to Winning Validator obj")

// 	} else {

// 		if WinningValidatorObj.WinnerValidator.ValidatorPublicKey == validator.CurrentValidator.ValidatorPublicKey {
// 			forgeTheBlock(WinningValidatorObj, len(validator.ValidatorsLstObj))
// 		}

// 	}

// 	if errStr != "" {
// 		w.WriteHeader(http.StatusInternalServerError)
// 		w.Write([]byte(errStr))
// 	} else {
// 		sendJson, _ := json.Marshal(WinningValidatorObj)
// 		w.Header().Set("Content-Type", "application/json")
// 		w.WriteHeader(http.StatusOK)
// 		w.Write(sendJson)
// 	}
// }
