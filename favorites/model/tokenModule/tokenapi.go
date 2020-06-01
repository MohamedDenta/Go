package tokenModule

import (
	"encoding/json"
	"fmt"
	"net/http"
	"unicode/utf8"

	"../logpkg"
	"../responses"

	"strings"
	//"time"

	"../account"
	"../admin"
	"../broadcastTcp"
	"../globalPkg"
	"../token"
	"../transactionModule"
)

//RegisteringNewTokenAPI Create new token
func RegisteringNewTokenAPI(w http.ResponseWriter, req *http.Request) {
	now, userIP := globalPkg.SetLogObj(req)
	logobj := logpkg.LogStruct{"_", now, userIP, "macAdress", "RegisteringNewTokenAPI", "tokenModule", "_", "_", "_", 0}

	TokenObj := token.StructToken{}
	decoder := json.NewDecoder(req.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&TokenObj)

	if err != nil {
		responseObj := responses.FindResponseByID("1")
		globalPkg.SendError(w, responseObj.EngResponse)
		globalPkg.WriteLog(logobj, responseObj.EngResponse, "failed")
		return

	}

	var validate = false
	var check = false
	var errorfound string
	//check for validate Token
	errorfound, check = validateToken(TokenObj, false)
	if check == true {
		errorfound = ""
		validate = true
	} else {
		globalPkg.SendError(w, errorfound)
		globalPkg.WriteLog(logobj, errorfound, "failed")
		return

	}

	//check if user balance cover total token amount
	if !validateUserAmount(TokenObj) {
		responseObj := responses.FindResponseByID("89")
		globalPkg.SendError(w, responseObj.EngResponse)
		globalPkg.WriteLog(logobj, responseObj.EngResponse, "failed")
		return

	}

	if validate == true {

		//The token data has been sent for validation
		errorfound, check = ValidatingNewToken(TokenObj)

		if check == true {
			// this variable will store the Ino Token amount that will calculated to the new token.
			// like this Ino Tokens (stored in TotalSupply temporarily) * TokenValue to get new TotalSupply.
			inoTokenAmount := (TokenObj.TokensTotalSupply * TokenObj.TokenValue) / globalPkg.GlobalObj.InoCoinToDollarRatio // IMPORTANT
			// TODO: token value is of dollar, to know how much to cut from ino token you'll have to
			// (TokensTotalSupply * TokenValue) / inoToDollarRatio.
			// input will be (TokensTotalSupply * TokenValue) / inoToDollarRatio. and output will be desired TokensTotalSupply from api.
			// TODO: change the validation om ino token accordingly to this formula.

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
			TokenObj.TokenID, _ = globalPkg.ConvertIntToFixedLengthString(index, globalPkg.GlobalObj.TokenIDStringFixedLength)

			// creating Tx that have sender and receiver = token.InitiatorAddress  .. but tokenID = "1" for sender
			// and tokenID = TokenObj.TokenID for receiver
			inoTokenID, _ := globalPkg.ConvertIntToFixedLengthString(0, globalPkg.GlobalObj.TokenIDStringFixedLength)
			tokenIcondata := TokenObj.TokenID + "_" + TokenObj.TokenIcon
			TokenObj.TokenIcon = ""

			//size accept 20000

			if utf8.RuneCountInString(tokenIcondata) > 20000 {
				responseObj := responses.FindResponseByID("90")
				globalPkg.SendError(w, responseObj.EngResponse)
				globalPkg.WriteLog(logobj, responseObj.EngResponse, "failed")
				return
			}
			tx1 := transactionModule.CreateTokenTx(TokenObj, inoTokenAmount, inoTokenID, true)

			//broadcast tx1 transaction
			broadcastTcp.BoardcastingTCP(tx1, "addTokenTransaction", "transaction")

			//approve the token to add it to database and broadcast token
			broadcastTcp.BoardcastingTCP(TokenObj, "addtoken", "token")
			var count = 0
		iconimage:
			tokendata := token.FindTokenByid(TokenObj.TokenID)
			if tokendata.TokenID != "" {
				broadcastTcp.BoardcastingTokenImgUDP(tokenIcondata, "addtokenimg", "addtokenimg")
			} else {
				count++
				goto iconimage
			}

			// fmt.Println("=====    Count waiting transaction and token created  :    ", count)
			//success message
			responseObj := responses.FindResponseByID("91")
			globalPkg.SendResponseMessage(w, responseObj.EngResponse)
			globalPkg.WriteLog(logobj, responseObj.EngResponse, "success")

		} else {
			responseObj := responses.FindResponseByID("92")
			globalPkg.SendError(w, responseObj.EngResponse+errorfound)
			globalPkg.WriteLog(logobj, responseObj.EngResponse, "failed")
			return
		}

	} else {

		responseObj := responses.FindResponseByID("93")
		globalPkg.SendError(w, responseObj.EngResponse)
		globalPkg.WriteLog(logobj, responseObj.EngResponse, "failed")
	}
}

//UpdatingTokenAPI update token data exact ID,name,symbol token
func UpdatingTokenAPI(w http.ResponseWriter, req *http.Request) {
	now, userIP := globalPkg.SetLogObj(req)
	logobj := logpkg.LogStruct{"_", now, userIP, "macAdress", "UpdatingTokenAPI", "tokenModule", "_", "_", "_", 0}

	TokenObj := token.StructToken{}
	decoder := json.NewDecoder(req.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&TokenObj)

	if err != nil {
		responseObj := responses.FindResponseByID("1")
		globalPkg.SendError(w, responseObj.EngResponse)
		globalPkg.WriteLog(logobj, responseObj.EngResponse, "failed")
		return
	}
	var validate = false
	var check = false
	var errorfound string
	//check for validate Token for updating
	errorfound, check = validateTokenForUpdating(TokenObj)
	if check == true {
		errorfound = ""
		validate = true
	} else {
		globalPkg.SendError(w, errorfound)
		globalPkg.WriteLog(logobj, errorfound, "failed")
		return
	}
	// Get the token old data from the database using its ID
	tokenOldObj := token.FindTokenByid(TokenObj.TokenID)
	if tokenOldObj.TokenID == "" {
		responseObj := responses.FindResponseByID("94")
		globalPkg.SendError(w, responseObj.EngResponse)
		globalPkg.WriteLog(logobj, responseObj.EngResponse, "failed")
		return
	}
	//check on reissuability updates
	if tokenOldObj.Reissuability == false && TokenObj.Reissuability == true {
		// data exist in contract ID or user public key use private token
		if TokenObj.ContractID != "" || len(TokenObj.UserPublicKey) != 0 {
			responseObj := responses.FindResponseByID("95")
			globalPkg.SendError(w, responseObj.EngResponse)
			globalPkg.WriteLog(logobj, responseObj.EngResponse, "failed")
			return
		}
	}

	if tokenOldObj.Reissuability == true && TokenObj.TokensTotalSupply > tokenOldObj.TokensTotalSupply {
		//check if user balance cover total token amount
		if !validateUserAmount(TokenObj) {
			responseObj := responses.FindResponseByID("96")
			globalPkg.SendError(w, responseObj.EngResponse)
			globalPkg.WriteLog(logobj, responseObj.EngResponse, "failed")
			return
		}
	}

	//check on change on type of token updates
	if tokenOldObj.TokenType == "public" && TokenObj.TokenType == "private" {

		if utf8.RuneCountInString(tokenOldObj.ContractID) > 4 || utf8.RuneCountInString(tokenOldObj.ContractID) <= 60 {
			responseObj := responses.FindResponseByID("97")
			globalPkg.SendError(w, responseObj.EngResponse)
			globalPkg.WriteLog(logobj, responseObj.EngResponse, "failed")
			return
		}
	}

	if tokenOldObj.TokenType == "private" && TokenObj.TokenType == "public" {

		if utf8.RuneCountInString(TokenObj.ContractID) < 4 || utf8.RuneCountInString(TokenObj.ContractID) > 60 {
			responseObj := responses.FindResponseByID("98")
			globalPkg.SendError(w, responseObj.EngResponse)
			globalPkg.WriteLog(logobj, responseObj.EngResponse, "failed")
			return
		}
	}

	if validate == true {
		//approve the token to update it to database and broadcast update token
		tokenIcondata := TokenObj.TokenID + "_" + TokenObj.TokenIcon
		TokenObj.TokenIcon = ""
		if utf8.RuneCountInString(tokenIcondata) > 20000 {
			responseObj := responses.FindResponseByID("90")
			globalPkg.SendError(w, responseObj.EngResponse)
			globalPkg.WriteLog(logobj, responseObj.EngResponse, "failed")
			return
		}
		broadcastTcp.BoardcastingTCP(TokenObj, "updatetoken", "token")
		broadcastTcp.BoardcastingTokenImgUDP(tokenIcondata, "addtokenimg", "addtokenimg")

		responseObj := responses.FindResponseByID("99")
		globalPkg.SendResponseMessage(w, responseObj.EngResponse)
		globalPkg.WriteLog(logobj, responseObj.EngResponse, "success")

	} else {
		responseObj := responses.FindResponseByID("100")
		globalPkg.SendError(w, responseObj.EngResponse)
		globalPkg.WriteLog(logobj, responseObj.EngResponse, "failed")
	}
}

//ExploringUserTokensAPI Providing info for user about his tokens
func ExploringUserTokensAPI(w http.ResponseWriter, req *http.Request) {
	//log
	now, userIP := globalPkg.SetLogObj(req)
	logobj := logpkg.LogStruct{"_", now, userIP, "macAdress", "ExploringUserTokensAPI", "tokenModule", "_", "_", "_", 0}

	var accountPasswordAndPubKey transactionModule.AccountPasswordAndPubKey

	decoder := json.NewDecoder(req.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&accountPasswordAndPubKey)
	if err != nil {
		responseObj := responses.FindResponseByID("1")
		globalPkg.SendError(w, responseObj.EngResponse)
		globalPkg.WriteLog(logobj, responseObj.EngResponse, "failed")

		return
	}
	//check public key and password is valid
	accObj := account.GetAccountByAccountPubicKey(accountPasswordAndPubKey.PublicKey)
	if accObj.AccountPublicKey == "" || accObj.AccountPassword == "" || accObj.AccountName == "" {
		responseObj := responses.FindResponseByID("10")
		globalPkg.SendError(w, responseObj.EngResponse)
		globalPkg.WriteLog(logobj, responseObj.EngResponse, "failed")
		return
	} else if accountPasswordAndPubKey.Password != accObj.AccountPassword {
		responseObj := responses.FindResponseByID("11")
		globalPkg.SendError(w, responseObj.EngResponse)
		globalPkg.WriteLog(logobj, responseObj.EngResponse, "failed")

		return
	}

	tokens := token.GetAllTokens() //get tokens from ledger
	exploretokenObj := TokenExplore{}
	tokenList := []TokenExplore{} //append all tokens types private or public and only return to user token id and token name

	//if user public key create token type public,private
	for _, taken := range tokens {
		if taken.InitiatorAddress == accountPasswordAndPubKey.PublicKey {
			// TODO: Please, Reconsider this condition.
			if taken.TokenType == "private" || taken.TokenType == "public" {
				exploretokenObj.TokenID = taken.TokenID
				exploretokenObj.TokenName = taken.TokenName
				tokenList = append(tokenList, exploretokenObj)
			}
		}
	}
	tokenids := transactionModule.GetTokenIDusedbyrecieverpk(accountPasswordAndPubKey.PublicKey)

	for _, tokenid := range tokenids {
		tokenObj2 := token.FindTokenByid(tokenid)
		if tokenObj2.TokenType == "public" {
			containtokenid := ContainstokenID(tokenList, tokenObj2.TokenID)
			if !containtokenid {
				exploretokenObj.TokenID = tokenObj2.TokenID
				exploretokenObj.TokenName = tokenObj2.TokenName
				tokenList = append(tokenList, exploretokenObj)
			}
		}
	}
	//Get all public tokens where this user is one of  their holders . Table Contact ID
	if len(tokenList) == 0 {
		responseObj := responses.FindResponseByID("101")
		globalPkg.SendError(w, responseObj.EngResponse)
		globalPkg.WriteLog(logobj, responseObj.EngResponse, "failed")

	} else {
		sendJSON, _ := json.Marshal(tokenList)
		globalPkg.SendResponse(w, sendJSON)
	}
}

// RefundToken  either with Inotoken or with fiat currency
func RefundToken(w http.ResponseWriter, req *http.Request) {

	//log
	now, userIP := globalPkg.SetLogObj(req)
	logobj := logpkg.LogStruct{"_", now, userIP, "macAdress", "RefundToken", "tokenModule", "_", "_", "_", 0}

	tokenTransactionObj := transactionModule.RefundDigitalWalletTx{}
	decoder := json.NewDecoder(req.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&tokenTransactionObj)

	if err != nil {
		responseObj := responses.FindResponseByID("1")
		globalPkg.SendError(w, responseObj.EngResponse)
		globalPkg.WriteLog(logobj, responseObj.EngResponse, "failed")
		return
	}
	//check for validate data transfer Token
	errorfound := tokenTransactionObj.ValidateRefundTransaction()
	// errorfound := transactionModule.ValidateRefundTransaction(tokenTransactionObj)

	if errorfound == "" {
		tokenObj := token.FindTokenByid(tokenTransactionObj.TokenID)

		transactionObj := transactionModule.RefundTokenTx(tokenObj, tokenTransactionObj)
		// TODO: update token to decrease the amount of refunded token??
		transactionObj.TransactionTime = globalPkg.UTCtime()
		// fmt.Println("=======================================================")
		// fmt.Println(transactionObj)
		// fmt.Println("=======================================================")
		transactionObj.TransactionID = ""
		transactionObj.TransactionID = globalPkg.CreateHash(transactionObj.TransactionTime, fmt.Sprintf("%s", transactionObj), 3)

		tokenObj.TokensTotalSupply -= tokenTransactionObj.Amount + (tokenTransactionObj.Amount * globalPkg.GlobalObj.TransactionRefundFee)
		broadcastTcp.BoardcastingTCP(transactionObj, "addTokenTransaction", "transaction")

		message := fmt.Sprintf(
			"Your Refund transaction with %v of Token ID %v has been refunded successfully to your account",
			tokenTransactionObj.Amount, tokenTransactionObj.TokenID,
		)
		sendJSON, _ := json.Marshal(message)
		globalPkg.SendResponse(w, sendJSON)
		globalPkg.WriteLog(logobj, "refund success", "success")

	} else {
		globalPkg.SendError(w, errorfound)
		globalPkg.WriteLog(logobj, errorfound, "failed")

	}
}

//GetAllTokenssAPI get all tokens
func GetAllTokenssAPI(w http.ResponseWriter, req *http.Request) {
	//log
	now, userIP := globalPkg.SetLogObj(req)
	logobj := logpkg.LogStruct{"_", now, userIP, "macAdress", "GetAllTokensAPI", "tokenModule", "_", "_", "_", 0}

	Adminobj := admin.Admin{}
	decoder := json.NewDecoder(req.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&Adminobj)

	if err != nil {
		responseObj := responses.FindResponseByID("1")
		globalPkg.SendError(w, responseObj.EngResponse)
		globalPkg.WriteLog(logobj, responseObj.EngResponse, "failed")
		return
	}
	if admin.ValidationAdmin(Adminobj) {
		sendJSON, _ := json.Marshal(token.GetAllTokens())
		globalPkg.SendResponse(w, sendJSON)
		globalPkg.WriteLog(logobj, string(sendJSON), "success")
	} else {
		responseObj := responses.FindResponseByID("2")
		globalPkg.SendError(w, responseObj.EngResponse)
		globalPkg.WriteLog(logobj, responseObj.EngResponse, "failed")
	}
}

//GettokennameAPI get token name
func GettokennameAPI(w http.ResponseWriter, req *http.Request) {
	//log
	now, userIP := globalPkg.SetLogObj(req)
	logobj := logpkg.LogStruct{"_", now, userIP, "macAdress", "GettokennameAPI", "tokenModule", "_", "_", "_", 0}

	Tokenname := globalPkg.JSONString{}
	decoder := json.NewDecoder(req.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&Tokenname)

	if err != nil {
		responseObj := responses.FindResponseByID("1")
		globalPkg.SendError(w, responseObj.EngResponse)
		globalPkg.WriteLog(logobj, responseObj.EngResponse, "failed")
		return
	}

	tokenObj := token.FindTokenByTokenName(Tokenname.Name)
	if tokenObj.TokenName != "" {
		sendJSON, _ := json.Marshal(tokenObj)
		globalPkg.SendResponse(w, sendJSON)
		globalPkg.WriteLog(logobj, string(sendJSON), "success")
		return
	}
	responseObj := responses.FindResponseByID("102")
	globalPkg.SendError(w, responseObj.EngResponse)
	globalPkg.WriteLog(logobj, responseObj.EngResponse, "failed")

}

//RegisteringBillingTokenAPI Create new billing token
func RegisteringBillingTokenAPI(w http.ResponseWriter, req *http.Request) {
	now, userIP := globalPkg.SetLogObj(req)
	logobj := logpkg.LogStruct{"", now, userIP, "macAdress", "RegisteringBillingTokenAPI", "tokenModule", "", "", "", 0}

	TokenObj := token.StructToken{}
	decoder := json.NewDecoder(req.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&TokenObj)

	if err != nil {
		responseObj := responses.FindResponseByID("1")
		globalPkg.SendError(w, responseObj.EngResponse)
		globalPkg.WriteLog(logobj, responseObj.EngResponse, "failed")
		return

	}

	var validate = false
	var check = false
	var errorfound string
	//check for validate Token
	errorfound, check = validateToken(TokenObj, true)
	if check == true {
		errorfound = ""
		validate = true
	} else {
		globalPkg.SendError(w, errorfound)
		globalPkg.WriteLog(logobj, errorfound, "failed")
		return
	}
	// check is account billing
	if !isAccountBilling(TokenObj.InitiatorAddress) {
		responseObj := responses.FindResponseByID("103")
		globalPkg.SendError(w, responseObj.EngResponse)
		globalPkg.WriteLog(logobj, responseObj.EngResponse, "failed")

		return
	}
	//check if user balance cover total token amount
	if !validateUserAmount(TokenObj) {
		responseObj := responses.FindResponseByID("89")
		globalPkg.SendError(w, responseObj.EngResponse)
		globalPkg.WriteLog(logobj, responseObj.EngResponse, "failed")

		return
	}
	if validate == true {
		//The token data has been sent for validation
		errorfound, check = ValidatingNewToken(TokenObj)

		if check == true {
			// this variable will store the Ino Token amount that will calculated to the new token.
			// like this Ino Tokens (stored in TotalSupply temporarily) * TokenValue to get new TotalSupply.
			inoTokenAmount := (TokenObj.TokensTotalSupply * TokenObj.TokenValue) / globalPkg.GlobalObj.InoCoinToDollarRatio // IMPORTANT
			// TODO: token value is of dollar, to know how much to cut from ino token you'll have to
			// (TokensTotalSupply * TokenValue) / inoToDollarRatio.
			// input will be (TokensTotalSupply * TokenValue) / inoToDollarRatio. and output will be desired TokensTotalSupply from api.
			// TODO: change the validation om ino token accordingly to this formula.

			index := newTokenIndex()
			TokenObj.TokenID, _ = globalPkg.ConvertIntToFixedLengthString(index, globalPkg.GlobalObj.TokenIDStringFixedLength)

			// creating Tx that have sender and receiver = token.InitiatorAddress  .. but tokenID = "1" for sender
			// and tokenID = TokenObj.TokenID for receiver
			inoTokenID, _ := globalPkg.ConvertIntToFixedLengthString(0, globalPkg.GlobalObj.TokenIDStringFixedLength)
			tokenIcondata := TokenObj.TokenID + "_" + TokenObj.TokenIcon
			TokenObj.TokenIcon = ""
			tx1 := transactionModule.CreateTokenTx(TokenObj, inoTokenAmount, inoTokenID, false)
			//broadcast tx1 transaction
			broadcastTcp.BoardcastingTCP(tx1, "addTokenTransaction", "transaction")

			//approve the token to add it to database and broadcast token
			broadcastTcp.BoardcastingTCP(TokenObj, "addtoken", "token")
			broadcastTcp.BoardcastingTokenImgUDP(tokenIcondata, "addtokenimg", "addtokenimg")

			//success message
			responseObj := responses.FindResponseByID("91")
			globalPkg.SendResponseMessage(w, responseObj.EngResponse)
			globalPkg.WriteLog(logobj, responseObj.EngResponse, "failed")
			return
		} else {
			responseObj := responses.FindResponseByID("92")
			globalPkg.SendResponseMessage(w, responseObj.EngResponse)
			globalPkg.WriteLog(logobj, responseObj.EngResponse, "failed")

			return
		}

	} else {
		responseObj := responses.FindResponseByID("93")
		globalPkg.SendResponseMessage(w, responseObj.EngResponse)
		globalPkg.WriteLog(logobj, responseObj.EngResponse, "failed")

	}
}

//GetAllResponsesAPI get all responses from database
func GetAllResponsesAPI(w http.ResponseWriter, req *http.Request) {

	Adminobj := admin.Admin{}

	decoder := json.NewDecoder(req.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&Adminobj)

	if err != nil {
		globalPkg.SendError(w, "please enter your correct request ")
		return
	}
	if admin.ValidationAdmin(Adminobj) {
		jsonObj, _ := json.Marshal(responses.GetAllResponses())
		globalPkg.SendResponse(w, jsonObj)
	} else {
		globalPkg.SendError(w, "you are not the admin ")
	}
}
