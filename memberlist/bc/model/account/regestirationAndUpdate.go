package account

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"log"
	"net/http"
	"net/smtp"
	"time"

	"../accountdb"
	"../broadcastTcp"
	"../errorpk"
	"../globalPkg"
	"../logpkg"
	// nexmo "gopkg.in/njern/gonexmo.v2"
)

var randomTable = [...]byte{'1', '2', '3', '4', '5', '6', '7', '8', '9', '0'}

//Make random string code that i use it to Verify User
func encodeToString(max int) string {
	buffer := make([]byte, max)
	_, err := io.ReadAtLeast(rand.Reader, buffer, max)
	if err != nil {
		errorpk.AddError("account encodeToString", "the string is more than the max", "runtime error")
	}

code:
	for index := 0; index < len(buffer); index++ {
		buffer[index] = randomTable[int(buffer[index])%len(randomTable)]
	}
	for _, userObj := range userobjlst {
		if userObj.Confirmation_code == string(buffer) {
			goto code
		}
	}
	return string(buffer)
}

//UpdateconfirmAtribute func to check if user first time loginthen update objList Array
func UpdateconfirmAtribute(userobj User) {
	var found bool
	var user User
	for _, user = range userobjlst {
		if user.Confirmation_code == userobj.Confirmation_code {
			found = true
			break
		}
	}
	if found == true {
	}
}

//send_SMS send SMS Using nexmo API
// func send_SMS(Phone_Number string, confirmation_code string) bool {

// 	nexmoClient, _ := nexmo.NewClient("53db0133", "iW59RoOYLrUBQ8yZ")

// 	// Test if it works by retrieving your account balance
// 	balance, err := nexmoClient.Account.GetBalance()
// 	log.Println(balance)
// 	message := &nexmo.SMSMessage{
// 		From: "go-nexmo",
// 		To:   Phone_Number,
// 		Type: nexmo.Text,
// 		Text: "Wellcom at your Wallet,your verfy code is: " + confirmation_code,
// 	}

// 	messageResponse, err := nexmoClient.SMS.Send(message)
// 	if err != nil {
// 		return false
// 	}

// 	log.Println("messageResponse: ", messageResponse)
// 	/*if messageResponse == "[{ "+ Phone_Number+"     Non White-listed Destination - rejected}]"{

// 	return false
// 	}*/
// 	log.Println("ERRRRROR :", err)
// 	return true
// }

//sendEmail send confirmation Email using Stmp
func sendEmail(Body string, Email string) {
	//mime := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n";
	from := "noreply@inovatian.com" ///// "inovatian.tech@gmail.com"
	pass := "ino13579$"             /////your passward   ////

	to := Email //Email of User

	msg := "From: " + from + "\n" +
		"To: " + to + "\n" +
		"Subject: Inovatian Digital Wallet Verification\n\n" + Body

	///confirmation link

	err := smtp.SendMail("mail.inovatian.com:26",
		smtp.PlainAuth("", from, pass, "mail.inovatian.com"),
		from, []string{to}, []byte(msg))

	if err != nil {
		log.Printf("smtp error: %s", err)
		return
	}

	log.Println("sent, visit", Email)
}

//sendConfirmationEmailORSMS send mail or SMS
/**send mail or SMS**/
func sendConfirmationEmailORSMS(userObj User) (string, User) {

	if userObj.Account.AccountEmail == "" {

		// flag := send_SMS(userObj.Account.AccountPhoneNumber, userObj.Confirmation_code)
		// if !flag {

		// 	//
		// 	return "sms not send NO internet connection", userObj
		// }

	} else {
		userObj.PathApi = globalPkg.RandomPath()
		fmt.Println("//i/    ", userObj.PathApi)
		body := "Dear " + userObj.Account.AccountName + `,
Thank you for joining Inovatian&#39;s InoChain, your request has been processed and your wallet has been created successfully.
Your confirmation code is: ` + userObj.Confirmation_code + `
Please follow the following link to activate your wallet:
(If this link is not clickable, please copy and paste into a new browser) 
` +
			globalPkg.GlobalObj.Downloadfileip + "/" + userObj.PathApi + "?confirmationcode=" + userObj.Confirmation_code +
			`
This is a no-reply email; for any enquiries please contact info@inovatian.com
If you did not create this wallet, please disregard this email.
Regards,
Inovatian Team`
		b := html.UnescapeString(body)
		sendEmail(b, userObj.Account.AccountEmail)

	}
	return "", userObj
}

//sendConfirmationEmailORSMS send mail or SMS
/**send mail or SMS**/
func sendConfirmationEmailORSMS2(userObj User) (string, User) {

	if userObj.Account.AccountEmail == "" {

		// flag := send_SMS(userObj.Account.AccountPhoneNumber, userObj.Confirmation_code)
		// if !flag {

		// 	//
		// 	return "sms not send NO internet connection", userObj
		// }

	} else {
		userObj.PathApi = globalPkg.RandomPath()
		fmt.Println("//i/    ", userObj.PathApi)
		body := "Dear " + userObj.Account.AccountName + `,
Thank you for joining Inovatian&#39;s InoChain, your request has been processed and your wallet has been created successfully.
Your confirmation code is: ` + userObj.Confirmation_code + `
This is a no-reply email; for any enquiries please contact info@inovatian.com
If you did not create this wallet, please disregard this email.
Regards,
Inovatian Team`
		b := html.UnescapeString(body)
		sendEmail(b, userObj.Account.AccountEmail)

	}
	return "", userObj
}

//TO CHECK iF USER rEGISTER AND NOT CONFIRMED oR USER REQUEST TO UPDATE HIS aCCCOUNT AND NOT CONFIRMED YET
func userStatus(user User) (int, string) { //check if user found in userobj list
	var errorfound string
	errorfound = ""
	var index int
	index = -1
	for i, UserObj := range userobjlst {
		if UserObj.Account.AccountName == user.Account.AccountName && UserObj.Account.AccountEmail == user.Account.AccountEmail && UserObj.Account.AccountPhoneNumber == user.Account.AccountPhoneNumber && UserObj.Method == "POST" {
			errorfound = "this user  registered and not confirmed"
			index = -2
			break
		}
		if UserObj.Account.AccountName == user.Account.AccountName && UserObj.Account.AccountEmail == user.Account.AccountEmail && UserObj.Account.AccountPhoneNumber == user.Account.AccountPhoneNumber && UserObj.Method == "PUT" {
			errorfound = "user  Found"
			index = i
			break
		}
		if UserObj.Account.AccountEmail == user.Account.AccountEmail && user.Account.AccountEmail != "" && UserObj.Method == "POST" {
			errorfound = "this email registered and not confirmed"
			index = -2
			break
		}

		if UserObj.Account.AccountPhoneNumber == user.Account.AccountPhoneNumber && user.Account.AccountPhoneNumber != "" && UserObj.Method == "POST" {
			errorfound = "this phon registered and not confirmed"
			index = -2
			break
		}
		if UserObj.Account.AccountName == user.Account.AccountName && UserObj.Method == "POST" {
			errorfound = "this userName  registered and not confirmed"
			index = -2
			break
		}

		if UserObj.Account.AccountName == user.Account.AccountName && UserObj.Method == "PUT" {
			errorfound = "UserName Found"
			index = i
			break
		}
		if UserObj.Account.AccountEmail == user.Account.AccountEmail && user.Account.AccountEmail != "" && UserObj.Method == "PUT" {
			errorfound = "Email found"
			index = i
			break
		}
		if UserObj.Account.AccountPhoneNumber == user.Account.AccountPhoneNumber && user.Account.AccountPhoneNumber != "" && UserObj.Method == "PUT" {
			errorfound = "Phone Found "
			index = i
			break
		}

	}
	return index, errorfound
}

//validate if User Enter Data valid then check if User exist before
func userValidation(userObj User) string {
	accountStruct := userObj.Account
	var MessageErr string
	if userObj.Method == "POST" {
		MessageErr = checkingIfAccountExixtsBeforeRegister(accountStruct)
	}
	if userObj.Method == "PUT" {
		MessageErr = checkingIfAccountExixtsBeforeUpdating(accountStruct)

	}

	if MessageErr != "" {
		return MessageErr
	}
	_, found := userStatus(userObj)
	if found != "" {
		return found
	}
	return ""
}

func ServiceRegisterAPI(w http.ResponseWriter, req *http.Request) {

	now, userIP := globalPkg.SetLogObj(req)
	logStruct := logpkg.LogStruct{"_", now, userIP, "macAdress", "ServiceRegisterAPI", "Account", "", "", "_", 0}

	userObj := User{}
	userObj.Method = "POST"
	decoder := json.NewDecoder(req.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&userObj.Account)

	if err != nil {
		globalPkg.SendError(w, "please enter your correct request ")
		globalPkg.WriteLog(logStruct, "failed to decode Object", "failed")
		return
	}
	InputData := userObj.Account.AccountName + "," + userObj.Account.AccountEmail + "," + userObj.Account.AccountPhoneNumber + userObj.Account.AccountPassword
	logStruct.InputData = InputData

	//check username and email is lowercase
	userObj.Account.AccountEmail = convertStringTolowerCaseAndtrimspace(userObj.Account.AccountEmail)
	userObj.Account.AccountName = convertStringTolowerCaseAndtrimspace(userObj.Account.AccountName)
	//check if account exist or Any field found before
	var Error string
	var accountStruct accountdb.AccountStruct
	Error = userValidation(userObj)
	if Error != "" {
		globalPkg.SendError(w, Error)
		globalPkg.WriteLog(logStruct, "error in validate user :"+Error+"\n", "failed")
		return
	}

	userObj.Account.AccountRole = "service"
	// userObj.Account.AccountTokenID, _ = globalPkg.ConvertIntToFixedLengthString(2, globalPkg.GlobalObj.TokenIDStringFixedLength) // second token id
	AccountTokenid, _ := globalPkg.ConvertIntToFixedLengthString(2, globalPkg.GlobalObj.TokenIDStringFixedLength) // second token id
	userObj.Account.AccountTokenID = append(userObj.Account.AccountTokenID, AccountTokenid)
	userObj.Account.AccountInitialUserName = userObj.Account.AccountName
	userObj.Account.AccountInitialPassword = userObj.Account.AccountPassword
	userObj = createPublicAndPrivate(userObj)

	userObj.Account.AccountLastUpdatedTime = globalPkg.UTCtime()
	accountStruct = userObj.Account

	var current time.Time
	current = globalPkg.UTCtime()

	userObj.CurrentTime = current
	broadcastTcp.BoardcastingTCP(accountStruct, "POST", "account")
	sendJson, _ := json.Marshal(accountStruct)
	globalPkg.SendResponse(w, sendJson)

	globalPkg.WriteLog(logStruct, "service successfully registered"+"\n", "success")
}

//UserRegister End point create new Account
func UserRegister(w http.ResponseWriter, req *http.Request) {
    if req.Method == "POST" {
	now, userIP := globalPkg.SetLogObj(req)
	logStruct := logpkg.LogStruct{"_", now, userIP, "macAdress", "UserRegister", "Account", "", "", "_", 0}

	userObj := User{}
	RandomCode := encodeToString(globalPkg.GlobalObj.MaxConfirmcode)
	userObj.Confirmation_code = RandomCode
	userObj.Method = "POST"
	decoder := json.NewDecoder(req.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&userObj.Account)
	if err != nil {
		globalPkg.SendError(w, "please enter your correct request ")
		globalPkg.WriteLog(logStruct, "failed to decode Object", "failed")
		return
	}

	InputData := userObj.Account.AccountName + "," + userObj.Account.AccountEmail + "," + userObj.Account.AccountPhoneNumber + userObj.Account.AccountPassword
	logStruct.InputData = InputData

	//check username and email is lowercase
	userObj.Account.AccountEmail = convertStringTolowerCaseAndtrimspace(userObj.Account.AccountEmail)
	userObj.Account.AccountName = convertStringTolowerCaseAndtrimspace(userObj.Account.AccountName)
	//check if account exist or Any feild found before
	var Error string
	var accountStruct accountdb.AccountStruct
	Error = userValidation(userObj)
	if Error != "" {
		globalPkg.SendError(w, Error)
		globalPkg.WriteLog(logStruct, "error in validate user :"+Error+"\n", "failed")
		return
	}
	accountStruct = userObj.Account
	accountStruct.AccountLastUpdatedTime = globalPkg.UTCtime()
	accountStruct.AccountInitialUserName = accountStruct.AccountName
	accountStruct.AccountInitialPassword = accountStruct.AccountPassword

	var current time.Time
	current = globalPkg.UTCtime()

	userObj.CurrentTime = current
	fmt.Println("registration :   ", userObj.CurrentTime)
	Error, userObj = sendConfirmationEmailORSMS(userObj)
	if Error != "" {
		globalPkg.SendError(w, "sms not send NO internet connection ")
		globalPkg.WriteLog(logStruct, "sms not send NO internet connection "+Error+"\n", "failed")
		return
	}

	broadcastTcp.BoardcastingTCP(userObj, "adduser", "account module")
	sendJson, _ := json.Marshal(accountStruct)
	globalPkg.SendResponse(w, sendJson)
	globalPkg.WriteLog(logStruct, "user successfully registered"+"\n", "success")
    }
}

//UpdateAccountInfo End Point this Api call by front End to make user to update his account info
func UpdateAccountInfo(w http.ResponseWriter, req *http.Request) {
	now, userIP := globalPkg.SetLogObj(req)
	logobj := logpkg.LogStruct{"", now, userIP, "macAdress", "UpdateUserInfo", "AccountModule", "", "", "_", 0}
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

	// 	logobj = logpkg.LogStruct{Logindex, now, userIP, "macAdress", "UpdateUserInfo", "AccountModule", "", "", "_", 0}
	// }
	// logobj = logfunc.ReplaceLog(logobj, "UpdateUserInfo", "AccountModule")

	user := User{}
	user.Method = "PUT"
	RandomCode := encodeToString(globalPkg.GlobalObj.MaxConfirmcode)
	decoder := json.NewDecoder(req.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&user)

	if err != nil {
		globalPkg.SendError(w, "please enter your correct request ")
		globalPkg.WriteLog(logobj, "faild to decode object", "failed")
		return
	}

	InputData := user.Account.AccountName + "," + user.Account.AccountEmail + "," + user.Account.AccountPhoneNumber + user.Account.AccountPassword
	logobj.InputData = InputData

	//approve username & email is lowercase and trim
	user.Account.AccountEmail = convertStringTolowerCaseAndtrimspace(user.Account.AccountEmail)
	user.Account.AccountName = convertStringTolowerCaseAndtrimspace(user.Account.AccountName)

	var accountObj accountdb.AccountStruct
	accountObj = accountdb.FindAccountByAccountPublicKey(user.Account.AccountPublicKey)
	if accountObj.AccountPassword != user.Oldpassword {
		globalPkg.SendError(w, "Invalid Pasword")
		globalPkg.WriteLog(logobj, "Invalid Pasword", "failed")

		logobj.OutputData = "Invalid password"
		logobj.Process = "failed"
		// logobj.Count = logobj.Count + 1

		// broadcastTcp.BoardcastingTCP(logobj, "", "AddAndUpdateLog")
		return
	}

	//check if account exist
	var accountStruct accountdb.AccountStruct
	accountStruct = user.Account
	MessageErr := checkingIfAccountExixtsBeforeUpdating(accountStruct)
	if MessageErr != "" {
		globalPkg.SendNotFound(w, MessageErr)
		globalPkg.WriteLog(logobj, MessageErr, "failed")
		logobj.OutputData = MessageErr
		logobj.Process = "failed"
		logobj.Count = logobj.Count + 1

		return
	}

	index, ErrorFound := userStatus(user)

	if index == -2 {
		globalPkg.SendError(w, ErrorFound)
		globalPkg.WriteLog(logobj, ErrorFound, "failed")
		logobj.OutputData = ErrorFound
		logobj.Process = "failed"
		logobj.Count = logobj.Count + 1

		return
	}
	if index != -1 {
		RemoveUserFromtemp(index) ///remove old Request
	}

	user.Confirmation_code = RandomCode
	current := globalPkg.UTCtime()
	user.CurrentTime = current

	if user.Account.AccountEmail == accountObj.AccountEmail && accountObj.AccountEmail != "" {
		accountObj.AccountName = accountStruct.AccountName
		accountObj.AccountPassword = accountStruct.AccountPassword
		accountObj.AccountPhoneNumber = accountStruct.AccountPhoneNumber
		accountObj.AccountAddress = accountStruct.AccountAddress

		broadcastTcp.BoardcastingTCP(accountObj, "PUT", "account")
		sendJson, _ := json.Marshal(accountObj)
		globalPkg.SendResponse(w, sendJson)
		globalPkg.WriteLog(logobj, accountObj.AccountName+"Update success", "success")

		// if logobj.Count > 0 {
		// 	logobj.Count = 0
		// 	logobj.OutputData = accountObj.AccountName + "Update success"
		// 	logobj.Process = "success"
		// 	broadcastTcp.BoardcastingTCP(logobj, "", "AddAndUpdateLog")

		// }

		return
	}

	ErrMessage, _ := sendConfirmationEmailORSMS2(user)
	if ErrMessage != "" {
		globalPkg.SendError(w, "sms not send NO internet connection ")
		globalPkg.WriteLog(logobj, "sms not send NO internet connection", "failed")
		//logobj.Count = logobj.Count + 1

		//broadcastTcp.BoardcastingTCP(logobj, "", "AddAndUpdateLog")
		return
	}
	broadcastTcp.BoardcastingTCP(user, "adduser", "account module") ////Ass updated user in temp

	sendJson, _ := json.Marshal(accountStruct)
	globalPkg.SendResponse(w, sendJson)
	log.Printf("this is your data: %#v\n", user)
	globalPkg.WriteLog(logobj, "user", "success")
	// if logobj.Count > 0 {
	// 	logobj.Count = 0
	// 	broadcastTcp.BoardcastingTCP(logobj, "", "AddAndUpdateLog")

	// }
}

func ServiceUpdateAPI(w http.ResponseWriter, req *http.Request) {
	now, userIP := globalPkg.SetLogObj(req)
	user := User{}
	user.Method = "PUT"
	logobj := logpkg.LogStruct{"_", now, userIP, "macAdress", "UpdateAccountInfo", "AccountModule", "", "", "_", 0}
	decoder := json.NewDecoder(req.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&user)

	if err != nil {
		globalPkg.SendError(w, "please enter your correct request ")
		globalPkg.WriteLog(logobj, "Failed to Decode Object", "failed")
		return
	}
	InputData := user.Account.AccountName + "," + user.Account.AccountEmail + "," + user.Account.AccountPhoneNumber + user.Account.AccountPassword
	logobj.InputData = InputData

	//approve username & email is lowercase and trim
	user.Account.AccountEmail = convertStringTolowerCaseAndtrimspace(user.Account.AccountEmail)
	user.Account.AccountName = convertStringTolowerCaseAndtrimspace(user.Account.AccountName)

	var accountObj accountdb.AccountStruct
	accountObj = accountdb.FindAccountByAccountPublicKey(user.Account.AccountPublicKey)
	if accountObj.AccountPassword != user.Oldpassword {
		globalPkg.SendError(w, "Invalid Password")
		globalPkg.WriteLog(logobj, "invalid password", "failed")
		return
	}
	//check if account exist
	var accountStruct accountdb.AccountStruct
	accountStruct = user.Account
	current := globalPkg.UTCtime()
	user.CurrentTime = current

	accountObj.AccountName = accountStruct.AccountName
	accountObj.AccountPassword = accountStruct.AccountPassword
	accountObj.AccountPhoneNumber = accountStruct.AccountPhoneNumber
	accountObj.AccountAddress = accountStruct.AccountAddress
	accountObj.AccountEmail = accountStruct.AccountEmail
	accountObj.AccountRole = "service"
	broadcastTcp.BoardcastingTCP(accountObj, "PUT", "account")
	tempAcc := accountObj
	tempAcc.AccountInitialUserName = ""
	tempAcc.AccountInitialPassword = ""
	tempAcc.AccountRole = ""
	tempAcc.AccountLastUpdatedTime = time.Time{}
	tempAcc.AccountBalance = ""
	tempAcc.BlocksLst = nil
	tempAcc.SessionID = ""

	sendJson, _ := json.Marshal(tempAcc)
	globalPkg.SendResponse(w, sendJson)
	globalPkg.WriteLog(logobj, accountObj.AccountName+"Update success", "success")
}
