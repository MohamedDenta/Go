package transactionModule

import (
	"fmt"

	"../account"
	"../accountdb"
	"../block"
	"../cryptogrpghy"
	"../globalPkg"
	globalpkg "../globalPkg"
	"../token"
	"../transaction"
)

/*---------------function to get the account balance---------------------*/
func GetAccountBalance(public_key string) map[string]float64 {
	tokensBalance := map[string]float64{}

	unspentTxs, _ := GetUnspentAndSpentTxs(public_key)
	fmt.Println("\n GetAccountBalance unspentTxs:", unspentTxs)
	for _, unspnetTx := range unspentTxs {
		_, tokenExist := tokensBalance[unspnetTx.TokenID]
		if !tokenExist {
			tokensBalance[unspnetTx.TokenID] = unspnetTx.InputValue
		} else {
			tokensBalance[unspnetTx.TokenID] += unspnetTx.InputValue
		}
	}

	return tokensBalance
}

// TODO: get spent and unspent transactions from get_unspent_transactions function
// Todo: Check the fee
func GetTransactionsByPublicKey(accountObj accountdb.AccountStruct) map[string][]jsonTransactions {
	var normalTxs, TokenCreationTxs, RefundedTokenTxs []jsonTransactions
	//var inoTokenID, _ = globalPkg.ConvertIntToFixedLengthString(0, globalPkg.GlobalObj.TokenIDStringFixedLength)

	returnedTransactions := map[string][]jsonTransactions{}

	//fmt.Println("account blocks", accountObj.BlocksLst)

	for _, blockObj := range accountObj.BlocksLst {
		decryptedIndex := cryptogrpghy.KeyDecrypt(globalpkg.EncryptAccount, blockObj)
		blockInfo := block.GetBlockInfoByID(decryptedIndex)
		for _, transactionObj := range blockInfo.BlockTransactions {
			var isSender bool
			var returnedTransaction jsonTransactions
			var inputToken string

			for _, transactionInput := range transactionObj.TransactionInput {
				inputToken = transactionInput.TokenID

				if transactionInput.SenderPublicKey == accountObj.AccountPublicKey {
					isSender = true
					returnedTransaction.Sender = accountObj.AccountName
				} else {
					isSender = false
					returnedTransaction.Sender = account.GetAccountByAccountPubicKey(transactionInput.SenderPublicKey).AccountName
				}
			}
			for _, transactionOutPut := range transactionObj.TransactionOutPut {
				isReceiver := transactionOutPut.RecieverPublicKey == accountObj.AccountPublicKey
				// it's not a fee.
				if !transactionOutPut.IsFee {
					//if he's the sender and receiver (either it's token creation Tx or Refund Tx)
					returnedTransaction.TransactionDate = transactionObj.TransactionTime
					returnedTransaction.Receiver = accountObj.AccountName
					returnedTransaction.Amount += transactionOutPut.OutPutValue
					if transactionObj.Type == 1 {
						if transactionOutPut.TokenID != inputToken {
							returnedTransaction.TokenID = transactionOutPut.TokenID
							TokenCreationTxs = append(TokenCreationTxs, returnedTransaction)
						}
					} else if transactionObj.Type == 2 || transactionObj.Type == 3 {
						if transactionOutPut.TokenID != inputToken {
							returnedTransaction.TokenID = inputToken
							RefundedTokenTxs = append(RefundedTokenTxs, returnedTransaction)
						}
					} else {
						returnedTransaction.TokenID = transactionOutPut.TokenID
						if isSender && !isReceiver {
							returnedTransaction.Receiver = account.GetAccountByAccountPubicKey(transactionOutPut.RecieverPublicKey).AccountName
							normalTxs = append(normalTxs, returnedTransaction)
						} else if !isSender && isReceiver {
							returnedTransaction.Receiver = accountObj.AccountName
							normalTxs = append(normalTxs, returnedTransaction)
						}
					}
				} else if transactionOutPut.RecieverPublicKey == accountObj.AccountPublicKey && transactionOutPut.IsFee {
					// TODO: decide what todo when this condition execute to get balance for the Inovatian account.
				}
			}
		}
	}
	// get transactions from transaction pool.
	for _, transactionObj := range transaction.GetPendingTransactions() {
		var isSender bool
		var returnedTransaction jsonTransactions
		var inputToken string

		for _, transactionInput := range transactionObj.TransactionInput {
			inputToken = transactionInput.TokenID

			if transactionInput.SenderPublicKey == accountObj.AccountPublicKey {
				isSender = true
				returnedTransaction.Sender = accountObj.AccountName
			} else {
				isSender = false
				returnedTransaction.Sender = account.GetAccountByAccountPubicKey(transactionInput.SenderPublicKey).AccountName
			}
		}
		for _, transactionOutPut := range transactionObj.TransactionOutPut {
			isReceiver := transactionOutPut.RecieverPublicKey == accountObj.AccountPublicKey
			// it's not a fee.
			if !transactionOutPut.IsFee {
				//if he's the sender and receiver (either it's token creation Tx or Refund Tx)
				returnedTransaction.TransactionDate = transactionObj.TransactionTime
				returnedTransaction.Receiver = accountObj.AccountName
				returnedTransaction.Amount += transactionOutPut.OutPutValue
				if transactionObj.Type == 1 {
					if transactionOutPut.TokenID != inputToken {
						returnedTransaction.TokenID = transactionOutPut.TokenID
						TokenCreationTxs = append(TokenCreationTxs, returnedTransaction)
					}
				} else if transactionObj.Type == 2 || transactionObj.Type == 3 {
					if transactionOutPut.TokenID != inputToken {
						returnedTransaction.TokenID = inputToken
						RefundedTokenTxs = append(RefundedTokenTxs, returnedTransaction)
					}
				} else {
					returnedTransaction.TokenID = transactionOutPut.TokenID
					if isSender && !isReceiver {
						returnedTransaction.Receiver = account.GetAccountByAccountPubicKey(transactionOutPut.RecieverPublicKey).AccountName
						normalTxs = append(normalTxs, returnedTransaction)
					} else if !isSender && isReceiver {
						returnedTransaction.Receiver = accountObj.AccountName
						normalTxs = append(normalTxs, returnedTransaction)
					}
				}
			} else if transactionOutPut.RecieverPublicKey == accountObj.AccountPublicKey && transactionOutPut.IsFee {
				// TODO: decide what todo when this condition execute to get balance for the Inovatian account.
			}
		}
	}
	returnedTransactions["normal"] = normalTxs
	returnedTransactions["refunded"] = RefundedTokenTxs
	returnedTransactions["tokenCreation"] = TokenCreationTxs
	return returnedTransactions
}

func GetTransactionsByTokenID(accountObj accountdb.AccountStruct, tokenID string) map[string][]jsonTransactions {
	allTransactions := GetTransactionsByPublicKey(accountObj)
	returnedTxs := map[string][]jsonTransactions{}
	for key, transactionLst := range allTransactions {
		tmp := transactionLst[:0]
		for _, transactionObj := range transactionLst {
			if transactionObj.TokenID == tokenID {
				tmp = append(tmp, transactionObj)
			}
		}
		returnedTxs[key] = tmp
	}
	return returnedTxs
}

// TODO: get spent and unspent transactions from get_unspent_transactions function. (tried and Totally FAILED)
// Todo: check the fee for the transaction and make changes to correctly calcualte the balance.
func GetAccountBalanceStatement(accountObj accountdb.AccountStruct, tokenID string) map[string]*jsonAccountBalanceStatement {
	var inoTokenID, _ = globalPkg.ConvertIntToFixedLengthString(0, globalPkg.GlobalObj.TokenIDStringFixedLength)
	var allTransactions map[string][]jsonTransactions
	if tokenID != "" {
		allTransactions = GetTransactionsByTokenID(accountObj, tokenID)
	} else {
		allTransactions = GetTransactionsByPublicKey(accountObj)
	}
	tokensBalance := map[string]*jsonAccountBalanceStatement{}
	//fmt.Println("allTransactions", allTransactions)
	for key, transactionLst := range allTransactions {
		for _, transactionObj := range transactionLst {
			_, tokenExist := tokensBalance[transactionObj.TokenID]
			if !tokenExist {
				tokensBalance[transactionObj.TokenID] = &jsonAccountBalanceStatement{0.0, 0.0, 0.0}
			}
			tokenObj := token.FindTokenByid(transactionObj.TokenID)
			if transactionObj.Sender != transactionObj.Receiver {
				if transactionObj.Sender == accountObj.AccountName {
					tokensBalance[transactionObj.TokenID].TotalSent += transactionObj.Amount // + float32(tokenObj.TokenValue) * globalPkg.GlobalObj.TransactionFee
				} else if transactionObj.Receiver == accountObj.AccountName {
					tokensBalance[transactionObj.TokenID].TotalReceived += transactionObj.Amount
				}
			} else {
				if key == "tokenCreation" {
					tokensBalance[inoTokenID].TotalSent += (transactionObj.Amount * tokenObj.TokenValue) / globalPkg.GlobalObj.InoCoinToDollarRatio // + globalPkg.GlobalObj.TransactionFee
					fmt.Println("\n *** tokensBalance[inoTokenID].TotalSent: ", tokensBalance[inoTokenID].TotalSent)
					tokensBalance[transactionObj.TokenID].TotalReceived += transactionObj.Amount
				} else if key == "refunded" {
					//amountInToken := (transactionObj.Amount * float32(tokenObj.TokenValue))
					tokensBalance[inoTokenID].TotalReceived += transactionObj.Amount                               // transactionObj.Amount * float32(tokenObj.TokenValue)
					tokensBalance[transactionObj.TokenID].TotalSent += transactionObj.Amount / tokenObj.TokenValue // - ( amountInToken * globalPkg.GlobalObj.TransactionRefundFee )
				}
			}
		}
	}
	for _, tokenBalance := range tokensBalance {
		if tokenBalance.TotalReceived != 0 && tokenBalance.TotalSent != 0 {
			tokenBalance.TotalBalance = tokenBalance.TotalReceived - tokenBalance.TotalSent
		}
		if tokenBalance.TotalReceived != 0 && tokenBalance.TotalSent == 0 {
			tokenBalance.TotalBalance = tokenBalance.TotalReceived
		}
	}
	return tokensBalance
}

// GetTokenIDusedbyrecieverpk  get token ids that public key one of its holder
func GetTokenIDusedbyrecieverpk(publickey string) []string {
	blocklist := block.GetBlockchain()
	var tokenidoutput []string
	for _, blockObj := range blocklist {
		for _, transactionObj := range blockObj.BlockTransactions {
			for _, transactionOutPutObj := range transactionObj.TransactionOutPut {
				if publickey == transactionOutPutObj.RecieverPublicKey {
					tokenidoutput = append(tokenidoutput, transactionOutPutObj.TokenID)
				}
			}
		}
	}
	return tokenidoutput
}

//CheckBalance Check User Balance
func CheckBalance(amount float64, publickey string, tokenid string) (string, bool) {
	balance := GetAccountBalance(publickey)
	tokenBalanceVal, exist := balance[tokenid]
	if !exist || amount > tokenBalanceVal {
		errorFound := "You do not have balance for the token with ID of " + tokenid
		return errorFound, false
	}
	return "", true
}
