package transactionModule

import (
	"rsapk"
	"unicode/utf8"

	"../account"
	"../accountdb"
	"../cryptogrpghy"
	"../globalPkg"
	"../responses"
	"../service"
	"../token"
	"../transaction"

	// "../transaction"

	"fmt"
	"strings"
	"time"
)

//AccountPasswordAndPubKey check pk,password
type AccountPasswordAndPubKey struct {
	Password  string
	PublicKey string
}

//SendData receiver , amount
type SendData struct {
	ReceiverName string
	Amount       int
}

//fltr notification omer
type fltr struct {
	Str string
	Lst string
}

//Notify session id , message
type Notify struct {
	SessionID string
	Message   string
}

//publickey ownership
type publickey struct {
	Publickey string
}

//MixedOwnership mixed token id , transaction object
type MixedOwnership struct {
	TokenID     string
	Transaction DigitalwalletTransaction
}

//OwnershipTX ownship ship tx
type OwnershipTX struct {
	Token       token.StructToken
	Transaction DigitalwalletTransaction
}

//DigitalwalletTransaction digital wallet transaction
type DigitalwalletTransaction struct {
	Validator string
	Sender    string
	Receiver  string
	TokenID   string
	Amount    float64
	Time      time.Time
	Signature string
	ServiceId string
}

//MixedTxStruct mixed of transaction , digital tx
type MixedTxStruct struct {
	TxObj        transaction.Transaction
	DigitalTxObj DigitalwalletTransaction
}

//RefundDigitalWalletTx refund digital wallet tx
type RefundDigitalWalletTx struct {
	Sender       string
	Receiver     string
	TokenID      string
	FlatCurrency bool
	Amount       float64
	Time         time.Time
	Signature    string
}

// ValidateTransaction validation tx data
func (digitalwalletTransactionObj DigitalwalletTransaction) ValidateTransaction() string {

	firstaccount := accountdb.GetFirstAccount()

	// time differnce between the received digital wallet transaction time and the server's time.
	timeDifference := time.Now().UTC().Sub(digitalwalletTransactionObj.Time.UTC()).Seconds()
	// fmt.Println("*************** transaction validation Time ************", timeDifference)
	if timeDifference > float64(globalPkg.GlobalObj.TxValidationTimeInSeconds) {
		responseObj := responses.FindResponseByID("124")
		return responseObj.EngResponse
	}

	var publickey *rsapk.PublicKey
	//sender is null and the reciever is inovatian	firstaccount:= account.GetFirstAccount()
	if digitalwalletTransactionObj.Sender == "" && firstaccount.AccountPublicKey == digitalwalletTransactionObj.Receiver {
		if !checkIfAccountIsActive(digitalwalletTransactionObj.Receiver) {
			responseObj := responses.FindResponseByID("125")
			return responseObj.EngResponse

		}
		receiverPK := account.FindpkByAddress(digitalwalletTransactionObj.Receiver).Publickey
		publickey = cryptogrpghy.ParsePEMtoRSApublicKey(receiverPK)
	} else {
		tokensBalance := GetAccountBalance(digitalwalletTransactionObj.Sender)
		tokenBalanceVal, exist := tokensBalance[digitalwalletTransactionObj.TokenID]
		if !exist {
			responseObj := responses.FindResponseByID("119")
			return responseObj.EngResponse + digitalwalletTransactionObj.TokenID

		}
		if checkIfAccountIsActive(digitalwalletTransactionObj.Receiver) && checkIfAccountIsActive(digitalwalletTransactionObj.Sender) {
			if tokenBalanceVal <= digitalwalletTransactionObj.Amount+globalPkg.GlobalObj.TransactionFee {
				responseObj := responses.FindResponseByID("126")
				return responseObj.EngResponse + digitalwalletTransactionObj.TokenID

			}
		} else {
			responseObj := responses.FindResponseByID("127")
			return responseObj.EngResponse

		}
		senderPK := account.FindpkByAddress(digitalwalletTransactionObj.Sender).Publickey
		publickey = cryptogrpghy.ParsePEMtoRSApublicKey(senderPK)
	}

	// fmt.Println("digitalWalletTx time", digitalwalletTransactionObj.Time.String())
	signatureData := digitalwalletTransactionObj.Sender + digitalwalletTransactionObj.Receiver +
		fmt.Sprintf("%f", digitalwalletTransactionObj.Amount) //+ digitalwalletTransactionObj.Time.UTC().Format("2006-01-02T03:04:05+00:00")
	// fmt.Println("publickey", publickey)
	validSig := cryptogrpghy.VerifyPKCS1v15(digitalwalletTransactionObj.Signature, signatureData, *publickey)

	if validSig {
		return ""
		// } else if !validSig {
		// 	return ""
		// 	// }
	}

	responseObj := responses.FindResponseByID("128")
	return responseObj.EngResponse
}

//ValidateRefundTransaction validate refund transaction
func (digitalWalletTx RefundDigitalWalletTx) ValidateRefundTransaction() string {

	if digitalWalletTx.Amount < 1 {
		responseObj := responses.FindResponseByID("54")
		return responseObj.EngResponse

	}
	if digitalWalletTx.Sender == "" || digitalWalletTx.Time.IsZero() || digitalWalletTx.TokenID == "" ||
		digitalWalletTx.Amount == 0.0 || digitalWalletTx.Signature == "" {
		responseObj := responses.FindResponseByID("130")
		return responseObj.EngResponse

	}
	if digitalWalletTx.FlatCurrency && digitalWalletTx.Receiver == "" {
		responseObj := responses.FindResponseByID("131")
		return responseObj.EngResponse

	}
	if utf8.RuneCountInString(digitalWalletTx.TokenID) < globalPkg.GlobalObj.TokenIDStringFixedLength || utf8.RuneCountInString(digitalWalletTx.TokenID) > globalPkg.GlobalObj.TokenIDStringFixedLength {
		responseObj := responses.FindResponseByID("132")
		return responseObj.EngResponse

	}
	// if utf8.RuneCountInString(digitalWalletTx.Signature) < 100 && utf8.RuneCountInString(digitalWalletTx.Signature) >= 200 {
	// 	responseObj := responses.FindResponseByID("133")
	// 	return responseObj.EngResponse

	// }
	// time differnce between the received digital wallet transaction time and the server's time.
	timeDifference := time.Now().UTC().Sub(digitalWalletTx.Time.UTC()).Seconds()
	if timeDifference > float64(globalPkg.GlobalObj.TxValidationTimeInSeconds) {
		responseObj := responses.FindResponseByID("124")
		return responseObj.EngResponse

	}
	tokensBalance := GetAccountBalance(digitalWalletTx.Sender)
	tokenBalanceVal, exist := tokensBalance[digitalWalletTx.TokenID]
	if !exist {
		responseObj := responses.FindResponseByID("119")
		return responseObj.EngResponse

	}
	decimals := strings.Split(fmt.Sprintf("%f", digitalWalletTx.Amount), ".")[1]
	if utf8.RuneCountInString(decimals) > 6 {
		responseObj := responses.FindResponseByID("134")
		return responseObj.EngResponse

	}
	if checkIfAccountIsActive(digitalWalletTx.Sender) {
		if digitalWalletTx.FlatCurrency {
			if !checkIfAccountIsActive(digitalWalletTx.Receiver) {
				responseObj := responses.FindResponseByID("21")
				return responseObj.EngResponse

			}
			refundTokenValue := token.FindTokenByid(digitalWalletTx.TokenID).TokenValue
			// convert the refunded token amount to inoToken amount.
			toInoToken := digitalWalletTx.Amount * refundTokenValue
			// convert the refund value (in InoToken value) to dollar value.
			refundValue := toInoToken * globalPkg.GlobalObj.InoCoinToDollarRatio
			// fmt.Println("\n toInoToken:", toInoToken)
			// fmt.Println("\n refundValue:", refundValue)
			if refundValue < toInoToken {
				inoTokenID, _ := globalPkg.ConvertIntToFixedLengthString(0, globalPkg.GlobalObj.TokenIDStringFixedLength)
				receiverTokensBalance := GetAccountBalance(digitalWalletTx.Receiver)
				receiverTokenBalanceVal, exist := receiverTokensBalance[inoTokenID]
				inoTokenName := token.FindTokenByid(inoTokenID).TokenName
				if !exist {
					responseObj := responses.FindResponseByID("59")
					return responseObj.EngResponse + inoTokenName

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
					responseObj := responses.FindResponseByID("59")
					return responseObj.EngResponse + inoTokenName

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
				responseObj := responses.FindResponseByID("128")
				return responseObj.EngResponse

			}
		} else {
			responseObj := responses.FindResponseByID("126")
			return responseObj.EngResponse + digitalWalletTx.TokenID

		}
	} else {
		responseObj := responses.FindResponseByID("127")
		return responseObj.EngResponse

	}
}

//ValidateTransactionToken validation token for transfer transaction token
func ValidateTransactionToken(tokenTransactionObj DigitalwalletTransaction, isFee bool) (string, bool) {

	var errorfound string

	if utf8.RuneCountInString(tokenTransactionObj.TokenID) != globalPkg.GlobalObj.TokenIDStringFixedLength {
		errorfound = fmt.Sprintf("token ID must be equal to %v characters", globalPkg.GlobalObj.TokenIDStringFixedLength)
		return errorfound, false
	}
	if isFee && tokenTransactionObj.Amount < 1 {
		responseObj := responses.FindResponseByID("54")
		return responseObj.EngResponse, false

	}
	//validate fields not fields
	if tokenTransactionObj.Receiver == "" || tokenTransactionObj.TokenID == "" ||
		(tokenTransactionObj.Amount == 0.0 && isFee) || tokenTransactionObj.Signature == "" || tokenTransactionObj.Time.IsZero() {
		responseObj := responses.FindResponseByID("130")
		errorfound = responseObj.EngResponse

		return errorfound, false
	}

	firstaccount := accountdb.GetFirstAccount()
	if tokenTransactionObj.Sender == "" && firstaccount.AccountPublicKey != tokenTransactionObj.Receiver {
		responseObj := responses.FindResponseByID("137")
		return responseObj.EngResponse, false

	}

	if isFee && tokenTransactionObj.Receiver == tokenTransactionObj.Sender {
		responseObj := responses.FindResponseByID("138")
		errorfound = responseObj.EngResponse

		return errorfound, false
	}
	//validate send signature < 200
	// if utf8.RuneCountInString(tokenTransactionObj.Signature) < 100 && utf8.RuneCountInString(tokenTransactionObj.Signature) >= 200 {
	// 	responseObj := responses.FindResponseByID("133")
	// 	errorfound = responseObj.EngResponse

	// 	return errorfound, false
	// }
	if time.Since(tokenTransactionObj.Time).Seconds() > float64(globalPkg.GlobalObj.TxValidationTimeInSeconds) {
		responseObj := responses.FindResponseByID("124")
		return responseObj.EngResponse, false

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
		responseObj := responses.FindResponseByID("139")
		errorfound = responseObj.EngResponse

		return errorfound, false
	}
	decimals := strings.Split(fmt.Sprintf("%f", tokenTransactionObj.Amount), ".")[1]
	if len(decimals) > 6 {
		responseObj := responses.FindResponseByID("134")
		return responseObj.EngResponse, false

	}

	errorfound = tokenTransactionObj.ValidateTransaction()
	if errorfound != "" {
		return errorfound, false
	}

	return "", true
}

// ValidateServiceTransaction validate service transaction
func (digitalwalletTransactionObj DigitalwalletTransaction) ValidateServiceTransaction() string {
	// time differnce between the received digital wallet transaction time and the server's time.
	timeDifference := time.Now().UTC().Sub(digitalwalletTransactionObj.Time.UTC()).Seconds()
	if timeDifference > float64(globalPkg.GlobalObj.TxValidationTimeInSeconds) {
		responseObj := responses.FindResponseByID("124")
		return responseObj.EngResponse

	}
	//  service1  :=Service.ServiceStruct{ID: "ServiceId", Mbytes : true}

	// Service.ServiceCreateOUpdate(service1)
	tokensBalance := GetAccountBalance(digitalwalletTransactionObj.Sender)
	tokenBalanceVal, exist := tokensBalance[digitalwalletTransactionObj.TokenID]
	if !exist {
		responseObj := responses.FindResponseByID("119")
		return responseObj.EngResponse + digitalwalletTransactionObj.TokenID

	}
	if checkIfAccountIsActive(digitalwalletTransactionObj.Receiver) && checkIfAccountIsActive(digitalwalletTransactionObj.Sender) {
		if tokenBalanceVal >= digitalwalletTransactionObj.Amount+globalPkg.GlobalObj.TransactionFee {
			senderPK := account.FindpkByAddress(digitalwalletTransactionObj.Sender).Publickey
			public_key := cryptogrpghy.ParsePEMtoRSApublicKey(senderPK)

			// fmt.Println("digitalWalletTx time", digitalwalletTransactionObj.Time.String())
			signatureData := digitalwalletTransactionObj.Sender + digitalwalletTransactionObj.Receiver +
				fmt.Sprintf("%f", digitalwalletTransactionObj.Amount) + digitalwalletTransactionObj.Time.UTC().Format("2006-01-02 03:04:05 PM -0000")

			validSig := cryptogrpghy.VerifyPKCS1v15(digitalwalletTransactionObj.Signature, signatureData, *public_key)
			servicevalid := false
			reciveracc := account.GetAccountByAccountPubicKey(digitalwalletTransactionObj.Receiver)
			// fmt.Println("digitalwalletTransactionObj.WalletService=-=--=--=-=-=-=-=-=-=", digitalwalletTransactionObj.ServiceId)
			if digitalwalletTransactionObj.ServiceId == "" {
				if reciveracc.AccountRole != "service" {
					// fmt.Println(" reciveracc.AccountRole=-=-=-=-=-=-==-=-=--", reciveracc.AccountRole)
					servicevalid = true
				} else {
					// fmt.Println(" reciveracc.AccountRole=-=-=-=-=-=-==-=-=--", reciveracc.AccountRole)
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

								// obj = service.CalculateAmountAndCost(obj)
								(&obj).CalculateAmountAndCost()
								cost := obj.Calculation + globalPkg.GlobalObj.TransactionFee
								if cost == digitalwalletTransactionObj.Amount {
									servicevalid = true
									break
								} else {
									responseObj := responses.FindResponseByID("60")
									return responseObj.EngResponse

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
						responseObj := responses.FindResponseByID("10")
						return responseObj.EngResponse

					}
					if idflag == false {
						servicevalid = false
						responseObj := responses.FindResponseByID("61")
						return responseObj.EngResponse

					}
					//check the calculate of the bytes = amount
				} else {
					servicevalid = false
					responseObj := responses.FindResponseByID("140")
					return responseObj.EngResponse

				}
			}
			// fmt.Println("servicevalid=-=-=-=-=-=-=-=-=-=-",servicevalid)
			if validSig && servicevalid {
				return ""
				// } else if !validSig {
				// 	return ""
			} else if !servicevalid {
				responseObj := responses.FindResponseByID("146")
				return responseObj.EngResponse

			}
			// } else {
			// 	return "You are not allowed to do this transaction"
			// }
		} else {
			responseObj := responses.FindResponseByID("59")
			return responseObj.EngResponse + digitalwalletTransactionObj.TokenID

		}
	} else {
		responseObj := responses.FindResponseByID("127")
		return responseObj.EngResponse

	}
	return ""
}
