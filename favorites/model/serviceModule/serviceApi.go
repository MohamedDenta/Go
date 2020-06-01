package serviceModule

import (
	"../account"
	"../accountdb"
	"../broadcastTcp"
	"../globalPkg"
	"../logfunc"
	"../logpkg"
	"../responses"
	"../service"
	"../transactionModule"
	"../validator"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

//InquiryNewInternetServiceCost add and validate the service
func InquiryNewInternetServiceCost(w http.ResponseWriter, req *http.Request) {

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

		logobj = logpkg.LogStruct{Logindex, now, userIP, "macAdress", "InquiryNewInternetServiceCost", "serviceModule", "", "", "_", 0}
	}
	logobj = logfunc.ReplaceLog(logobj, "InquiryNewInternetServiceCost", "serviceModule")

	serviveobj := service.ServiceStruct{}
	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&serviveobj)

	if err != nil {
		responseObj := responses.FindResponseByID("1")
		globalPkg.SendError(w, responseObj.EngResponse)
		globalPkg.WriteLog(logobj, responseObj.EngResponse, "failed")
		return
	}
	//
	if serviveobj.Duration < 1 {
		responseObj := responses.FindResponseByID("52")
		globalPkg.SendError(w, responseObj.EngResponse)
		globalPkg.WriteLog(logobj, responseObj.EngResponse, "failed")
		return
	}
	if serviveobj.Bandwidth < 1 {
		responseObj := responses.FindResponseByID("53")
		globalPkg.SendError(w, responseObj.EngResponse)
		globalPkg.WriteLog(logobj, responseObj.EngResponse, "failed")

		return
	}
	if serviveobj.Amount < 1 {
		responseObj := responses.FindResponseByID("54")
		globalPkg.SendError(w, responseObj.EngResponse)
		globalPkg.WriteLog(logobj, responseObj.EngResponse, "failed")
		return
	}
	if serviveobj.PublicKey == "" || serviveobj.Password == "" {
		responseObj := responses.FindResponseByID("10")
		globalPkg.SendError(w, responseObj.EngResponse)
		globalPkg.WriteLog(logobj, responseObj.EngResponse, "failed")
		return
	}
	serviveobj.Time = now
	inoTokenID, _ := globalPkg.ConvertIntToFixedLengthString(0, globalPkg.GlobalObj.TokenIDStringFixedLength)
	accountobj := account.GetAccountByAccountPubicKey(serviveobj.PublicKey)
	if accountobj.AccountPassword != serviveobj.Password {
		responseObj := responses.FindResponseByID("11")
		logobj.OutputData = "Invalid password"
		logobj.Count = logobj.Count + 1
		broadcastTcp.BoardcastingTCP(logobj, "", "AddAndUpdateLog")
		globalPkg.SendError(w, responseObj.EngResponse)
		globalPkg.WriteLog(logobj, responseObj.EngResponse, "failed")

		return
	} else if accountobj.AccountPublicKey != serviveobj.PublicKey {
		responseObj := responses.FindResponseByID("10")
		logobj.OutputData = "Invalid publickey"
		logobj.Count = logobj.Count + 1
		broadcastTcp.BoardcastingTCP(logobj, "", "AddAndUpdateLog")
		globalPkg.SendError(w, responseObj.EngResponse)
		globalPkg.WriteLog(logobj, responseObj.EngResponse, "failed")
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
	// fmt.Printf("eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee%v", length)

	serviveobj.ID = accountobj.AccountIndex + "_" + length
	if serviveobj.Bandwidth >= 1024 {

		responseObj := responses.FindResponseByID("53")
		logobj.Count = logobj.Count + 1
		logobj.OutputData = "Invalid bandwidth"
		broadcastTcp.BoardcastingTCP(logobj, "", "AddAndUpdateLog")
		globalPkg.SendError(w, responseObj.EngResponse)
		globalPkg.WriteLog(logobj, responseObj.EngResponse, "failed")
		return
	}
	if serviveobj.Day == true && serviveobj.Duration > 31 {
		responseObj := responses.FindResponseByID("52")
		logobj.Count = logobj.Count + 1
		logobj.OutputData = "Invalid duration"
		broadcastTcp.BoardcastingTCP(logobj, "", "AddAndUpdateLog")
		globalPkg.SendError(w, responseObj.EngResponse)
		globalPkg.WriteLog(logobj, responseObj.EngResponse, "failed")
		return
	}
	if serviveobj.Day == false && serviveobj.Duration > 1440 {
		responseObj := responses.FindResponseByID("52")
		logobj.Count = logobj.Count + 1
		logobj.OutputData = "Invalid duration"
		logobj.Process = "failed"
		broadcastTcp.BoardcastingTCP(logobj, "", "AddAndUpdateLog")
		globalPkg.SendError(w, responseObj.EngResponse)
		globalPkg.WriteLog(logobj, responseObj.EngResponse, "failed")
		return
	}

	(&serviveobj).CalculateAmountAndCost()
	//Balance := transactionModule.GetAccountBalance(accountobj.AccountPublicKey)
	inoTokenBalance := transactionModule.GetAccountBalanceStatement(accountobj, inoTokenID)

	var Balance float64
	_, tokenExist := inoTokenBalance[inoTokenID]
	if !tokenExist {
		responseObj := responses.FindResponseByID("59")
		logobj.Count = logobj.Count + 1
		logobj.Process = "failed"
		logobj.OutputData = "you don not have balance for inoToken."
		broadcastTcp.BoardcastingTCP(logobj, "", "AddAndUpdateLog")
		globalPkg.SendError(w, responseObj.EngResponse)
		globalPkg.WriteLog(logobj, responseObj.EngResponse, "failed")
		return
	}
	Balance = inoTokenBalance[inoTokenID].TotalBalance

	// fmt.Printf("kkkkkkkkkkkkkkk", Balance)
	cost := serviveobj.Calculation + globalPkg.GlobalObj.TransactionFee
	if cost > Balance {
		responseObj := responses.FindResponseByID("59")
		logobj.Count = logobj.Count + 1
		logobj.OutputData = "your balance can't satisfy your request"
		broadcastTcp.BoardcastingTCP(logobj, "", "AddAndUpdateLog")
		globalPkg.SendResponseMessage(w, responseObj.EngResponse)
		globalPkg.WriteLog(logobj, responseObj.EngResponse, "failed")
		return
	}
	//Calculation := fmt.Sprintf("%f", serviveobj.Calculation)
	stringCost := fmt.Sprintf("%f", cost)
	msg := "The user has inquired for a service of : " + stringCost + " with ID :" + serviveobj.ID
	response := InquiryResponse{serviveobj.ID, stringCost, msg}
	jsonObj, _ := json.Marshal(response)
	globalPkg.SendResponse(w, jsonObj)
	globalPkg.WriteLog(logobj, string(jsonObj), "success")
	if logobj.Count > 0 {
		logobj.Process = "success"
		logobj.Count = 0
		broadcastTcp.BoardcastingTCP(logobj, "", "AddAndUpdateLog")

	}
	// globalPkg.SendResponseMessage(w, msg)
	broadcastTcp.BoardcastingTCP(serviveobj, "Tmp", "Add Service")
	// fmt.Println("---All", service.GetAllservice())
	return
}

//PurchaseService -----------------------------------------------------
func PurchaseService(w http.ResponseWriter, req *http.Request) {
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

		logobj = logpkg.LogStruct{Logindex, now, userIP, "macAdress", "PurchaseService", "ServiceModule", "", "", "_", 0}
	}
	logobj = logfunc.ReplaceLog(logobj, "PurchaseService", "ServiceModule")

	var PurchaseServiceObj PurchaseServiceStruct
	var VoucherRespobj service.VoucherResponse
	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&PurchaseServiceObj)

	if err != nil {
		responseObj := responses.FindResponseByID("1")
		globalPkg.SendError(w, responseObj.EngResponse)
		globalPkg.WriteLog(logobj, responseObj.EngResponse, "failed")
		return
	}
	if PurchaseServiceObj.Transactionobj.Amount < 1 {
		responseObj := responses.FindResponseByID("60")
		globalPkg.SendError(w, responseObj.EngResponse)
		globalPkg.WriteLog(logobj, responseObj.EngResponse, "failed")

		return
	}
	accountobj := account.GetAccountByAccountPubicKey(PurchaseServiceObj.Transactionobj.Sender)
	if accountobj.AccountPassword != PurchaseServiceObj.Password {
		fmt.Println("account", PurchaseServiceObj.Transactionobj.Sender)
		fmt.Println("PurchaseServiceObj", PurchaseServiceObj.Password)
		responseObj := responses.FindResponseByID("10")
		logobj.Count = logobj.Count + 1
		logobj.OutputData = "Invalid password"
		logobj.Process = "failed"
		broadcastTcp.BoardcastingTCP(logobj, "", "AddAndUpdateLog")
		globalPkg.SendError(w, responseObj.EngResponse)
		globalPkg.WriteLog(logobj, responseObj.EngResponse, "failed")
		return
	}

	serviceobj, found := service.CheckRequestID(PurchaseServiceObj.ID, PurchaseServiceObj.Transactionobj.Sender)
	if !found {
		responseObj := responses.FindResponseByID("61")
		globalPkg.SendError(w, responseObj.EngResponse)
		logobj.Count = logobj.Count + 1
		logobj.OutputData = "Invalid Request"
		logobj.Process = "failed"
		broadcastTcp.BoardcastingTCP(logobj, "", "AddAndUpdateLog")
		globalPkg.WriteLog(logobj, responseObj.EngResponse, "failed")
		return
	}
	inoTokenID, _ := globalPkg.ConvertIntToFixedLengthString(0, globalPkg.GlobalObj.TokenIDStringFixedLength)
	inoTokenBalance := transactionModule.GetAccountBalanceStatement(accountobj, inoTokenID)
	var Balance float64
	_, tokenExist := inoTokenBalance[inoTokenID]
	if !tokenExist {
		responseObj := responses.FindResponseByID("59")
		globalPkg.SendError(w, responseObj.EngResponse)
		logobj.Count = logobj.Count + 1
		logobj.OutputData = "you do not have balance for inoToken."
		logobj.Process = "failed"
		broadcastTcp.BoardcastingTCP(logobj, "", "AddAndUpdateLog")
		globalPkg.WriteLog(logobj, responseObj.EngResponse, "failed")
		return
	}
	Balance = inoTokenBalance[inoTokenID].TotalBalance
	cost := serviceobj.Calculation + globalPkg.GlobalObj.TransactionFee
	if cost > Balance {

		jsonObj, _ := json.Marshal(VoucherRespobj)
		globalPkg.SendResponse(w, jsonObj)
		globalPkg.WriteLog(logobj, string(jsonObj), "failed")
		logobj.OutputData = string(jsonObj)
		logobj.Process = "failed"
		logobj.Count = logobj.Count + 1

		broadcastTcp.BoardcastingTCP(logobj, "", "AddAndUpdateLog")

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

	st2 := service.Voucher{"create-voucher", "1", strconv.Itoa(serviceobj.Amount), "2048", "4096", strconv.Itoa(serviceobj.Bandwidth), MBytes, serviceobj.ID, strconv.Itoa(serviceobj.M)}
	Createresponse, createstatus := RequestServiceAPI(st2, PurchaseServiceObj.ID)

	if createstatus {
		serviceobj.VoutcherId = Createresponse.Code
		serviceobj.CreateTime = Createresponse.Create_time
		/////////////////Add ew Transacrtion
		//var transactionobj transaction.DigitalwalletTransaction
		PurchaseServiceObj.Transactionobj.ServiceId = serviceobj.ID
		PurchaseServiceObj.Transactionobj.Validator = validator.CurrentValidator.ValidatorIP
		// transjson, _ := json.Marshal(PurchaseServiceObj.Transactionobj)
		// response := transactionModule.ValidateServiceTransaction(PurchaseServiceObj.Transactionobj)
		response := PurchaseServiceObj.Transactionobj.ValidateServiceTransaction()
		//globalPkg.SendRequestAndGetResponse(transjson, validator.CurrentValidator.ValidatorIP+"/2e4a9d667ad5e3cef02eae9", "POST", &transactionobj)

		// fmt.Println("transaction Response", response)
		if response != "" {
			globalPkg.SendError(w, response)
			globalPkg.WriteLog(logobj, response, "failed")
			logobj.OutputData = response
			logobj.Count = logobj.Count + 1
			logobj.Process = "failed"
			broadcastTcp.BoardcastingTCP(logobj, "", "AddAndUpdateLog")

			return
		} else {
			transactionObj, errorf := transactionModule.DigitalwalletToUTXOTrans(PurchaseServiceObj.Transactionobj, true)
			if errorf != "" {
				globalPkg.SendError(w, errorf)
				globalPkg.WriteLog(logobj, errorf, "failed")
				return
			}
			// transactionObj := transactionModule.DigitalwalletToUTXOTrans(PurchaseServiceObj.Transactionobj)
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
			if logobj.Count > 0 {
				logobj.Count = 0
				logobj.OutputData = string(jsonObj)

				logobj.Process = "success"
				broadcastTcp.BoardcastingTCP(logobj, "", "AddAndUpdateLog")

			}
			broadcastTcp.BoardcastingTCP(serviceobj, "DB", "Add Service")
			return
		}
	} else {
		responseObj := responses.FindResponseByID("62")
		globalPkg.SendError(w, responseObj.EngResponse)
		globalPkg.WriteLog(logobj, responseObj.EngResponse, "failed")
	}
}

//GetAllPurchasedServices Get all purchased services End point
func GetAllPurchasedServices(w http.ResponseWriter, req *http.Request) {
	now, userIP := globalPkg.SetLogObj(req)
	logobj := logpkg.LogStruct{"_", now, userIP, "macAdress", "GetAllPurchasedServices", "serviceModule", "_", "_", "_", 0}
	UserKeysobj := service.UserKeys{}
	decoder := json.NewDecoder(req.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&UserKeysobj)

	if err != nil {
		responseObj := responses.FindResponseByID("1")
		globalPkg.SendError(w, responseObj.EngResponse)
		globalPkg.WriteLog(logobj, responseObj.EngResponse, "failed")
		return
	}

	accountobj := account.GetAccountByAccountPubicKey(UserKeysobj.PublicKey)
	if accountobj.AccountPassword != UserKeysobj.Password {
		responseObj := responses.FindResponseByID("10")
		globalPkg.SendError(w, responseObj.EngResponse)
		globalPkg.WriteLog(logobj, responseObj.EngResponse, "failed")
		return
	}
	//-----call Api Created by Alaa to get All servic using Acccount index
	services := service.ServiceStructGetByPrefix(accountobj.AccountIndex)
	if len(services) == 0 {
		responseObj := responses.FindResponseByID("63")
		globalPkg.SendResponseMessage(w, responseObj.EngResponse)
		globalPkg.WriteLog(logobj, responseObj.EngResponse, "warning")
		return
	}
	jsonObj, _ := json.Marshal(services)
	globalPkg.SendResponse(w, jsonObj)
	globalPkg.WriteLog(logobj, string(jsonObj), "success")
	return

}

//CheckVoucherStatus check voucher status
func CheckVoucherStatus(w http.ResponseWriter, req *http.Request) {

	now, userIP := globalPkg.SetLogObj(req)
	logobj := logpkg.LogStruct{"_", now, userIP, "macAdress", "CheckVoucherStatus", "serviceModule", "_", "_", "_", 0}
	Vocherstatusobj := service.Vocherstatus{}
	decoder := json.NewDecoder(req.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&Vocherstatusobj)

	if err != nil {
		responseObj := responses.FindResponseByID("1")
		globalPkg.SendError(w, responseObj.EngResponse)
		globalPkg.WriteLog(logobj, responseObj.EngResponse, "failed")
		return
	}
	//	var abj  ResponseCreateVoucher//service.VocherstatusResponse

	Create_time := service.GetCreateTime(Vocherstatusobj.VoucherID) //"2019-01-02 03:04:05 PM -0000"
	if Create_time == -1 {
		responseObj := responses.FindResponseByID("64")
		globalPkg.SendError(w, responseObj.EngResponse)
		globalPkg.WriteLog(logobj, responseObj.EngResponse, "failed")
		return
	}
	accountobj := account.GetAccountByAccountPubicKey(Vocherstatusobj.PublicKey)
	if accountobj.AccountPassword != Vocherstatusobj.Password {
		responseObj := responses.FindResponseByID("10")
		globalPkg.SendError(w, responseObj.EngResponse)
		globalPkg.WriteLog(logobj, responseObj.EngResponse, "failed")
		return
	}
	// fmt.Println("*------Create_time", Create_time)
	ResponseVoucher, VoucherIDStatus, ok := GetVoucherDataAPI(Create_time)

	if VoucherIDStatus && ok {
		if ResponseVoucher.Used == 0 {
			responseObj := responses.FindResponseByID("65")
			globalPkg.SendResponseMessage(w, responseObj.EngResponse)
			globalPkg.WriteLog(logobj, responseObj.EngResponse, "warning")
		} else {
			jsonObj, _ := json.Marshal(ResponseVoucher)
			globalPkg.SendResponse(w, jsonObj)
			globalPkg.WriteLog(logobj, string(jsonObj), "success")
		}
		return
	} else if VoucherIDStatus && !ok {
		responseObj := responses.FindResponseByID("62")
		globalPkg.SendError(w, responseObj.EngResponse)
		globalPkg.WriteLog(logobj, responseObj.EngResponse, "failed")
		return
	} else if !VoucherIDStatus && ok {
		responseObj := responses.FindResponseByID("64")
		globalPkg.SendError(w, responseObj.EngResponse)
		globalPkg.WriteLog(logobj, responseObj.EngResponse, "failed")
		return
	}

	return

}

// GetAllNamesandPKsForServiceAccount Api to get All Accounts names and pk for users whose role is service
func GetAllNamesandPKsForServiceAccount(w http.ResponseWriter, req *http.Request) {
	now, userIP := globalPkg.SetLogObj(req)
	logobj := logpkg.LogStruct{"_", now, userIP, "macAdress", "GetAllNamesandPKsForServiceAccount", "serviceModule", "_", "_", "_", 0}

	UserKeysobj := service.UserKeys{}
	decoder := json.NewDecoder(req.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&UserKeysobj)
	// TODO : create log object

	if err != nil {
		responseObj := responses.FindResponseByID("1")
		globalPkg.SendError(w, responseObj.EngResponse)
		globalPkg.WriteLog(logobj, responseObj.EngResponse, "failed")
		return
	}

	accountobj := account.GetAccountByAccountPubicKey(UserKeysobj.PublicKey)
	if accountobj.AccountPassword != UserKeysobj.Password {
		responseObj := responses.FindResponseByID("10")
		globalPkg.SendError(w, responseObj.EngResponse)
		globalPkg.WriteLog(logobj, responseObj.EngResponse, "failed")
		return
	}
	Accounts := accountdb.GetNamesandPKsForServiceAccount()
	jsonObj, _ := json.Marshal(Accounts)
	globalPkg.SendResponse(w, jsonObj)

	return

}

// RequestServiceAPI request a service
func RequestServiceAPI(data service.Voucher, publickey string) (service.ResponseCreateVoucher, bool) {
	//fmt.Println("putchasedataaa", data)
	is_fine := true
	cre := service.ServCredentials{"Administrator", "adeelakramrox2"}
	cred_json, _ := json.Marshal(cre)
	vou_json, _ := json.Marshal(data)
	r, c := service.CreateVoucher(cred_json, vou_json, globalPkg.GlobalObj.ServiceLogin, globalPkg.GlobalObj.ServiceCreateVoutcher, "POST")
	if c != 200 {
		is_fine = false
	} else {
		is_fine = true
	}
	return r, is_fine
}

// GetVoucherDataAPI get service data
func GetVoucherDataAPI(create_time int64) (service.ResponseCreateVoucher, bool, bool) {
	is_fine := true
	var v service.ResponseCreateVoucher
	v.Create_time = create_time
	cre := service.ServCredentials{"Administrator", "adeelakramrox2"}
	jsonObj, _ := json.Marshal(cre)
	// fmt.Println(jsonObj)
GET_VOUCHER:
	v, code := service.GetVoucherData(v)
	if code == http.StatusUnauthorized {
		r := service.ServiceLogin(globalPkg.GlobalObj.ServiceLogin, "POST", jsonObj) // login
		if r != 200 {
			// can not login
			fmt.Println(r)
			is_fine = false
			return v, is_fine, true
		}
		goto GET_VOUCHER
	}
	ok := true
	fmt.Println(v, code)
	switch code {
	case 200:
		is_fine = true
		ok = true
		break
	case 5:
		is_fine = true
		ok = false
		break
	case 4:
		is_fine = false
		ok = true
		break
	default:
		is_fine = true
		ok = false
		break
	}
	return v, is_fine, ok
}
