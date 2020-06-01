package account

import (
	"encoding/json"
	"net/http"
	"time"

	"../accountdb"
	"../broadcastTcp"
	"../globalPkg"
	"../logpkg"
)

var resetPassReq []ResetPasswordData

//ResetPasswordData struct
type ResetPasswordData struct {
	Code        string
	Email       string
	Phonnum     string
	Newpassword string
	CurrentTime time.Time
	PathApi     string
}

//ForgetPasswordReturn  return data from forget password api
type ForgetPasswordReturn struct {
	Code    string
	PathAPI string
}

//SetResetPasswordData func
func SetResetPasswordData(resetPasswordDataObj []ResetPasswordData) {
	resetPassReq = resetPasswordDataObj
}

//GetResetPasswordData func
func GetResetPasswordData() []ResetPasswordData {
	return resetPassReq
}

//findInResetPassPool CHECK IF USER mAKE rEQUEST BEFORE TO RESET HIS PASSWORD
func findInResetPassPool(userResetpass ResetPasswordData) (int, bool) { //check if User found in userobj list
	var errorfound bool
	var index int
	index = -1
	for i, U := range resetPassReq {

		if U.Email == userResetpass.Email && userResetpass.Email != "" && U.Code == userResetpass.Code {
			errorfound = true
			index = i
			break
		}
		if U.Phonnum == userResetpass.Phonnum && userResetpass.Phonnum != "" && U.Code == userResetpass.Code {
			errorfound = true
			index = i
			break
		}
		errorfound = false
	}
	return index, errorfound
}

//ResetPassword user can reset his password
func ResetPassword(w http.ResponseWriter, req *http.Request) {
	//log
	now, userIP := globalPkg.SetLogObj(req)
	logobj := logpkg.LogStruct{"", now, userIP, "macAdress", "ResetPassword", "AccountModule", "_", "_", "_", 0}
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

	// 	logobj = logpkg.LogStruct{Logindex, now, userIP, "macAdress", "ResetPassword", "AccountModule", "_", "_", "_", 0}
	// } else {
	// 	logobj = logfunc.ReplaceLog(logobj, "ResetPassword", "AccountModule")
	// }
	ResetPasswordDataObj := ResetPasswordData{}

	// check on path url
	existurl := false
	for _, resetObj := range resetPassReq {
		p := "/" + resetObj.PathApi

		if req.URL.Path == p {
			existurl = true
			break
		}
	}

	if existurl == false {
		globalPkg.SendError(w, "this page not found")
		logobj.OutputData = "this page not found"
		logobj.Process = "faild"
		logpkg.WriteOnlogFile(logobj)
		// logobj.Count = logobj.Count + 1

		// broadcastTcp.BoardcastingTCP(logobj, "", "AddAndUpdateLog")

		return
	}

	Data := ResetPasswordData{}
	decoder := json.NewDecoder(req.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&ResetPasswordDataObj)
	if err != nil {
		globalPkg.SendError(w, "please enter your correct request ")
		globalPkg.WriteLog(logobj, "please enter your correct request", "failed")

		return
	}
	InputData := Data.Email
	logobj.InputData = InputData
	//email is lowercase
	ResetPasswordDataObj.Email = convertStringTolowerCaseAndtrimspace(ResetPasswordDataObj.Email)
	var AccountObj accountdb.AccountStruct
	i, found := findInResetPassPool(ResetPasswordDataObj)

	if found == false {
		globalPkg.SendNotFound(w, "Invalid Data")
		globalPkg.WriteLog(logobj, "invalid data", "failed")
		logobj.OutputData = "Invalid Data"
		logobj.Process = "failed"
		// logobj.Count = logobj.Count + 1

		// broadcastTcp.BoardcastingTCP(logobj, "", "AddAndUpdateLog")

		return

	}
	/////----Data should be removed from list
	Data = resetPassReq[i]

	if len(ResetPasswordDataObj.Newpassword) != 64 {
		globalPkg.SendError(w, " invalid password length ")
		globalPkg.WriteLog(logobj, "invalid password length ", "failed")
		return
	}
	sub := now.Sub(Data.CurrentTime).Seconds()
	if sub > 3000 {
		globalPkg.SendError(w, "Time out ")
		logobj.OutputData = "Time out "
		logobj.Process = "faild"
		globalPkg.WriteLog(logobj, "time out", "failed")

		// logobj.Count = logobj.Count + 1

		// broadcastTcp.BoardcastingTCP(logobj, "", "AddAndUpdateLog")

		return
	}
	if ResetPasswordDataObj.Email != "" {
		AccountObj = getAccountByEmail(ResetPasswordDataObj.Email)
		AccountObj.AccountPassword = ResetPasswordDataObj.Newpassword
		broadcastTcp.BoardcastingTCP(AccountObj, "Resetpass", "account")
		globalPkg.SendResponseMessage(w, "your password successfully changed")
		globalPkg.WriteLog(logobj, "your password successfully changed", "success")

		// if logobj.Count > 0 {
		// 	logobj.OutputData = "your password successfully changed"
		// 	logobj.Process = "success"
		// 	// logobj.Count = 0

		// 	// broadcastTcp.BoardcastingTCP(logobj, "", "AddAndUpdateLog")

		// }

		return

	}

	if ResetPasswordDataObj.Phonnum != "" {
		AccountObj = getAccountByPhone(ResetPasswordDataObj.Email)
		AccountObj.AccountPassword = ResetPasswordDataObj.Newpassword
		broadcastTcp.BoardcastingTCP(AccountObj, "Resetpass", "account") //	updateAccount(AccountObj)
		globalPkg.SendResponseMessage(w, "your password successfully changed")
		globalPkg.WriteLog(logobj, "your password successfully changed", "success")
		// if logobj.Count > 0 {
		// 	logobj.OutputData = "your password successfully changed"
		// 	logobj.Process = "success"
		// 	logobj.Count = 0

		// 	broadcastTcp.BoardcastingTCP(logobj, "", "AddAndUpdateLog")

		// }

		return

	}

}

//ForgetPassword user can make Request For Remmeber the password
func ForgetPassword(w http.ResponseWriter, req *http.Request) {
	//log
	now, userIP := globalPkg.SetLogObj(req)
	logobj := logpkg.LogStruct{"_", now, userIP, "macAdress", "ForgetPassword", "AccountModule", "_", "_", "_", 0}

	RandomCode := encodeToString(globalPkg.GlobalObj.MaxConfirmcode)
	current := globalPkg.UTCtime()
	Confirmation_code := RandomCode
	contact := ResetPasswordData{}
	decoder := json.NewDecoder(req.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&contact)
	if err != nil {
		globalPkg.SendError(w, "please enter your correct request ")
		globalPkg.WriteLog(logobj, "please enter your correct request", "failed")
		return
	}
	// ForgetPassword email is lowercase and trim
	contact.Email = convertStringTolowerCaseAndtrimspace(contact.Email)
	RP := ResetPasswordData{}
	RP.Code = Confirmation_code
	RP.CurrentTime = current
	accountObjbyEmail := getAccountByEmail(contact.Email)
	accountObjByPhone := getAccountByPhone(contact.Phonnum)

	forgetObj := ForgetPasswordReturn{}
	if accountObjbyEmail.AccountPublicKey != "" && contact.Email != "" {

		RP.Email = contact.Email
		RP.PathApi = globalPkg.RandomPath()
		broadcastTcp.BoardcastingTCP(RP, "addRestPassword", "account module")
		//addResetpassObjInTemp(RP)
		//body email for forget password
		body := "Dear " + accountObjbyEmail.AccountName + `,
		You recently requested to reset your password for your Inovation Corporation account.
		Your confirmation code is: ` + RP.Code + `.
		if you did not request a password reset, please ignore this email or reply to let us know.
		
		Regards,
		Inovatian Team`

		sendEmail(body, contact.Email)
		forgetObj.Code = RP.Code
		forgetObj.PathAPI = RP.PathApi
		jsonObj, _ := json.Marshal(forgetObj)
		globalPkg.SendResponse(w, jsonObj)
		// globalPkg.SendResponseMessage(w, Confirmation_code)
		globalPkg.WriteLog(logobj, "success send confirmation code"+RP.Code, "success")
		return
	}

	if accountObjByPhone.AccountPublicKey != "" && contact.Phonnum != "" {

		RP.Phonnum = contact.Phonnum
		RP.PathApi = globalPkg.RandomPath()
		broadcastTcp.BoardcastingTCP(RP, "addRestPassword", "account module")

		// send_SMS(contact.Phonnum, RP.Code)
		forgetObj.Code = RP.Code
		forgetObj.PathAPI = RP.PathApi
		// globalPkg.SendResponseMessage(w, Confirmation_code)
		jsonObj, _ := json.Marshal(forgetObj)
		globalPkg.SendResponse(w, jsonObj)
		globalPkg.WriteLog(logobj, "success send confirmation code"+RP.Code, "success")
		return
	}

	globalPkg.SendError(w, "invalid Email Or Phone")
	globalPkg.WriteLog(logobj, "invalid Email Or Phone"+RP.Code, "failed")

}

//UpdateResetpassObjInTemp for reset password
func UpdateResetpassObjInTemp(index int, ResetpassObj ResetPasswordData) {
	resetPassReq = append(resetPassReq[:index], resetPassReq[index+1:]...)
	resetPassReq = append(resetPassReq, ResetpassObj)
}
