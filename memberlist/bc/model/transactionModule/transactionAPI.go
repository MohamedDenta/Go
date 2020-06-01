package transactionModule

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"../account"
	"../accountdb"
	"../admin"
	"../broadcastTcp"
	"../globalPkg"
	"../logpkg"
	"../token"
	"../transaction"
)

type AccountPasswordAndPubKey struct {
	Password  string
	PublicKey string
}
type SendData struct {
	ReceiverName string
	Amount       int
}

// notification omer
type fltr struct {
	Str string
	Lst string
}
type Notify struct {
	SessionID string
	Message   string
}

/*----------------endpoint to get all the transactions----------------- */
func GetAllTransactionsAPI(w http.ResponseWriter, req *http.Request) {
	//log
	now, userIP := globalPkg.SetLogObj(req)
	logobj := logpkg.LogStruct{"_", now, userIP, "macAdress", "AddNewTransaction", "transactionModule", "_", "_", "_", 0}

	Adminobj := admin.Admin{}

	decoder := json.NewDecoder(req.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&Adminobj)

	if err != nil {
		globalPkg.SendError(w, "please enter your correct request")
		globalPkg.WriteLog(logobj, "please enter your correct request", "failed")
		return
	}
	// logobj.InputData = Adminobj.AdminUsername + Adminobj.AdminPassword
	logobj.InputData = Adminobj.UsernameAdmin + Adminobj.PasswordAdmin
	if admin.ValidationAdmin(Adminobj) {
		sendJSON, _ := json.Marshal(transaction.GetPendingTransactions())
		globalPkg.SendResponse(w, sendJSON)
		globalPkg.WriteLog(logobj, "get all transaction success", "success")
	} else {
		globalPkg.SendError(w, "you are not admin")
		globalPkg.WriteLog(logobj, "you are not admin", "failed")
	}
}

//AddNewTransaction add new transaction api----------------- */
func AddNewTransaction(w http.ResponseWriter, req *http.Request) {
	//log
	now, userIP := globalPkg.SetLogObj(req)
	logobj := logpkg.LogStruct{"", now, userIP, "macAdress", "AddNewTransation", "transactionModule", "", "", "_", 0}
	// found, logobj := logpkg.CheckIfLogFound(userIP)

	// if found && now.Sub(logobj.Currenttime).Seconds() > globalPkg.GlobalObj.DeleteAccountTimeInseacond {

	// 	logobj.Count = 0
	// 	broadcastTcp.BoardcastingTCP(logobj, "", "AddAndUpdateLog")

	// }
	// if found && logobj.Count >= 10 {

	// 	globalPkg.SendError(w, "your Account have been blocked")
	// 	return
	// }

	// if !found {

	// 	Logindex := userIP.String() + "_" + logfunc.NewLogIndex()

	// 	logobj = logpkg.LogStruct{Logindex, now, userIP, "macAdress", "AddNewTransation", "transactionModule", "", "", "_", 0}
	// }
	// logobj = logfunc.ReplaceLog(logobj, "AddNewTransation", "transactionModule")

	// type StringData struct {
	// 	EncryptedData string
	// 	SessionID     string
	// }
	// var encryptedDigitalWalletTx StringData
	var errStr string
	var digitalWalletTransaction transaction.DigitalwalletTransaction
	var transactionObj transaction.Transaction
	//json.NewDecoder(req.Body).Decode(&digitalwalletTransactionObj)

	decoder := json.NewDecoder(req.Body)
	decoder.DisallowUnknownFields()
	// err := decoder.Decode(&encryptedDigitalWalletTx)
	err := decoder.Decode(&digitalWalletTransaction)

	if err != nil {
		globalPkg.SendError(w, "  please enter your correct request")
		globalPkg.WriteLog(logobj, "please enter your correct request", "failed")
		fmt.Printf("\n \n ******** add new transaction error is : %v ********** \n \n", err)
		return
	}
	fmt.Println("----------transaction time before :", time.Now())
	// decrypt digital wallet transaction.
	// digitalWalletTransaction, err = DecryptDigitalWalletTx(encryptedDigitalWalletTx.EncryptedData)
	// if err != nil {
	// 	globalPkg.SendError(w, err.Error())
	// 	return
	// }
	// fmt.Println("----------transaction time after 1 :", time.Now())

	// validate the digital wallet transaction.
	errStr, noError := ValidateTransactionToken(digitalWalletTransaction)

	fmt.Println("----------transaction time after 2 :", time.Now())

	// validate the digital wallet transaction.
	firstaccount := accountdb.GetFirstAccount()
	if noError {
		if digitalWalletTransaction.Sender == "" && firstaccount.AccountPublicKey == digitalWalletTransaction.Receiver {
			transactionObj = addcoins(digitalWalletTransaction)
		} else {
			transactionObj = DigitalwalletToUTXOTrans(digitalWalletTransaction)
		}
		mixedObj := transaction.MixedTxStruct{TxObj: transactionObj, DigitalTxObj: digitalWalletTransaction}

		//go func() {
		//	res1 := broadcastTcp.BoardcastingTCP(mixedObj, "addTransaction", "transaction")
		//	fmt.Println("transaction api res1", res1)
		//}()
		//fmt.Println("\n )())()()( transaction api transactionObj", transactionObj)
		fmt.Println("----------transaction time before broadcast :", time.Now())

		res, _ := broadcastTcp.BoardcastingTCP(mixedObj, "addTransaction", "transaction")

		fmt.Println("transaction api res", res)

		fmt.Println("----------transaction time after broadcast :", time.Now())
		if !res.Valid {
			noError = res.Valid
			errStr = "there was a double spend transaction"
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
		// if logobj.Count > 0 {
		// 	logobj.Count = 0
		// 	broadcastTcp.BoardcastingTCP(logobj, "", "AddAndUpdateLog")

		// }
		fmt.Println("----------transaction time after all :", time.Now())

		// globalPkg.SendResponseMessage(w, "Your transaction with "+tmp+" coin has been sent successfully to "+obj.AccountName)
		return
		//_, err := w.Write([]byte("Your transaction with " + tmp + " coin has been sent successfully to " + obj.AccountName))
		//fmt.Println("************************\n write err", err)
	} else {
		fmt.Println(errStr)
		globalPkg.SendError(w, errStr)
		globalPkg.WriteLog(logobj, errStr, "failed")
		// logobj.Count = logobj.Count + 1
		// fmt.Println("----------transaction time before log :", time.Now())

		// broadcastTcp.BoardcastingTCP(logobj, "", "AddAndUpdateLog")

		// fmt.Println("----------transaction time after log :", time.Now())

	}
}

// GetBalance endpoint to get the balance---------------- */
func GetBalance(w http.ResponseWriter, req *http.Request) {
	now, userIP := globalPkg.SetLogObj(req)
	logobj := logpkg.LogStruct{"_", now, userIP, "macAdress", "GetBAlance", "transactionModuleApi", "_", "_", "_", 0}

	var accountPasswordAndPubKey AccountPasswordAndPubKey

	//json.NewDecoder(req.Body).Decode(&accountPasswordAndPubKey)
	decoder := json.NewDecoder(req.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&accountPasswordAndPubKey)

	if err != nil {
		globalPkg.SendError(w, "  please enter your correct request")
		globalPkg.WriteLog(logobj, "please enter your correct request", "failed")
		return
	}
	//if err := json.NewDecoder(req.Body).Decode(&accountPasswordAndPubKey); err != nil {
	//	errorpk.AddError("GetTransactionByPublicKey API Transaction module package ", "can't decode json to accountPasswordAndPubKey struct")
	//	w.WriteHeader(http.StatusServiceUnavailable)
	//	w.Write([]byte("Service is Unavailable"))
	//}

	accObj := account.GetAccountByAccountPubicKey(accountPasswordAndPubKey.PublicKey)
	if accObj.AccountPublicKey == "" || accObj.AccountPassword == "" || accObj.AccountName == "" {
		globalPkg.SendNotFound(w, "The Account with this Public key is not found")
		globalPkg.WriteLog(logobj, "The Account with this Public key is not found", "failed")

		return
		//w.Write([]byte("The Account with this Public key is not found"))
	} else if !checkIfAccountIsActive(accountPasswordAndPubKey.PublicKey) {
		globalPkg.SendError(w, "The Account with this Public key is not active")
		globalPkg.WriteLog(logobj, "The Account with this Public key is not active", "failed")
		return
		//w.Write([]byte("The Account with this Public key is not active"))
	} else if accountPasswordAndPubKey.Password != accObj.AccountPassword {
		globalPkg.SendError(w, "The given password is incorrect")
		globalPkg.WriteLog(logobj, "The given password is incorrect", "failed")
		return
		//w.Write([]byte("The given password is incorrect"))
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
	//w.WriteHeader(http.StatusOK)
	//_ = json.NewEncoder(w).Encode(GetAccountBalanceStatement(accObj))
}

//GetTransactionByPublicKey used to get all transactions linked to the account by using the provided account PubKey
func GetTransactionByPublicKey(w http.ResponseWriter, req *http.Request) {
	now, userIP := globalPkg.SetLogObj(req)
	logobj := logpkg.LogStruct{"_", now, userIP, "macAdress", "GetTransactionByPublicKey", "transactionModule", "_", "_", "_", 0}

	var accountPasswordAndPubKey AccountPasswordAndPubKey
	logobj.InputData = accountPasswordAndPubKey.PublicKey + "," + accountPasswordAndPubKey.Password
	//json.NewDecoder(req.Body).Decode(&accountPasswordAndPubKey)
	decoder := json.NewDecoder(req.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&accountPasswordAndPubKey)

	if err != nil {
		globalPkg.SendError(w, "please enter your correct request")
		globalPkg.WriteLog(logobj, "please enter your correct request", "failed")

		return
	}
	//if err := json.NewDecoder(req.Body).Decode(&accountPasswordAndPubKey); err != nil {
	//	errorpk.AddError("GetTransactionByPublicKey API Transaction module package ", "can't decode json to accountPasswordAndPubKey struct")
	//	w.WriteHeader(http.StatusServiceUnavailable)
	//	w.Write([]byte("Service is Unavailable"))
	//}
	accObj := account.GetAccountByAccountPubicKey(accountPasswordAndPubKey.PublicKey)
	if accObj.AccountPublicKey == "" || accObj.AccountPassword == "" || accObj.AccountName == "" {
		globalPkg.SendNotFound(w, "The Account with this Public key is not found")
		globalPkg.WriteLog(logobj, "The Account with this Public key is not found", "failed")
		return
		//w.Write([]byte("The Account with this Public key is not found"))
	} else if !checkIfAccountIsActive(accountPasswordAndPubKey.PublicKey) {
		globalPkg.SendError(w, "The Account with this Public key is not active")
		globalPkg.WriteLog(logobj, "The Account with this Public key is not active", "failed")
		return
		//w.Write([]byte("The Account with this Public key is not active"))
	} else if accountPasswordAndPubKey.Password != accObj.AccountPassword {
		globalPkg.SendError(w, "The given password is incorrect")
		globalPkg.WriteLog(logobj, "The given password is incorrect", "failed")
		return
		//w.Write([]byte("The given password is incorrect"))
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
		globalPkg.SendError(w, "  please enter your correct request")
		globalPkg.WriteLog(logobj, "please enter your correct request", "failed")
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
		globalPkg.SendError(w, "you are not admin")
		globalPkg.WriteLog(logobj, "you are not admin", "failed")
	}
}

func GetAllTransactionDbAPI(w http.ResponseWriter, req *http.Request) {
	//log
	now, userIP := globalPkg.SetLogObj(req)
	logobj := logpkg.LogStruct{"_", now, userIP, "macAdress", "GetAllTransactionDbAPI", "transactionModule", "_", "_", "_", 0}

	Adminobj := admin.Admin{}

	decoder := json.NewDecoder(req.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&Adminobj)

	if err != nil {
		globalPkg.SendError(w, "  please enter your correct request")
		globalPkg.WriteLog(logobj, "please enter your correct request", "failed")
		return
	}

	if admin.ValidationAdmin(Adminobj) {
		sendJSON, _ := json.Marshal(transaction.GetAllTransaction())
		globalPkg.SendResponse(w, sendJSON)
		globalPkg.WriteLog(logobj, string(sendJSON), "success")
	} else {
		globalPkg.SendError(w, "you are not admin")
		globalPkg.WriteLog(logobj, "you are not admin", "failed")
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
		globalPkg.SendError(w, "please enter your correct request")
		globalPkg.WriteLog(logobj, "please enter your correct request", "failed")
		return
	}
	//validate initiator address if exist in account data
	accountobj := account.GetAccountByAccountPubicKey(tokenObj.InitiatorAddress)
	if accountobj.AccountPublicKey == "" {
		globalPkg.SendError(w, "please enter valid initiator address")
		globalPkg.WriteLog(logobj, "please enter valid initiator address", "failed")
		return
	}
	if accountobj.AccountPassword != tokenObj.Password {
		globalPkg.SendError(w, "The given password is incorrect.")
		globalPkg.WriteLog(logobj, "The given password is incorrect.", "failed")
		return
	}

	tokenobj := token.FindTokenByTokenName(tokenObj.TokenName)
	if tokenobj.TokenName != "" {
		if tokenobj.InitiatorAddress != tokenObj.InitiatorAddress {
			globalPkg.SendError(w, "you don't the creator of this token")
			globalPkg.WriteLog(logobj, "you don't the creator of this token", "failed")
			return
		}
	} else {
		globalPkg.SendError(w, "this token name not exists")
		globalPkg.WriteLog(logobj, "this token name not exists", "failed")
		return
	}

	if account.ContainstokenID(accountobj.AccountTokenID, tokenobj.TokenID) {
		x := GetTransactionsByTokenID(accountobj, tokenobj.TokenID)
		sendJSON, _ := json.Marshal(x)
		globalPkg.SendResponse(w, sendJSON)
		globalPkg.WriteLog(logobj, string(sendJSON), "success")
		return
	}
}
