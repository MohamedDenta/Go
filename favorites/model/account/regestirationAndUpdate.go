package account

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"log"
	"net"
	"net/http"
	"net/smtp"
	"time"

	"../accountdb"
	"../broadcastTcp"
	"../errorpk"
	"../globalPkg"
	"../logfunc"
	"../logpkg"
	"../responses"
	// "../validator"

	nexmo "gopkg.in/njern/gonexmo.v2"
)

var randomTable = [...]byte{'1', '2', '3', '4', '5', '6', '7', '8', '9', '0'}

//Tempregister temp struct to save requested register api in it
type Tempregister struct {
	IP       net.IP    //ip user
	Count    int       //count for number of request register
	Count2   int       //count2 for number of request register with fake email
	LastTime time.Time //store the time make request register
}

//Tamparr store for all request register
var Tamparr []Tempregister

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
		if userObj.Confirmationcode == string(buffer) {
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
		if user.Confirmationcode == userobj.Confirmationcode {
			found = true
			break
		}
	}
	if found == false {
		fmt.Println("wrong confirmation code")
	}
}

//sendSMS send SMS Using nexmo API
func sendSMS(PhoneNumber string, confirmationcode string) bool {

	nexmoClient, _ := nexmo.NewClient("53db0133", "iW59RoOYLrUBQ8yZ")

	// Test if it works by retrieving your account balance
	balance, err := nexmoClient.Account.GetBalance()
	log.Println(balance)
	message := &nexmo.SMSMessage{
		From: "go-nexmo",
		To:   PhoneNumber,
		Type: nexmo.Text,
		Text: "Wellcom at your Wallet,your verfy code is: " + confirmationcode,
	}

	messageResponse, err := nexmoClient.SMS.Send(message)
	if err != nil {
		return false
	}

	log.Println("messageResponse: ", messageResponse)
	/*if messageResponse == "[{ "+ Phone_Number+"     Non White-listed Destination - rejected}]"{

	return false
	}*/
	log.Println("ERRRRROR :", err)
	return true
}

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
func sendConfirmationEmailORSMS(userObj User) (string, User) {

	if userObj.Account.AccountEmail == "" {

		flag := sendSMS(userObj.Account.AccountPhoneNumber, userObj.Confirmationcode)
		if !flag {
			return "sms not send NO internet connection", userObj
		}

	} else {
		userObj.PathAPI = globalPkg.RandomPath()
		fmt.Println("//i/    ", userObj.PathAPI)
		body := "Dear " + userObj.Account.AccountName + `,
		Thank you for joining Inovatian&#39;s InoChain, your request has been processed and your wallet has been created successfully.
		Your confirmation code is: ` + userObj.Confirmationcode + `
		Please follow the following link to activate your wallet:
		(If this link is not clickable, please copy and paste into a new browser) 
		` + globalPkg.GlobalObj.Downloadfileip + "/" + userObj.PathAPI + "?confirmationcode=" + userObj.Confirmationcode +
			`
		This is a no-reply email. for any enquiries please contact info@inovatian.com
		If you did not create this wallet, please disregard this email.
		Regards,
		Inovatian Team`

		b := html.UnescapeString(body)
		sendEmail(b, userObj.Account.AccountEmail)

	}
	return "", userObj
}

//userStatus TO CHECK iF USER rEGISTER AND NOT CONFIRMED oR USER REQUEST TO UPDATE HIS aCCCOUNT AND NOT CONFIRMED YET
func userStatus(user User) (int, string) { //check if user found in userobj list
	var errorfound string
	errorfound = ""
	var index int
	index = -1
	for i, UserObj := range userobjlst {
		if UserObj.Account.AccountName == user.Account.AccountName && UserObj.Account.AccountEmail == user.Account.AccountEmail && UserObj.Account.AccountPhoneNumber == user.Account.AccountPhoneNumber && UserObj.Method == "POST" {
			responseObj := responses.FindResponseByID("153")
			errorfound = responseObj.EngResponse
			index = -2
			break
		}
		if UserObj.Account.AccountName == user.Account.AccountName && UserObj.Account.AccountEmail == user.Account.AccountEmail && UserObj.Account.AccountPhoneNumber == user.Account.AccountPhoneNumber && UserObj.Method == "PUT" {
			responseObj := responses.FindResponseByID("154")
			errorfound = responseObj.EngResponse
			index = i
			break
		}
		if UserObj.Account.AccountEmail == user.Account.AccountEmail && user.Account.AccountEmail != "" && UserObj.Method == "POST" {
			responseObj := responses.FindResponseByID("155")
			errorfound = responseObj.EngResponse
			index = -2
			break
		}

		if UserObj.Account.AccountPhoneNumber == user.Account.AccountPhoneNumber && user.Account.AccountPhoneNumber != "" && UserObj.Method == "POST" {
			responseObj := responses.FindResponseByID("156")
			errorfound = responseObj.EngResponse
			index = -2
			break
		}
		if UserObj.Account.AccountName == user.Account.AccountName && UserObj.Method == "POST" {
			responseObj := responses.FindResponseByID("157")
			errorfound = responseObj.EngResponse
			index = -2
			break
		}

		if UserObj.Account.AccountName == user.Account.AccountName && UserObj.Method == "PUT" {
			responseObj := responses.FindResponseByID("33")
			errorfound = responseObj.EngResponse
			index = i
			break
		}
		if UserObj.Account.AccountEmail == user.Account.AccountEmail && user.Account.AccountEmail != "" && UserObj.Method == "PUT" {
			responseObj := responses.FindResponseByID("31")
			errorfound = responseObj.EngResponse
			index = i
			break
		}
		if UserObj.Account.AccountPhoneNumber == user.Account.AccountPhoneNumber && user.Account.AccountPhoneNumber != "" && UserObj.Method == "PUT" {
			responseObj := responses.FindResponseByID("32")
			errorfound = responseObj.EngResponse
			index = i
			break
		}

	}
	return index, errorfound
}

//userValidation validate if User Enter Data valid then check if User exist before
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
// ServiceRegisterAPI service register 
func ServiceRegisterAPI(w http.ResponseWriter, req *http.Request) {

	now, userIP := globalPkg.SetLogObj(req)
	logStruct := logpkg.LogStruct{"_", now, userIP, "macAdress", "ServiceRegisterAPI", "Account", "", "", "_", 0}

	userObj := User{}
	userObj.Method = "POST"
	decoder := json.NewDecoder(req.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&userObj.Account)

	if err != nil {
		responseObj := responses.FindResponseByID("1")
		globalPkg.SendError(w, responseObj.EngResponse)
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
	sendJSON, _ := json.Marshal(accountStruct)
	globalPkg.SendResponse(w, sendJSON)

	globalPkg.WriteLog(logStruct, "service successfully registered"+"\n", "success")
}

// //addTemp add object temp request register in temp array
// func addTemp(tempObj Tempregister) int {
// 	Tamparr = append(Tamparr, tempObj) //append
// 	lastindex := len(Tamparr) - 1      //get index that object append on it
// 	return lastindex
// }

// //findIPTemp find ip on temp arr
// func findIPTemp(ip net.IP) int {
// 	for index, obj := range Tamparr {
// 		//compare ip with ip on temp
// 		if obj.IP.Equal(ip) {
// 			return index //return index of object
// 		}
// 	}
// 	return -1
// }

// //DeleteRequestRegister  go routine func to delete requested register temp after 20 minute 20 *60 =1200 second
// func DeleteRequestRegister() {
// 	for {
// 		time.Sleep(time.Second * 1200) //delete object request from temp after 20 minute
// 		for index, temprequest := range Tamparr {
// 			// if time.Now().UTC().Second() > temprequest.LastTime.Second() {
// 			// compare time now with last time in temp request register in seconds
// 			if globalPkg.UTCtime().Second() > temprequest.LastTime.Second() {
// 				Tamparr = append(Tamparr[:index], Tamparr[index+1:]...) //delete index from temp arr
// 			}
// 		}
// 	}
// }

// //GetTempRequestedRegister return values of temp array requested register
// func GetTempRequestedRegister() []Tempregister {
// 	return Tamparr
// }

//UserRegister End point create new Account
// func UserRegister(w http.ResponseWriter, req *http.Request) {
// 	now, userIP := globalPkg.SetLogObj(req)
// 	logStruct := logpkg.LogStruct{"", now, userIP, "macAdress", "UserRegister", "Account", "", "", "", 0}

// 	tempObj := Tempregister{}
// 	var id int //id to know the index of object append
// 	index := findIPTemp(userIP)

// 	if index == -1 {
// 		tempObj.IP = userIP
// 		tempObj.Count = tempObj.Count + 1
// 		tempObj.LastTime = now
// 		id = addTemp(tempObj) //id to know the index of object append
// 		fmt.Println("id", id)
// 	} else {
// 		if Tamparr[index].Count > 5 {
// 			Tamparr[index].Count = (Tamparr[index].Count) + 1
// 			Tamparr[index].LastTime = now.Add(time.Minute * time.Duration(5)) //add to time now + 5 minutes
// 			globalPkg.SendError(w, "you had reached the max number of registration  please wait 5 minutes ")
// 			globalPkg.WriteLog(logStruct, "you had reached the max number of registration  please wait 5 minute", "failed")
// 			return
// 		}
// 		Tamparr[index].Count = (Tamparr[index].Count) + 1
// 		Tamparr[index].LastTime = now.Add(time.Minute * time.Duration(5))
// 	}
// 	userObj := User{}
// 	RandomCode := encodeToString(globalPkg.GlobalObj.MaxConfirmcode)
// 	userObj.Confirmation_code = RandomCode
// 	userObj.Method = "POST"
// 	decoder := json.NewDecoder(req.Body)
// 	decoder.DisallowUnknownFields()
// 	err := decoder.Decode(&userObj.Account)
// 	if err != nil {
// 		responseObj := responses.FindResponseByID("1")
//globalPkg.SendError(w, responseObj.EngResponse)
// 		globalPkg.WriteLog(logStruct, "failed to decode Object", "failed")
// 		return
// 	}

// 	InputData := userObj.Account.AccountName + "," + userObj.Account.AccountEmail + "," + userObj.Account.AccountPhoneNumber + userObj.Account.AccountPassword
// 	logStruct.InputData = InputData

// 	//check username and email is lowercase
// 	userObj.Account.AccountEmail = convertStringTolowerCaseAndtrimspace(userObj.Account.AccountEmail)
// 	userObj.Account.AccountName = convertStringTolowerCaseAndtrimspace(userObj.Account.AccountName)
// 	//check if account exist or Any feild found before
// 	var Error string
// 	var accountStruct accountdb.AccountStruct
// 	Error = userValidation(userObj)
// 	if Error != "" {
// 		globalPkg.SendError(w, Error)
// 		globalPkg.WriteLog(logStruct, "error in validate user :"+Error+"\n", "failed")
// 		return
// 	}

// 	accountStruct = userObj.Account

// 	var str string       //string hold error
// 	var count2global int //count2 global hold value of count2 to compare count2 is greater than or equal 5
// 	//checkagain:
// 	result, errors := mailck.Check("noreply@mancke.net", accountStruct.AccountEmail)

// 	if errors != nil {
// 		str = errors.Error()                                                               //convert errors to string
// 		if str == "421 service not available (connection refused, too many connections)" { //if connection busy
// 			fmt.Println("connection refused !")
// 			//	goto checkagain

// 		}
// 	}
// 	switch {
// 	// email valid email already exist in mail server is real
// 	case result.IsValid():
// 		fmt.Println("the mailserver accepts mails for this mailbox")
// 	case result.IsError():
// 		fmt.Println("  ------   *  result.IsError()   ", result)
// 		fmt.Println("we can't say for sure if the address is valid or not")
// 		if result.Message == "The target mailserver responded with an error." {

// 			if index == -1 { //add count2 for first time
// 				Tamparr[id].Count2 = Tamparr[id].Count2 + 1
// 				count2global = Tamparr[id].Count2
// 			} else {
// 				Tamparr[id].Count2 = Tamparr[id].Count2 + 1
// 				count2global = Tamparr[id].Count2
// 			}

// 			if count2global >= 5 { //count2 greaterthan 5
// 				Tamparr[index].LastTime = now.Add(time.Minute * time.Duration(15)) //add time now after 15
// 				globalPkg.SendError(w, "Bad request email after you time")
// 				globalPkg.WriteLog(logStruct, "Bad request email after you time", "failed")
// 				return
// 			}
// 		}

// 	case result.IsInvalid():
// 		fmt.Println("Invalid email ------  ")

// 		if index == -1 { //add count2 for first time
// 			Tamparr[id].Count2 = Tamparr[id].Count2 + 1
// 			count2global = Tamparr[id].Count2
// 		} else {
// 			Tamparr[id].Count2 = Tamparr[id].Count2 + 1
// 			count2global = Tamparr[id].Count2
// 		}

// 		if count2global >= 5 { //count2 greaterthan 5
// 			Tamparr[index].LastTime = now.Add(time.Minute * time.Duration(15)) //add time now after 15
// 			globalPkg.SendError(w, "Bad request email after you time")
// 			globalPkg.WriteLog(logStruct, "Bad request email after you time", "failed")
// 			return
// 		}
// 		//reason for invalid email
// 		switch result {
// 		case mailck.InvalidDomain:
// 			fmt.Println("domain is invalid")
// 		case mailck.InvalidSyntax:
// 			fmt.Println("e-mail address syntax is invalid")
// 		}
// 	}

// 	accountStruct.AccountLastUpdatedTime = globalPkg.UTCtime()
// 	accountStruct.AccountInitialUserName = accountStruct.AccountName
// 	accountStruct.AccountInitialPassword = accountStruct.AccountPassword

// 	var current time.Time
// 	current = globalPkg.UTCtime()

// 	userObj.CurrentTime = current
// 	fmt.Println("registration :   ", userObj.CurrentTime)
// 	Error, userObj = sendConfirmationEmailORSMS(userObj)
// 	if Error != "" {
// 		globalPkg.SendError(w, "sms not send NO internet connection ")
// 		globalPkg.WriteLog(logStruct, "sms not send NO internet connection "+Error+"\n", "failed")
// 		return
// 	}

// 	broadcastTcp.BoardcastingTCP(userObj, "adduser", "account module")
// 	sendJson, _ := json.Marshal(accountStruct)
// 	globalPkg.SendResponse(w, sendJson)
// 	globalPkg.WriteLog(logStruct, "user successfully registered"+"\n", "success")
// }

//UserRegister End point create new Account
func UserRegister(w http.ResponseWriter, req *http.Request) {
	now, userIP := globalPkg.GetIP(req)
	logStruct := logpkg.LogStruct{"", now, userIP, "macAdress", "UserRegister", "Account", "", "", "", 0}

	userObj := User{}
	RandomCode := encodeToString(globalPkg.GlobalObj.MaxConfirmcode)
	userObj.Confirmationcode = RandomCode
	userObj.Method = "POST"
	decoder := json.NewDecoder(req.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&userObj.Account)
	if err != nil {
		responseObj := responses.FindResponseByID("1")
		globalPkg.SendError(w, responseObj.EngResponse)
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
		responseObj := responses.FindResponseByID("18")
		globalPkg.SendError(w, responseObj.EngResponse)
		globalPkg.WriteLog(logStruct, "sms not send NO internet connection "+Error+"\n", "failed")
		return
	}

	broadcastTcp.BoardcastingTCP(userObj, "adduser", "account module")
	sendJSON, _ := json.Marshal(accountStruct)
	globalPkg.SendResponse(w, sendJSON)
	globalPkg.WriteLog(logStruct, "user successfully registered"+"\n", "success")
}

//UpdateAccountInfo End Point this Api call by front End to make user to update his account info
func UpdateAccountInfo(w http.ResponseWriter, req *http.Request) {
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
		logobj = logpkg.LogStruct{Logindex, now, userIP, "macAdress", "UpdateUserInfo", "AccountModule", "", "", "_", 0}
	}
	logobj = logfunc.ReplaceLog(logobj, "UpdateUserInfo", "AccountModule")

	user := User{}
	user.Method = "PUT"
	RandomCode := encodeToString(globalPkg.GlobalObj.MaxConfirmcode)
	decoder := json.NewDecoder(req.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&user)

	if err != nil {
		responseObj := responses.FindResponseByID("1")
		globalPkg.SendError(w, responseObj.EngResponse)
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
	if accountObj.AccountPublicKey != user.Account.AccountPublicKey {
		responseObj := responses.FindResponseByID("10")
		globalPkg.SendError(w, responseObj.EngResponse)
		globalPkg.WriteLog(logobj, "Invalid Public key", "failed")
		logobj.OutputData = "Invalid Public key"
		logobj.Process = "failed"
		logobj.Count = logobj.Count + 1
		broadcastTcp.BoardcastingTCP(logobj, "", "AddAndUpdateLog")
		return
	}
	if accountObj.AccountPassword != user.Oldpassword {
		responseObj := responses.FindResponseByID("11")
		globalPkg.SendError(w, responseObj.EngResponse)
		globalPkg.WriteLog(logobj, "Invalid Pasword", "failed")
		logobj.OutputData = "Invalid password"
		logobj.Process = "failed"
		logobj.Count = logobj.Count + 1

		broadcastTcp.BoardcastingTCP(logobj, "", "AddAndUpdateLog")
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

	user.Confirmationcode = RandomCode
	current := globalPkg.UTCtime()
	user.CurrentTime = current

	if user.Account.AccountEmail == accountObj.AccountEmail && accountObj.AccountEmail != "" {
		accountObj.AccountName = accountStruct.AccountName
		accountObj.AccountPassword = accountStruct.AccountPassword
		accountObj.AccountPhoneNumber = accountStruct.AccountPhoneNumber
		accountObj.AccountAddress = accountStruct.AccountAddress

		broadcastTcp.BoardcastingTCP(accountObj, "PUT", "account")
		sendJSON, _ := json.Marshal(accountObj)
		globalPkg.SendResponse(w, sendJSON)
		globalPkg.WriteLog(logobj, accountObj.AccountName+"Update success", "success")

		if logobj.Count > 0 {
			logobj.Count = 0
			logobj.OutputData = accountObj.AccountName + "Update success"
			logobj.Process = "success"
			broadcastTcp.BoardcastingTCP(logobj, "", "AddAndUpdateLog")

		}

		return
	}

	ErrMessage, _ := sendConfirmationEmailORSMS(user)
	if ErrMessage != "" {
		responseObj := responses.FindResponseByID("18")
		globalPkg.SendError(w, responseObj.EngResponse)
		globalPkg.WriteLog(logobj, "sms not send NO internet connection", "failed")
		return
	}
	broadcastTcp.BoardcastingTCP(user, "adduser", "account module") ////Ass updated user in temp

	sendJSON, _ := json.Marshal(accountStruct)
	globalPkg.SendResponse(w, sendJSON)
	log.Printf("this is your data: %#v\n", user)
	globalPkg.WriteLog(logobj, "user", "success")
	if logobj.Count > 0 {
		logobj.Count = 0
		broadcastTcp.BoardcastingTCP(logobj, "", "AddAndUpdateLog")

	}
}
//ServiceUpdateAPI service update
func ServiceUpdateAPI(w http.ResponseWriter, req *http.Request) {
	now, userIP := globalPkg.GetIP(req)
	user := User{}
	user.Method = "PUT"
	logobj := logpkg.LogStruct{"_", now, userIP, "macAdress", "UpdateAccountInfo", "AccountModule", "", "", "_", 0}
	decoder := json.NewDecoder(req.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&user)

	if err != nil {
		responseObj := responses.FindResponseByID("1")
		globalPkg.SendError(w, responseObj.EngResponse)
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
		responseObj := responses.FindResponseByID("11")
		globalPkg.SendError(w, responseObj.EngResponse)
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

	sendJSON, _ := json.Marshal(tempAcc)
	globalPkg.SendResponse(w, sendJSON)
	globalPkg.WriteLog(logobj, accountObj.AccountName+"Update success", "success")
}
// BillingAccountRegisterAPI billing register 
func BillingAccountRegisterAPI(w http.ResponseWriter, req *http.Request) {
	now, userIP := globalPkg.GetIP(req)
	logStruct := logpkg.LogStruct{"", now, userIP, "macAdress", "ServiceRegisterAPI", "Account", "", "", "", 0}
	userObj := User{}
	userObj.Method = "POST"
	decoder := json.NewDecoder(req.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&userObj.Account)
	if err != nil {
		responseObj := responses.FindResponseByID("1")
		globalPkg.SendError(w, responseObj.EngResponse)
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
	userObj.Account.AccountRole = "billing"
	userObj.Account.AccountStatus = true
	userObj.Account.AccountInitialUserName = userObj.Account.AccountName
	userObj.Account.AccountInitialPassword = userObj.Account.AccountPassword
	userObj = createPublicAndPrivate(userObj)

	userObj.Account.AccountLastUpdatedTime = globalPkg.UTCtime()
	accountStruct = userObj.Account

	var current time.Time
	current = globalPkg.UTCtime()

	userObj.CurrentTime = current
	broadcastTcp.BoardcastingTCP(accountStruct, "POST", "account")
	sendJSON, _ := json.Marshal(accountStruct)
	globalPkg.SendResponse(w, sendJSON)
	globalPkg.WriteLog(logStruct, "account successfully registered"+"\n", "success")
	accountdb.OwnershipCreate(accountdb.AccountOwnershipStruct{AccountIndex: userObj.Account.AccountIndex, Owner: true}) // ownership
}
