package serviceModule

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	// "../logfunc"

	"../logpkg"

	"net/http"
	//"time"

	"../transactionModule"

	"../globalPkg"

	// "model/validator"

	"../transaction"

	"../broadcastTcp"
	//	transaction "../transaction"

	"../account"
	"../service"
	"../accountdb"
)

type InquiryResponse struct {
	ID     string
	Amount string
	Msg    string
}
type PurchaseResponse struct {
	ID  string
	Msg string
}

type PurchaseServiceStruct struct {
	ID string

	Password       string
	Transactionobj transaction.DigitalwalletTransaction
}

/////End point
//***********************************add and validate the service **************************//
func InquiryNewInternetServiceCost(w http.ResponseWriter, req *http.Request) {

	now, userIP := globalPkg.SetLogObj(req)
	logobj := logpkg.LogStruct{"", now, userIP, "macAdress", "InquiryNewInternetServiceCost", "serviceModule", "", "", "_", 0}
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

	// 	logobj = logpkg.LogStruct{Logindex, now, userIP, "macAdress", "InquiryNewInternetServiceCost", "serviceModule", "", "", "_", 0}
	// }
	// logobj = logfunc.ReplaceLog(logobj, "InquiryNewInternetServiceCost", "serviceModule")

	serviveobj := service.ServiceStruct{}
	decoder := json.NewDecoder(req.Body)
	//	decoder.DisallowUnknownFields()
	err := decoder.Decode(&serviveobj)

	if err != nil {
		globalPkg.SendError(w, "please enter your correct request ")
		globalPkg.WriteLog(logobj, "please enter your correct request", "failed")
		return
	}
	//
	if serviveobj.Duration < 1 {
		globalPkg.SendError(w, "please,check Duration shouldn't be less than 1")

		return
	}
	if serviveobj.Bandwidth < 1 {
		globalPkg.SendError(w, "please,check bandwidth shouldn't be less than 1")

		return
	}
	if serviveobj.Amount < 1 {
		globalPkg.SendError(w, "please,check Amount shouldn't be less than 1")

		return
	}
	if serviveobj.PublicKey == "" || serviveobj.Password == "" {
		globalPkg.SendError(w, "Empty Attribute publickey or password")
		globalPkg.WriteLog(logobj, "Empty Attribute public key or password", "failed")
		return
	}
	serviveobj.Time = now
	inoTokenID, _ := globalPkg.ConvertIntToFixedLengthString(0, globalPkg.GlobalObj.TokenIDStringFixedLength)
	accountobj := account.GetAccountByAccountPubicKey(serviveobj.PublicKey)
	if accountobj.AccountPassword != serviveobj.Password {
		globalPkg.SendError(w, "Invalid password")
		globalPkg.WriteLog(logobj, "Invalid password", "failed")
		// logobj.OutputData = "Invalid password"
		// logobj.Count = logobj.Count + 1

		// broadcastTcp.BoardcastingTCP(logobj, "", "AddAndUpdateLog")

		return
	} else if accountobj.AccountPublicKey != serviveobj.PublicKey {
		globalPkg.SendError(w, "Invalid publickey")
		globalPkg.WriteLog(logobj, "Invalid publickey", "failed")
		// logobj.OutputData = "Invalid publickey"
		// logobj.Count = logobj.Count + 1

		// broadcastTcp.BoardcastingTCP(logobj, "", "AddAndUpdateLog")
		return
	}

	txs := service.ServiceStructGetlastPrefix(accountobj.AccountIndex) //transactionModule.GetTransactionsByTokenID(accountobj, inoTokenID)
	length := ""
	index := 0
	if txs.ID == "" {
		length, _ = globalPkg.ConvertIntToFixedLengthString(0, 13)
	} else {
		res := strings.Split(txs.ID, "_")
		/*if len(res) == 2 {
			index = globalPkg.ConvertFixedLengthStringtoInt(res[1]) + 1
		} else if len(res) > 2 {
			index = globalPkg.ConvertFixedLengthStringtoInt(res[2]) + 1
		}*/
		index = globalPkg.ConvertFixedLengthStringtoInt(res[len(res)-1]) + 1
		length, _ = globalPkg.ConvertIntToFixedLengthString(index, 13)
	}

	// length ,_ = globalPkg.ConvertIntToFixedLengthString(0,13)
	fmt.Printf("eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee%v", length)

	serviveobj.ID = accountobj.AccountIndex + "_" + length
	if serviveobj.Bandwidth >= 1024 {
		globalPkg.SendError(w, "Invalid BandWidth")
		globalPkg.WriteLog(logobj, "invalid bandwidth", "failed")
		// logobj.OutputData = "Invalid bandwidth"
		// logobj.Count = logobj.Count + 1

		// broadcastTcp.BoardcastingTCP(logobj, "", "AddAndUpdateLog")

		return
	}
	if serviveobj.Day == true && serviveobj.Duration > 31 {
		globalPkg.SendError(w, "Invalid Duration")
		globalPkg.WriteLog(logobj, "invalid duration", "failed")
		// logobj.OutputData = "Invalid duration"
		// logobj.Count = logobj.Count + 1

		// broadcastTcp.BoardcastingTCP(logobj, "", "AddAndUpdateLog")
		return
	}
	if serviveobj.Day == false && serviveobj.Duration > 1440 {
		globalPkg.SendError(w, "Invalid Duration")
		globalPkg.WriteLog(logobj, "invalid duration", "failed")
		// logobj.OutputData = "Invalid duration"
		// logobj.Process = "failed"
		// logobj.Count = logobj.Count + 1

		// broadcastTcp.BoardcastingTCP(logobj, "", "AddAndUpdateLog")
		return
	}
	serviveobj = service.CalculateAmountAndCost(serviveobj)
	//Balance := transactionModule.GetAccountBalance(accountobj.AccountPublicKey)
	inoTokenBalance := transactionModule.GetAccountBalanceStatement(accountobj, inoTokenID)

	var Balance float64
	_, tokenExist := inoTokenBalance[inoTokenID]
	if !tokenExist {
		globalPkg.SendError(w, "you do not have balance for inoToken.")
		globalPkg.WriteLog(logobj, "you don not have balance for inoToken.", "failed")
		// logobj.Process = "failed"
		// logobj.OutputData = "you don not have balance for inoToken."
		// logobj.Count = logobj.Count + 1

		// broadcastTcp.BoardcastingTCP(logobj, "", "AddAndUpdateLog")
		return
	}
	Balance = inoTokenBalance[inoTokenID].TotalBalance

	fmt.Printf("kkkkkkkkkkkkkkk", Balance)
	cost := serviveobj.Calculation + globalPkg.GlobalObj.TransactionFee
	if cost > Balance {
		globalPkg.SendResponseMessage(w, "your balance can't satisfy your request")
		globalPkg.WriteLog(logobj, "your balance can't satisfy your request", "failed")
		// logobj.OutputData = "your balance can't satisfy your request"
		// logobj.Count = logobj.Count + 1

		// broadcastTcp.BoardcastingTCP(logobj, "", "AddAndUpdateLog")
		return

	}
	//Calculation := fmt.Sprintf("%f", serviveobj.Calculation)
	stringCost := fmt.Sprintf("%f", cost)
	msg := "The user has inquired for a service of : " + stringCost + " with ID :" + serviveobj.ID
	response := InquiryResponse{serviveobj.ID, stringCost, msg}
	jsonObj, _ := json.Marshal(response)
	globalPkg.SendResponse(w, jsonObj)
	globalPkg.WriteLog(logobj, string(jsonObj), "success")
	// if logobj.Count > 0 {
	// 	logobj.Process = "success"
	// 	logobj.Count = 0
	// 	broadcastTcp.BoardcastingTCP(logobj, "", "AddAndUpdateLog")

	// }
	// globalPkg.SendResponseMessage(w, msg)
	broadcastTcp.BoardcastingTCP(serviveobj, "Tmp", "Add Service")
	fmt.Println("---All", service.GetAllservice())
	return
}

//___________________________________________
//
//-----------------------------------------------------
func PurchaseService(w http.ResponseWriter, req *http.Request) {
	now, userIP := globalPkg.SetLogObj(req)
	logobj := logpkg.LogStruct{"", now, userIP, "macAdress", "PurchaseService", "ServiceModule", "", "", "_", 0}
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

	// 	logobj = logpkg.LogStruct{Logindex, now, userIP, "macAdress", "PurchaseService", "ServiceModule", "", "", "_", 0}
	// }
	// logobj = logfunc.ReplaceLog(logobj, "PurchaseService", "ServiceModule")

	var PurchaseServiceObj PurchaseServiceStruct
	var VoucherRespobj service.VoucherResponse
	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&PurchaseServiceObj)

	if err != nil {
		globalPkg.SendError(w, "please enter your correct request ")
		globalPkg.WriteLog(logobj, "please enter your correct request", "failed")
		return
	}
	if PurchaseServiceObj.Transactionobj.Amount < 1 {
		globalPkg.SendError(w, "invalid Amount")
		return
	}
	accountobj := account.GetAccountByAccountPubicKey(PurchaseServiceObj.Transactionobj.Sender)
	if accountobj.AccountPassword != PurchaseServiceObj.Password {
		fmt.Println("account", PurchaseServiceObj.Transactionobj.Sender)
		fmt.Println("PurchaseServiceObj", PurchaseServiceObj.Password)
		globalPkg.SendError(w, "Invalid public key or password")
		globalPkg.WriteLog(logobj, "Invalid public key or password", "failed")
		// logobj.OutputData = "Invalid password"
		// logobj.Process = "failed"
		// logobj.Count = logobj.Count + 1

		// broadcastTcp.BoardcastingTCP(logobj, "", "AddAndUpdateLog")

		return

	}
	serviceobj, found := service.CheckRequestId(PurchaseServiceObj.ID, PurchaseServiceObj.Transactionobj.Sender)
	if !found {
		globalPkg.SendError(w, "Invalid Request")
		globalPkg.WriteLog(logobj, "Invalid Request", "failed")
		// logobj.OutputData = "Invalid Request"
		// logobj.Process = "failed"
		// logobj.Count = logobj.Count + 1

		// broadcastTcp.BoardcastingTCP(logobj, "", "AddAndUpdateLog")
		return
	}
	inoTokenID, _ := globalPkg.ConvertIntToFixedLengthString(0, globalPkg.GlobalObj.TokenIDStringFixedLength)
	inoTokenBalance := transactionModule.GetAccountBalanceStatement(accountobj, inoTokenID)
	var Balance float64
	_, tokenExist := inoTokenBalance[inoTokenID]
	if !tokenExist {
		globalPkg.SendError(w, "you do not have balance for inoToken.")
		globalPkg.WriteLog(logobj, "you do not have balance for inoToken.", "failed")
		// logobj.OutputData = "you do not have balance for inoToken."
		// logobj.Process = "failed"
		// logobj.Count = logobj.Count + 1

		// broadcastTcp.BoardcastingTCP(logobj, "", "AddAndUpdateLog")

		return
	}
	Balance = inoTokenBalance[inoTokenID].TotalBalance
	cost := serviceobj.Calculation + globalPkg.GlobalObj.TransactionFee
	if cost > Balance {

		jsonObj, _ := json.Marshal(VoucherRespobj)
		globalPkg.SendResponse(w, jsonObj)
		globalPkg.WriteLog(logobj, string(jsonObj), "failed")
		// logobj.OutputData = string(jsonObj)
		// logobj.Process = "failed"
		// logobj.Count = logobj.Count + 1

		// broadcastTcp.BoardcastingTCP(logobj, "", "AddAndUpdateLog")

		return
	}
	/*VoucherRequestobj := VoucherRequest{"create-voucher", "1",serviceobj.Bandwidth, "2048", "4096",serviceobj.Mbytes , true, "4 Mbps/2 Mbps 1 GB 6 Hr Package Voucher","1"
	jsonObj, _ := json.Marshal(VoucherRequestobj)
	VoucherStructobj.Voucher=globalPkg.SendRequestAndGetResponse(jsonObj,"https://192.168.1.27:8443/api/s/default/cmd/hotspot","POST",VoucherRequest)
	*/
	var MBytes string
	if serviceobj.Mbytes == true {
		MBytes = "true"
	} else {
		MBytes = "false"
	}

	st2 := globalPkg.Voucher{"create-voucher", "1", strconv.Itoa(serviceobj.Amount), "2048", "4096", strconv.Itoa(serviceobj.Bandwidth), MBytes, serviceobj.ID, strconv.Itoa(serviceobj.M)}
	Createresponse, createstatus := globalPkg.RequestServiceAPI(st2, PurchaseServiceObj.ID)

	if createstatus {
		serviceobj.VoutcherId = Createresponse.Code
		serviceobj.CreateTime = Createresponse.Create_time
		/////////////////Add ew Transacrtion
		//var transactionobj transaction.DigitalwalletTransaction
		PurchaseServiceObj.Transactionobj.ServiceId = serviceobj.ID
		// transjson, _ := json.Marshal(PurchaseServiceObj.Transactionobj)
		response := transactionModule.ValidateServiceTransaction(PurchaseServiceObj.Transactionobj) //globalPkg.SendRequestAndGetResponse(transjson, validator.CurrentValidator.ValidatorIP+"/2e4a9d667ad5e3cef02eae9", "POST", &transactionobj)

		fmt.Println("transaction Response", response)
		if response != "" {
			globalPkg.SendError(w, response)
			globalPkg.WriteLog(logobj, response, "failed")
			// logobj.OutputData = response
			// logobj.Count = logobj.Count + 1
			// logobj.Process = "failed"
			// broadcastTcp.BoardcastingTCP(logobj, "", "AddAndUpdateLog")

			return
		} else {
			PurchaseServiceObj.Transactionobj.Amount -= globalPkg.GlobalObj.TransactionFee
			transactionObj := transactionModule.DigitalwalletToUTXOTrans(PurchaseServiceObj.Transactionobj)
			//var transactionObjlst []transaction.Transaction
			//transactionObjlst = append(transactionObjlst, transactionObj)
			//fmt.Println("transaction obj lst????", transactionObjlst)
			// var lst []string
			// lst = append(lst, transactionObj.TransactionTime.Format("2006-01-02 03:04:05 PM -0000"))

			broadcastTcp.BoardcastingTCP(transactionObj, "addTokenTransaction", "transaction")
			respons := PurchaseResponse{serviceobj.VoutcherId, "service transaction added successfuly"}
			jsonObj, _ := json.Marshal(respons)
			globalPkg.SendResponse(w, jsonObj)
			globalPkg.WriteLog(logobj, string(jsonObj), "success")
			// if logobj.Count > 0 {
			// 	logobj.Count = 0
			// 	logobj.OutputData = string(jsonObj)

			// 	logobj.Process = "success"
			// 	broadcastTcp.BoardcastingTCP(logobj, "", "AddAndUpdateLog")

			// }
			broadcastTcp.BoardcastingTCP(serviceobj, "DB", "Add Service")
			return
		}
	} else {
		globalPkg.SendError(w, "website is under maintenance")
		globalPkg.WriteLog(logobj, "website is under maintenance", "failed")

	}
}

//--------------------------------------------
// Get all purchased services End point
//-----------------------------------------------------------------------
func GetAllPurchasedServices(w http.ResponseWriter, req *http.Request) {
	now, userIP := globalPkg.SetLogObj(req)
	logobj := logpkg.LogStruct{"_", now, userIP, "macAdress", "GetAllPurchasedServices", "serviceModule", "_", "_", "_", 0}
	UserKeysobj := service.UserKeys{}
	decoder := json.NewDecoder(req.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&UserKeysobj)

	if err != nil {
		globalPkg.SendError(w, "please enter your correct request")
		globalPkg.WriteLog(logobj, "please enter your correct request", "failed")
		return
	}

	accountobj := account.GetAccountByAccountPubicKey(UserKeysobj.PublicKey)
	if accountobj.AccountPassword != UserKeysobj.Password {

		globalPkg.SendError(w, "Invalid public key or password")
		globalPkg.WriteLog(logobj, "Invalid public key or password", "failed")
		return
	}
	//-----call Api Created by Alaa to get All servic using Acccount index
	services := service.ServiceStructGetByPrefix(accountobj.AccountIndex)
	if len(services) == 0 {

		globalPkg.SendResponseMessage(w, "you don't have services")
		globalPkg.WriteLog(logobj, "you don't have services", "warning")
		return
	}
	jsonObj, _ := json.Marshal(services)
	globalPkg.SendResponse(w, jsonObj)
	globalPkg.WriteLog(logobj, string(jsonObj), "success")
	return

}

/*----------------------------------------------------------------------------------------*/
/*   */
//\\//\\//\\//\\//\\//\\//\\//\\///\\/\/\/\/\/\\/\/\///\/\/\/\/\/\/\\\\/\/\\/\/\/\/\/\/\/\/\/\/\/\/\/\/\/\
//-----------------------------------------------------------------------------------------------------------
func CheckVoucherStatus(w http.ResponseWriter, req *http.Request) {

	now, userIP := globalPkg.SetLogObj(req)
	logobj := logpkg.LogStruct{"_", now, userIP, "macAdress", "CheckVoucherStatus", "serviceModule", "_", "_", "_", 0}
	Vocherstatusobj := service.Vocherstatus{}
	decoder := json.NewDecoder(req.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&Vocherstatusobj)

	if err != nil {
		globalPkg.SendError(w, "please enter your correct request")
		globalPkg.WriteLog(logobj, "Failed to decode object", "Failed")
		return
	}
	//	var abj  ResponseCreateVoucher//service.VocherstatusResponse

	Create_time := service.GetCreateTime(Vocherstatusobj.VoucherID) //"2019-01-02 03:04:05 PM -0000"
	if Create_time == -1 {
		globalPkg.SendError(w, "Invalid VoucherID")
		globalPkg.WriteLog(logobj, "Invalid VoucherID", "Failed")
		return
	}
	accountobj := account.GetAccountByAccountPubicKey(Vocherstatusobj.PublicKey)
	if accountobj.AccountPassword != Vocherstatusobj.Password {
		globalPkg.SendError(w, "Invalid public key or password")
		globalPkg.WriteLog(logobj, "Invalid public key or password", "Failed")
		return
	}
	fmt.Println("*------Create_time", Create_time)
	ResponseVoucher, VoucherIDStatus, ok := globalPkg.GetVoucherDataAPI(Create_time)

	if VoucherIDStatus && ok {
		if ResponseVoucher.Used == 0 {
			globalPkg.SendResponseMessage(w, "this Voucher is not used yet")
			globalPkg.WriteLog(logobj, "this Voucher is not used yet", "warning")
		} else {
			jsonObj, _ := json.Marshal(ResponseVoucher)
			globalPkg.SendResponse(w, jsonObj)
			globalPkg.WriteLog(logobj, string(jsonObj), "success")
		}
		return
	} else if VoucherIDStatus && !ok {
		globalPkg.SendError(w, "website is under maintenance")
		globalPkg.WriteLog(logobj, "website is under maintenance", "failed")
		return
	} else if !VoucherIDStatus && ok {
		globalPkg.SendError(w, "wrong  voucher")
		globalPkg.WriteLog(logobj, "wrong voucher", "failed")
		return
	}

	return

}

//---------------------------------------------------------------------------------------
//  Api to get All Accounts names and pk for users whose role is service
//------------------------------------------------------------------------------------------
func GetAllNamesandPKsForServiceAccount(w http.ResponseWriter, req *http.Request) {
	now, userIP := globalPkg.SetLogObj(req)
	logobj := logpkg.LogStruct{"_", now, userIP, "macAdress", "GetAllNamesandPKsForServiceAccount", "serviceModule", "_", "_", "_", 0}

	UserKeysobj := service.UserKeys{}
	decoder := json.NewDecoder(req.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&UserKeysobj)
	// TODO : create log object

	if err != nil {
		globalPkg.SendError(w, "please enter your correct request")
		globalPkg.WriteLog(logobj, "Failed to decode object", "Failed")
		return
	}

	accountobj := account.GetAccountByAccountPubicKey(UserKeysobj.PublicKey)
	if accountobj.AccountPassword != UserKeysobj.Password {

		globalPkg.SendError(w, "Invalid public key or password")
		globalPkg.WriteLog(logobj, "Invalid public key or password", "Failed")
		return
	}
	Accounts := accountdb.GetNamesandPKsForServiceAccount()
	jsonObj, _ := json.Marshal(Accounts)
	globalPkg.SendResponse(w, jsonObj)

	return

}
