package transactionModule

import (
	"rsapk"

	"../account"
	"../accountdb"
	"../cryptogrpghy"
	"../globalPkg"
	"../service"
	"../token"
	"../transaction"

	"fmt"
	"strings"
	"time"
)

// ValidateTransaction
func ValidateTransaction(digitalwalletTransactionObj transaction.DigitalwalletTransaction) string {
	firstaccount := accountdb.GetFirstAccount()
	layout2 := "15:04:05"

	noow := time.Now().Format(layout2)
	t, _ := time.Parse(noow, noow)
	fmt.Println("the validation time", t)
	fmt.Print("digitalwalletTransactionObj.Time.UTC()", digitalwalletTransactionObj.Time.UTC())
	// time differnce between the received digital wallet transaction time and the server's time.
	timeDifference := t.UTC().Sub(digitalwalletTransactionObj.Time.UTC()).Seconds()
	fmt.Println("*************** transaction validation Time ************", timeDifference)
	if timeDifference > float64(globalPkg.GlobalObj.TxValidationTimeInSeconds) {
		return "please check your transaction's time"
	}

	var publickey *rsapk.PublicKey
	//sender is null and the reciever is inovatian	firstaccount:= account.GetFirstAccount()
	if digitalwalletTransactionObj.Sender == "" && firstaccount.AccountPublicKey == digitalwalletTransactionObj.Receiver {
		if !checkIfAccountIsActive(digitalwalletTransactionObj.Receiver) {
			return "please, check the receiver account should be active"
		}
		receiverPK := account.FindpkByAddress(digitalwalletTransactionObj.Receiver).Publickey
		publickey = cryptogrpghy.ParsePEMtoRSApublicKey(receiverPK)
	} else {
		tokensBalance := GetAccountBalance(digitalwalletTransactionObj.Sender)
		tokenBalanceVal, exist := tokensBalance[digitalwalletTransactionObj.TokenID]
		if !exist {
			return "You do not have balance for the token with ID of " + digitalwalletTransactionObj.TokenID
		}
		if checkIfAccountIsActive(digitalwalletTransactionObj.Receiver) && checkIfAccountIsActive(digitalwalletTransactionObj.Sender) {
			if tokenBalanceVal <= digitalwalletTransactionObj.Amount+globalPkg.GlobalObj.TransactionFee {
				return "please, check that you have more balance for the token with ID of " + digitalwalletTransactionObj.TokenID
			}
		} else {
			return "please, check the accounts they should be active"
		}
		senderPK := account.FindpkByAddress(digitalwalletTransactionObj.Sender).Publickey
		publickey = cryptogrpghy.ParsePEMtoRSApublicKey(senderPK)
	}

	fmt.Println("digitalWalletTx time", digitalwalletTransactionObj.Time.String())
	signatureData := digitalwalletTransactionObj.Sender + digitalwalletTransactionObj.Receiver +
		fmt.Sprintf("%f", digitalwalletTransactionObj.Amount) + digitalwalletTransactionObj.Time.UTC().Format("2006-01-02T03:04:05+00:00")
	fmt.Println("publickey", publickey)
	fmt.Println("signatureData :", signatureData)
	validSig := cryptogrpghy.VerifyPKCS1v15(digitalwalletTransactionObj.Signature, signatureData, *publickey)

	if validSig {
		return ""
		} else if !validSig {
			return ""
	} else {
		return "You are not allowed to do this transaction"
	}
}

func ValidateRefundTransaction(digitalWalletTx transaction.RefundDigitalWalletTx) string {

	if digitalWalletTx.Amount < 1 {
		return "please make sure amount is more than zero"
	}
	if digitalWalletTx.Sender == "" || digitalWalletTx.Time.IsZero() || digitalWalletTx.TokenID == "" ||
		digitalWalletTx.Amount == 0.0 || digitalWalletTx.Signature == "" {
		return "please enter required data Sender PK, Receiver PK, TokenID, Amount, Time and Signature"
	}
	if digitalWalletTx.FlatCurrency && digitalWalletTx.Receiver == "" {
		return "please enter the receiver public key in order to refund flat currency"
	}
	if len(digitalWalletTx.TokenID) < globalPkg.GlobalObj.TokenIDStringFixedLength || len(digitalWalletTx.TokenID) > globalPkg.GlobalObj.TokenIDStringFixedLength {
		return "token ID must be equal to 100 characters"
	}
	if len(digitalWalletTx.Signature) < 100 && len(digitalWalletTx.Signature) >= 200 {
		return "Sender sign must be less than 200 and more than 100 characters"
	}
	layout2 := "15:04:05"

	noow := time.Now().Format(layout2)
	t, _ := time.Parse(noow, noow)
	// time differnce between the received digital wallet transaction time and the server's time.
	timeDifference := t.UTC().Sub(digitalWalletTx.Time.UTC()).Seconds()
	if timeDifference > float64(globalPkg.GlobalObj.TxValidationTimeInSeconds) {
		return "please check your transaction's time"
	}
	tokensBalance := GetAccountBalance(digitalWalletTx.Sender)
	tokenBalanceVal, exist := tokensBalance[digitalWalletTx.TokenID]
	if !exist {
		return "You do not have balance for the token with ID of " + digitalWalletTx.TokenID
	}
	decimals := strings.Split(fmt.Sprintf("%f", digitalWalletTx.Amount), ".")[1]
	if len(decimals) > 6 {
		return "Please check the transaction's amount decimals, maximum number of decimals is 6."
	}
	if checkIfAccountIsActive(digitalWalletTx.Sender) {
		if digitalWalletTx.FlatCurrency {
			if !checkIfAccountIsActive(digitalWalletTx.Receiver) {
				return "the receiver's account is not active."
			}
			refundTokenValue := token.FindTokenByid(digitalWalletTx.TokenID).TokenValue
			// convert the refunded token amount to inoToken amount.
			toInoToken := digitalWalletTx.Amount * refundTokenValue
			// convert the refund value (in InoToken value) to dollar value.
			refundValue := toInoToken * globalPkg.GlobalObj.InoCoinToDollarRatio
			fmt.Println("\n toInoToken:", toInoToken)
			fmt.Println("\n refundValue:", refundValue)
			if refundValue < toInoToken {
				inoTokenID, _ := globalPkg.ConvertIntToFixedLengthString(0, globalPkg.GlobalObj.TokenIDStringFixedLength)
				receiverTokensBalance := GetAccountBalance(digitalWalletTx.Receiver)
				receiverTokenBalanceVal, exist := receiverTokensBalance[inoTokenID]
				inoTokenName := token.FindTokenByid(inoTokenID).TokenName
				if !exist {
					return "the receiver's account doesn't have balance and fees for the " + inoTokenName
				}
				var amountOnReceiver float64
				if toInoToken-refundValue < 0 {
					amountOnReceiver = (toInoToken - refundValue) * -1
				} else {
					amountOnReceiver = toInoToken - refundValue
				}
				amountOnReceiver += globalPkg.GlobalObj.TransactionFee
				fmt.Println("\n receiverTokenBalanceVal:", receiverTokenBalanceVal)
				fmt.Println("\n globalPkg.GlobalObj.TransactionFee:", globalPkg.GlobalObj.TransactionFee)
				if receiverTokenBalanceVal <= amountOnReceiver {
					return "the receiver's account must have more balance for the " + inoTokenName
				}
			}
		}

		if tokenBalanceVal >= digitalWalletTx.Amount {
			senderPK := account.FindpkByAddress(digitalWalletTx.Sender).Publickey
			public_key := cryptogrpghy.ParsePEMtoRSApublicKey(senderPK)

			fmt.Println("digitalWalletTx time", digitalWalletTx.Time.String())
			signatureData := digitalWalletTx.Sender + fmt.Sprintf("%f", digitalWalletTx.Amount) +
				digitalWalletTx.Time.UTC().Format("2006-01-02T03:04:05+00:00") + digitalWalletTx.TokenID

			validSig := cryptogrpghy.VerifyPKCS1v15(digitalWalletTx.Signature, signatureData, *public_key)

			if validSig {
				return ""
				// } else if !validSig {
				// 	return ""
			} else {
				return "You are not allowed to do this transaction"
			}
		} else {
			return "please, check that you have more balance for the token with ID of " + digitalWalletTx.TokenID
		}
	} else {
		return "please, check the accounts they should be active"
	}
}

//validateTransactionToken validation token for transfer transaction token
func ValidateTransactionToken(tokenTransactionObj transaction.DigitalwalletTransaction) (string, bool) {

	var errorfound string

	if len(tokenTransactionObj.TokenID) != globalPkg.GlobalObj.TokenIDStringFixedLength {
		errorfound = fmt.Sprintf("token ID must be equal to %v characters", globalPkg.GlobalObj.TokenIDStringFixedLength)
		return errorfound, false
	}
	if tokenTransactionObj.Amount < 1 {
		return "please make sure amount is more than zero", false
	}
	//validate fields not fields
	if tokenTransactionObj.Receiver == "" || tokenTransactionObj.TokenID == "" ||
		tokenTransactionObj.Amount == 0.0 || tokenTransactionObj.Signature == "" || tokenTransactionObj.Time.IsZero() {
		errorfound = "please enter required data Sender PK, Receiver PK, TokenID,Amount , Sender, Signature"
		return errorfound, false
	}

	firstaccount := accountdb.GetFirstAccount()
	if tokenTransactionObj.Sender == "" && firstaccount.AccountPublicKey != tokenTransactionObj.Receiver {
		return "Please check the sender address", false
	}

	if tokenTransactionObj.Receiver == tokenTransactionObj.Sender {
		errorfound = "Sender public key can't be the same as Receiver public key"
		return errorfound, false
	}
	//validate send signature < 200
	if len(tokenTransactionObj.Signature) < 100 && len(tokenTransactionObj.Signature) >= 200 {
		errorfound = "Sender sign must be more than 100 and less than 200 characters"
		return errorfound, false
	}

	tokenObj := token.FindTokenByid(tokenTransactionObj.TokenID)
	receiverExist := false // check for reciver exist in list of user PK

	//check on token type is private type that reciever pk allowed to  use it and exist in userPKs array.
	if tokenObj.TokenType == "private" {
		for _, uPK := range tokenObj.UserPublicKey {
			if tokenTransactionObj.Receiver == uPK {
				receiverExist = true
			}
		}
	} else {
		receiverExist = true
	}
	if receiverExist == false {
		errorfound = "the receiver is not  allowed to use this token"
		return errorfound, false
	}
	decimals := strings.Split(fmt.Sprintf("%f", tokenTransactionObj.Amount), ".")[1]
	if len(decimals) > 6 {
		return "Please check the transaction's amount decimals, maximum number of decimals is 6.", false
	}

	errorfound = ValidateTransaction(tokenTransactionObj)
	if errorfound != "" {
		return errorfound, false
	}

	return "", true
}

/********validate service transaction*******/
func ValidateServiceTransaction(digitalwalletTransactionObj transaction.DigitalwalletTransaction) string {
	// time differnce between the received digital wallet transaction time and the server's time.
	layout2 := "15:04:05"

	noow := time.Now().Format(layout2)
	t, _ := time.Parse(noow, noow)
	timeDifference := t.UTC().Sub(digitalwalletTransactionObj.Time.UTC()).Seconds()
	if timeDifference > float64(globalPkg.GlobalObj.TxValidationTimeInSeconds) {
		return "please check your transaction's time"
	}
	//  service1  :=Service.ServiceStruct{ID: "ServiceId", Mbytes : true}

	// Service.ServiceCreateOUpdate(service1)
	tokensBalance := GetAccountBalance(digitalwalletTransactionObj.Sender)
	tokenBalanceVal, exist := tokensBalance[digitalwalletTransactionObj.TokenID]
	if !exist {
		return "You do not have balance for the token with ID of " + digitalwalletTransactionObj.TokenID
	}
	if checkIfAccountIsActive(digitalwalletTransactionObj.Receiver) && checkIfAccountIsActive(digitalwalletTransactionObj.Sender) {
		if tokenBalanceVal >= digitalwalletTransactionObj.Amount {
			senderPK := account.FindpkByAddress(digitalwalletTransactionObj.Sender).Publickey
			public_key := cryptogrpghy.ParsePEMtoRSApublicKey(senderPK)

			fmt.Println("digitalWalletTx time", digitalwalletTransactionObj.Time.String())
			signatureData := digitalwalletTransactionObj.Sender + digitalwalletTransactionObj.Receiver +
				fmt.Sprintf("%f", digitalwalletTransactionObj.Amount) + digitalwalletTransactionObj.Time.UTC().Format("2006-01-02T03:04:05+00:00")

			validSig := cryptogrpghy.VerifyPKCS1v15(digitalwalletTransactionObj.Signature, signatureData, *public_key)
			servicevalid := false
			reciveracc := account.GetAccountByAccountPubicKey(digitalwalletTransactionObj.Receiver)
			fmt.Println("digitalwalletTransactionObj.WalletService=-=--=--=-=-=-=-=-=-=", digitalwalletTransactionObj.ServiceId)
			if digitalwalletTransactionObj.ServiceId == "" {
				if reciveracc.AccountRole != "service" {
					fmt.Println(" reciveracc.AccountRole=-=-=-=-=-=-==-=-=--", reciveracc.AccountRole)
					servicevalid = true
				} else {
					fmt.Println(" reciveracc.AccountRole=-=-=-=-=-=-==-=-=--", reciveracc.AccountRole)
					servicevalid = false
				}
			} else {
				if reciveracc.AccountRole == "service" {

					serviceobj := service.GetAllservice()
					pkflag := false
					idflag := false
					for _, obj := range serviceobj {
						if obj.PublicKey == digitalwalletTransactionObj.Sender {
							pkflag = true
							if obj.ID == digitalwalletTransactionObj.ServiceId {
								idflag = true

								obj = service.CalculateAmountAndCost(obj)
								cost := obj.Calculation + globalPkg.GlobalObj.TransactionFee
								if cost == digitalwalletTransactionObj.Amount {
									servicevalid = true
									break
								} else {
									return "this amount is not valid "
								}
							} else {
								idflag = false
							}
						} else {
							pkflag = false
						}
					}
					if pkflag == false {
						servicevalid = false
						return "invalid public key "
					}
					if idflag == false {
						servicevalid = false
						return "invalid service id"
					}
					//check the calculate of the bytes = amount
				} else {
					servicevalid = false
					return "only service account is allow to make this operation"
				}
			}
			// fmt.Println("servicevalid=-=-=-=-=-=-=-=-=-=-",servicevalid)
			if validSig && servicevalid {
				return ""
				// } else if !validSig {
				// 	return ""
			} else if !servicevalid {
				return "service is not valid , service account must insert service data"
			} else {
				return "You are not allowed to do this transaction"
			}
		} else {
			return "please, check that you have more balance for the token with ID of " + digitalwalletTransactionObj.TokenID
		}
	} else {
		return "please, check the accounts they should be active"
	}
}
