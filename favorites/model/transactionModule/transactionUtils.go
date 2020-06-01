package transactionModule

import (
	"net/http"
	"strings"
	"unicode/utf8"

	"../globalPkg"
	"../responses"
	"../token"

	"../account"
	"../accountdb"
	"../broadcastTcp"
	"../cryptogrpghy"
	"../logpkg"
	"../transaction"

	// "../transaction"

	"encoding/json"
	"errors"
	"fmt"
	"time"
)

// func ValidateTx2(digitalWalletTx transaction.DigitalwalletTransaction, tx transaction.Transaction) string {
// 	if errStr := ValidateTransaction(digitalWalletTx); errStr == "" {

// 		outputSum := 0.0
// 		for _, outputObj := range tx.TransactionOutPut {
// 			outputSum += outputObj.OutPutValue
// 		}
// 		inputSum := 0.0
// 		for _, inputObj := range tx.TransactionInput {
// 			inputSum += inputObj.InputValue
// 		}
// 		if outputSum == inputSum {
// 			allOldTx := transaction.GetAllTransactionForPK(digitalWalletTx.Sender)
// 			inputexist := false
// 			for _, txObj := range allOldTx {
// 				oldInputTx, _ := json.Marshal(txObj.TransactionInput)
// 				newInputTx, _ := json.Marshal(tx.TransactionInput)
// 				if bytes.Compare(oldInputTx, newInputTx) == 0 {
// 					inputexist = true
// 				}
// 			}
// 			if !inputexist {
// 				return "true"
// 			} else {
// 				errorpk.AddError("ValidateTx2 Transaction module", "input is exist", "Validation Error")
// 				return "input is exist"
// 			}

// 		} else {
// 			errorpk.AddError("ValidateTx2 Transaction module", "digitalWalletTx is rong", "Validation Error")
// 			return "digitalWalletTx is rong"
// 		}

// 	} else {
// 		errorpk.AddError("ValidateTx2 Transaction module", errStr, "Validation Error")
// 		return errStr
// 	}
// }

// // AddTxToTransactionPool
// // TODO: benchmark the function.
// func ValidateTx(digitalWalletTx transaction.DigitalwalletTransaction, tx transaction.Transaction) string {
// 	if errStr := ValidateTransaction(digitalWalletTx); errStr == "" {
// 		unspentTxInputs, _ := GetUnspentAndSpentTxs(digitalWalletTx.Sender)
// 		// make a map of the unspent inputs Ids.
// 		unspentTxInputsIds := make(map[string]struct{}, len(unspentTxInputs))
// 		for _, unspentTxInput := range unspentTxInputs {
// 			unspentTxInputsIds[unspentTxInput.InputID] = struct{}{}
// 		}
// 		// check to see if every Tx Input id is in the mapped unspent inputs ids.
// 		var validTxInputs bool
// 		for _, txInput := range tx.TransactionInput {
// 			if _, ok := unspentTxInputsIds[txInput.InputID]; ok {
// 				validTxInputs = ok
// 			}
// 		}
// 		// if it's a valid Tx inputs, then check the amount of inputs == outputs.
// 		if validTxInputs {
// 			var inputSum, outputSum, transactionAmount float64
// 			for _, txInput := range tx.TransactionInput {
// 				inputSum += txInput.InputValue
// 			}
// 			for _, txOutput := range tx.TransactionOutPut {
// 				outputSum += txOutput.OutPutValue
// 				if txOutput.RecieverPublicKey == digitalWalletTx.Receiver {
// 					transactionAmount += txOutput.OutPutValue
// 				}
// 			}
// 			// if it's equal, then check if the Tx sender & receiver PubKeys are the same in digitalWalletTx.
// 			if inputSum == outputSum && transactionAmount == digitalWalletTx.Amount {
// 				var validSender, validReceiver bool
// 				for _, txInput := range tx.TransactionInput {
// 					if txInput.SenderPublicKey == digitalWalletTx.Sender {
// 						validSender = true
// 						break
// 					}
// 				}
// 				for _, txOutput := range tx.TransactionOutPut {
// 					if txOutput.RecieverPublicKey == digitalWalletTx.Receiver {
// 						validReceiver = true
// 						break
// 					}
// 				}
// 				if validReceiver && validSender {
// 					//transaction.AddTransaction(tx)
// 					return "true"
// 				} else {
// 					var note = "Tx receiver & sender public keys doesn't match digitalWalletTx receiver & sender public keys"
// 					errorpk.AddError("ValidateTx Transaction module", note, "Validation Error")
// 					return note
// 				}
// 			} else {
// 				var note = "Tx inputs isn't equal to the Tx outpus"
// 				errorpk.AddError("ValidateTx Transaction module", note, "Validation Error")
// 				return note
// 			}
// 		} else {
// 			var note = "Tx inputs doesn't match with the unspent Txs Inputs"
// 			errorpk.AddError("ValidateTx Transaction module", note, "Validation Error")
// 			return note
// 		}
// 	} else {

// 		errorpk.AddError("ValidateTx Transaction module", errStr, "Validation Error")
// 		return errStr
// 	}
// }

// checkIfAccountIsActive return the AccountStatus of the account with publicKey, else returns false.
func checkIfAccountIsActive(publicKey string) bool {

	// fmt.Println("account.GetAccountByAccountPubicKey(publicKey))", account.GetAccountByAccountPubicKey(publicKey).AccountStatus)
	// fmt.Println("-------------------------------------------------------")
	if (account.GetAccountByAccountPubicKey(publicKey)).AccountPublicKey != "" {
		return (account.GetAccountByAccountPubicKey(publicKey)).AccountStatus
	}

	return false

}

// DecryptDigitalWalletTx takes a string that consists of first 172 characters are encrypted key and the rest is the
// encrypted data. key is encrypted with RSA, data is encrypted with AES. it will first decrypt the key then decrypt the
// data using this key. and return transaction.DigitalwalletTransaction object.
func DecryptDigitalWalletTx(encryptedDigitalWalletTx string) (DigitalwalletTransaction, error) {

	if utf8.RuneCountInString(encryptedDigitalWalletTx) > 172 {
		encryptedKey := encryptedDigitalWalletTx[:172]
		fmt.Println("encryptedKey: ", encryptedKey)
		encryptedData := encryptedDigitalWalletTx[172:]
		fmt.Println("encryptedData: ", encryptedData)

		// hashedKey := string(cryptogrpghy.RSADEC(validator.CurrentValidator.ValidatorPrivateKey, encryptedKey)) // fmt.Sprintf("%x", cryptogrpghy.RSADEC(validator.CurrentValidator.ValidatorPrivateKey, encryptedKey))

		hashedKey, err := cryptogrpghy.Decrypt(globalPkg.RSAPublic, globalPkg.RSAPrivate, encryptedKey)
		if err != nil {
			return DigitalwalletTransaction{}, err
		}
		// encodedData := cryptogrpghy.KeyDecrypt(hashedKey, encryptedData)
		// fmt.Println("hashedKey :", hashedKey, "lenght :", len(hashedKey))
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

		var data DigitalwalletTransaction

		if err = json.Unmarshal([]byte(encodedData), &data); err != nil {
			fmt.Println("err json.Unmarshal", err)
			return DigitalwalletTransaction{}, errors.New("please enter a valid transaction data")
		}

		// fmt.Println("data", data)
		return data, nil
	}
	return DigitalwalletTransaction{}, errors.New("please enter a valid encrypted string")
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

//transferRefundRemaineder transfer remainder (difference in value) of refund Tx between Inovatian and service accounts, either it's a profit or loss for Inovatian.
func transferRefundRemaineder(tx DigitalwalletTransaction, profit bool) {
	if profit {
		tx.Receiver = accountdb.GetFirstAccount().AccountPublicKey
	} else {
		tx.Sender = accountdb.GetFirstAccount().AccountPublicKey
	}
	tx.Time = time.Now()
	transactionObj, _ := DigitalwalletToUTXOTrans(tx, true)

	broadcastTcp.BoardcastingTCP(transactionObj, "addTokenTransaction", "transaction")
	fmt.Println("transferRefundRemaineder transactionObj", transactionObj)
}

//makeTransferTX ownership functions --
func makeTransferTX(txData DigitalwalletTransaction, PublicKey, tknID string) string {
	// validate the digital wallet transaction.
	errStr, fine := ValidateTransactionToken(txData, false)
	if !fine {
		return errStr
	}
	transactionObj, _ := DigitalwalletToUTXOTrans(txData, false)
	mixedObj := MixedTxStruct{TxObj: transactionObj, DigitalTxObj: txData}
	res, _ := broadcastTcp.BoardcastingTCP(mixedObj, "addTransaction", "transaction")
	fmt.Println("transaction api res", res)
	if !res.Valid {
		responseObj := responses.FindResponseByID("1")
		errStr = responseObj.EngResponse

		// errStr = "there's a double spend transaction"
	} else {
		// update ownership for normal account
		ret := updateownershipfield(PublicKey, res.TxID, false, false) // update account owner
		if ret != "" {
			return ret
		}
		// update ownership for billing account
		ret2 := updateownershipfield(txData.Receiver, tknID, true, false) // receiver
		if ret2 != "" {
			return ret2
		}
		ret3 := updateownershipfield(txData.Sender, tknID, true, true) // sender
		if ret3 != "" {
			return ret3
		}
	}
	return errStr
}

//updateownershipfield update owner field
func updateownershipfield(publickey, txid string, owner, sender bool) string {
	acc := account.GetAccountByAccountPubicKey(publickey)
	ownershipObj, fine := accountdb.FindOwnershipByKey(acc.AccountIndex)
	// fmt.Println("history", ownershipObj)
	if !fine {
		// fmt.Println("not fine ...")
		if owner && sender {
			responseObj := responses.FindResponseByID("1")
			return responseObj.EngResponse
			// return "account has not any ownerships"
		} else if owner == false {
			responseObj := responses.FindResponseByID("1")
			return responseObj.EngResponse
			// return "this user not registered to any node"
		}
	}
	if !owner {
		ownershipObj.HistoryOwnership = append(ownershipObj.HistoryOwnership, ownershipObj.Ownership...)
		ownershipObj.Ownership = nil
		ownershipObj.Ownership = append(ownershipObj.Ownership, txid)
	} else {
		if sender {
			ownershipObj.Ownership = nil
			ownershipObj.Ownership = removeTxIDFromLst(ownershipObj.Ownership, txid)
			// fmt.Println("ownershipObj.Ownership ", ownershipObj.Ownership)
			ownershipObj.HistoryOwnership = append(ownershipObj.HistoryOwnership, txid)
		} else {
			ownershipObj.Ownership = append(ownershipObj.Ownership, txid)
		}
	}
	// ret := accountdb.OwnershipCreate(ownershipObj)
	// if !ret {
	// 	return "error in update ownership"
	// }
	broadcastTcp.BoardcastingTCP(ownershipObj, "", "create ownership")
	return ""
}

// removeTxIDFromLst remove tx id from tx
func removeTxIDFromLst(OwnershipLst []string, txid string) []string {
	var arr []string
	for _, item := range OwnershipLst {
		if item != txid {
			// OwnershipLst = append(OwnershipLst[i:], OwnershipLst[i+1:]...)
			arr = append(arr, item)
			fmt.Println("OwnershipLst ", OwnershipLst)
			//return OwnershipLst
		}
	}
	return arr
}

//createownershipfield create ownership field
func createownershipfield(publickey, txid string, owner bool) bool {

	acc := account.GetAccountByAccountPubicKey(publickey)
	var ownership accountdb.AccountOwnershipStruct
	ownership.AccountIndex = acc.AccountIndex
	ownership.Ownership = append(ownership.Ownership, txid)
	ownership.Owner = owner
	fmt.Println("fine ...")
	// account not found , create a new record
	broadcastTcp.BoardcastingTCP(ownership, "", "create ownership")
	//ret := accountdb.OwnershipCreate(ownership)
	return true
}

//getTransactionByID get transaction by id
func getTransactionByID(txID string) transaction.TransactionDB {
	txs := transaction.GetAllTransaction()
	for _, tx := range txs {
		if tx.TransactionID == txID {
			return tx
		}
	}
	return transaction.TransactionDB{}
}

//getAccountFromTransaction get Account from tx
func getAccountFromTransaction(tx transaction.TransactionDB) accountdb.AccountStruct {
	for _, out := range tx.TransactionOutPut {
		if !out.IsFee {
			return account.GetAccountByAccountPubicKey(out.RecieverPublicKey)
		}
	}
	return accountdb.AccountStruct{}
}

//verifyAccount verify account
func verifyAccount(w *http.ResponseWriter, logobj *logpkg.LogStruct, accObj *accountdb.AccountStruct) bool {
	// check if account found ?
	if accObj.AccountName == "" {
		responseObj := responses.FindResponseByID("142")
		globalPkg.SendError(*w, responseObj.EngResponse)
		globalPkg.WriteLog(*logobj, responseObj.EngResponse, "failed")

		return false
	}
	// check if account billing ?
	if accObj.AccountRole != "billing" {
		responseObj := responses.FindResponseByID("103")
		globalPkg.SendError(*w, responseObj.EngResponse)
		globalPkg.WriteLog(*logobj, responseObj.EngResponse, "failed")

		return false
	}
	return true
}

//isAccountBilling check if account billing
func isAccountBilling(sender string) bool {
	S := account.GetAccountByAccountPubicKey(sender)
	if S.AccountRole == "billing" {
		return true
	}
	return false
}

//createOwnershipToken create ownership token
func createOwnershipToken(tokenData token.StructToken) (string, string) {
	errorStr, fine := validateOwnershipToken(tokenData)
	if !fine {
		return errorStr, ""
	}
	lastInx := getLastTokenIndex()
	index := 0
	if lastInx != "-1" {
		// TODO : split LastIndex
		res := strings.Split(lastInx, "_")
		if len(res) == 2 {
			index = globalPkg.ConvertFixedLengthStringtoInt(res[1]) + 1
		} else {
			index = globalPkg.ConvertFixedLengthStringtoInt(lastInx) + 1
		}
	}
	inoTokenAmount := (tokenData.TokensTotalSupply * tokenData.TokenValue) / globalPkg.GlobalObj.InoCoinToDollarRatio

	tokenData.TokenID, _ = globalPkg.ConvertIntToFixedLengthString(index, globalPkg.GlobalObj.TokenIDStringFixedLength)
	inoTokenID, _ := globalPkg.ConvertIntToFixedLengthString(0, globalPkg.GlobalObj.TokenIDStringFixedLength)
	tokenIcondata := tokenData.TokenID + "_" + tokenData.TokenIcon
	tokenData.TokenIcon = ""
	tokenData.Ownership = true
	tx1 := CreateTokenTx(tokenData, inoTokenAmount, inoTokenID, true)
	//broadcast tx1 transaction
	broadcastTcp.BoardcastingTCP(tx1, "addTokenTransaction", "transaction")
	//approve the token to add it to database and broadcast token
	broadcastTcp.BoardcastingTCP(tokenData, "addtoken", "token")
	broadcastTcp.BoardcastingTokenImgUDP(tokenIcondata, "addtokenimg", "addtokenimg")
	return "", tokenData.TokenID
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

// validateOwnershipToken validate ownership token
func validateOwnershipToken(tokenData token.StructToken) (string, bool) {
	var errorfound string
	//validate token name ==20
	if utf8.RuneCountInString(tokenData.TokenName) < 4 || utf8.RuneCountInString(tokenData.TokenName) > 20 {
		responseObj := responses.FindResponseByID("67")
		errorfound = responseObj.EngResponse

		return errorfound, false
	}
	//validate token symbol == 4
	if utf8.RuneCountInString(tokenData.TokenSymbol) > 4 {
		responseObj := responses.FindResponseByID("68")
		errorfound = responseObj.EngResponse

		return errorfound, false
	}
	tokens := token.GetAllTokens()
	for _, tokenDataOld := range tokens {
		if tokenDataOld.TokenName == tokenData.TokenName {
			responseObj := responses.FindResponseByID("85")
			errorfound = responseObj.EngResponse

			return errorfound, false
		}
		if tokenDataOld.TokenSymbol == tokenData.TokenSymbol {
			responseObj := responses.FindResponseByID("86")
			errorfound = responseObj.EngResponse

			return errorfound, false
		}
		if tokenDataOld.Ownership && tokenData.InitiatorAddress == tokenDataOld.InitiatorAddress {
			responseObj := responses.FindResponseByID("143")
			errorfound = responseObj.EngResponse

			return errorfound, false
		}
	}
	// validate description if empty or == 100
	if utf8.RuneCountInString(tokenData.Description) == 0 || utf8.RuneCountInString(tokenData.Description) <= 100 {
		errorfound = ""
	} else {
		responseObj := responses.FindResponseByID("69")
		errorfound = responseObj.EngResponse

		return errorfound, false
	}
	//validate initiator address if empty
	if tokenData.InitiatorAddress == "" {
		responseObj := responses.FindResponseByID("71")
		errorfound = responseObj.EngResponse

		return errorfound, false
	}
	//validate initiator address if exist in account data
	accountobj := account.GetAccountByAccountPubicKey(tokenData.InitiatorAddress)
	if accountobj.AccountPublicKey == "" {
		responseObj := responses.FindResponseByID("71")
		errorfound = responseObj.EngResponse

		return errorfound, false
	}
	if accountobj.AccountPassword != tokenData.Password {
		responseObj := responses.FindResponseByID("11")
		errorfound = responseObj.EngResponse

		return errorfound, false
	}

	// validate Tokens Total Supply equal to 1
	if tokenData.TokensTotalSupply != 1 {
		responseObj := responses.FindResponseByID("73")
		errorfound = responseObj.EngResponse

		return errorfound, false
	}
	// validate Tokens Value equal to 1
	if tokenData.TokenValue != 1 {
		responseObj := responses.FindResponseByID("74")
		errorfound = responseObj.EngResponse

		return errorfound, false
	}
	// validate usage type
	if tokenData.UsageType != "security" {
		responseObj := responses.FindResponseByID("75")
		errorfound = responseObj.EngResponse

		return errorfound, false
	}
	//validate token precision from 0 to 5
	if tokenData.Precision != 0 {
		responseObj := responses.FindResponseByID("79")
		errorfound = responseObj.EngResponse

		return errorfound, false
	}
	//validate Tokens TokenType is mandatory public
	if tokenData.TokenType != "public" {
		responseObj := responses.FindResponseByID("144")
		errorfound = responseObj.EngResponse

		return errorfound, false
	}

	// Dynamic price	If the price of token is dynamic it gets its value from bidding platform.
	// Bidding platform API URL.
	//  based on ValueDynamic  True or false
	if tokenData.ValueDynamic == true {
		//for example value
		biddingplatformValue := 5.5
		tokenData.Dynamicprice = biddingplatformValue
	}
	//token.GetAllTokens()
	return "", true
}

//createOwnershipTransaction create ownership transaction
func createOwnershipTransaction(txData DigitalwalletTransaction, publickey, ownershipID string) string {
	// validate the digital wallet transaction.
	errStr, fine := ValidateTransactionToken(txData, false)
	if !fine {
		return errStr
	}
	transactionObj, _ := DigitalwalletToUTXOTrans(txData, false)
	transactionObj.OwnershipID = ownershipID // store token id for normal account that have this ownership
	mixedObj := MixedTxStruct{TxObj: transactionObj, DigitalTxObj: txData}
	res, _ := broadcastTcp.BoardcastingTCP(mixedObj, "addTransaction", "transaction")
	fmt.Println("transaction api res", res)
	if !res.Valid {
		responseObj := responses.FindResponseByID("1")
		errStr = responseObj.EngResponse
		// errStr = "there's a double spend transaction"
	} else {
		// update ownership for normal account
		ret := createownershipfield(publickey, res.TxID, false) // update account owner
		// update ownership for billing account
		ret2 := createownershipfield(txData.Receiver, ownershipID, true)
		if !ret || !ret2 {
			responseObj := responses.FindResponseByID("145")
			return responseObj.EngResponse

		}
	}
	return errStr
}
