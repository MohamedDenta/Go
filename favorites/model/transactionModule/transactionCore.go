package transactionModule

import (
	"bytes"
	"encoding/json"
	"strings"

	"../account"
	"../accountdb"
	"../block"
	"../cryptogrpghy"
	"../errorpk"
	"../globalPkg"
	"../responses"
	"../token"
	"../transaction"
	"../validator"

	"fmt"
	"sort"
	"time"

	//alaa
	"../globalfinctiontransaction"
)

//jsonTransactions json transaction return
type jsonTransactions struct {
	TransactionDate time.Time
	Sender          string
	Receiver        string
	TokenID         string
	Amount          float64
}

//jsonTransactionsfortokens json transaction fro token creator see transaction
type jsonTransactionsfortokens struct {
	TransactionDate time.Time
	Sender          string
	Receiver        string
	Amount          float64
}

//jsonAccountBalanceStatement account balance statement
type jsonAccountBalanceStatement struct {
	TotalReceived float64
	TotalSent     float64
	TotalBalance  float64
}

//BalanceAccount return account balance of token name and balance
type BalanceAccount struct {
	Tokenname string
	Balance   *jsonAccountBalanceStatement
}

// GetUnspentAndSpentTxs gets the account's inputs & outputs and filter through them to get the unspent and spent inputs.
func GetUnspentAndSpentTxs(publicKey string) ([]transaction.TXInput, []transaction.TXInput) {

	// unfiltered inputs got from getTxInputs function (it's actually outputs. check getTxInputs function for more info).
	// you know like literally it's called Unspent Transaction Outputs
	var unfilteredInputs []transaction.TXInput
	var spentTxInputs []transaction.TXInput
	var unSpentTxInputs []transaction.TXInput

	accountObj := account.GetAccountByAccountPubicKey(publicKey)

	transactionPool := transaction.GetPendingTransactions()
	// loop over block list
	for _, blockObj := range accountObj.BlocksLst {
		blockObj = cryptogrpghy.KeyDecrypt("123456789", blockObj)
		blockObj := block.GetBlockInfoByID(blockObj)
		// loop over transactions in every blockObj the account is associated with.
		// linear search
		for _, transactionObj := range blockObj.BlockTransactions {
			// get all inputs
			spent, unspent := getTxInputs(transactionObj, accountObj.AccountPublicKey)
			spentTxInputs = append(spentTxInputs, spent...)
			unfilteredInputs = append(unfilteredInputs, unspent...)
		}
	}
	for _, transactionObj := range transactionPool {
		for _, txInput := range transactionObj.TransactionInput {
			for _, blockTxInput := range unfilteredInputs {
				if txInput.InputID != blockTxInput.InputID {
					spent, unspent := getTxInputs(transactionObj, accountObj.AccountPublicKey)
					spentTxInputs = append(spentTxInputs, spent...)
					unfilteredInputs = append(unfilteredInputs, unspent...)
				}
			}
		}
	}
	// filter through the inputs to get the unspent inputs.
	for _, unfilteredInput := range unfilteredInputs {
		spent := false
		for index, spentTxInput := range spentTxInputs {
			if spentTxInput.InputID == unfilteredInput.InputID && spentTxInput.InputValue == unfilteredInput.InputValue {
				spentTxInputs = append(spentTxInputs[:index], spentTxInputs[index+1:]...)
				spent = true
				break
			}
		}
		if !spent {
			unSpentTxInputs = append(unSpentTxInputs, unfilteredInput)
		}
	}
	return unSpentTxInputs, spentTxInputs
}

// makeTxInputs gets the unspent inputs and put them in transaction if there's an input with exact value of transaction
// amount + fees. else it will sort the UTXO ascending by value then sum and add each input by value until the
// sum >= transaction amount + fees. it also calculate the remainder of balance to the sender if the sum > amountWithFee
func makeTxInputs(tx *transaction.Transaction, tokenID, senderPK string, amountWithFee float64) {
	// usinputs, _ := GetUnspentAndSpentTxs(senderPK)
	usinputs := GetUnspentByTokenID(senderPK, tokenID)
	// for _, usinputObj := range usinputs {
	// 	if usinputObj.TokenID == tokenID && usinputObj.InputValue == amountWithFee {
	// 		tx.TransactionInput = append(
	// 			tx.TransactionInput, transaction.TXInput{
	// 				InputID: usinputObj.InputID, InputValue: usinputObj.InputValue,
	// 				SenderPublicKey: senderPK, TokenID: tokenID,
	// 			},
	// 		)
	// 		return
	// 	}
	// }

	// sort the unspent spent inputs ascending by InputValue
	sort.SliceStable(usinputs, func(k, j int) bool {
		return usinputs[k].InputValue < usinputs[j].InputValue
	},
	)
	var sum float64

	for _, unspentInput := range usinputs {
		// if unspentInput.TokenID == tokenID {
		tx.TransactionInput = append(
			tx.TransactionInput, transaction.TXInput{
				InputID: unspentInput.InputID, InputValue: unspentInput.InputValue,
				SenderPublicKey: senderPK, TokenID: tokenID,
			},
		)
		sum = sum + unspentInput.InputValue
		if amountWithFee <= sum {
			break
		}
		// }
	}
	if sum > amountWithFee {
		tx.TransactionOutPut = append(
			tx.TransactionOutPut, transaction.TXOutput{
				OutPutValue:       sum - amountWithFee,
				RecieverPublicKey: senderPK, TokenID: tokenID,
			},
		)
	}
}

// DigitalwalletToUTXOTrans transform digitalWalletTx to UTXO Transaction of Token transfer operation.
func DigitalwalletToUTXOTrans(digitalWalletTx DigitalwalletTransaction, isFee bool) (transaction.Transaction, string) {

	var transactionobj transaction.Transaction
	var inoTokenID, _ = globalPkg.ConvertIntToFixedLengthString(0, globalPkg.GlobalObj.TokenIDStringFixedLength)
	var feesAccIndex, _ = globalPkg.ConvertIntToFixedLengthString(1, globalPkg.GlobalObj.StringFixedLength)
	var feesAccPublicKey = account.GetAccountByIndex(feesAccIndex).AccountPublicKey
	transactionobj.ServiceId = digitalWalletTx.ServiceId
	transactionobj.Validator = digitalWalletTx.Validator

	//feeValue := globalPkg.GlobalObj.TransactionFee
	//var amountWithFee = digitalWalletTx.Amount + feeValue
	feeValue := globalPkg.GlobalObj.TransactionFee
	if !isFee {
		feeValue = float64(0.0)
	}
	var amountWithFee = digitalWalletTx.Amount + feeValue
	// output with transaction amount subtracted with the validator's fee.
	transactionobj.TransactionOutPut = append(
		transactionobj.TransactionOutPut, transaction.TXOutput{
			OutPutValue: digitalWalletTx.Amount,
			TokenID:     digitalWalletTx.TokenID, RecieverPublicKey: digitalWalletTx.Receiver,
		},
	)
	// output with validator's fee for the 2 cases, first one is from InoToken to InoToken,
	// second one is from any other token to the same type of token.
	if digitalWalletTx.TokenID == inoTokenID {
		transactionobj.TransactionOutPut = append(
			transactionobj.TransactionOutPut, transaction.TXOutput{
				OutPutValue: feeValue, IsFee: true, TokenID: digitalWalletTx.TokenID,
				RecieverPublicKey: feesAccPublicKey,
			},
		)
	} else {
		tokenValue := token.FindTokenByid(digitalWalletTx.TokenID).TokenValue
		feeValue = tokenValue * globalPkg.GlobalObj.TransactionFee
		if !isFee {
			feeValue = float64(0.0)
		}
		transactionobj.TransactionOutPut = append(
			transactionobj.TransactionOutPut, transaction.TXOutput{
				OutPutValue: feeValue, IsFee: true, TokenID: digitalWalletTx.TokenID,
				RecieverPublicKey: feesAccPublicKey,
			},
		)
	}

	makeTxInputs(&transactionobj, digitalWalletTx.TokenID, digitalWalletTx.Sender, amountWithFee)

	// fmt.Println("\n @)@))@)@) Core transactionobj", transactionobj)

	transactionobj.Type = 0
	transactionobj.TransactionID = ""
	transactionobj.TransactionTime = globalPkg.UTCtime()
	transactionobj.TransactionTimeUser, _ = time.Parse("2006-01-02 03:04:05 PM -0000", time.Now().Format("2006-01-02 03:04:05 PM -0000"))
	transactionobj.TransactionID = globalfinctiontransaction.TransacionIDFromTemp(globalPkg.CreateHash(transactionobj.TransactionTime, fmt.Sprintf("%s", transactionobj), 3)) //globalPkg.CreateHash(transactionobj.TransactionTime, fmt.Sprintf("%s", transactionobj), 3)

	//save signature in database
	var SignObj transaction.SignatureDB
	var transobject transaction.Transaction //empty struct of transaction
	//return transaction id and time by signture
	signobjects := transaction.FindsignatureBySender(digitalWalletTx.Signature)
	// if find this signature
	if signobjects.Signature != "" {
		// loop on all transaction id and get trans id
		for _, transactionid := range signobjects.TransactionID {
			//loop on map and return key , value. k for transaction id , v for transaction time
			for k, v := range transactionid {
				// convert digital wallet time to utc and  format
				timedigitalwallet := globalPkg.UTCtimefield(digitalWalletTx.Time)
				// get transaction by id
				transObj := transaction.GettransactionByID(k)
				for _, transInputObj := range transObj.TransactionInput {
					// compare sender and collecting of seconds in transaction db and input of user
					if transInputObj.SenderPublicKey == digitalWalletTx.Sender && v.Unix() == timedigitalwallet.Unix() {
						InputIDArray := strings.Split(transactionobj.TransactionID, "_")

						globalfinctiontransaction.MapTempRollBack(transactionobj.Validator, InputIDArray[1])
						responseObj := responses.FindResponseByID("120")
						errorfound := responseObj.EngResponse

						return transobject, errorfound
					}
				}
			}
		}
		// map to add transaction id with time in record signature
		maptransid := map[string]time.Time{
			transactionobj.TransactionID: globalPkg.UTCtimefield(digitalWalletTx.Time),
		}
		// append map on array of transaction ids
		signobjects.TransactionID = append(signobjects.TransactionID, maptransid)
		transaction.SaveSignature(signobjects)
	} else {
		//signature not find
		SignObj.Signature = digitalWalletTx.Signature
		m := make(map[string]time.Time) //map for trans id and time

		m[transactionobj.TransactionID] = globalPkg.UTCtimefield(digitalWalletTx.Time)
		SignObj.TransactionID = append(SignObj.TransactionID, m)
		transaction.SaveSignature(SignObj)
	}

	return transactionobj, ""
}

// CreateTokenTx create token tx
func CreateTokenTx(token token.StructToken, senderAmount float64, senderTokenID string, notBilling bool) transaction.Transaction {

	var inoTokenID, _ = globalPkg.ConvertIntToFixedLengthString(0, globalPkg.GlobalObj.TokenIDStringFixedLength)
	var feesAccIndex, _ = globalPkg.ConvertIntToFixedLengthString(1, globalPkg.GlobalObj.StringFixedLength)
	var feesAccPublicKey = account.GetAccountByIndex(feesAccIndex).AccountPublicKey
	var transactionobj transaction.Transaction
	// output with transaction amount subtracted with the validator's fee.
	transactionobj.TransactionOutPut = append(
		transactionobj.TransactionOutPut, transaction.TXOutput{
			OutPutValue: token.TokensTotalSupply, TokenID: token.TokenID,
			RecieverPublicKey: token.InitiatorAddress,
		},
		// output with validator's fee itself.
		transaction.TXOutput{
			OutPutValue: globalPkg.GlobalObj.TransactionFee, IsFee: true, TokenID: senderTokenID,
			RecieverPublicKey: feesAccPublicKey,
		},
	)
	//var amountWithFee = senderAmount + globalPkg.GlobalObj.TransactionFee
	transactionobj.NotBilling = notBilling
	var amountWithFee = senderAmount
	if notBilling {
		amountWithFee = amountWithFee + globalPkg.GlobalObj.TransactionFee
	}
	makeTxInputs(&transactionobj, inoTokenID, token.InitiatorAddress, amountWithFee)

	transactionobj.TransactionTime = globalPkg.UTCtime()
	transactionobj.Validator = validator.CurrentValidator.ValidatorIP
	transactionobj.TransactionTimeUser, _ = time.Parse("2006-01-02 03:04:05 PM -0000", time.Now().Format("2006-01-02 03:04:05 PM -0000"))
	transactionobj.Type = 1
	transactionobj.TransactionID = ""
	transactionobj.TransactionID = globalfinctiontransaction.TransacionIDFromTemp(globalPkg.CreateHash(transactionobj.TransactionTime, fmt.Sprintf("%s", transactionobj), 3)) //globalPkg.CreateHash(transactionobj.TransactionTime, fmt.Sprintf("%s", transactionobj), 3)

	return transactionobj
}

// TODO: the user wants to refund for example 500 of his own token. at first it will be calculated by the tokenValue and

//RefundTokenTx refund token transaction
func RefundTokenTx(token token.StructToken, refundDwTx RefundDigitalWalletTx) transaction.Transaction {
	var transactionobj transaction.Transaction
	var refundFeesAccIndex, _ = globalPkg.ConvertIntToFixedLengthString(2, globalPkg.GlobalObj.StringFixedLength)
	var refundFeesAccPublicKey = account.GetAccountByIndex(refundFeesAccIndex).AccountPublicKey
	var inoTokenID, _ = globalPkg.ConvertIntToFixedLengthString(0, globalPkg.GlobalObj.TokenIDStringFixedLength)
	tokenFee := refundDwTx.Amount * globalPkg.GlobalObj.TransactionRefundFee

	// TODO: This is a right calculation..
	//  output will be the inoToken (regardless of it's sent to point of sale flat currency (service account) or inoToken to the person who refund)
	//  and input will be the token desired to refund from.
	// TODO: if this refundValue calculation is less or more than senderAmount * float32(token.TokenValue) . then
	// TODO: the difference either less(profit) or more(loss) will be refrenced on the inovatian account.
	// convert the refunded token amount to inoToken amount.
	toInoToken := refundDwTx.Amount * token.TokenValue
	// convert the refund value (in InoToken value) to dollar value.
	refundValue := toInoToken * globalPkg.GlobalObj.InoCoinToDollarRatio

	var amountWithFee = refundDwTx.Amount + globalPkg.GlobalObj.TransactionRefundFee

	// output with transaction amount subtracted with the validator's fee.
	if refundDwTx.FlatCurrency {
		transactionobj.TransactionOutPut = append(
			transactionobj.TransactionOutPut, transaction.TXOutput{
				OutPutValue: refundValue, TokenID: inoTokenID, RecieverPublicKey: refundDwTx.Receiver,
			},
		)
		if refundValue < toInoToken {
			var tx DigitalwalletTransaction
			tx.Sender = refundDwTx.Receiver
			tx.TokenID = inoTokenID
			if toInoToken-refundValue < 0 {
				tx.Amount = (toInoToken - refundValue) * -1
			} else {
				tx.Amount = toInoToken - refundValue
			}

			transferRefundRemaineder(tx, true) // profit for the Inovatian account
		} else if refundValue > toInoToken {
			// TODO: create input after this output for the inovatian stating the loss of ( (refundValue + tokenFee) - toInoToken) ???
			// TODO: solved, just make a new Tx with its input from service account and output is for account who refuned this Tx.
			var tx DigitalwalletTransaction
			tx.Receiver = refundDwTx.Receiver
			tx.TokenID = inoTokenID
			if refundValue-toInoToken < 0 {
				tx.Amount = (refundValue - toInoToken) * -1
			} else {
				tx.Amount = refundValue - toInoToken
			}
			// add the transaction fee to be the amount
			tx.Amount += globalPkg.GlobalObj.TransactionFee
			transferRefundRemaineder(tx, false) // loss for the Inovatian account
		}
		transactionobj.Type = 2
	} else {
		transactionobj.Type = 3

		transactionobj.TransactionOutPut = append(
			transactionobj.TransactionOutPut, transaction.TXOutput{
				OutPutValue: toInoToken, TokenID: inoTokenID,
				RecieverPublicKey: refundDwTx.Sender,
			},
		)
	}
	transactionobj.TransactionOutPut = append(
		transactionobj.TransactionOutPut, transaction.TXOutput{
			OutPutValue: tokenFee, IsFee: true, TokenID: inoTokenID,
			RecieverPublicKey: refundFeesAccPublicKey,
		},
	)

	makeTxInputs(&transactionobj, refundDwTx.TokenID, refundDwTx.Sender, amountWithFee)
	transactionobj.TransactionTime = globalPkg.UTCtime()
	transactionobj.TransactionTimeUser, _ = time.Parse("2006-01-02 03:04:05 PM -0000", time.Now().Format("2006-01-02 03:04:05 PM -0000"))
	transactionobj.TransactionID = ""
	transactionobj.TransactionID = globalfinctiontransaction.TransacionIDFromTemp(globalPkg.CreateHash(transactionobj.TransactionTime, fmt.Sprintf("%s", transactionobj), 3)) //globalPkg.CreateHash(transactionobj.TransactionTime, fmt.Sprintf("%s", transactionobj), 3)

	return transactionobj
}

//addcoins add coins to first account
func addcoins(digitalwalletObj DigitalwalletTransaction) transaction.Transaction {

	firstTokenID, _ := globalPkg.ConvertIntToFixedLengthString(0, globalPkg.GlobalObj.TokenIDStringFixedLength)

	firstaccount := accountdb.GetFirstAccount()
	var transactionObj transaction.Transaction
	transactionObj.TransactionTime, _ = time.Parse("2006-01-02 03:04:05 PM -0000", time.Now().UTC().Format("2006-01-02 03:04:05 PM -0000"))
	transactionObj.TransactionTimeUser, _ = time.Parse("2006-01-02 03:04:05 PM -0000", time.Now().Format("2006-01-02 03:04:05 PM -0000"))
	transactionObj.TransactionOutPut = append(transactionObj.TransactionOutPut, transaction.TXOutput{
		OutPutValue: digitalwalletObj.Amount, RecieverPublicKey: firstaccount.AccountPublicKey,
		TokenID: firstTokenID,
	})

	transactionObj.TransactionID = ""
	transactionObj.TransactionID = globalfinctiontransaction.TransacionIDFromTemp(globalPkg.CreateHash(transactionObj.TransactionTime, fmt.Sprintf("%s", transactionObj), 3))
	return transactionObj
}

//ValidateTx2 validate transaction
func ValidateTx2(digitalWalletTx DigitalwalletTransaction, tx transaction.Transaction) string {
	return "true"
	fmt.Println("\n the broadcast handle transaction.Transaction:", tx)

	if errStr := digitalWalletTx.ValidateTransaction(); errStr == "" {
		fmt.Println("\n the broadcast handle transactionModule.ValidateTransaction:", errStr)

		inoAccPK := accountdb.GetFirstAccount().AccountPublicKey
		isAddingCoinsToIno := digitalWalletTx.Sender == "" && digitalWalletTx.Receiver != "" && digitalWalletTx.Receiver == inoAccPK

		if isAddingCoinsToIno {
			return "true"
		} else {
			outputSum := 0.0
			inoTokenID, _ := globalPkg.ConvertIntToFixedLengthString(0, globalPkg.GlobalObj.TokenIDStringFixedLength)
			for _, outputObj := range tx.TransactionOutPut {
				if outputObj.IsFee && inoTokenID != digitalWalletTx.TokenID {
					tokenValue := token.FindTokenByid(digitalWalletTx.TokenID).TokenValue
					outputSum += outputObj.OutPutValue / tokenValue
				} else {
					outputSum += outputObj.OutPutValue
				}
			}
			inputSum := 0.0
			for _, inputObj := range tx.TransactionInput {
				inputSum += inputObj.InputValue
			}
			// fmt.Println("-------------------------------------------------------outputSum", outputSum)
			// fmt.Println("-------------------------------------------------------inputSum", inputSum)

			if outputSum == inputSum {
				allOldTx := transaction.GetAllTransactionForPK(digitalWalletTx.Sender)
				inputexist := false
				for _, txObj := range allOldTx {
					oldInputTx, _ := json.Marshal(txObj.TransactionInput)
					newInputTx, _ := json.Marshal(tx.TransactionInput)
					if bytes.Compare(oldInputTx, newInputTx) == 0 {
						inputexist = true
					}
				}
				if !inputexist {
					var isDoubleSpend bool
					senderPK := digitalWalletTx.Sender

					if txPoolInputsIds, userExist := transaction.PendingValidationTxs[senderPK]; !userExist {
						for _, txInput := range tx.TransactionInput {
							transaction.PendingValidationTxs[senderPK] = append(transaction.PendingValidationTxs[senderPK], txInput.InputID)
						}
					} else {
						for _, txInput := range tx.TransactionInput {
							for _, txInputID := range txPoolInputsIds {
								if txInput.InputID == txInputID {
									fmt.Println("\n found double spend: \n", txInput.InputID, "\n", txInputID)
									isDoubleSpend = true
								}
							}
						}
					}

					if !isDoubleSpend {
						TransactionObj2 := tx
						TransactionObj2.TransactionID = ""
						fmt.Println("validator ", tx.Validator)
						//alaa
						getmap := globalfinctiontransaction.GetTransactionIndexTemMap()
						getindex := getmap[tx.Validator]
						hashtransaction := globalfinctiontransaction.CheckTransactionID(getindex, globalPkg.CreateHash(tx.TransactionTime, fmt.Sprintf("%s", TransactionObj2), 3), tx.Validator)
						fmt.Println("tx.TransactionID", tx.TransactionID)
						fmt.Println("TransactionObj2", hashtransaction)

						// alaa

						// getmap := globalfinctiontransaction.GetTransactionIndexTemMap()
						// getindex := getmap[tx.Validator]
						// hashtransaction := globalfinctiontransaction.CheckTransactionID(getindex, globalPkg.CreateHash(tx.TransactionTime, fmt.Sprintf("%s", TransactionObj2), 3), tx.Validator)
						// fmt.Println(" ----- transobj add    ****   ", tx)
						// fmt.Println("  _______   ^  hash  ^    ______", hashtransaction)

						// hashtransaction := globalPkg.CreateHash(tx.TransactionTime, fmt.Sprintf("%s", TransactionObj2), 3)
						// fmt.Println(" ----- transobj add    ****   ", tx)
						// fmt.Println("  _______   ^  hash  ^    ______", hashtransaction)

						if tx.TransactionID == hashtransaction {
							return "true"
						} else {
							responseObj := responses.FindResponseByID("141")
							return responseObj.EngResponse

						}
					} else {
						tx.DeleteTransaction()
						responseObj := responses.FindResponseByID("109")
						return responseObj.EngResponse

					}
				} else {
					errorpk.AddError("ValidateTx2 Transaction module", "input is exist", "Validation Error")
					return "input is exist"
				}

			} else {
				errorpk.AddError("ValidateTx2 Transaction module", "digitalWalletTx is rong", "Validation Error")
				return "digitalWalletTx is rong"
			}
		}
	} else {
		errorpk.AddError("ValidateTx2 Transaction module", errStr, "Validation Error")
		return errStr
	}
}

// GetAllTransactionsByTokenID all tx by token id
func GetAllTransactionsByTokenID(tokenID string) map[string]map[string][]jsonTransactionsfortokens {
	returnedTx := map[string][]jsonTransactionsfortokens{}
	tokennamemap := map[string]map[string][]jsonTransactionsfortokens{}
	var returnTransaction jsonTransactionsfortokens
	var normalTxs, TokenCreationTxs, RefundedTokenTxs []jsonTransactionsfortokens
	var senderpk string

	blocklist := block.GetBlockchain()
	for _, blockObj := range blocklist {

		for _, transactionObj := range blockObj.BlockTransactions {
			for _, transactionInputObj := range transactionObj.TransactionInput {
				senderpk = account.GetAccountByAccountPubicKey(transactionInputObj.SenderPublicKey).AccountName
			}

			for _, transactionOutPutObj := range transactionObj.TransactionOutPut {
				if transactionOutPutObj.TokenID == tokenID {
					returnTransaction.Sender = senderpk
					returnTransaction.Receiver = account.GetAccountByAccountPubicKey(transactionOutPutObj.RecieverPublicKey).AccountName
					returnTransaction.Amount = transactionOutPutObj.OutPutValue
					// returnTransaction.TransactionDate = transactionObj.TransactionTime
					returnTransaction.TransactionDate = transactionObj.TransactionTimeUser
					if transactionObj.Type == 0 {
						normalTxs = append(normalTxs, returnTransaction)
					}
					if transactionObj.Type == 1 {
						returnTransaction.Sender = ""
						TokenCreationTxs = append(TokenCreationTxs, returnTransaction)
					}
					if transactionObj.Type == 2 {
						RefundedTokenTxs = append(RefundedTokenTxs, returnTransaction)
					}

					break
				}
			}
		}
	}
	tokenname := token.FindTokenByid(tokenID).TokenName
	returnedTx["normal"] = normalTxs
	returnedTx["refunded"] = RefundedTokenTxs
	returnedTx["tokenCreation"] = TokenCreationTxs
	tokennamemap[tokenname] = returnedTx
	return tokennamemap
}

// GetUnspentByTokenID get unspent input for publickey and token id
func GetUnspentByTokenID(publicKey string, tokenid string) []transaction.TXInput {

	// unfiltered inputs got from getTxInputs function (it's actually outputs. check getTxInputs function for more info).
	// you know like literally it's called Unspent Transaction Outputs
	var unfilteredInputs []transaction.TXInput
	var spentTxInputs []transaction.TXInput
	var unSpentTxInputs []transaction.TXInput

	accountObj := account.GetAccountByAccountPubicKey(publicKey)

	transactionPool := transaction.GetPendingTransactions()
	// loop over block list
	for _, blockObj := range accountObj.BlocksLst {
		blockObj = cryptogrpghy.KeyDecrypt("123456789", blockObj)
		blockObj := block.GetBlockInfoByID(blockObj)
		// loop over transactions in every blockObj the account is associated with.
		// linear search
		for _, transactionObj := range blockObj.BlockTransactions {
			// get all inputs
			// spent, unspent := getTxInputs(transactionObj, accountObj.AccountPublicKey)
			spent, unspent := getTxInputsforToken(transactionObj, accountObj.AccountPublicKey, tokenid)
			spentTxInputs = append(spentTxInputs, spent...)
			unfilteredInputs = append(unfilteredInputs, unspent...)
		}
	}
	for _, transactionObj := range transactionPool {
		for _, txInput := range transactionObj.TransactionInput {
			for _, blockTxInput := range unfilteredInputs {
				if txInput.InputID != blockTxInput.InputID {
					// spent, unspent := getTxInputs(transactionObj, accountObj.AccountPublicKey)
					spent, unspent := getTxInputsforToken(transactionObj, accountObj.AccountPublicKey, tokenid)
					spentTxInputs = append(spentTxInputs, spent...)
					unfilteredInputs = append(unfilteredInputs, unspent...)
				}
			}
		}
	}
	// filter through the inputs to get the unspent inputs.
	for _, unfilteredInput := range unfilteredInputs {
		spent := false
		for index, spentTxInput := range spentTxInputs {
			if spentTxInput.InputID == unfilteredInput.InputID && spentTxInput.InputValue == unfilteredInput.InputValue {
				spentTxInputs = append(spentTxInputs[:index], spentTxInputs[index+1:]...)
				spent = true
				break
			}
		}
		if !spent {
			unSpentTxInputs = append(unSpentTxInputs, unfilteredInput)
		}
	}

	return unSpentTxInputs
}

// getTxInputsforToken get transaction input for token
func getTxInputsforToken(tx transaction.Transaction, PubKey string, tokenid string) (spentTxs, unspentTxs []transaction.TXInput) {
	for _, transactionOutPutObj := range tx.TransactionOutPut {
		if transactionOutPutObj.RecieverPublicKey != PubKey && transactionOutPutObj.IsFee {
			continue
		} else if transactionOutPutObj.RecieverPublicKey == PubKey && transactionOutPutObj.TokenID == tokenid {
			unspentTxs = append(unspentTxs, transaction.TXInput{
				InputID: tx.TransactionID, InputValue: transactionOutPutObj.OutPutValue,
				SenderPublicKey: transactionOutPutObj.RecieverPublicKey, TokenID: transactionOutPutObj.TokenID,
			})
		}
	}
	for _, transactionInputObj := range tx.TransactionInput {
		if transactionInputObj.SenderPublicKey == PubKey && transactionInputObj.TokenID == tokenid {
			spentTxs = append(spentTxs, transactionInputObj)
		}
	}
	return spentTxs, unspentTxs
}
