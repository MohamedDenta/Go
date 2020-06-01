package account

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"../logfunc"
	"../responses"

	"../accountdb"
	"../broadcastTcp"
	"../errorpk"
	"../globalPkg"
	"../logpkg"
)
// loginUser login user
type loginUser struct {
	EmailOrPhone string
	Password     string
	SessionID    string
	AuthValue    string
}
// savekey save pk
type savekey struct {
	PublicKey string
	Passsword string
	Email     string
}
// Login api 
func Login(w http.ResponseWriter, req *http.Request) {

	now, userIP := globalPkg.GetIP(req)

	found, logobj := logpkg.CheckIfLogFound(userIP)

	if found && now.Sub(logobj.Currenttime).Seconds() > globalPkg.GlobalObj.DeleteAccountTimeInseacond {

		logobj.Count = 0
		broadcastTcp.BoardcastingTCP(logobj, "", "AddAndUpdateLog")

	}
	if found && logobj.Count >= 10 {

		globalPkg.SendError(w, "your Account have been blocked")
		return
	}

	if !found {

		Logindex := userIP.String() + "_" + logfunc.NewLogIndex()

		logobj = logpkg.LogStruct{Logindex, now, userIP, "macAdress", "Login", "AccountModule", "", "", "_", 0}
	}
	logobj = logfunc.ReplaceLog(logobj, "Login", "AccountModule")

	var NewloginUser = loginUser{}
	var SessionObj accountdb.AccountSessionStruct

	decoder := json.NewDecoder(req.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&NewloginUser)
	if err != nil {
		responseObj := responses.FindResponseByID("1")
		globalPkg.SendError(w, responseObj.EngResponse)
		globalPkg.WriteLog(logobj, "failed to decode object", "failed")
		return
	}
	InputData := NewloginUser.EmailOrPhone + "," + NewloginUser.Password + "," + NewloginUser.AuthValue
	logobj.InputData = InputData
	//confirm email is lowercase and trim
	NewloginUser.EmailOrPhone = convertStringTolowerCaseAndtrimspace(NewloginUser.EmailOrPhone)
	if NewloginUser.EmailOrPhone == "" {
		responseObj := responses.FindResponseByID("22")
		globalPkg.SendError(w, responseObj.EngResponse)
		globalPkg.WriteLog(logobj, "please Enter your Email Or Phone", "failed")
		return
	}
	if NewloginUser.AuthValue == "" && NewloginUser.Password == "" {
		responseObj := responses.FindResponseByID("23")
		globalPkg.SendError(w, responseObj.EngResponse)
		globalPkg.WriteLog(logobj, "please Enter your password  Or Authvalue", "failed")
		return
	}

	var accountObj accountdb.AccountStruct
	var Email bool
	Email = false
	if strings.Contains(NewloginUser.EmailOrPhone, "@") && strings.Contains(NewloginUser.EmailOrPhone, ".") {
		Email = true
		accountObj = getAccountByEmail(NewloginUser.EmailOrPhone)
	} else {
		accountObj = getAccountByPhone(NewloginUser.EmailOrPhone)
	}

	//if account is not found whith data logged in with
	if accountObj.AccountPublicKey == "" && accountObj.AccountName == "" {
		responseObj := responses.FindResponseByID("22")
		globalPkg.SendError(w, responseObj.EngResponse)
		globalPkg.WriteLog(logobj, "Account not found please check your email or phone", "failed")
		logobj.Count = logobj.Count + 1
		logobj.OutputData = "Account not found please check your email or phone"
		logobj.Process = "failed"
		broadcastTcp.BoardcastingTCP(logobj, "", "AddAndUpdateLog")

		return
	}

	if accountObj.AccountIndex == "" && Email == true { //AccountPublicKey replaces with AccountIndex
		responseObj := responses.FindResponseByID("25")
		globalPkg.SendNotFound(w, responseObj.EngResponse)
		globalPkg.WriteLog(logobj, "Please,Check your account Email ", "failed")
	
		logobj.Count = logobj.Count + 1
		logobj.OutputData = "Please,Check your account Email "
		logobj.Process = "failed"
		broadcastTcp.BoardcastingTCP(logobj, "", "AddAndUpdateLog")

		return
	}
	if accountObj.AccountIndex == "" && Email == false {
		responseObj := responses.FindResponseByID("11")
		globalPkg.SendError(w, responseObj.EngResponse)
		globalPkg.WriteLog(logobj, "Please,Check your account phone ", "failed")
		logobj.OutputData = "Please,Check your account phone "
		logobj.Process = "failed"
		logobj.Count = logobj.Count + 1

		broadcastTcp.BoardcastingTCP(logobj, "", "AddAndUpdateLog")

		return
	}

	if (accountObj.AccountName == "" || (NewloginUser.Password != "" && accountObj.AccountPassword != NewloginUser.Password && Email == true) || (accountObj.AccountEmail != NewloginUser.EmailOrPhone && Email == true && NewloginUser.Password != "")) && NewloginUser.Password != "" {
		responseObj := responses.FindResponseByID("25")
		globalPkg.SendError(w, responseObj.EngResponse)
		globalPkg.WriteLog(logobj, "Please,Check your account Email or password", "failed")

		logobj.Count = logobj.Count + 1
		logobj.OutputData = "Please,Check your account Email or password"
		logobj.Process = "failed"
		broadcastTcp.BoardcastingTCP(logobj, "", "AddAndUpdateLog")

		return
	}
	if (accountObj.AccountName == "" || (accountObj.AccountAuthenticationValue != NewloginUser.AuthValue && Email == true) || (accountObj.AccountEmail != NewloginUser.EmailOrPhone && Email == true)) && NewloginUser.AuthValue != "" {
		responseObj := responses.FindResponseByID("25")
		globalPkg.SendError(w, responseObj.EngResponse)
		globalPkg.WriteLog(logobj, "Please,Check your account Email or AuthenticationValue", "failed")
		logobj.Count = logobj.Count + 1
		logobj.OutputData = "Please,Check your account Email or AuthenticationValues"
		logobj.Process = "failed"
		broadcastTcp.BoardcastingTCP(logobj, "", "AddAndUpdateLog")
		return

	}
	if (accountObj.AccountName == "" || (strings.TrimSpace(accountObj.AccountPhoneNumber) != "" && Email == false) || (accountObj.AccountPassword != NewloginUser.Password && Email == false) || (accountObj.AccountPhoneNumber != NewloginUser.EmailOrPhone && Email == false)) && NewloginUser.Password != "" {
		fmt.Println("i am a phone")
		responseObj := responses.FindResponseByID("27")
		globalPkg.SendError(w, responseObj.EngResponse)
		globalPkg.WriteLog(logobj, "Please,Check your account  phoneNAmber OR password", "failed")
		logobj.Count = logobj.Count + 1
		logobj.OutputData = "Please,Check your account  phoneNAmber OR password"
		logobj.Process = "failed"
		broadcastTcp.BoardcastingTCP(logobj, "", "AddAndUpdateLog")
		return

	}

	if (accountObj.AccountName == "" || (strings.TrimSpace(accountObj.AccountPhoneNumber) != "" && Email == false) || (accountObj.AccountPassword != NewloginUser.AuthValue && Email == false) || (accountObj.AccountPhoneNumber != NewloginUser.EmailOrPhone && Email == false)) && NewloginUser.AuthValue != "" {
		// fmt.Println("i am a phone")
		globalPkg.WriteLog(logobj, "Please,Check your account  phoneNAmber OR AuthenticationValue", "failed")

		logobj.Count = logobj.Count + 1
		logobj.OutputData = "Please,Check your account  phoneNAmber OR AuthenticationValue"
		logobj.Process = "failed"
		broadcastTcp.BoardcastingTCP(logobj, "", "AddAndUpdateLog")
		responseObj := responses.FindResponseByID("11")
		globalPkg.SendError(w, responseObj.EngResponse)
		return
	}

	if accountObj.AccountPublicKey == "" && accountObj.AccountName != "" { // first login
		var user User
		user = createPublicAndPrivate(user)
		// accountObj.AccountPublicKey = user.Account.AccountPublicKey
		broadcastTcp.BoardcastingTCP(accountObj, "POST", "account")
		accountObj.AccountPublicKey = user.Account.AccountPublicKey
		accountObj.AccountPrivateKey = user.Account.AccountPrivateKey
		sendJSON, _ := json.Marshal(accountObj)

		
		w.Header().Set("jwt-token", globalPkg.GenerateJwtToken(accountObj.AccountName, false)) // set jwt token
		
		globalPkg.SendResponse(w, sendJSON)
		SessionObj.SessionId = NewloginUser.SessionID
		SessionObj.AccountIndex = accountObj.AccountIndex
		//--search if sesssion found
		// session should be unique
		flag, _ := CheckIfsessionFound(SessionObj)

		if flag == true {

			broadcastTcp.BoardcastingTCP(SessionObj, "", "Delete Session")

		}
		broadcastTcp.BoardcastingTCP(SessionObj, "", "Add Session")

		return

	}
	fmt.Println(accountObj)
	SessionObj.SessionId = NewloginUser.SessionID
	SessionObj.AccountIndex = accountObj.AccountIndex
	//--search if sesssion found
	// session should be unique
	flag, _ := CheckIfsessionFound(SessionObj)

	if flag == true {
		broadcastTcp.BoardcastingTCP(SessionObj, "", "Delete Session")
	}
	broadcastTcp.BoardcastingTCP(SessionObj, "", "Add Session")
	globalPkg.WriteLog(logobj, accountObj.AccountName+","+accountObj.AccountPassword+","+accountObj.AccountEmail+","+accountObj.AccountRole, "success")
	if logobj.Count > 0 {
		logobj.Count = 0
		logobj.OutputData = accountObj.AccountName
		logobj.Process = "success"
		broadcastTcp.BoardcastingTCP(logobj, "", "AddAndUpdateLog")
	}

	sendJSON, _ := json.Marshal(accountObj)
	w.Header().Set("jwt-token", globalPkg.GenerateJwtToken(accountObj.AccountName, false)) // set jwt token
	globalPkg.SendResponse(w, sendJSON)

}
// SavePublickey save pk
func SavePublickey(w http.ResponseWriter, req *http.Request) {
	//log
	now, userIP := globalPkg.GetIP(req)
	logobj := logpkg.LogStruct{"", now, userIP, "macAdress", "SavePublicKey", "AccountModule", "", "", "", 0}

	var saveKeyReq savekey
	decoder := json.NewDecoder(req.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&saveKeyReq)
	errStr := ""
	if err != nil {
		responseObj := responses.FindResponseByID("1")
		globalPkg.SendError(w, responseObj.EngResponse)
		errStr = errorpk.AddError("SavePublickey AccountModuleAPI  "+req.Method, "can't convert body to saveKeyReq obj", "runtime error")
		globalPkg.WriteLog(logobj, "please enter your correct request", "failed")
		return
	}
	account := getAccountByEmail(saveKeyReq.Email)
	fmt.Println("Denta ", account.AccountPublicKey)
	if account.AccountEmail != saveKeyReq.Email || account.AccountPassword != saveKeyReq.Passsword {
		errStr = errorpk.AddError("SavePublickey", "error in email or password", "hack")
		responseObj := responses.FindResponseByID("25")
		globalPkg.SendError(w, responseObj.EngResponse)
		globalPkg.WriteLog(logobj, errStr, "failed")
		return
	}
	if !FindAdressInTemp(saveKeyReq.PublicKey) {
		errStr = errorpk.AddError("SavePublickey", "wrong public key", "hack")
		responseObj := responses.FindResponseByID("10")
		globalPkg.SendError(w, responseObj.EngResponse)
		globalPkg.WriteLog(logobj, errStr, "failed")
		return
	}
	account.AccountPublicKey = saveKeyReq.PublicKey
	account.AccountStatus = true
	broadcastTcp.BoardcastingTCP(account, "set public key", "account")

	sendJSON, _ := json.Marshal(account)
	globalPkg.SendResponse(w, sendJSON)

}
