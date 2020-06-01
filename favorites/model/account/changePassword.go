package account

import (
	"encoding/json"
	"net/http"
	"time"

	"../accountdb"
	"../broadcastTcp"
	"../globalPkg"
	"../logfunc"
	"../logpkg"
	"../responses"
)

// resetPassReq array of reset password
var resetPassReq []ResetPasswordData

//ResetPasswordData struct
type ResetPasswordData struct {
	Code        string
	Email       string
	Phonnum     string
	Newpassword string
	CurrentTime time.Time
	PathAPI     string
}

//ForgetPasswordReturn  return data from forget password api Reponse body
type ForgetPasswordReturn struct {
	Code    string
	PathAPI string
}

//SetResetPasswordData  set reset paaword func
func SetResetPasswordData(resetPasswordDataObj []ResetPasswordData) {
	resetPassReq = resetPasswordDataObj
}

//GetResetPasswordData get  reset paaword func
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
	now, userIP := globalPkg.GetIP(req)

	found, logobj := logpkg.CheckIfLogFound(userIP)

	if found && now.Sub(logobj.Currenttime).Seconds() > globalPkg.GlobalObj.DeleteAccountTimeInseacond {

		logobj.Count = 0
		broadcastTcp.BoardcastingTCP(logobj, "", "AddAndUpdateLog")

	}
	if found && logobj.Count >= 10 {
		responseObj := responses.FindResponseByID("6")
		globalPkg.SendError(w, responseObj.EngResponse)
		return
	}

	if !found {

		Logindex := userIP.String() + "_" + logfunc.NewLogIndex()

		logobj = logpkg.LogStruct{Logindex, now, userIP, "macAdress", "ResetPassword", "AccountModule", "_", "_", "_", 0}
	} else {
		logobj = logfunc.ReplaceLog(logobj, "ResetPassword", "AccountModule")
	}
	ResetPasswordDataObj := ResetPasswordData{}

	// check on path url
	existurl := false
	for _, resetObj := range resetPassReq {
		p := "/" + resetObj.PathAPI

		if req.URL.Path == p {
			existurl = true
			break
		}
	}

	if existurl == false {
		responseObj := responses.FindResponseByID("12")
		globalPkg.SendError(w, responseObj.EngResponse)
		logobj.OutputData = "this page is not found"
		logobj.Process = "faild"
		logpkg.WriteOnlogFile(logobj)
		logobj.Count = logobj.Count + 1

		broadcastTcp.BoardcastingTCP(logobj, "", "AddAndUpdateLog")
		return
	}

	Data := ResetPasswordData{}
	decoder := json.NewDecoder(req.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&ResetPasswordDataObj)
	if err != nil {
		responseObj := responses.FindResponseByID("1")
		globalPkg.SendError(w, responseObj.EngResponse)
		globalPkg.WriteLog(logobj, responseObj.EngResponse, "failed")

		return
	}
	InputData := Data.Email
	logobj.InputData = InputData
	//email is lowercase
	ResetPasswordDataObj.Email = convertStringTolowerCaseAndtrimspace(ResetPasswordDataObj.Email)
	var AccountObj accountdb.AccountStruct
	i, found := findInResetPassPool(ResetPasswordDataObj)

	if found == false {
		responseObj := responses.FindResponseByID("104")
		globalPkg.SendNotFound(w, responseObj.EngResponse)
		globalPkg.WriteLog(logobj, "invalid data", "failed")
		logobj.OutputData = "Invalid Data"
		logobj.Process = "failed"
		logobj.Count = logobj.Count + 1

		broadcastTcp.BoardcastingTCP(logobj, "", "AddAndUpdateLog")

		return

	}
	//Data should be removed from list
	Data = resetPassReq[i]

	if len(ResetPasswordDataObj.Newpassword) != 64 {
		responseObj := responses.FindResponseByID("7")
		globalPkg.SendError(w, responseObj.EngResponse)
		globalPkg.WriteLog(logobj, "invalid password length ", "failed")
		return
	}
	sub := now.Sub(Data.CurrentTime).Seconds()
	if sub > 3000 {
		responseObj := responses.FindResponseByID("8")
		globalPkg.SendError(w, responseObj.EngResponse)
		logobj.OutputData = "Time out "
		logobj.Process = "faild"
		globalPkg.WriteLog(logobj, "time out", "failed")
		logobj.Count = logobj.Count + 1
		broadcastTcp.BoardcastingTCP(logobj, "", "AddAndUpdateLog")
		return
	}
	if ResetPasswordDataObj.Email != "" {
		AccountObj = getAccountByEmail(ResetPasswordDataObj.Email)
		AccountObj.AccountPassword = ResetPasswordDataObj.Newpassword
		AccountObj.AccountLastUpdatedTime = now
		broadcastTcp.BoardcastingTCP(AccountObj, "Resetpass", "account")
		responseObj := responses.FindResponseByID("9")
		globalPkg.SendResponseMessage(w, responseObj.EngResponse)
		globalPkg.WriteLog(logobj, "your password successfully changed", "success")

		if logobj.Count > 0 {
			logobj.OutputData = "your password successfully changed"
			logobj.Process = "success"
			logobj.Count = 0

			broadcastTcp.BoardcastingTCP(logobj, "", "AddAndUpdateLog")
		}
		return
	}

	if ResetPasswordDataObj.Phonnum != "" {
		AccountObj = getAccountByPhone(ResetPasswordDataObj.Email)
		AccountObj.AccountPassword = ResetPasswordDataObj.Newpassword
		AccountObj.AccountLastUpdatedTime = now
		broadcastTcp.BoardcastingTCP(AccountObj, "Resetpass", "account") //	updateAccount(AccountObj)
		responseObj := responses.FindResponseByID("9")
		globalPkg.SendResponseMessage(w, responseObj.EngResponse)
		globalPkg.WriteLog(logobj, "your password successfully changed", "success")
		if logobj.Count > 0 {
			logobj.OutputData = "your password successfully changed"
			logobj.Process = "success"
			logobj.Count = 0

			broadcastTcp.BoardcastingTCP(logobj, "", "AddAndUpdateLog")
		}
		return
	}

}

//ForgetPassword user can make Request For Remmeber the password
func ForgetPassword(w http.ResponseWriter, req *http.Request) {
	//log
	now, userIP := globalPkg.GetIP(req)
	logobj := logpkg.LogStruct{"_", now, userIP, "macAdress", "ForgetPassword", "AccountModule", "_", "_", "_", 0}

	RandomCode := encodeToString(globalPkg.GlobalObj.MaxConfirmcode)
	current := globalPkg.UTCtime()
	ConfirmationCode := RandomCode
	contact := ResetPasswordData{}
	decoder := json.NewDecoder(req.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&contact)
	if err != nil {
		responseObj := responses.FindResponseByID("1")
		globalPkg.SendError(w, responseObj.EngResponse)
		globalPkg.WriteLog(logobj, "please enter your correct request", "failed")
		return
	}
	// ForgetPassword email is lowercase and trim
	contact.Email = convertStringTolowerCaseAndtrimspace(contact.Email)
	RP := ResetPasswordData{}
	RP.Code = ConfirmationCode
	RP.CurrentTime = current
	accountObjbyEmail := getAccountByEmail(contact.Email)
	accountObjByPhone := getAccountByPhone(contact.Phonnum)
	forgetObj := ForgetPasswordReturn{}
	if accountObjbyEmail.AccountPublicKey != "" && contact.Email != "" {

		RP.Email = contact.Email
		RP.PathAPI = globalPkg.RandomPath()
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
		forgetObj.PathAPI = RP.PathAPI
		jsonObj, _ := json.Marshal(forgetObj)
		globalPkg.SendResponse(w, jsonObj)
		//globalPkg.SendResponseMessage(w, Confirmation_code)
		globalPkg.WriteLog(logobj, "success send confirmation code"+RP.Code, "success")
		return
	}

	if accountObjByPhone.AccountPublicKey != "" && contact.Phonnum != "" {

		RP.Phonnum = contact.Phonnum
		RP.PathAPI = globalPkg.RandomPath()
		broadcastTcp.BoardcastingTCP(RP, "addRestPassword", "account module")
		//addResetpassObjInTemp(RP)
		sendSMS(contact.Phonnum, RP.Code)
		forgetObj.Code = RP.Code
		forgetObj.PathAPI = RP.PathAPI

		jsonObj, _ := json.Marshal(forgetObj)
		globalPkg.SendResponse(w, jsonObj)
		globalPkg.WriteLog(logobj, "success send confirmation code"+RP.Code, "success")
		return
	}
	responseObj := responses.FindResponseByID("11")
	globalPkg.SendError(w, responseObj.EngResponse)
	globalPkg.WriteLog(logobj, "invalid Email Or Phone"+RP.Code, "failed")

}

//UpdateResetpassObjInTemp for reset password
func UpdateResetpassObjInTemp(index int, ResetpassObj ResetPasswordData) {
	resetPassReq = append(resetPassReq[:index], resetPassReq[index+1:]...)
	resetPassReq = append(resetPassReq, ResetpassObj)
}
