package tokenModule

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"../logpkg"

	"strings"
	//"time"

	"../account"
	"../admin"
	"../broadcastTcp"
	"../globalPkg"
	"../token"
	"../transaction"
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
		globalPkg.SendError(w, "please enter your correct request")
		globalPkg.WriteLog(logobj, "please enter your correct request", "failed")
		return

	}

	var validate = false
	var check = false
	var errorfound string
	//check for validate Token
	errorfound, check = validateToken(TokenObj)
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
		globalPkg.SendError(w, "your balance amount less than total amount of tokens ")
		globalPkg.WriteLog(logobj, "your balance amount less than total amount of tokens ", "failed")
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
			if len(tokenIcondata) > 20000 {
				globalPkg.SendError(w, " max size of icon token is 10 kb")
				globalPkg.WriteLog(logobj, "max size of icon token is 10 kb", "failed")
				return
			}
			TokenObj.TokenIcon = ""
			TokenObj.TokenTime = globalPkg.UTCtime()
			tx1 := transactionModule.CreateTokenTx(TokenObj, inoTokenAmount, inoTokenID)

			//broadcast tx1 transaction
			broadcastTcp.BoardcastingTCP(tx1, "addTokenTransaction", "transaction")

			//approve the token to add it to database and broadcast token
			broadcastTcp.BoardcastingTCP(TokenObj, "addtoken", "token")
			broadcastTcp.BoardcastingTokenImgUDP(tokenIcondata, "addtokenimg", "addtokenimg")

			//success message
			// sendJSON, _ := json.Marshal("The token has been created")
			globalPkg.SendResponseMessage(w, "The token has been created")
			globalPkg.WriteLog(logobj, "The token has been created", "success")
		} else {
			globalPkg.SendError(w, "The token is invalid :  "+errorfound)
			globalPkg.WriteLog(logobj, "The token is invalid ", "failed")
			return
		}

	} else {
		globalPkg.SendError(w, "The application refused to create token.")
		globalPkg.WriteLog(logobj, "The application refused to create token.", "failed")
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
		globalPkg.SendError(w, "please enter your correct request")
		globalPkg.WriteLog(logobj, "please enter your correct request", "failed")
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
		globalPkg.SendError(w, "Token ID  don't exist before")
		globalPkg.WriteLog(logobj, "Token ID  don't exist before", "failed")
		return
	}
	//check on reissuability updates
	if tokenOldObj.Reissuability == false && TokenObj.Reissuability == true {
		// data exist in contract ID or user public key use private token
		if TokenObj.ContractID != "" || len(TokenObj.UserPublicKey) != 0 {
			globalPkg.SendError(w, "it rejects the update: there are token holders unlike this user.")
			globalPkg.WriteLog(logobj, "it rejects the update: there are token holders unlike this user.", "failed")
			return
		}
	}

	if tokenOldObj.Reissuability == true && TokenObj.TokensTotalSupply > tokenOldObj.TokensTotalSupply {
		//check if user balance cover total token amount
		if !validateUserAmount(TokenObj) {
			globalPkg.SendError(w, "it rejects the update: the user have not balance covers the new increase with the current price to the token.")
			globalPkg.WriteLog(logobj, "it rejects the update: the user have not balance covers the new increase with the current price to the token.", "failed")
			return
		}
	}

	//check on change on type of token updates
	if tokenOldObj.TokenType == "public" && TokenObj.TokenType == "private" {

		if len(tokenOldObj.ContractID) > 4 || len(tokenOldObj.ContractID) <= 60 {
			globalPkg.SendError(w, "it rejects the update:the token has association with any contract.")
			globalPkg.WriteLog(logobj, "it rejects the update:  the token has association with any contract.", "failed")
			return
		}
	}

	if tokenOldObj.TokenType == "private" && TokenObj.TokenType == "public" {

		if len(tokenOldObj.ContractID) < 4 || len(tokenOldObj.ContractID) > 60 {
			globalPkg.SendError(w, "please enter Contract ID associated with public token")
			globalPkg.WriteLog(logobj, "please enter Contract ID associated with public token", "failed")
			return
		}
	}

	if validate == true {
		//approve the token to update it to database and broadcast update token
		tokenIcondata := TokenObj.TokenID + "_" + TokenObj.TokenIcon
		if len(tokenIcondata) > 20000 {
			globalPkg.SendError(w, " max size of icon token is 10 kb")
			globalPkg.WriteLog(logobj, "max size of icon token is 10 kb", "failed")
			return
		}
		TokenObj.TokenIcon = ""
		broadcastTcp.BoardcastingTCP(TokenObj, "updatetoken", "token")
		broadcastTcp.BoardcastingTokenImgUDP(tokenIcondata, "addtokenimg", "addtokenimg")
		sendJSON, _ := json.Marshal("The token data has been updated")
		globalPkg.SendResponse(w, sendJSON)
		globalPkg.WriteLog(logobj, string(sendJSON), "success")
	} else {
		globalPkg.SendError(w, "The application refused to update token. ")
		globalPkg.WriteLog(logobj, "The application refused to update token. ", "failed")
	}
}

//ExploringUserTokensAPI Providing info for user about his tokens
func ExploringUserTokensAPI(w http.ResponseWriter, req *http.Request) {
	now, userIP := globalPkg.SetLogObj(req)
	logobj := logpkg.LogStruct{"_", now, userIP, "macAdress", "ExploringUserTokensAPI", "tokenModule", "_", "_", "_", 0}

	var accountPasswordAndPubKey transactionModule.AccountPasswordAndPubKey

	decoder := json.NewDecoder(req.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&accountPasswordAndPubKey)

	if err != nil {
		globalPkg.SendError(w, "please enter your correct request")
		globalPkg.WriteLog(logobj, "please enter your correct request", "failed")
		return
	}
	//check public key and password is valid
	accObj := account.GetAccountByAccountPubicKey(accountPasswordAndPubKey.PublicKey)
	if accObj.AccountPublicKey == "" || accObj.AccountPassword == "" || accObj.AccountName == "" {
		globalPkg.SendNotFound(w, "The Account with this Public key is not found.")
		globalPkg.WriteLog(logobj, "The Account with this Public key is not found.", "failed")
		return
	} else if accountPasswordAndPubKey.Password != accObj.AccountPassword {
		globalPkg.SendError(w, "The given password is incorrect.")
		globalPkg.WriteLog(logobj, "The given password is incorrect.", "failed")
		return
	}

	tokens := token.GetAllTokens()     //get tokens from ledger
	tokenList := []token.StructToken{} //append all tokens types private or public

	//if user public key create token type public,private
	for _, taken := range tokens {
		if taken.InitiatorAddress == accountPasswordAndPubKey.PublicKey {
			// TODO: Please, Reconsider this condition.
			if taken.TokenType == "private" || taken.TokenType == "public" {
				tokenList = append(tokenList, taken)
			}
		}
	}
	tokenids := transactionModule.GetTokenIDusedbyrecieverpk(accountPasswordAndPubKey.PublicKey)

	for _, tokenid := range tokenids {
		tokenObj2 := token.FindTokenByid(tokenid)
		if tokenObj2.TokenType == "public" {
			containtokenid := ContainstokenID(tokenList, tokenObj2.TokenID)
			if !containtokenid {
				tokenList = append(tokenList, tokenObj2)
			}
		}
	}

	//Get all public tokens where this user is one of  their holders . Table Contact ID
	if len(tokenList) == 0 {
		globalPkg.SendError(w, "The user has not create any token or use any public one.")
		globalPkg.WriteLog(logobj, "The user has not create any token or use any public one.", "failed")
	} else {
		sendJSON, _ := json.Marshal(tokenList)
		globalPkg.SendResponse(w, sendJSON)
		globalPkg.WriteLog(logobj, "The user tokens returned", "success")
	}
}

// RefundToken  either with Inotoken or with fiat currency
func RefundToken(w http.ResponseWriter, req *http.Request) {

	//log
	now, userIP := globalPkg.SetLogObj(req)
	logobj := logpkg.LogStruct{"_", now, userIP, "macAdress", "RefundToken", "tokenModule", "_", "_", "_", 0}

	tokenTransactionObj := transaction.RefundDigitalWalletTx{}
	decoder := json.NewDecoder(req.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&tokenTransactionObj)

	if err != nil {
		globalPkg.SendError(w, "please enter your correct request")
		globalPkg.WriteLog(logobj, "please enter your correct request", "failed")
		return
	}
	//check for validate data transfer Token
	errorfound := transactionModule.ValidateRefundTransaction(tokenTransactionObj)
	if errorfound == "" {
		tokenObj := token.FindTokenByid(tokenTransactionObj.TokenID)

		transactionObj := transactionModule.RefundTokenTx(tokenObj, tokenTransactionObj)
		// TODO: update token to decrease the amount of refunded token??
		transactionObj.TransactionTime = globalPkg.UTCtime()
		fmt.Println("=======================================================")
		fmt.Println(transactionObj)
		fmt.Println("=======================================================")
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
	} else {
		globalPkg.SendError(w, errorfound)

	}
}

func getLastTokenIndex() string {
	var Token token.StructToken
	Token = token.GetLastToken()
	if Token.TokenID == "" {
		return "-1"
	}

	return Token.TokenID
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
		globalPkg.SendError(w, "please enter your correct request")
		globalPkg.WriteLog(logobj, "please enter your correct request", "failed")
		return
	}
	if admin.ValidationAdmin(Adminobj) {
		sendJSON, _ := json.Marshal(token.GetAllTokens())
		globalPkg.SendResponse(w, sendJSON)
		globalPkg.WriteLog(logobj, string(sendJSON), "success")
	} else {
		globalPkg.SendError(w, "You are not admin")
		globalPkg.WriteLog(logobj, "You are not admin", "failed")
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
		globalPkg.SendError(w, "please enter your correct request")
		globalPkg.WriteLog(logobj, "please enter your correct request", "failed")
		return
	}

	tokenObj := token.FindTokenByTokenName(Tokenname.Name)
	if tokenObj.TokenName != "" {
		sendJSON, _ := json.Marshal(tokenObj)
		globalPkg.SendResponse(w, sendJSON)
		globalPkg.WriteLog(logobj, string(sendJSON), "success")
		return
	}
	globalPkg.SendError(w, "no token name")
	globalPkg.WriteLog(logobj, "no token name", "failed")
}

type Tokendata struct {
	Day        string
	Tokenvalue float64
}

//GetTokensValueLastdaysAPI get token values in last ten days --
func GetTokensValueLastdaysAPI(w http.ResponseWriter, req *http.Request) {
	//log
	now, userIP := globalPkg.SetLogObj(req)
	logobj := logpkg.LogStruct{"_", now, userIP, "macAdress", "GetTokensValueLastdaysAPI", "tokenModule", "_", "_", "_", 0}

	var TokenName globalPkg.JSONString

	decoder := json.NewDecoder(req.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&TokenName)

	if err != nil {
		globalPkg.SendError(w, "please enter your correct request ")
		globalPkg.WriteLog(logobj, "failed to decode admin object", "failed")
		return
	}

	tokenObj := token.FindTokenByTokenName(TokenName.Name)
	if tokenObj.TokenName != "" {
		data := make(map[string]float64)
		var day string

		// _,_,day	:= tokenObj.TokenTime.Date()
		now := globalPkg.UTCtime()
		diff := now.Sub(tokenObj.TokenTime)
		days := int(diff.Hours() / 24)
		if days == 0 {
			days += days + 1
		}
		for j := 0; j <= days; j++ {
			if j > 10 {
				break
			}
			_, _, d := globalPkg.UTCtime().Date()
			day = strconv.Itoa(d - j)
			data[day] = tokenObj.TokenValue
		}

		// Convert map to slice of key-value pairs.
		tkValue := []Tokendata{}
		for key, value := range data {
			var tk Tokendata
			tk.Day = key
			tk.Tokenvalue = value
			tkValue = append(tkValue, tk)
		}
		sendJSON, _ := json.Marshal(tkValue)
		globalPkg.SendResponse(w, sendJSON)
		globalPkg.WriteLog(logobj, string(sendJSON), "success")
		return
	}
	globalPkg.SendError(w, "no token name")
	globalPkg.WriteLog(logobj, "no token name", "failed")

}

//ContainstokenID Contains tells whether a contains x.
func ContainstokenID(TokenID []token.StructToken, tokenid string) bool {
	for _, n := range TokenID {
		if tokenid == n.TokenID {
			return true
		}
	}
	return false
}
