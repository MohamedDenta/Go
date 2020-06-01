package transactionValidation

import (
	"../token"
	"bytes"
	"encoding/json"
	"fmt"

	"../accountdb"
	"../errorpk"
	"../globalPkg"
	"../transaction"
	"../transactionModule"
)

func ValidateTx2(digitalWalletTx transaction.DigitalwalletTransaction, tx transaction.Transaction) string {
	return "true"
	fmt.Println("\n the broadcast handle transaction.Transaction:", tx)

	if errStr := transactionModule.ValidateTransaction(digitalWalletTx); errStr == "" {
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
			fmt.Println("-------------------------------------------------------outputSum", outputSum)
			fmt.Println("-------------------------------------------------------inputSum", inputSum)

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

						hashtransaction := globalPkg.CreateHash(tx.TransactionTime, fmt.Sprintf("%s", TransactionObj2), 3)
						fmt.Println(" ----- transobj add    ****   ", tx)
						fmt.Println("  _______   ^  hash  ^    ______", hashtransaction)

						if tx.TransactionID == hashtransaction {
							return "true"
						} else {
							return "hash not equal"
						}
					} else {
						transaction.DeleteTransaction(tx)
						return "there's a double spent transaction"
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
