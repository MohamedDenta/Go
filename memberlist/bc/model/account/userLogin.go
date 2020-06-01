package account

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"../accountdb"
	"../broadcastTcp"
	"../errorpk"
	"../globalPkg"
	"../logpkg"
)

type loginUser struct {
	EmailOrPhone string
	Password     string
	SessionID    string
	AuthValue    string
}

type savekey struct {
	PublicKey string
	Passsword string
	Email     string
}

//Login  End Point user can login to his Account using Email Or phone and password
func Login(w http.ResponseWriter, req *http.Request) {

	now, userIP := globalPkg.SetLogObj(req)
	logobj := logpkg.LogStruct{"", now, userIP, "macAdress", "Login", "AccountModule", "", "", "_", 0}
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

	// 	logobj = logpkg.LogStruct{Logindex, now, userIP, "macAdress", "Login", "AccountModule", "", "", "_", 0}
	// }
	// logobj = logfunc.ReplaceLog(logobj, "Login", "AccountModule")

	var NewloginUser = loginUser{}
	var SessionObj accountdb.AccountSessionStruct

	decoder := json.NewDecoder(req.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&NewloginUser)
	if err != nil {
		globalPkg.SendError(w, "please enter your correct request ")
		globalPkg.WriteLog(logobj, "failed to decode object", "failed")
		return
	}
	InputData := NewloginUser.EmailOrPhone + "," + NewloginUser.Password + "," + NewloginUser.AuthValue
	logobj.InputData = InputData
	//confirm email is lowercase and trim
	NewloginUser.EmailOrPhone = convertStringTolowerCaseAndtrimspace(NewloginUser.EmailOrPhone)
	if NewloginUser.EmailOrPhone == "" {
		globalPkg.WriteLog(logobj, "please Enter your Email Or Phone", "failed")
		globalPkg.SendNotFound(w, "please Enter your Email Or Phone")
		return
	}
	if NewloginUser.AuthValue == "" && NewloginUser.Password == "" {
		globalPkg.WriteLog(logobj, "please Enter your password  Or Authvalue", "failed")
		globalPkg.SendError(w, "please Enter your password  Or Authvalue")
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
		globalPkg.SendError(w, "Account not found please check your email or phone")
		globalPkg.WriteLog(logobj, "Account not found please check your email or phone", "failed")
		// logobj.Count = logobj.Count + 1
		// logobj.OutputData = "Account not found please check your email or phone"
		// logobj.Process = "failed"
		// broadcastTcp.BoardcastingTCP(logobj, "", "AddAndUpdateLog")

		return
	}

	if accountObj.AccountIndex == "" && Email == true { //AccountPublicKey replaces with AccountIndex
		globalPkg.WriteLog(logobj, "Please,Check your account Email ", "failed")
		globalPkg.SendNotFound(w, "Please,Check your account Email ")
		// logobj.Count = logobj.Count + 1
		// logobj.OutputData = "Please,Check your account Email "
		// logobj.Process = "failed"
		// broadcastTcp.BoardcastingTCP(logobj, "", "AddAndUpdateLog")

		return
	}
	if accountObj.AccountIndex == "" && Email == false {
		globalPkg.WriteLog(logobj, "Please,Check your account phone ", "failed")
		globalPkg.SendNotFound(w, "Please,Check your account phone ")
		// logobj.OutputData = "Please,Check your account phone "
		// logobj.Process = "failed"
		// logobj.Count = logobj.Count + 1

		// broadcastTcp.BoardcastingTCP(logobj, "", "AddAndUpdateLog")

		return
	}

	if (accountObj.AccountName == "" || (NewloginUser.Password != "" && accountObj.AccountPassword != NewloginUser.Password && Email == true) || (accountObj.AccountEmail != NewloginUser.EmailOrPhone && Email == true && NewloginUser.Password != "")) && NewloginUser.Password != "" {
		globalPkg.WriteLog(logobj, "Please,Check your account Email or password", "failed")
		globalPkg.SendError(w, "Please,Check your account Email or password")

		// logobj.Count = logobj.Count + 1
		// logobj.OutputData = "Please,Check your account Email or password"
		// logobj.Process = "failed"
		// broadcastTcp.BoardcastingTCP(logobj, "", "AddAndUpdateLog")

		return
	}
	if (accountObj.AccountName == "" || (accountObj.AccountAuthenticationValue != NewloginUser.AuthValue && Email == true) || (accountObj.AccountEmail != NewloginUser.EmailOrPhone && Email == true)) && NewloginUser.AuthValue != "" {
		globalPkg.WriteLog(logobj, "Please,Check your account Email or AuthenticationValue", "failed")

		// logobj.Count = logobj.Count + 1
		// logobj.OutputData = "Please,Check your account Email or AuthenticationValues"
		// logobj.Process = "failed"
		// broadcastTcp.BoardcastingTCP(logobj, "", "AddAndUpdateLog")
		globalPkg.SendError(w, "Please,Check your account Email or AuthenticationValue")
		return

	}
	if (accountObj.AccountName == "" || (strings.TrimSpace(accountObj.AccountPhoneNumber) != "" && Email == false) || (accountObj.AccountPassword != NewloginUser.Password && Email == false) || (accountObj.AccountPhoneNumber != NewloginUser.EmailOrPhone && Email == false)) && NewloginUser.Password != "" {
		fmt.Println("i am a phone")
		globalPkg.WriteLog(logobj, "Please,Check your account  phoneNAmber OR password", "failed")

		// logobj.Count = logobj.Count + 1
		// logobj.OutputData = "Please,Check your account  phoneNAmber OR password"
		// logobj.Process = "failed"
		// broadcastTcp.BoardcastingTCP(logobj, "", "AddAndUpdateLog")
		globalPkg.SendError(w, "Please,Check your account  phoneNAmber OR password")
		return

	}

	if (accountObj.AccountName == "" || (strings.TrimSpace(accountObj.AccountPhoneNumber) != "" && Email == false) || (accountObj.AccountPassword != NewloginUser.AuthValue && Email == false) || (accountObj.AccountPhoneNumber != NewloginUser.EmailOrPhone && Email == false)) && NewloginUser.AuthValue != "" {
		fmt.Println("i am a phone")
		globalPkg.WriteLog(logobj, "Please,Check your account  phoneNAmber OR AuthenticationValue", "failed")

		// logobj.Count = logobj.Count + 1
		// logobj.OutputData = "Please,Check your account  phoneNAmber OR AuthenticationValue"
		// logobj.Process = "failed"
		// broadcastTcp.BoardcastingTCP(logobj, "", "AddAndUpdateLog")
		globalPkg.SendError(w, "Please,Check your account  phoneNAmber OR AuthenticationValue")
		return

	}

	if accountObj.AccountPublicKey == "" && accountObj.AccountName != "" {
		var user User

		user = createPublicAndPrivate(user)

		broadcastTcp.BoardcastingTCP(accountObj, "POST", "account")
		accountObj.AccountPublicKey = user.Account.AccountPublicKey
		accountObj.AccountPrivateKey = user.Account.AccountPrivateKey
		sendJson, _ := json.Marshal(accountObj)

		//w.Header().Set("Content-Type", "application/json")
		w.Header().Set("jwt-token", globalPkg.GenerateJwtToken(accountObj.AccountName, false)) // set jwt token
		//w.WriteHeader(http.StatusOK)
		//w.Write(sendJson)
		globalPkg.SendResponse(w, sendJson)
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

	// if logobj.Count > 0 {
	// 	logobj.Count = 0
	// 	logobj.OutputData = accountObj.AccountName
	// 	logobj.Process = "success"
	// 	broadcastTcp.BoardcastingTCP(logobj, "", "AddAndUpdateLog")

	// }
	sendJson, _ := json.Marshal(accountObj)
	w.Header().Set("jwt-token", globalPkg.GenerateJwtToken(accountObj.AccountName, false)) // set jwt token
	globalPkg.SendResponse(w, sendJson)

}

//SavePublickey api called when user saves his private key
func SavePublickey(w http.ResponseWriter, req *http.Request) {
	//log
	now, userIP := globalPkg.SetLogObj(req)
	logobj := logpkg.LogStruct{"_", now, userIP, "macAdress", "SavePublicKey", "AccountModule", "", "", "_", 0}

	var saveKeyReq savekey
	decoder := json.NewDecoder(req.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&saveKeyReq)
	errStr := ""

	if err != nil {
		errStr = errorpk.AddError("SavePublickey AccountModuleAPI  "+req.Method, "can't convert body to saveKeyReq obj", "runtime error")
		globalPkg.SendError(w, "please enter your correct request make sure that your req contain : Publickey then Password then Email ")
		globalPkg.WriteLog(logobj, "please enter your correct request", "failed")
		return
	}
	account := getAccountByEmail(saveKeyReq.Email)
	if account.AccountEmail == saveKeyReq.Email && account.AccountPassword == saveKeyReq.Passsword {
		account.AccountPublicKey = saveKeyReq.PublicKey
		broadcastTcp.BoardcastingTCP(account, "set public key", "account")
	} else {
		globalPkg.SendError(w, "failed to get this email plz check if account exist an make sure to insert the right pass and email!")
		globalPkg.WriteLog(logobj, "failed to get this email plz check if account exist an make sure to insert the right pass and email!", "failed")
		return
	}
	if ifAccountExistsBefore(saveKeyReq.PublicKey) {
		globalPkg.SendError(w, "this public key exists before")
		globalPkg.WriteLog(logobj, errStr, "failed")
		return
	}
	// broadcastTcp.BoardcastingTCP(account, "set public key", "account")
	if errStr == "" {
		sendJson, _ := json.Marshal(account)
		globalPkg.SendResponse(w, sendJson)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
	} else {
		globalPkg.SendError(w, errStr)
		globalPkg.WriteLog(logobj, errStr, "failed")
	}

}
