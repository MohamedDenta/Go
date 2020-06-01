package transactionModule

import (
	"bytes"

	"../account"
	"../accountdb"
	"../broadcastTcp"
	"../cryptogrpghy"
	"../errorpk"
	"../transaction"
	"../validator"
	"../token"
	"../globalPkg"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

func ValidateTx2(digitalWalletTx transaction.DigitalwalletTransaction, tx transaction.Transaction) string {
	if errStr := ValidateTransaction(digitalWalletTx); errStr == "" {
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

		// outputSum := 0.0
		// for _, outputObj := range tx.TransactionOutPut {
		// 	outputSum += outputObj.OutPutValue
		// }
		// inputSum := 0.0
		// for _, inputObj := range tx.TransactionInput {
		// 	inputSum += inputObj.InputValue
		// }
		fmt.Println("outputSum--------", outputSum)
		fmt.Println("inputSum--------", inputSum)

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
				return "true"
			} else {
				errorpk.AddError("ValidateTx2 Transaction module", "input is exist", "Validation Error")
				return "input is exist"
			}

		} else {
			errorpk.AddError("ValidateTx2 Transaction module", "digitalWalletTx is rong", "Validation Error")
			return "digitalWalletTx is rong"
		}

	} else {
		errorpk.AddError("ValidateTx2 Transaction module", errStr, "Validation Error")
		return errStr
	}
}
// AddTxToTransactionPool
// TODO: benchmark the function.
func ValidateTx(digitalWalletTx transaction.DigitalwalletTransaction, tx transaction.Transaction) string {
	if errStr := ValidateTransaction(digitalWalletTx); errStr == "" {
		unspentTxInputs, _ := GetUnspentAndSpentTxs(digitalWalletTx.Sender)
		// make a map of the unspent inputs Ids.
		unspentTxInputsIds := make(map[string]struct{}, len(unspentTxInputs))
		for _, unspentTxInput := range unspentTxInputs {
			unspentTxInputsIds[unspentTxInput.InputID] = struct{}{}
		}
		// check to see if every Tx Input id is in the mapped unspent inputs ids.
		var validTxInputs bool
		for _, txInput := range tx.TransactionInput {
			if _, ok := unspentTxInputsIds[txInput.InputID]; ok {
				validTxInputs = ok
			}
		}
		// if it's a valid Tx inputs, then check the amount of inputs == outputs.
		if validTxInputs {
			var inputSum, outputSum, transactionAmount float64
			for _, txInput := range tx.TransactionInput {
				inputSum += txInput.InputValue
			}
			for _, txOutput := range tx.TransactionOutPut {
				outputSum += txOutput.OutPutValue
				if txOutput.RecieverPublicKey == digitalWalletTx.Receiver {
					transactionAmount += txOutput.OutPutValue
				}
			}
			// if it's equal, then check if the Tx sender & receiver PubKeys are the same in digitalWalletTx.
			if inputSum == outputSum && transactionAmount == digitalWalletTx.Amount {
				var validSender, validReceiver bool
				for _, txInput := range tx.TransactionInput {
					if txInput.SenderPublicKey == digitalWalletTx.Sender {
						validSender = true
						break
					}
				}
				for _, txOutput := range tx.TransactionOutPut {
					if txOutput.RecieverPublicKey == digitalWalletTx.Receiver {
						validReceiver = true
						break
					}
				}
				if validReceiver && validSender {
					//transaction.AddTransaction(tx)
					return "true"
				} else {
					var note = "Tx receiver & sender public keys doesn't match digitalWalletTx receiver & sender public keys"
					errorpk.AddError("ValidateTx Transaction module", note, "Validation Error")
					return note
				}
			} else {
				var note = "Tx inputs isn't equal to the Tx outpus"
				errorpk.AddError("ValidateTx Transaction module", note, "Validation Error")
				return note
			}
		} else {
			var note = "Tx inputs doesn't match with the unspent Txs Inputs"
			errorpk.AddError("ValidateTx Transaction module", note, "Validation Error")
			return note
		}
	} else {

		errorpk.AddError("ValidateTx Transaction module", errStr, "Validation Error")
		return errStr
	}
}

// checkIfAccountIsActive return the AccountStatus of the account with publicKey, else returns false.
func checkIfAccountIsActive(publicKey string) bool {

	fmt.Println("account.GetAccountByAccountPubicKey(publicKey))", account.GetAccountByAccountPubicKey(publicKey).AccountStatus)
	fmt.Println("-------------------------------------------------------")
	if (account.GetAccountByAccountPubicKey(publicKey)).AccountPublicKey != "" {
		return (account.GetAccountByAccountPubicKey(publicKey)).AccountStatus
	} else {
		return false
	}
}

// DecryptDigitalWalletTx takes a string that consists of first 172 characters are encrypted key and the rest is the
// encrypted data. key is encrypted with RSA, data is encrypted with AES. it will first decrypt the key then decrypt the
// data using this key. and return transaction.DigitalwalletTransaction object.
func DecryptDigitalWalletTx(encryptedDigitalWalletTx string) (transaction.DigitalwalletTransaction, error) {

	if len(encryptedDigitalWalletTx) > 172 {
		encryptedKey := encryptedDigitalWalletTx[:172]
		fmt.Println("encryptedKey: ", encryptedKey)
		encryptedData := encryptedDigitalWalletTx[172:]
		fmt.Println("encryptedData: ", encryptedData)

		// hashedKey := string(cryptogrpghy.RSADEC(validator.CurrentValidator.ValidatorPrivateKey, encryptedKey)) // fmt.Sprintf("%x", cryptogrpghy.RSADEC(validator.CurrentValidator.ValidatorPrivateKey, encryptedKey))

		hashedKey, err := cryptogrpghy.Decrypt(validator.CurrentValidator.ValidatorPublicKey, validator.CurrentValidator.ValidatorPrivateKey, encryptedKey)
		if err != nil {
			return transaction.DigitalwalletTransaction{}, err
		}
		fmt.Println("hashedKey :", hashedKey, "lenght :", len(hashedKey))
		encodedData := ""
		// encodedData := cryptogrpghy.KeyDecrypt(hashedKey, encryptedData)
		if len(hashedKey) < 40 {
			encodedData, err = cryptogrpghy.FAESDecrypt(encryptedData, hashedKey)
		} else {
			encodedData, err = cryptogrpghy.AESDecrypt(encryptedData, hashedKey)
		}

		if err != nil {
			fmt.Println("err :", err)
		}

		fmt.Println(encodedData)

		var data transaction.DigitalwalletTransaction

		if err = json.Unmarshal([]byte(encodedData), &data); err != nil {
			fmt.Println("err json.Unmarshal", err)
			return transaction.DigitalwalletTransaction{}, errors.New("please enter a valid transaction data")
		}
		fmt.Println("transaction data :", data)
		return data, nil
	}
	return transaction.DigitalwalletTransaction{}, errors.New("please enter a valid encrypted string")
}

// getTxInputs gets the outputs that have ReceiverPublicKey = PubKey and convert them to inputs(isn't checked yet to see
//  if it's a spent or an unspent input). then gets the inputs and check if SenderPublicKey = PubKey so it will be a spent input.
func getTxInputs(tx transaction.Transaction, PubKey string) (spentTxs, unspentTxs []transaction.TXInput) {
	for _, transactionOutPutObj := range tx.TransactionOutPut {
		if transactionOutPutObj.RecieverPublicKey != PubKey && transactionOutPutObj.IsFee {
			continue
		} else if transactionOutPutObj.RecieverPublicKey == PubKey {
			unspentTxs = append(unspentTxs, transaction.TXInput{
				InputID: tx.TransactionID, InputValue: transactionOutPutObj.OutPutValue,
				SenderPublicKey: transactionOutPutObj.RecieverPublicKey, TokenID: transactionOutPutObj.TokenID,
			})
		}
	}
	for _, transactionInputObj := range tx.TransactionInput {
		if transactionInputObj.SenderPublicKey == PubKey {
			spentTxs = append(spentTxs, transactionInputObj)
		}
	}
	return spentTxs, unspentTxs
}

// transfer remainder (difference in value) of refund Tx between Inovatian and service accounts, either it's a profit or loss for Inovatian.
func transferRefundRemaineder(tx transaction.DigitalwalletTransaction, profit bool) {
	if profit {
		tx.Receiver = accountdb.GetFirstAccount().AccountPublicKey
	} else {
		tx.Sender = accountdb.GetFirstAccount().AccountPublicKey
	}
	tx.Time = time.Now()
	transactionObj := DigitalwalletToUTXOTrans(tx)

	broadcastTcp.BoardcastingTCP(transactionObj, "addTokenTransaction", "transaction")
	fmt.Println("transferRefundRemaineder transactionObj", transactionObj)
}
