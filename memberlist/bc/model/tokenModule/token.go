package tokenModule

import (
	"fmt"

	"../account"
	"../accountdb"
	"../globalPkg"
	"../token"
	"../transaction"
	"../transactionModule"
)

//validateToken validation token for add token
func validateToken(tokenObj token.StructToken) (string, bool) {

	var errorfound string
	//validate token id ==100
	//if len(tokenObj.TokenID) != 100 {
	//	errorfound = "token ID must be 100 characters"
	//	return errorfound, false
	//}
	//validate token name ==20
	if len(tokenObj.TokenName) < 4 || len(tokenObj.TokenName) > 20 {
		errorfound = "token name must be more than 4 characters and less than or equal 20 characters"
		return errorfound, false
	}
	//validate token symbol <= 4
	if len(tokenObj.TokenSymbol) > 4 {
		errorfound = "token symbol should less than or equal to 4 characters"
		return errorfound, false
	}
	// validate icon url if empty or ==100
	// if len(tokenObj.IconURL) == 0 || len(tokenObj.IconURL) <= 100 {
	// 	errorfound = ""
	// } else {
	// 	errorfound = "Icon URL is optiaonal if enter it must be less or equal 100 characters"
	// 	return errorfound, false
	// }
	// validate description if empty or == 100
	if len(tokenObj.Description) == 0 || len(tokenObj.Description) <= 100 {
		errorfound = ""
	} else {
		errorfound = "Description is optiaonal if enter it must be less or equal 100 characters"
		return errorfound, false
	}
	//validate initiator address if empty
	if tokenObj.InitiatorAddress == "" {
		errorfound = "please enter initiator address (Public key)"
		return errorfound, false
	}
	//validate initiator address if exist in account data
	accountobj := account.GetAccountByAccountPubicKey(tokenObj.InitiatorAddress)
	fmt.Println("------------------    ", accountobj)
	if accountobj.AccountPublicKey == "" {
		errorfound = "please enter valid initiator address (Public key)"
		return errorfound, false
	}
	if accountobj.AccountPassword != tokenObj.Password {
		errorfound = "The given password is incorrect."
		return errorfound, false
	}

	//validate Tokens Total Supply less than or equal zero
	if tokenObj.TokensTotalSupply < 1 {
		errorfound = "please enter Tokens Total Supply more than zero"
		return errorfound, false
	}
	//validate Tokens value less than or equal zero
	if tokenObj.TokenValue <= 0.0 {
		errorfound = "please enter Tokens value more than zero"
		return errorfound, false
	}
	//validate token precision from 0 to 5
	if tokenObj.Precision < 0 || tokenObj.Precision > 5 {
		errorfound = "please enter Precision range from 0 to 5 "
		return errorfound, false
	}
	//validate Tokens UsageType is mandatory security or utility
	if tokenObj.UsageType == "security" || tokenObj.UsageType == "utility" {
		errorfound = ""
	} else {
		errorfound = "please enter UsageType security or utility"
		return errorfound, false
	}
	if tokenObj.UsageType == "security" && tokenObj.Precision != 0 {
		errorfound = "UsageType security  and must precision equal zero"
		return errorfound, false
	}
	//validate Tokens TokenType is mandatory public  or private
	if tokenObj.TokenType == "public" || tokenObj.TokenType == "private" {
		// check type token is public, validating for enter contact ID
		if tokenObj.TokenType == "public" {
			// validate ContractID if empty or ==60
			if len(tokenObj.ContractID) < 4 || len(tokenObj.ContractID) > 60 {
				errorfound = "enter ContractID must be more than 4 character and less than or equal 60 characters"
				return errorfound, false
			}
		}
		// check type token is Private , validating for enter pentential PK ,
		// enter the potential users public keys which can use this token
		accountList := accountdb.GetAllAccounts()
		if tokenObj.TokenType == "private" {
			//enter pentential PK which can use this token
			if len(tokenObj.UserPublicKey) != 0 {
				for _, pk := range tokenObj.UserPublicKey {
					if pk == tokenObj.InitiatorAddress {
						errorfound = "user create token can't be in user public key "
						return errorfound, false
					}
					if !containspk(accountList, pk) {
						errorfound = "this public key is not associated with any account"
						return errorfound, false
					}
				}
			} else {
				errorfound = "enter the potential users public keys which can use this token"
				return errorfound, false
			}
		}
	} else {
		errorfound = "please enter TokenType is public  or private"
		return errorfound, false
	}

	// Dynamic price	If the price of token is dynamic it gets its value from bidding platform.
	// Bidding platform API URL.
	//  based on ValueDynamic  True or false
	if tokenObj.ValueDynamic == true {
		//for example value
		biddingplatformValue := 5.5
		tokenObj.Dynamicprice = biddingplatformValue
	}
	return "", true
}

// Contains tells whether a contains x.
func containspk(a []accountdb.AccountStruct, pk string) bool {
	for _, n := range a {
		if pk == n.AccountPublicKey {
			return true
		}
	}
	return false
}

//validateUserAmount validate user balance cover total amount
func validateUserAmount(tokenObj token.StructToken) bool {
	inoTokenID, _ := globalPkg.ConvertIntToFixedLengthString(0, globalPkg.GlobalObj.TokenIDStringFixedLength)
	//total amount of tokens
	//TokenTotalAmount := float64(tokenObj.TokensTotalSupply) * tokenObj.TokenValue

	txFee := globalPkg.GlobalObj.TransactionFee
	// this will be the the amount in Ino Token(id = inoTokenID) which will be spent in order to create the token
	TokenAmount := (tokenObj.TokensTotalSupply*tokenObj.TokenValue)/globalPkg.GlobalObj.InoCoinToDollarRatio + txFee
	//get account balance from unspent transactions outputs
	accountBalance := transactionModule.GetAccountBalance(tokenObj.InitiatorAddress)
	// get the Ino Token balance from the map of tokens balance of the account
	balance, exist := accountBalance[inoTokenID]
	// check if Ino Token balance exist and is bigger than what he will spent
	if exist && balance > TokenAmount {
		return true
	}
	return false
}

//ValidatingNewToken validate new token
func ValidatingNewToken(TokenObj token.StructToken) (string, bool) {

	var errorfound string
	//check if token name , symbol exist before
	tokens := token.GetAllTokens()
	for _, tokenobjOld := range tokens {
		if tokenobjOld.TokenName == TokenObj.TokenName {
			errorfound = "token name exist before"
			return errorfound, false
		}
		if tokenobjOld.TokenSymbol == TokenObj.TokenSymbol {
			errorfound = "token symbol exist before"
			return errorfound, false
		}
	}
	//check about active token depend on ** total supply
	return "", true
}

//AddToken add token to db
func AddToken(tokenObj token.StructToken) (string, bool) {
	message := ""
	check := true
	if !token.TokenCreate(tokenObj) {
		message = "Can't create token"
		check = false
	}
	return message, check
}

//validateTokenForUpdating validation token for Update token
func validateTokenForUpdating(tokenObj token.StructToken) (string, bool) {

	var errorfound string
	//validate initiator address if exist in account data
	accountobj := account.GetAccountByAccountPubicKey(tokenObj.InitiatorAddress)
	if accountobj.AccountPublicKey == "" {
		errorfound = "please enter valid initiator address (Public key)"
		return errorfound, false
	}
	if accountobj.AccountPassword != tokenObj.Password {
		errorfound = "The given password is incorrect."
		return errorfound, false
	}
	//validate Tokens Total Supply less than or equal zero
	if tokenObj.TokensTotalSupply < 1 {
		errorfound = "please enter Tokens Total Supply more than zero"
		return errorfound, false
	}
	//validate Tokens value less than or equal zero
	if tokenObj.TokenValue <= 0.0 {
		errorfound = "please enter Tokens value more than zero"
		return errorfound, false
	}
	//validate token precision from 0 to 5
	if tokenObj.Precision < 0 || tokenObj.Precision > 5 {
		errorfound = "please enter Precision range from 0 to 5 "
		return errorfound, false
	}

	//validate Tokens TokenType is mandatory public  or private
	if tokenObj.TokenType == "public" || tokenObj.TokenType == "private" {
		// check type token is public, optianal enter contact ID
		if tokenObj.TokenType == "public" {
			// validate ContractID if empty or ==60
			if len(tokenObj.ContractID) < 4 || len(tokenObj.ContractID) > 60 {
				errorfound = "enter ContractID must be more than 4 character and less than or equal 60 characters"
				return errorfound, false
			}
		}
		// check type token is Private , optianal enter pentential PK ,enter the potential users public keys which can use this token
		accountList := accountdb.GetAllAccounts()
		if tokenObj.TokenType == "private" {
			//enter pentential PK which can use this token
			if len(tokenObj.UserPublicKey) != 0 {
				for _, pk := range tokenObj.UserPublicKey {
					if pk == tokenObj.InitiatorAddress {
						errorfound = "user create token can't be in user public key "
						return errorfound, false
					}
					if !containspk(accountList, pk) {
						errorfound = "this public key is not associated with any account"
						return errorfound, false
					}
				}
			} else {
				errorfound = "enter the potential users public keys which can use this token"
				return errorfound, false
			}
		}
	} else {
		errorfound = "please enter TokenType is public  or private"
		return errorfound, false
	}

	return "", true
}

//validateTransactionToken validation token for transfer transaction token
func validateTransactionToken(tokenTransactionObj transaction.DigitalwalletTransaction) (string, bool) {

	var errorfound string

	//validate token id ==100
	//if len(tokenTransactionObj.TokenID) != 100 {
	//	errorfound = "token ID must be 100 characters"
	//	return errorfound, false
	//}
	//validate fields not fields
	if tokenTransactionObj.Sender == "" || tokenTransactionObj.Amount == 0.0 || tokenTransactionObj.Signature == "" {
		errorfound = "please enter required data  Sender PK,Receiver PK,Amount,Sender sign"
		return errorfound, false
	}

	//validate send signature ==200
	//if len(tokenTransactionObj.Signature) != 200 {
	//	errorfound = "Sender sign must be 200 characters"
	//	return errorfound, false
	//}

	errorfound = transactionModule.ValidateTransaction(tokenTransactionObj)
	if errorfound != "" {
		return errorfound, false
	}

	return "", true
}
