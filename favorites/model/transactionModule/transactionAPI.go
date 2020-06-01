package transactionModule

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"../transaction"

	"../logfunc"
	"../validator"

	"../account"
	"../accountdb"
	"../admin"
	"../broadcastTcp"
	"../globalPkg"
	"../logpkg"
	"../responses"
	"../token"
)

// GetAllTransactionsAPI endpoint to get all the transactions
func GetAllTransactionsAPI(w http.ResponseWriter, req *http.Request) {
	//log
	now, userIP := globalPkg.SetLogObj(req)
	logobj := logpkg.LogStruct{"_", now, userIP, "macAdress", "AddNewTransaction", "transactionModule", "_", "_", "_", 0}

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
	// logobj.InputData = Adminobj.AdminUsername + Adminobj.AdminPassword
	logobj.InputData = Adminobj.UsernameAdmin + Adminobj.PasswordAdmin
	if admin.ValidationAdmin(Adminobj) {
		sendJSON, _ := json.Marshal(transaction.GetPendingTransactions())
		globalPkg.SendResponse(w, sendJSON)
		globalPkg.WriteLog(logobj, "get all transaction success", "success")
	} else {
		responseObj := responses.FindResponseByID("2")
		globalPkg.SendError(w, responseObj.EngResponse)
		globalPkg.WriteLog(logobj, responseObj.EngResponse, "failed")
	}
}

//AddNewTransaction add new transaction api
func AddNewTransaction(w http.ResponseWriter, req *http.Request) {
	//log
	now, userIP := globalPkg.SetLogObj(req)

	found, logobj := logpkg.CheckIfLogFound(userIP)

	if found && now.Sub(logobj.Currenttime).Seconds() > globalPkg.GlobalObj.DeleteAccountTimeInseacond {

		logobj.Count = 0
		broadcastTcp.BoardcastingTCP(logobj, "", "AddAndUpdateLog")

	}
	if found && logobj.Count >= 10 {
		responseObj := responses.FindResponseByID("6")
		globalPkg.SendError(w, responseObj.EngResponse)
		globalPkg.WriteLog(logobj, responseObj.EngResponse, "failed")

		return
	}

	if !found {

		Logindex := userIP.String() + "_" + logfunc.NewLogIndex()

		logobj = logpkg.LogStruct{Logindex, now, userIP, "macAdress", "AddNewTransation", "transactionModule", "", "", "_", 0}
	}
	logobj = logfunc.ReplaceLog(logobj, "AddNewTransation", "transactionModule")

	type StringData struct {
		EncryptedData string
		SessionID     string
	}
	var encryptedDigitalWalletTx StringData
	var errStr string
	var digitalWalletTransaction DigitalwalletTransaction
	var transactionObj transaction.Transaction

	decoder := json.NewDecoder(req.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&encryptedDigitalWalletTx)
	// err := decoder.Decode(&digitalWalletTransaction)

	if err != nil {
		responseObj := responses.FindResponseByID("1")
		globalPkg.SendError(w, responseObj.EngResponse)
		globalPkg.WriteLog(logobj, responseObj.EngResponse, "failed")
		fmt.Printf("\n \n ******** add new transaction error is : %v ********** \n \n", err)
		return
	}
	// decrypt digital wallet transaction.
	digitalWalletTransaction, err = DecryptDigitalWalletTx(encryptedDigitalWalletTx.EncryptedData)
	digitalWalletTransaction.Validator = validator.CurrentValidator.ValidatorIP
	if err != nil {
		globalPkg.SendError(w, err.Error())
		return
	}

	// validate the digital wallet transaction.
	errStr, noError := ValidateTransactionToken(digitalWalletTransaction, true)
	// validate the digital wallet transaction.
	firstaccount := accountdb.GetFirstAccount()
	if noError {
		if digitalWalletTransaction.Sender == "" && firstaccount.AccountPublicKey == digitalWalletTransaction.Receiver {
			transactionObj = addcoins(digitalWalletTransaction)
		} else {

			transactionObj, errStr = DigitalwalletToUTXOTrans(digitalWalletTransaction, true)
			if errStr != "" {
				globalPkg.SendError(w, errStr)
				globalPkg.WriteLog(logobj, errStr, "failed")
				return
			}
		}
		mixedObj := MixedTxStruct{TxObj: transactionObj, DigitalTxObj: digitalWalletTransaction}

		//go func() {
		//	res1 := broadcastTcp.BoardcastingTCP(mixedObj, "addTransaction", "transaction")
		//	fmt.Println("transaction api res1", res1)
		//}()
		//fmt.Println("\n )())()()( transaction api transactionObj", transactionObj)
		var res broadcastTcp.TxBroadcastResponse
		broadcastTcp.BoardcastingTCP(mixedObj, "addTransaction+"+transactionObj.TransactionID, "transaction")

		fmt.Println("transaction api res", res)

		var sumofRemovedvalidator int
		validatorlist := validator.GetAllValidators()
		for _, validatObj := range validatorlist {
			if validatObj.ValidatorRemove == false {
				sumofRemovedvalidator++
			}
		}
		// fmt.Println("  sum of validator    :   ", sumofRemovedvalidator)
		var transReponse broadcastTcp.TransactionReponse
		//loop on reponse transaction until get reponse transaction id
		for {
			transReponse = broadcastTcp.GetMap(transactionObj.TransactionID)
			if transReponse.Count != 0 {
				if transReponse.Count == sumofRemovedvalidator {
					if transReponse.Status == true {
						res.Valid = true
					} else {
						res.Valid = false
					}
					goto breakout
				}

			}

		}
		// to break from infinite loop
	breakout:

		if !res.Valid {
			noError = res.Valid
			responseObj := responses.FindResponseByID("109")
			errStr = responseObj.EngResponse

		}
	}
	logobj.InputData = digitalWalletTransaction.Sender + "," + digitalWalletTransaction.Receiver + "," + strconv.FormatFloat(float64(digitalWalletTransaction.Amount), 'f', 6, 64)

	if noError {

		obj := account.GetAccountByAccountPubicKey(digitalWalletTransaction.Receiver)
		//if obj.SessionID != encryptedDigitalWalletTx.SessionID {
		//	globalPkg.SendError(w, "Invalid SessionId")
		//	return
		//}

		//	fmt.Println("***************88888*************obbbbj??", obj)
		// sendJson, _ := json.Marshal(transactionObj)
		//w.WriteHeader(http.StatusOK)
		tmp := strconv.FormatFloat(float64(digitalWalletTransaction.Amount), 'f', 6, 64)
		tokenObj := token.FindTokenByid(digitalWalletTransaction.TokenID)

		// notification
		// this part is for transaction notification
		sessionLst := account.Getaccountsessionid(digitalWalletTransaction.Receiver) //get all sesions for the the reciever identified by his public key

		// var message Notify
		// var flatterSessionSdList []string
		// message.Message = logobj.OutputData // declared before ~= "your trans with 15 coin sended to Omar"
		// for _, sID := range sessionLst {    //range over current session lst and send notification to them all
		// 	s := strings.Split(sID, "_")
		// 	message.SessionID = s[0]
		// 	if s[1] == "flatter" {
		// 		flatterSessionSdList = append(flatterSessionSdList, message.SessionID)
		// 	} else {
		// 		msg, err := json.Marshal(message)
		// 		if err != nil {
		// 			fmt.Println(err)
		// 			return
		// 		}
		// 		globalPkg.SendRequest(msg, globalPkg.GlobalObj.DigitalwalletIpNotfication, "POST")
		// 	}
		// }
		// end of notification
		var timp fltr
		timp.Str = "Your transaction with " + tmp + " coin has been sent successfully to " + obj.AccountName
		timp.Lst = sessionLst
		sendJSON, _ := json.Marshal(timp)
		globalPkg.SendResponse(w, sendJSON)

		globalPkg.WriteLog(logobj, fmt.Sprintf(
			"Your transaction with %v of %v Token has been sent successfully to %v.",
			digitalWalletTransaction.Amount, tokenObj.TokenName, obj.AccountInitialUserName,
		), "success")
		if logobj.Count > 0 {
			logobj.Count = 0
			broadcastTcp.BoardcastingTCP(logobj, "", "AddAndUpdateLog")

		}
		// globalPkg.SendResponseMessage(w, "Your transaction with "+tmp+" coin has been sent successfully to "+obj.AccountName)
		// return
		//_, err := w.Write([]byte("Your transaction with " + tmp + " coin has been sent successfully to " + obj.AccountName))
		//fmt.Println("************************\n write err", err)
	} else {
		fmt.Println(errStr)
		globalPkg.SendError(w, errStr)
		globalPkg.WriteLog(logobj, errStr, "failed")
		logobj.Count = logobj.Count + 1

		broadcastTcp.BoardcastingTCP(logobj, "", "AddAndUpdateLog")

	}
}

// GetBalance endpoint to get the balance
func GetBalance(w http.ResponseWriter, req *http.Request) {
	now, userIP := globalPkg.SetLogObj(req)
	logobj := logpkg.LogStruct{"_", now, userIP, "macAdress", "GetBAlance", "transactionModuleApi", "_", "_", "_", 0}

	var accountPasswordAndPubKey AccountPasswordAndPubKey

	decoder := json.NewDecoder(req.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&accountPasswordAndPubKey)

	if err != nil {
		responseObj := responses.FindResponseByID("1")
		globalPkg.SendError(w, responseObj.EngResponse)
		globalPkg.WriteLog(logobj, responseObj.EngResponse, "failed")
		return
	}
	//if err := json.NewDecoder(req.Body).Decode(&accountPasswordAndPubKey); err != nil {
	//	errorpk.AddError("GetTransactionByPublicKey API Transaction module package ", "can't decode json to accountPasswordAndPubKey struct")
	//	w.WriteHeader(http.StatusServiceUnavailable)
	//	w.Write([]byte("Service is Unavailable"))
	//}

	accObj := account.GetAccountByAccountPubicKey(accountPasswordAndPubKey.PublicKey)
	if accObj.AccountPublicKey == "" || accObj.AccountPassword == "" || accObj.AccountName == "" {
		responseObj := responses.FindResponseByID("10")
		globalPkg.SendNotFound(w, responseObj.EngResponse)
		globalPkg.WriteLog(logobj, responseObj.EngResponse, "failed")
		return

	} else if !checkIfAccountIsActive(accountPasswordAndPubKey.PublicKey) {
		responseObj := responses.FindResponseByID("127")
		globalPkg.SendError(w, responseObj.EngResponse)
		globalPkg.WriteLog(logobj, responseObj.EngResponse, "failed")
		return
	} else if accountPasswordAndPubKey.Password != accObj.AccountPassword {
		responseObj := responses.FindResponseByID("11")
		globalPkg.SendError(w, responseObj.EngResponse)
		globalPkg.WriteLog(logobj, responseObj.EngResponse, "failed")
		return

	}
	//logobj.OutputData = GetAccountBalanceStatement(accountPasswordAndPubKey.PublicKey)

	// logobj.Process = "success"
	// logpkg.WriteOnlogFile(logobj)
	balanceObj := GetAccountBalanceStatement(accObj, "")

	var balance []BalanceAccount
	var BalanceAccountObj BalanceAccount
	for key, value := range balanceObj {
		tokenObj := token.FindTokenByid(key)
		BalanceAccountObj.Tokenname = tokenObj.TokenName
		BalanceAccountObj.Balance = value
		balance = append(balance, BalanceAccountObj)
	}
	sendJSON, _ := json.Marshal(balance)
	globalPkg.SendResponse(w, sendJSON)
	globalPkg.WriteLog(logobj, "get balance success", "success")

}

//GetTransactionByPublicKey used to get all transactions linked to the account by using the provided account PubKey
func GetTransactionByPublicKey(w http.ResponseWriter, req *http.Request) {
	now, userIP := globalPkg.SetLogObj(req)
	logobj := logpkg.LogStruct{"_", now, userIP, "macAdress", "GetTransactionByPublicKey", "transactionModule", "_", "_", "_", 0}

	var accountPasswordAndPubKey AccountPasswordAndPubKey
	logobj.InputData = accountPasswordAndPubKey.PublicKey + "," + accountPasswordAndPubKey.Password

	decoder := json.NewDecoder(req.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&accountPasswordAndPubKey)

	if err != nil {
		responseObj := responses.FindResponseByID("1")
		globalPkg.SendError(w, responseObj.EngResponse)
		globalPkg.WriteLog(logobj, responseObj.EngResponse, "failed")
		return
	}
	//if err := json.NewDecoder(req.Body).Decode(&accountPasswordAndPubKey); err != nil {
	//	errorpk.AddError("GetTransactionByPublicKey API Transaction module package ", "can't decode json to accountPasswordAndPubKey struct")
	//	w.WriteHeader(http.StatusServiceUnavailable)
	//	w.Write([]byte("Service is Unavailable"))
	//}
	accObj := account.GetAccountByAccountPubicKey(accountPasswordAndPubKey.PublicKey)
	if accObj.AccountPublicKey == "" || accObj.AccountPassword == "" || accObj.AccountName == "" {
		responseObj := responses.FindResponseByID("10")
		globalPkg.SendNotFound(w, responseObj.EngResponse)
		globalPkg.WriteLog(logobj, responseObj.EngResponse, "failed")
		return
		//w.Write([]byte("The Account with this Public key is not found"))
	} else if !checkIfAccountIsActive(accountPasswordAndPubKey.PublicKey) {
		responseObj := responses.FindResponseByID("127")
		globalPkg.SendNotFound(w, responseObj.EngResponse)
		globalPkg.WriteLog(logobj, responseObj.EngResponse, "failed")

		return

	} else if accountPasswordAndPubKey.Password != accObj.AccountPassword {
		responseObj := responses.FindResponseByID("11")
		globalPkg.SendNotFound(w, responseObj.EngResponse)
		globalPkg.WriteLog(logobj, responseObj.EngResponse, "failed")
		return

	}

	TransactionMap := GetTransactionsByPublicKey(accObj)

	for key, trasLst := range TransactionMap {

		for index, transactionObj := range trasLst {
			transactionObj.TokenID = (token.FindTokenByid(transactionObj.TokenID)).TokenName
			TransactionMap[key][index] = transactionObj
		}

	}
	sendJSON, _ := json.Marshal(TransactionMap)
	globalPkg.SendResponse(w, sendJSON)
	globalPkg.WriteLog(logobj, "get transaction by pk success", "success")
}

//GetTransactionDbByIdAPI get transaction DB by ID
func GetTransactionDbByIdAPI(w http.ResponseWriter, r *http.Request) {
	now, userIP := globalPkg.SetLogObj(r)
	logobj := logpkg.LogStruct{"_", now, userIP, "macAdress", "GetTransactionDbByIdAPI", "transactionModule", "_", "_", "_", 0}

	adminObj := admin.Admin{}
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&adminObj)

	if err != nil {
		responseObj := responses.FindResponseByID("1")
		globalPkg.SendNotFound(w, responseObj.EngResponse)
		globalPkg.WriteLog(logobj, responseObj.EngResponse, "failed")
		return
	}

	if admin.ValidationAdmin(adminObj) {
		TransactionKey := fmt.Sprintf("%v", adminObj.ObjectInterface)
		// fmt.Println(" ",TransactionKey)
		tx := transaction.GetTransactionByKey(TransactionKey)
		sendJSON, _ := json.Marshal(tx)
		globalPkg.SendResponse(w, sendJSON)
		globalPkg.WriteLog(logobj, string(sendJSON), "success")
	} else {
		responseObj := responses.FindResponseByID("2")
		globalPkg.SendNotFound(w, responseObj.EngResponse)
		globalPkg.WriteLog(logobj, responseObj.EngResponse, "failed")
	}
}

//GetAllTransactionDbAPI get all transaction database
func GetAllTransactionDbAPI(w http.ResponseWriter, req *http.Request) {
	//log
	now, userIP := globalPkg.SetLogObj(req)
	logobj := logpkg.LogStruct{"_", now, userIP, "macAdress", "GetAllTransactionDbAPI", "transactionModule", "_", "_", "_", 0}

	Adminobj := admin.Admin{}

	decoder := json.NewDecoder(req.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&Adminobj)

	if err != nil {
		responseObj := responses.FindResponseByID("1")
		globalPkg.SendNotFound(w, responseObj.EngResponse)
		globalPkg.WriteLog(logobj, responseObj.EngResponse, "failed")
		return
	}

	if admin.ValidationAdmin(Adminobj) {
		sendJSON, _ := json.Marshal(transaction.GetAllTransaction())
		globalPkg.SendResponse(w, sendJSON)
		globalPkg.WriteLog(logobj, string(sendJSON), "success")
	} else {
		responseObj := responses.FindResponseByID("2")
		globalPkg.SendNotFound(w, responseObj.EngResponse)
		globalPkg.WriteLog(logobj, responseObj.EngResponse, "failed")
	}
}

//GetAllTransactionforOneTokenAPI token initiator see all transaction for this token
func GetAllTransactionforOneTokenAPI(w http.ResponseWriter, req *http.Request) {
	//log
	now, userIP := globalPkg.SetLogObj(req)
	logobj := logpkg.LogStruct{"", now, userIP, "macAdress", "GetAllTransactionforOneTokenAPI", "transactionModule", "", "", "", 0}

	tokenObj := token.StructToken{}
	decoder := json.NewDecoder(req.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&tokenObj)
	if err != nil {
		responseObj := responses.FindResponseByID("1")
		globalPkg.SendNotFound(w, responseObj.EngResponse)
		globalPkg.WriteLog(logobj, responseObj.EngResponse, "failed")
		return
	}
	//validate initiator address if exist in account data
	accountobj := account.GetAccountByAccountPubicKey(tokenObj.InitiatorAddress)
	if accountobj.AccountPublicKey == "" {
		responseObj := responses.FindResponseByID("71")
		globalPkg.SendNotFound(w, responseObj.EngResponse)
		globalPkg.WriteLog(logobj, responseObj.EngResponse, "failed")
		return
	}
	if accountobj.AccountPassword != tokenObj.Password {
		responseObj := responses.FindResponseByID("11")
		globalPkg.SendNotFound(w, responseObj.EngResponse)
		globalPkg.WriteLog(logobj, responseObj.EngResponse, "failed")
		return
	}

	tokenobj := token.FindTokenByTokenName(tokenObj.TokenName)
	if tokenobj.TokenName != "" {
		if tokenobj.InitiatorAddress != tokenObj.InitiatorAddress {
			responseObj := responses.FindResponseByID("110")
			globalPkg.SendNotFound(w, responseObj.EngResponse)
			globalPkg.WriteLog(logobj, responseObj.EngResponse, "failed")

			return
		}
	} else {
		responseObj := responses.FindResponseByID("111")
		globalPkg.SendNotFound(w, responseObj.EngResponse)
		globalPkg.WriteLog(logobj, responseObj.EngResponse, "failed")

		return
	}

	if account.ContainstokenID(accountobj.AccountTokenID, tokenobj.TokenID) {
		returnedTxs := GetAllTransactionsByTokenID(tokenobj.TokenID)
		sendJSON, _ := json.Marshal(returnedTxs)
		globalPkg.SendResponse(w, sendJSON)
		globalPkg.WriteLog(logobj, string(sendJSON), "success")
		return
	}
}

// CreateOwnershipAPI for creating ownershipToken
func CreateOwnershipAPI(w http.ResponseWriter, req *http.Request) {
	var data OwnershipTX
	//log
	now, userIP := globalPkg.SetLogObj(req)
	found, logobj := logpkg.CheckIfLogFound(userIP)
	if found && now.Sub(logobj.Currenttime).Seconds() > globalPkg.GlobalObj.DeleteAccountTimeInseacond {
		logobj.Count = 0
		broadcastTcp.BoardcastingTCP(logobj, "", "AddAndUpdateLog")
	}
	if found && logobj.Count >= 10 {
		responseObj := responses.FindResponseByID("6")
		globalPkg.SendNotFound(w, responseObj.EngResponse)
		globalPkg.WriteLog(logobj, responseObj.EngResponse, "failed")
		return
	}
	if !found {
		Logindex := userIP.String() + "_" + logfunc.NewLogIndex()
		logobj = logpkg.LogStruct{Logindex, now, userIP, "macAdress", "CreateOwnershipAPI", "transactionModule", "", "", "_", 0}
	}
	logobj = logfunc.ReplaceLog(logobj, "CreateOwnershipAPI", "transactionModule")
	decoder := json.NewDecoder(req.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&data)
	if err != nil {
		responseObj := responses.FindResponseByID("1")
		globalPkg.SendNotFound(w, responseObj.EngResponse)
		globalPkg.WriteLog(logobj, responseObj.EngResponse, "failed")
		return
	}
	tokenData := data.Token
	txData := data.Transaction
	s, fine := ValidateTransactionToken(txData, false)
	if !fine {
		globalPkg.SendError(w, s)
		globalPkg.WriteLog(logobj, s, "failed")
		return
	}
	res, id := createOwnershipToken(tokenData)
	if res != "" {
		globalPkg.SendError(w, res)
		globalPkg.WriteLog(logobj, res, "failed")
		return
	}

	res = createOwnershipTransaction(txData, tokenData.InitiatorAddress, id)
	if res != "" {
		globalPkg.SendError(w, res)
		globalPkg.WriteLog(logobj, res, "failed")
		return
	}
	responseObj := responses.FindResponseByID("112")
	globalPkg.SendResponseMessage(w, responseObj.EngResponse)
	globalPkg.WriteLog(logobj, responseObj.EngResponse, "success")
}

// TransferOwnershipAPI transfer ownership from billing account to another
func TransferOwnershipAPI(w http.ResponseWriter, req *http.Request) {
	//log
	now, userIP := globalPkg.SetLogObj(req)
	found, logobj := logpkg.CheckIfLogFound(userIP)
	if found && now.Sub(logobj.Currenttime).Seconds() > globalPkg.GlobalObj.DeleteAccountTimeInseacond {
		logobj.Count = 0
		broadcastTcp.BoardcastingTCP(logobj, "", "AddAndUpdateLog")
	}
	if found && logobj.Count >= 10 {
		responseObj := responses.FindResponseByID("6")
		globalPkg.SendError(w, responseObj.EngResponse)
		globalPkg.WriteLog(logobj, responseObj.EngResponse, "failed")
		return
	}
	if !found {
		Logindex := userIP.String() + "_" + logfunc.NewLogIndex()
		logobj = logpkg.LogStruct{Logindex, now, userIP, "macAdress", "TransferOwnership", "transactionModule", "", "", "_", 0}
	}
	logobj = logfunc.ReplaceLog(logobj, "TransferOwnership", "transactionModule")
	var data MixedOwnership
	decoder := json.NewDecoder(req.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&data)
	if err != nil {
		responseObj := responses.FindResponseByID("1")
		globalPkg.SendError(w, responseObj.EngResponse)
		globalPkg.WriteLog(logobj, responseObj.EngResponse, "failed")
		return
	}
	tkn := token.FindTokenByid(data.TokenID)
	if isAccountBilling(tkn.InitiatorAddress) {
		responseObj := responses.FindResponseByID("113")
		globalPkg.SendError(w, responseObj.EngResponse)
		globalPkg.WriteLog(logobj, responseObj.EngResponse, "failed")

		return
	}
	txData := data.Transaction
	ret := isAccountBilling(txData.Sender)
	ret2 := isAccountBilling(txData.Receiver)
	if !ret || !ret2 {
		responseObj := responses.FindResponseByID("114")
		globalPkg.SendError(w, responseObj.EngResponse)
		globalPkg.WriteLog(logobj, responseObj.EngResponse, "failed")

		return
	}
	errStr := makeTransferTX(txData, tkn.InitiatorAddress, data.TokenID)
	if errStr != "" {
		globalPkg.SendError(w, errStr)
		globalPkg.WriteLog(logobj, errStr, "failed")
		return
	}
	responseObj := responses.FindResponseByID("115")
	globalPkg.SendResponseMessage(w, responseObj.EngResponse)
	globalPkg.WriteLog(logobj, responseObj.EngResponse, "success")

}

// GetAllOwnershipForBillingAccountAPI get all ownerships for some billing  account
func GetAllOwnershipForBillingAccountAPI(w http.ResponseWriter, req *http.Request) {
	//log
	now, userIP := globalPkg.SetLogObj(req)
	logobj := logpkg.LogStruct{"", now, userIP, "macAdress", "GetAllOwnershipForBillingAccountAPI", "transactionModule", "", "", "", 0}
	var b publickey
	decoder := json.NewDecoder(req.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&b)
	if err != nil {
		responseObj := responses.FindResponseByID("1")
		globalPkg.SendError(w, responseObj.EngResponse)
		globalPkg.WriteLog(logobj, responseObj.EngResponse, "failed")
		return
	}
	accObj := account.GetAccountByAccountPubicKey(b.Publickey)
	// check if account is valid
	if !verifyAccount(&w, &logobj, &accObj) {
		return
	}
	// get all ownershipes
	// read txIds from them
	// search for transaction whose receiver = pk
	// return token , account
	if !isAccountBilling(b.Publickey) {
		responseObj := responses.FindResponseByID("116")
		globalPkg.SendError(w, responseObj.EngResponse)
		globalPkg.WriteLog(logobj, responseObj.EngResponse, "failed")

		return
	}
	ownerAccount := account.GetAccountByAccountPubicKey(b.Publickey)
	// fmt.Println("2@2 ", ownerAccount)
	OwnershipStructObj, isfine := accountdb.FindOwnershipByKey(ownerAccount.AccountIndex)
	if !isfine {
		responseObj := responses.FindResponseByID("117")
		globalPkg.SendError(w, responseObj.EngResponse)
		globalPkg.WriteLog(logobj, responseObj.EngResponse, "failed")
		return
	}
	if len(OwnershipStructObj.Ownership) < 1 {
		responseObj := responses.FindResponseByID("117")
		globalPkg.SendResponseMessage(w, responseObj.EngResponse)
		globalPkg.WriteLog(logobj, responseObj.EngResponse, "success")

		return
	}
	sendJSON, _ := json.Marshal(OwnershipStructObj.Ownership)
	globalPkg.SendResponse(w, sendJSON)
	globalPkg.WriteLog(logobj, string(sendJSON), "success")

}

//GetUserOwnerAPI get user Owner api
func GetUserOwnerAPI(w http.ResponseWriter, req *http.Request) {
	//log
	now, userIP := globalPkg.SetLogObj(req)
	logobj := logpkg.LogStruct{"", now, userIP, "macAdress", "GetUserOwnerAPI", "transactionModule", "", "", "", 0}
	var b publickey
	decoder := json.NewDecoder(req.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&b)
	if err != nil {
		responseObj := responses.FindResponseByID("1")
		globalPkg.SendError(w, responseObj.EngResponse)
		globalPkg.WriteLog(logobj, responseObj.EngResponse, "failed")
		return
	}
	acc := account.GetAccountByAccountPubicKey(b.Publickey)
	owner, fine := accountdb.FindOwnershipByKey(acc.AccountIndex)
	if !fine {
		responseObj := responses.FindResponseByID("118")
		globalPkg.SendError(w, responseObj.EngResponse)
		globalPkg.WriteLog(logobj, responseObj.EngResponse, "failed")

		return
	}
	var index int
	if len(owner.Ownership) == 0 {
		index = 0
	} else {
		index = len(owner.Ownership) - 1
	}
	txID := owner.Ownership[index] // transaction id
	fmt.Println("txID ", txID)
	tx := getTransactionByID(txID)

	var ownerAccount accountdb.AccountStruct
	ownerAccount = getAccountFromTransaction(tx)
	if ownerAccount.AccountName == "" {
		responseObj := responses.FindResponseByID("118")
		globalPkg.SendError(w, responseObj.EngResponse)
		globalPkg.WriteLog(logobj, responseObj.EngResponse, "failed")
		return
	}

	JSON, _ := json.Marshal(ownerAccount)
	globalPkg.WriteLog(logobj, "get the owner of user", "success")
	globalPkg.SendResponse(w, JSON)
}
