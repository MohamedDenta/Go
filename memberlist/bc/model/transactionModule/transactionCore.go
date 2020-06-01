package transactionModule

import (
	"../account"
	"../accountdb"
	"../block"
	"../cryptogrpghy"
	"../globalPkg"
	globalpkg "../globalPkg"
	"../token"
	"../transaction"

	"fmt"
	"sort"
	"time"
)

type jsonTransactions struct {
	TransactionDate time.Time
	Sender          string
	Receiver        string
	TokenID         string
	Amount          float64
}

type jsonAccountBalanceStatement struct {
	TotalReceived float64
	TotalSent     float64
	TotalBalance  float64
}

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

	//fmt.Println("\n \n ****************** the accountObj of inov:", accountObj)

	transactionPool := transaction.GetPendingTransactions()
	// loop over block list
	for _, blockObj := range accountObj.BlocksLst {
		blockObj = cryptogrpghy.KeyDecrypt(globalpkg.EncryptAccount, blockObj)
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
	usinputs, _ := GetUnspentAndSpentTxs(senderPK)

	for _, usinputObj := range usinputs {
		if usinputObj.TokenID == tokenID && usinputObj.InputValue == amountWithFee {
			tx.TransactionInput = append(
				tx.TransactionInput, transaction.TXInput{
					InputID: usinputObj.InputID, InputValue: usinputObj.InputValue,
					SenderPublicKey: senderPK, TokenID: tokenID,
				},
			)
			return
		}
	}
	// sort the unspent spent inputs ascending by InputValue
	sort.SliceStable(usinputs, func(k, j int) bool {
		return usinputs[k].InputValue < usinputs[j].InputValue
	},
	)
	var sum float64

	for _, unspentInput := range usinputs {
		if unspentInput.TokenID == tokenID {
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
		}
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
func DigitalwalletToUTXOTrans(digitalWalletTx transaction.DigitalwalletTransaction) transaction.Transaction {

	var transactionobj transaction.Transaction
	var inoTokenID, _ = globalPkg.ConvertIntToFixedLengthString(0, globalPkg.GlobalObj.TokenIDStringFixedLength)
	var feesAccIndex, _ = globalPkg.ConvertIntToFixedLengthString(1, globalPkg.GlobalObj.StringFixedLength)
	var feesAccPublicKey = account.GetAccountByIndex(feesAccIndex).AccountPublicKey
	transactionobj.ServiceId = digitalWalletTx.ServiceId

	feeValue := globalPkg.GlobalObj.TransactionFee
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
		transactionobj.TransactionOutPut = append(
			transactionobj.TransactionOutPut, transaction.TXOutput{
				OutPutValue: feeValue, IsFee: true, TokenID: digitalWalletTx.TokenID,
				RecieverPublicKey: feesAccPublicKey,
			},
		)
	}

	//
	makeTxInputs(&transactionobj, digitalWalletTx.TokenID, digitalWalletTx.Sender, amountWithFee)

	fmt.Println("\n @)@))@)@) Core transactionobj", transactionobj)

	transactionobj.Type = 0
	transactionobj.TransactionID = ""
	transactionobj.TransactionTime = globalPkg.UTCtime()
	transactionobj.TransactionID = globalPkg.CreateHash(transactionobj.TransactionTime, fmt.Sprintf("%s", transactionobj), 3)

	return transactionobj
}

// CreateTokenTx
func CreateTokenTx(token token.StructToken, senderAmount float64, senderTokenID string) transaction.Transaction {

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
	var amountWithFee = senderAmount + globalPkg.GlobalObj.TransactionFee

	makeTxInputs(&transactionobj, inoTokenID, token.InitiatorAddress, amountWithFee)

	transactionobj.TransactionTime = globalPkg.UTCtime()
	transactionobj.Type = 1
	transactionobj.TransactionID = ""
	transactionobj.TransactionID = globalPkg.CreateHash(transactionobj.TransactionTime, fmt.Sprintf("%s", transactionobj), 3)

	return transactionobj
}

// TODO: the user wants to refund for example 500 of his own token. at first it will be calculated by the tokenValue and

func RefundTokenTx(token token.StructToken, refundDwTx transaction.RefundDigitalWalletTx) transaction.Transaction {
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
			var tx transaction.DigitalwalletTransaction
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
			var tx transaction.DigitalwalletTransaction
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
	transactionobj.TransactionID = ""
	transactionobj.TransactionID = globalPkg.CreateHash(transactionobj.TransactionTime, fmt.Sprintf("%s", transactionobj), 3)

	return transactionobj
}

func addcoins(digitalwalletObj transaction.DigitalwalletTransaction) transaction.Transaction {
	fmt.Println("increase balance ----------------")
	firstTokenID, _ := globalPkg.ConvertIntToFixedLengthString(0, globalPkg.GlobalObj.TokenIDStringFixedLength)
	// token.TokenCreate(
	// 	token.StructToken{
	// 		TokenID: firstTokenID, TokensTotalSupply: float32(conf.Server.InitialMinerCoins), TokenName: "InoToken",
	// 	},
	// )
	firstaccount := accountdb.GetFirstAccount()
	var transactionObj transaction.Transaction
	transactionObj.TransactionTime, _ = time.Parse("2006-01-02 03:04:05 PM -0000", time.Now().UTC().Format("2006-01-02 03:04:05 PM -0000"))
	transactionObj.TransactionOutPut = append(transactionObj.TransactionOutPut, transaction.TXOutput{
		OutPutValue: digitalwalletObj.Amount, RecieverPublicKey: firstaccount.AccountPublicKey,
		TokenID: firstTokenID,
	})
	// txIdHash := cryptogrpghy.ClacHash([]byte(accountObj.AccountPublicKey + strconv.FormatFloat(conf.Server.InitialMinerCoins, 'f', 6, 64) + transactionObj.TransactionTime.String()))
	// transactionObj.TransactionID = hex.EncodeToString(txIdHash[:])
	// transaction.AddTransaction(transactionObj)
	transactionObj.TransactionID = ""
	transactionObj.TransactionID = globalPkg.CreateHash(transactionObj.TransactionTime, fmt.Sprintf("%s", transactionObj), 3)
	fmt.Println("======================   **    =============================")
	// transaction.AddTransaction(transactionObj)
	fmt.Println("======================   **    =============================")
	return transactionObj
}
