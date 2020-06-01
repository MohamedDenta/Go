package account

import (
	"crypto/rand"
	"encoding/json"
	"net/http"
	"rsapk"
	"time"

	"../accountdb"
	"../broadcastTcp"
	"../cryptogrpghy"
	"../globalPkg"
	"../logpkg"
	"../responses"
	"../validator"
)

type stringmessage struct {
	Message string
}

//cryptoData  pk , private
type cryptoData struct {
	PublicKey  string
	PrivateKey string
}

//-TO MAKE PROCESS Hapened in both confirmation Api check if confirmation code valid
//if invald the func return please check your varification code
//if User.Method ==post this mean that User confirm to register
//then we create public and private key and pass User to func AddAccount
//if User.Method ==put thise mean that User confirm to update his info
//then we pass the User account to UpdateAccount func
func confirmationProcess(userobject User, confirmationcode string, now time.Time) (string, User) {
	var flag bool
	flag = false
	for _, User := range userobjlst {
		if User.Confirmationcode == confirmationcode {
			userobject = User
			flag = true
			break
		}
	}

	if flag != true {
		responseObj := responses.FindResponseByID("14")
		return responseObj.EngResponse,userobject
	}

	if userobject.Method == "POST" {
		var accountobj accountdb.AccountStruct
		//create public and private key
		userobject.Account.AccountInitialUserName = userobject.Account.AccountName
		userobject.Account.AccountInitialPassword = userobject.Account.AccountPassword
		userobject.Account.AccountStatus = true
		userobject.Account.AccountIndex = NewIndex()
		userobject = createPublicAndPrivate(userobject)
		accountobj = userobject.Account
		accountobj.AccountPrivateKey = ""
		accountobj.AccountLastUpdatedTime = globalPkg.UTCtime()
		broadcastTcp.BoardcastingTCP(accountobj, "POST", "account")
		accountobj.AccountPrivateKey = userobject.Account.AccountPrivateKey
	}

	if userobject.Method == "PUT" {
		var accountobj accountdb.AccountStruct
		accountobj = accountdb.FindAccountByAccountPublicKey(userobject.Account.AccountPublicKey)
		if accountobj.AccountPublicKey == "" {
			responseObj := responses.FindResponseByID("10")
			return responseObj.EngResponse,userobject
		}
		updatedaccountobj := userobject.Account

		accountobj.AccountName = updatedaccountobj.AccountName
		accountobj.AccountPassword = updatedaccountobj.AccountPassword
		accountobj.AccountPhoneNumber = updatedaccountobj.AccountPhoneNumber
		accountobj.AccountAddress = updatedaccountobj.AccountAddress
		accountobj.AccountEmail = updatedaccountobj.AccountEmail
		accountobj.AccountPublicKey = updatedaccountobj.AccountPublicKey
		accountobj.AccountLastUpdatedTime = globalPkg.UTCtime()
		broadcastTcp.BoardcastingTCP(accountobj, "PUT", "account")

	}
	return "", userobject
}

//ConfirmatinByEmail calls when user press on confirmation link delivered using Email
//if Data Valid then redirect user to login page
func ConfirmatinByEmail(w http.ResponseWriter, req *http.Request) {
	now, userIP := globalPkg.SetLogObj(req)
	logobj := logpkg.LogStruct{"_", now, userIP, "macAdress", "ConfirmationByEmail", "AccountModule", "_", "_", "_", 0}

	var userObj User
	// check on path url
	existurl := false
	for _, userj := range userobjlst {
		p := "/" + userj.PathAPI
		// fmt.Println("   --- path --    ", req.URL.Path, "   ====    ", p)
		if req.URL.Path == p {
			existurl = true
			break
		}
	}

	if existurl == false {
		responseObj := responses.FindResponseByID("12")
		globalPkg.SendError(w, responseObj.EngResponse)
		logobj.OutputData = "this page not found"
		logobj.Process = "faild"
		logpkg.WriteOnlogFile(logobj)
		return
	}
	keys, ok := req.URL.Query()["confirmationcode"] // values.Get("confirmationcode") //return parameter from url

	if !ok || len(keys) == 0 {
		responseObj := responses.FindResponseByID("1")
		globalPkg.SendNotFound(w, responseObj.EngResponse)
		globalPkg.WriteLog(logobj, "please check your parameters", "failed")
		return
	}

	var flag bool
	for _, User := range userobjlst {
		if existurl == true {
			if User.Confirmationcode == keys[0] {

				userObj = User
				flag = true
				break
			}
		}
	}

	if flag != true {
		responseObj := responses.FindResponseByID("14")
		globalPkg.SendError(w, responseObj.EngResponse)
		globalPkg.WriteLog(logobj, "please,check Your verification Code", "failed")
		return
	}
	if now.Sub(userObj.CurrentTime).Seconds() > globalPkg.GlobalObj.DeleteAccountTimeInseacond {
		responseObj := responses.FindResponseByID("8")
		globalPkg.SendError(w, responseObj.EngResponse)
		globalPkg.WriteLog(logobj, "timeout", "failed")
		return
	}
	//userobjlst[index].Confirmed = true
	if userObj.Method == "POST" {
		userObj.Account.AccountInitialUserName = userObj.Account.AccountName
		userObj.Account.AccountInitialPassword = userObj.Account.AccountPassword
		userObj.Account.AccountIndex = NewIndex()
		userObj.Account.AccountStatus = true
		//userObj.Account.Confirmed = false
		userObj.Account.AccountLastUpdatedTime = globalPkg.UTCtime()
		broadcastTcp.BoardcastingTCP(userObj.Account, "POST", "account")
	}

	if userObj.Method == "PUT" {
		Message, _ := confirmationProcess(userObj, keys[0], userObj.CurrentTime)
		if Message != "" {
			http.Redirect(w, req, validator.DigitalWalletIpObj.DigitalwalletIp+":"+validator.DigitalWalletIpObj.Digitalwalletport+"/auth/register", http.StatusSeeOther)
			globalPkg.WriteLog(logobj, Message, "failed")
			return
		}
	}
	http.Redirect(w, req, validator.DigitalWalletIpObj.DigitalwalletIp+":"+validator.DigitalWalletIpObj.Digitalwalletport+"/auth/login", http.StatusSeeOther)
	globalPkg.WriteLog(logobj, "redirect user to login page", "success")
}

//ConfirmationMessage called to confim new user and save save new user in database
func ConfirmationMessage(w http.ResponseWriter, req *http.Request) {
	now, userIP := globalPkg.SetLogObj(req)
	logobj := logpkg.LogStruct{"_", now, userIP, "macAdress", "ConfirmationMessage", "AccountModule", "", "", "_", 0}
	confmationCode := stringmessage{}
	decoder := json.NewDecoder(req.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&confmationCode)
	if err != nil {
		responseObj := responses.FindResponseByID("1")
		globalPkg.SendError(w, responseObj.EngResponse)
		globalPkg.WriteLog(logobj, "failed to decode object", "failed")
		return
	}

	logobj.InputData = confmationCode.Message

	user := User{}

	var message string
	message, user = confirmationProcess(user, confmationCode.Message, now)

	if message != "" {
		globalPkg.SendError(w, message)
		globalPkg.WriteLog(logobj, message, "failed")
		return
	}
	if user.Method == "POST" {

		user.Account.AccountInitialUserName = user.Account.AccountName
		user.Account.AccountInitialPassword = user.Account.AccountPassword
		user.Account.AccountStatus = true
		user.Account.AccountStatus = false
		//user.Account.Confirmed = false
		user.Account.AccountLastUpdatedTime = globalPkg.UTCtime()
		broadcastTcp.BoardcastingTCP(user.Account, "POST", "account")
		cryptoDataObj := cryptoData{}

		cryptoDataObj.PublicKey = user.Account.AccountPublicKey
		cryptoDataObj.PrivateKey = user.Account.AccountPrivateKey

		sendJSON, _ := json.Marshal(cryptoDataObj)

		globalPkg.SendResponse(w, sendJSON)
		globalPkg.WriteLog(logobj, "Congratulations", "success")
		return
	}
	if user.Method == "PUT" {
		responseObj := responses.FindResponseByID("15")
		globalPkg.SendResponseMessage(w, responseObj.EngResponse)
		globalPkg.WriteLog(logobj, "Congratulations", "success")
	}

}

//createPublicAndPrivate create pk , private
func createPublicAndPrivate(userobj User) User {
	userobj.Account.AccountIndex = NewIndex()
	bitSize := globalPkg.GlobalObj.RSAKeyBitSize
	reader := rand.Reader

	// save pk and address in db
	var savePKObj SavePKStruct
	savePKObj.Index = userobj.Account.AccountIndex
	//infinite loop to get unique public key
	for {

		key, err := rsapk.GenerateKey(reader, bitSize)
		cryptogrpghy.CheckError(err)
		pk := cryptogrpghy.GetPublicPEMKey(key.PublicKey)
		pk2 := []byte(pk)
		add := cryptogrpghy.Address(pk2)

		// save pk, address
		savePKObj.Publickey = pk
		savePKObj.Address = string(add)
		savePKObj.CurrentTime = globalPkg.UTCtime()
		var accountobj accountdb.AccountStruct
		accountobj = GetAccountByAccountPubicKey(savePKObj.Address)
		if accountobj.AccountPublicKey == "" { //not found this public key in account DB

			userobj.Account.AccountPublicKey = string(add)
			userobj.Account.AccountPrivateKey = cryptogrpghy.GetPrivatePEMKey(key)

			broadcastTcp.BoardcastingTCP(savePKObj, "", "savepk")
			// break
			goto outsidefor
		}
	}
outsidefor:
	return userobj
}
