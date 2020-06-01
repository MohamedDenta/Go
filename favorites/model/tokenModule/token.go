package tokenModule

import (
	"unicode/utf8"
	"strings"
	"../account"
	"../accountdb"
	"../globalPkg"
	"../responses"
	"../token"
	"../transactionModule"
)

//TokenExplore explore user token need to see only token name and token id
type TokenExplore struct {
	TokenID   string
	TokenName string
}

//validateToken validation token for add token
func validateToken(tokenObj token.StructToken, isOwnership bool) (string, bool) {

	var errorfound string
	//validate token id ==100
	//if len(tokenObj.TokenID) != 100 {
	//	errorfound = "token ID must be 100 characters"
	//	return errorfound, false
	//}
	//validate token name ==20
	if utf8.RuneCountInString(tokenObj.TokenName) < 4 || utf8.RuneCountInString(tokenObj.TokenName) > 20 {
		responseObj := responses.FindResponseByID("67")
		errorfound = responseObj.EngResponse
		return errorfound, false
	}
	//validate token symbol == 4
	if utf8.RuneCountInString(tokenObj.TokenSymbol) > 4 {
		responseObj := responses.FindResponseByID("68")
		errorfound = responseObj.EngResponse
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
	if utf8.RuneCountInString(tokenObj.Description) == 0 || utf8.RuneCountInString(tokenObj.Description) <= 100 {
		errorfound = ""
	} else {
		responseObj := responses.FindResponseByID("69")
		errorfound = responseObj.EngResponse

		return errorfound, false
	}
	//validate initiator address if empty
	if tokenObj.InitiatorAddress == "" {
		responseObj := responses.FindResponseByID("71")
		errorfound = responseObj.EngResponse

		return errorfound, false
	}
	//validate initiator address if exist in account data
	accountobj := account.GetAccountByAccountPubicKey(tokenObj.InitiatorAddress)
	// fmt.Println("------------------    ", accountobj)
	if accountobj.AccountPublicKey == "" {
		responseObj := responses.FindResponseByID("71")
		errorfound = responseObj.EngResponse

		return errorfound, false
	}
	if accountobj.AccountPassword != tokenObj.Password {
		responseObj := responses.FindResponseByID("11")
		errorfound = responseObj.EngResponse

		return errorfound, false
	}

	//validate Tokens Total Supply less than or equal zero
	if tokenObj.TokensTotalSupply < 1 {
		responseObj := responses.FindResponseByID("72")
		errorfound = responseObj.EngResponse

		return errorfound, false
	}
	if isOwnership {
		// validate Tokens Total Supply equal to 1
		if tokenObj.TokensTotalSupply != 1 {
			responseObj := responses.FindResponseByID("73")
			errorfound = responseObj.EngResponse

			return errorfound, false
		}
		// validate Tokens Value equal to 1
		if tokenObj.TokenValue != 1 {
			responseObj := responses.FindResponseByID("74")
			errorfound = responseObj.EngResponse

			return errorfound, false
		}
		// validate usage type
		if tokenObj.UsageType != "security" {
			responseObj := responses.FindResponseByID("75")
			errorfound = responseObj.EngResponse

			return errorfound, false
		}
	}
	//validate Tokens value less than or equal zero
	if tokenObj.TokenValue <= 0.0 {
		responseObj := responses.FindResponseByID("76")
		errorfound = responseObj.EngResponse

		return errorfound, false
	}
	//validate token precision from 0 to 5
	if tokenObj.Precision < 0 || tokenObj.Precision > 5 {
		responseObj := responses.FindResponseByID("77")
		errorfound = responseObj.EngResponse

		return errorfound, false
	}
	//validate Tokens UsageType is mandatory security or utility
	if tokenObj.UsageType == "security" || tokenObj.UsageType == "utility" {
		errorfound = ""
	} else {
		responseObj := responses.FindResponseByID("78")
		errorfound = responseObj.EngResponse

		return errorfound, false
	}
	if tokenObj.UsageType == "security" && tokenObj.Precision != 0 {
		responseObj := responses.FindResponseByID("79")
		errorfound = responseObj.EngResponse

		return errorfound, false
	}
	//validate Tokens TokenType is mandatory public  or private
	if tokenObj.TokenType == "public" || tokenObj.TokenType == "private" {
		// check type token is public, validating for enter contact ID
		if tokenObj.TokenType == "public" {
			// validate ContractID if empty or ==60
			if utf8.RuneCountInString(tokenObj.ContractID) < 4 || utf8.RuneCountInString(tokenObj.ContractID) > 60 {
				responseObj := responses.FindResponseByID("80")
				errorfound = responseObj.EngResponse

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
						responseObj := responses.FindResponseByID("81")
						errorfound = responseObj.EngResponse

						return errorfound, false
					}
					if !containspk(accountList, pk) {
						responseObj := responses.FindResponseByID("10")
						errorfound = responseObj.EngResponse

						return errorfound, false
					}
				}
			} else {
				responseObj := responses.FindResponseByID("83")
				errorfound = responseObj.EngResponse

				return errorfound, false
			}
		}
	} else {
		responseObj := responses.FindResponseByID("84")
		errorfound = responseObj.EngResponse

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
	txFee := globalPkg.GlobalObj.TransactionFee
	// this will be the the amount in Ino Token(id = inoTokenID) which will be spent in order to create the token
	TokenAmount := ((tokenObj.TokensTotalSupply + txFee) * tokenObj.TokenValue) / globalPkg.GlobalObj.InoCoinToDollarRatio
	//total amount of tokens
	//TokenTotalAmount := float64(tokenObj.TokensTotalSupply) * tokenObj.TokenValue

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
			responseObj := responses.FindResponseByID("85")
			errorfound = responseObj.EngResponse

			return errorfound, false
		}
		if tokenobjOld.TokenSymbol == TokenObj.TokenSymbol {
			responseObj := responses.FindResponseByID("86")
			errorfound = responseObj.EngResponse

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
		responseObj := responses.FindResponseByID("93")
		message = responseObj.EngResponse

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
		responseObj := responses.FindResponseByID("71")
		errorfound = responseObj.EngResponse

		return errorfound, false
	}
	if accountobj.AccountPassword != tokenObj.Password {
		responseObj := responses.FindResponseByID("11")
		errorfound = responseObj.EngResponse

		return errorfound, false
	}
	//validate Tokens Total Supply less than or equal zero
	if tokenObj.TokensTotalSupply < 1 {
		responseObj := responses.FindResponseByID("72")
		errorfound = responseObj.EngResponse

		return errorfound, false
	}
	//validate Tokens value less than or equal zero
	if tokenObj.TokenValue <= 0.0 {
		responseObj := responses.FindResponseByID("76")
		errorfound = responseObj.EngResponse

		return errorfound, false
	}
	//validate token precision from 0 to 5
	if tokenObj.Precision < 0 || tokenObj.Precision > 5 {
		responseObj := responses.FindResponseByID("77")
		errorfound = responseObj.EngResponse

		return errorfound, false
	}

	//validate Tokens TokenType is mandatory public  or private
	if tokenObj.TokenType == "public" || tokenObj.TokenType == "private" {
		// check type token is public, optianal enter contact ID
		if tokenObj.TokenType == "public" {
			// validate ContractID if empty or ==60
			if utf8.RuneCountInString(tokenObj.ContractID) < 4 || utf8.RuneCountInString(tokenObj.ContractID) > 60 {
				responseObj := responses.FindResponseByID("80")
				errorfound = responseObj.EngResponse

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
						responseObj := responses.FindResponseByID("81")
						errorfound = responseObj.EngResponse

						return errorfound, false
					}
					if !containspk(accountList, pk) {
						responseObj := responses.FindResponseByID("10")
						errorfound = responseObj.EngResponse

						return errorfound, false
					}
				}
			} else {
				responseObj := responses.FindResponseByID("83")
				errorfound = responseObj.EngResponse

				return errorfound, false
			}
		}
	} else {
		responseObj := responses.FindResponseByID("84")
		errorfound = responseObj.EngResponse

		return errorfound, false
	}

	return "", true
}

//validateTransactionToken validation token for transfer transaction token
func validateTransactionToken(tokenTransactionObj transactionModule.DigitalwalletTransaction) (string, bool) {

	var errorfound string

	if tokenTransactionObj.Sender == "" || tokenTransactionObj.Amount == 0.0 || tokenTransactionObj.Signature == "" {
		responseObj := responses.FindResponseByID("130")
		errorfound = responseObj.EngResponse

		return errorfound, false
	}

	//validate send signature ==200
	//if len(tokenTransactionObj.Signature) != 200 {
	//	errorfound = "Sender sign must be 200 characters"
	//	return errorfound, false
	//}

	// errorfound = transactionModule.ValidateTransaction(tokenTransactionObj)
	errorfound = tokenTransactionObj.ValidateTransaction()

	if errorfound != "" {
		return errorfound, false
	}

	return "", true
}
//isAccountBilling is account billing  
func isAccountBilling(address string) bool {
	acc := account.GetAccountByAccountPubicKey(address)
	return acc.AccountRole == "billing"
}
//getLastTokenIndex get last token index 
func getLastTokenIndex() string {
	var Token token.StructToken
	Token = token.GetLastToken()
	if Token.TokenID == "" {
		return "-1"
	}

	return Token.TokenID
}

//ContainstokenID Contains tells whether a contains x.
func ContainstokenID(TokenID []TokenExplore, tokenid string) bool {
	for _, n := range TokenID {
		if tokenid == n.TokenID {
			return true
		}
	}
	return false
}
//newTokenIndex create new index token 
func newTokenIndex() int{
LastIndex := getLastTokenIndex()
			index := 0
			if LastIndex != "-1" {
				// TODO : split LastIndex
				res := strings.Split(LastIndex, "_")
				if len(res) == 2 {
					index = globalPkg.ConvertFixedLengthStringtoInt(res[1]) + 1
				} else {
					index = globalPkg.ConvertFixedLengthStringtoInt(LastIndex) + 1
				}
			}
			return index
}