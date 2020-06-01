package adminModule

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"
	"unicode/utf8"

	admin "../admin"
	"../block"
	"../broadcastTcp"
	globalPkg "../globalPkg"
	"../logfunc"
	logpkg "../logpkg"
	"../responses"
	validator "../validator"
)
// txdata data represent transaction number and time
type txdata struct {
	Time              string
	TransactionNumber string
}

// GetAllAdminsAPI get all Admins info
func GetAllAdminsAPI(w http.ResponseWriter, req *http.Request) {
	// write log struct
	now, userIP := globalPkg.SetLogObj(req)
	logobj := logpkg.LogStruct{"_", now, userIP, "macAdress", "GetAllAdminsAPI", "adminModule", "_", "_", "_", 0}

	AdminObj := admin.AdminStruct{}
	decoder := json.NewDecoder(req.Body)

	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&AdminObj); err != nil {
		responseObj := responses.FindResponseByID("1")
		globalPkg.SendError(w, responseObj.EngResponse)
		globalPkg.WriteLog(logobj, responseObj.EngResponse, "failed")
		return
	}
	
	if AdminObj.AdminUsername == "" || AdminObj.AdminPassword == "" {
		responseObj := responses.FindResponseByID("11")
		globalPkg.SendError(w, responseObj.EngResponse)
		globalPkg.WriteLog(logobj, responseObj.EngResponse, "failed")
		return
	}
	Adminexist := admin.GetAdminsByUsername(AdminObj.AdminUsername)
	if AdminObj.AdminUsername == Adminexist.AdminUsername && AdminObj.AdminPassword == Adminexist.AdminPassword {
		lst := admin.GetAllAdmins()
		for index := range lst {
			lst[index].AdminPassword = ""
			lst[index].SuperAdminPassword = ""
		}
		sendJSON, _ := json.Marshal(lst)
		globalPkg.SendResponse(w, sendJSON)
		globalPkg.WriteLog(logobj, "list of all Admin ", "success")

	} else {
		responseObj := responses.FindResponseByID("11")
		globalPkg.SendError(w, responseObj.EngResponse)
		globalPkg.WriteLog(logobj, responseObj.EngResponse, "failed")

	}

}

//AddNewAdmin register for new admin
func AddNewAdmin(w http.ResponseWriter, req *http.Request) {
	now, userIP := globalPkg.SetLogObj(req)
	logobj := logpkg.LogStruct{"_", now, userIP, "macAdress", "AddNewAdmin", "adminModule", "_", "_", "_", 0}

	AdminObj := admin.AdminStruct{}

	decoder := json.NewDecoder(req.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&AdminObj); err != nil {
		responseObj := responses.FindResponseByID("1")
		globalPkg.SendError(w, responseObj.EngResponse)
		globalPkg.WriteLog(logobj, responseObj.EngResponse, "failed")
		return
	}
	logobj.InputData = AdminObj.AdminUsername
	if AdminObj.SuperAdminUsername == "" || AdminObj.SuperAdminPassword == "" {
		responseObj := responses.FindResponseByID("30")
		globalPkg.SendError(w, responseObj.EngResponse)
		globalPkg.WriteLog(logobj, responseObj.EngResponse, "failed")
		return
	}
	if utf8.RuneCountInString(AdminObj.AdminPassword) != 64 {
		responseObj := responses.FindResponseByID("11")
		globalPkg.SendError(w, responseObj.EngResponse)
		globalPkg.WriteLog(logobj, responseObj.EngResponse, "failed")
		return
	}
	if !admin.ValidationAdmin(admin.Admin{AdminObj.SuperAdminUsername, AdminObj.SuperAdminPassword, ""}) {
		responseObj := responses.FindResponseByID("2")
		globalPkg.SendError(w, responseObj.EngResponse)
		globalPkg.WriteLog(logobj, responseObj.EngResponse, "failed")
		return
	}

	if admin.CheckAdminExistsBefore(AdminObj.AdminUsername) {
		responseObj := responses.FindResponseByID("33")
		globalPkg.SendError(w, responseObj.EngResponse)
		globalPkg.WriteLog(logobj, responseObj.EngResponse, "failed")
		return
	}
	//check for email , phone is exist before
	errorfound := AdminObj.DataFound()
	if errorfound != "" {
		globalPkg.SendError(w, errorfound)
		globalPkg.WriteLog(logobj, errorfound, "failed")
		return
	}

	for _, adminValidator := range AdminObj.Validatorlst {
		exist := false
		for _, validatorobj := range validator.ValidatorsLstObj {
			if validatorobj.ValidatorIP == adminValidator {
				exist = true
				break
			}
		}
		if exist == false {
			responseObj := responses.FindResponseByID("34")
			globalPkg.SendError(w, responseObj.EngResponse)
			globalPkg.WriteLog(logobj, responseObj.EngResponse, "failed")
			return
		}
	}

	if AdminObj.AdminStartDate.Before(time.Now().UTC()) || AdminObj.AdminEndDate.Before(AdminObj.AdminStartDate) {
		responseObj := responses.FindResponseByID("35")
		globalPkg.SendError(w, responseObj.EngResponse)
		globalPkg.WriteLog(logobj, responseObj.EngResponse, "failed")
		return
	}
	AdminObj.AdminStartDate = globalPkg.UTCtimefield(AdminObj.AdminStartDate) //globalPkg.UTCtime(AdminObj.AdminStartDate)
	AdminObj.AdminEndDate = globalPkg.UTCtimefield(AdminObj.AdminEndDate)     //globalPkg.UTCtime()
	AdminObj.AdminLastUpdateTime = globalPkg.UTCtime()

	// index
	index:= admin.NewAdminIndex()

	AdminObj.AdminID, _ = globalPkg.ConvertIntToFixedLengthString(index, globalPkg.GlobalObj.StringFixedLength)
	i, _ := strconv.Atoi(AdminObj.AdminID)

	var currentIndex = ""
	if i > 0 {
		currentIndex = admin.GetHash([]byte(validator.CurrentValidator.ValidatorIP)) + "_" + AdminObj.AdminID
	} else {
		currentIndex = AdminObj.AdminID
	}

	AdminObj.AdminID = currentIndex

	broadcastTcp.BoardcastingTCP(AdminObj, "addadmin", "admin")
	
	sendJSON, _ := json.Marshal(AdminObj)
	globalPkg.SendResponse(w, sendJSON)
	globalPkg.WriteLog(logobj, AdminObj.AdminUsername, "success")
}

//LoginAdmin to login admin
func LoginAdmin(w http.ResponseWriter, req *http.Request) {

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

		logobj = logpkg.LogStruct{Logindex, now, userIP, "macAdress", "LoginAdmin", "adminModule", "_", "_", "_", 0}
	}
	logobj = logfunc.ReplaceLog(logobj, "LoginAdmin", "adminModule")

	AdminObj := admin.AdminStruct{}
	decoder := json.NewDecoder(req.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&AdminObj); err != nil {
		responseObj := responses.FindResponseByID("1")
		globalPkg.SendError(w, responseObj.EngResponse)
		globalPkg.WriteLog(logobj, responseObj.EngResponse, "failed")
		return
	}
	InputData := AdminObj.AdminUsername + "," + AdminObj.AdminEmail
	logobj.InputData = InputData
	logobj.InputData = AdminObj.AdminUsername
	if AdminObj.AdminUsername == "" || AdminObj.AdminPassword == "" {
		responseObj := responses.FindResponseByID("11")
		globalPkg.SendError(w, responseObj.EngResponse)
		globalPkg.WriteLog(logobj, responseObj.EngResponse, "failed")
		return
	}
	Adminexist := admin.GetAdmins(AdminObj)
	if AdminObj.AdminUsername == Adminexist.AdminUsername && AdminObj.AdminPassword == Adminexist.AdminPassword {
		Adminexist.SuperAdminPassword = ""
		sendJSON, _ := json.Marshal(Adminexist)
		w.Header().Set("jwt-token", globalPkg.GenerateJwtToken(Adminexist.AdminUsername, true)) // set jwt token
		globalPkg.SendResponse(w, sendJSON)
		if logobj.Count > 0 {
			logobj.Count = 0
			broadcastTcp.BoardcastingTCP(logobj, "", "AddAndUpdateLog")
		}
		globalPkg.WriteLog(logobj, Adminexist.AdminUsername, "success")
	} else {
		responseObj := responses.FindResponseByID("11")
		globalPkg.SendError(w, responseObj.EngResponse)
		logobj.Count = logobj.Count + 1
		broadcastTcp.BoardcastingTCP(logobj, "", "AddAndUpdateLog")
		globalPkg.WriteLog(logobj, responseObj.EngResponse, "failed")

	}
}

//UpdateAdmin update admin info
func UpdateAdmin(w http.ResponseWriter, req *http.Request) {
	now, userIP := globalPkg.SetLogObj(req)
	logobj := logpkg.LogStruct{"_", now, userIP, "macAdress", "updateAdmin", "adminModule", "_", "_", "_", 0}

	AdminObj := admin.AdminStruct{}
	decoder := json.NewDecoder(req.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&AdminObj); err != nil {
		responseObj := responses.FindResponseByID("1")
		globalPkg.SendError(w, responseObj.EngResponse)
		globalPkg.WriteLog(logobj, responseObj.EngResponse, "failed")
		return
	}
	//	if utf8.RuneCountInString(AdminObj.AdminPassword) != 64 {
	if utf8.RuneCountInString(AdminObj.AdminPassword) != 64 {
		responseObj := responses.FindResponseByID("11")
		globalPkg.SendError(w, responseObj.EngResponse)
		globalPkg.WriteLog(logobj, responseObj.EngResponse, "failed")
		return
	}

	exist := true
	username := AdminObj.OldUsername
	existsAdminObj := admin.GetAdminsByUsername(username)

	if AdminObj.OldUsername == "" || existsAdminObj.AdminUsername != AdminObj.OldUsername || AdminObj.OldPassword == "" || existsAdminObj.AdminPassword != AdminObj.OldPassword {
		exist = false
		responseObj := responses.FindResponseByID("2")
		globalPkg.SendError(w, responseObj.EngResponse)
		globalPkg.WriteLog(logobj, responseObj.EngResponse, "failed")
		return
	}
    //  check if admin data exist before
	adminlist := admin.GetAllAdmins()
	for _, admin := range adminlist {
		if admin.AdminUsername != existsAdminObj.AdminUsername {
			if admin.AdminUsername == AdminObj.AdminUsername {
				responseObj := responses.FindResponseByID("31")
				globalPkg.SendError(w, responseObj.EngResponse)
				globalPkg.WriteLog(logobj, responseObj.EngResponse, "failed")
				return
			}
			if admin.AdminEmail == AdminObj.AdminEmail {
				responseObj := responses.FindResponseByID("32")
				globalPkg.SendError(w, responseObj.EngResponse)
				globalPkg.WriteLog(logobj, responseObj.EngResponse, "failed")
				return
			} 
			if admin.AdminPhone == AdminObj.AdminPhone {
				responseObj := responses.FindResponseByID("33")
				globalPkg.SendError(w, responseObj.EngResponse)
				globalPkg.WriteLog(logobj, responseObj.EngResponse, "failed")
				return
			}
		}
	}

	if exist == true {
		//update except the  AdminStartDate,
		AdminObj.AdminID = existsAdminObj.AdminID
		AdminObj.AdminStartDate = existsAdminObj.AdminStartDate
		AdminObj.AdminEndDate = existsAdminObj.AdminEndDate
		AdminObj.AdminActive = existsAdminObj.AdminActive
		AdminObj.AdminRole = existsAdminObj.AdminRole
		AdminObj.Validatorlst = existsAdminObj.Validatorlst
		AdminObj.ValiatorIPtoDeactive = existsAdminObj.ValiatorIPtoDeactive
		AdminObj.SuperAdminUsername = existsAdminObj.SuperAdminUsername
		AdminObj.SuperAdminPassword = existsAdminObj.SuperAdminPassword
		AdminObj.AdminLastUpdateTime = globalPkg.UTCtime()
		broadcastTcp.BoardcastingTCP(AdminObj, "updateadmin", "admin")
		responseObj := responses.FindResponseByID("40")
		globalPkg.SendResponseMessage(w, responseObj.EngResponse)
		globalPkg.WriteLog(logobj, responseObj.EngResponse, "success")
		return
	}
	responseObj := responses.FindResponseByID("41")
	globalPkg.SendError(w, responseObj.EngResponse)
	globalPkg.WriteLog(logobj, responseObj.EngResponse, "failed")
	return
}

//UpdatesuperAdmin update all admin info
func UpdatesuperAdmin(w http.ResponseWriter, req *http.Request) {
	now, userIP := globalPkg.SetLogObj(req)
	logobj := logpkg.LogStruct{"_", now, userIP, "macAdress", "updateAdmin", "adminModule", "_", "_", "_", 0}

	AdminObj := admin.AdminStruct{}
	decoder := json.NewDecoder(req.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&AdminObj); err != nil {
		responseObj := responses.FindResponseByID("1")
		globalPkg.SendError(w, responseObj.EngResponse)
		globalPkg.WriteLog(logobj, responseObj.EngResponse, "failed")
		return
	}
	//password hash 64 character
	if utf8.RuneCountInString(AdminObj.AdminPassword) != 64 {
		responseObj := responses.FindResponseByID("11")
		globalPkg.SendError(w, responseObj.EngResponse)
		globalPkg.WriteLog(logobj, responseObj.EngResponse, "failed")
		return
	}
	//this data like name ,email,phone not be empty
	if AdminObj.AdminUsername == "" || AdminObj.AdminEmail == "" || AdminObj.AdminPhone == "" || AdminObj.AdminRole == "" {
		responseObj := responses.FindResponseByID("42")
		globalPkg.SendError(w, responseObj.EngResponse)
		globalPkg.WriteLog(logobj, responseObj.EngResponse, "failed")
		return
	}

	exist := true
	username := AdminObj.OldUsername
	existsAdminObj := admin.GetAdminsByUsername(username)

	// check on username and password for super admin
	// if AdminObj.SuperAdminUsername == "" || existsAdminObj.SuperAdminUsername != AdminObj.SuperAdminUsername || AdminObj.SuperAdminPassword == "" || existsAdminObj.SuperAdminPassword != AdminObj.SuperAdminPassword {
	if AdminObj.SuperAdminUsername == "" || AdminObj.SuperAdminPassword == "" { //if any superadmin can update
		exist = false
		responseObj := responses.FindResponseByID("30")
		globalPkg.SendError(w, responseObj.EngResponse)
		globalPkg.WriteLog(logobj, responseObj.EngResponse, "failed")
		return
	}
	//check if superadmin of super admin can update not any super admin make update
	if existsAdminObj.SuperAdminUsername != AdminObj.SuperAdminUsername || existsAdminObj.SuperAdminPassword != AdminObj.SuperAdminPassword {
		existsSuperAdminforAdminObj := admin.GetAdminsByUsername(existsAdminObj.SuperAdminUsername)
		if existsSuperAdminforAdminObj.SuperAdminUsername != AdminObj.SuperAdminUsername {
			responseObj := responses.FindResponseByID("44")
			globalPkg.SendError(w, responseObj.EngResponse)
			globalPkg.WriteLog(logobj, responseObj.EngResponse, "failed")
			return
		}
	}

	superusername := AdminObj.SuperAdminUsername
	existsSuperAdminObj := admin.GetAdminsByUsername(superusername)
	// only superadmin can do update or use this api
	if existsSuperAdminObj.AdminRole != "SuperAdmin" {
		responseObj := responses.FindResponseByID("44")
		globalPkg.SendError(w, responseObj.EngResponse)
		globalPkg.WriteLog(logobj, responseObj.EngResponse, "failed")
		return
	}
	//super admin can increase  his the period of end date t
	if existsAdminObj.AdminRole != "SuperAdmin" {
		if AdminObj.AdminEndDate.After(existsSuperAdminObj.AdminEndDate) {
			responseObj := responses.FindResponseByID("46")
			globalPkg.SendError(w, responseObj.EngResponse)
			globalPkg.WriteLog(logobj, responseObj.EngResponse, "failed")
			return
		}
	}
	//  check validator ip in validator list or not
	for _, adminValidator := range AdminObj.Validatorlst {
		exist := false
		for _, validatorobj := range validator.ValidatorsLstObj {
			if validatorobj.ValidatorIP == adminValidator {
				exist = true
				break
			}
		}
		if exist == false {
			responseObj := responses.FindResponseByID("34")
			globalPkg.SendError(w, responseObj.EngResponse)
			globalPkg.WriteLog(logobj, responseObj.EngResponse, "failed")
			return
		}
	}
	// check for unique email or phone and name
	adminlist := admin.GetAllAdmins()
	for _, admin := range adminlist {
		if admin.AdminUsername != existsAdminObj.AdminUsername {
			if admin.AdminUsername == AdminObj.AdminUsername {
				responseObj := responses.FindResponseByID("31")
				globalPkg.SendError(w, responseObj.EngResponse)
				globalPkg.WriteLog(logobj, responseObj.EngResponse, "failed")
				return
			}
			if admin.AdminEmail == AdminObj.AdminEmail{
				responseObj := responses.FindResponseByID("32")
				globalPkg.SendError(w, responseObj.EngResponse)
				globalPkg.WriteLog(logobj, responseObj.EngResponse, "failed")
				return
			}
			if admin.AdminPhone == AdminObj.AdminPhone {
				responseObj := responses.FindResponseByID("33")
				globalPkg.SendError(w, responseObj.EngResponse)
				globalPkg.WriteLog(logobj, responseObj.EngResponse, "failed")
				return
			}
		}
	}
	if exist == true {

		//update except the  AdminStartDate,
		AdminObj.AdminID = existsAdminObj.AdminID
		// AdminObj.AdminStartDate = globalPkg.UTCtimefield(AdminObj.AdminStartDate)
		AdminObj.AdminStartDate = existsAdminObj.AdminStartDate //can't update start date
		AdminObj.AdminEndDate = globalPkg.UTCtimefield(AdminObj.AdminEndDate)
		AdminObj.AdminLastUpdateTime = globalPkg.UTCtime()

		broadcastTcp.BoardcastingTCP(AdminObj, "updateadmin", "admin")
		responseObj := responses.FindResponseByID("40")
		globalPkg.SendResponseMessage(w, responseObj.EngResponse)
		globalPkg.WriteLog(logobj, responseObj.EngResponse, "success")
		return
	}
	responseObj := responses.FindResponseByID("41")
	globalPkg.SendError(w, responseObj.EngResponse)
	globalPkg.WriteLog(logobj, responseObj.EngResponse, "failed")
	return

}

//GetAlltransactionPerMonthAPI endpoint to get All transaction Per Month API
func GetAlltransactionPerMonthAPI(w http.ResponseWriter, req *http.Request) {
	//log
	now, userIP := globalPkg.SetLogObj(req)
	logobj := logpkg.LogStruct{"_", now, userIP, "macAdress", "GetAlltransactionPerMonthAPI", "adminModule", "", "", "_", 0}

	Adminobj := admin.Admin{}
	decoder := json.NewDecoder(req.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&Adminobj)

	if err != nil {
		responseObj := responses.FindResponseByID("1")
		globalPkg.SendError(w, responseObj.EngResponse)
		globalPkg.WriteLog(logobj, responseObj.EngResponse, "failed")
		return
	}
	if admin.ValidationAdmin(Adminobj) {
		data := make(map[string]int)
		var monthYear string

		blocklst := block.GetBlockchain()
		for _, block := range blocklst {
			y, m, _ := block.BlockTimeStamp.Date()
			monthYear = strconv.Itoa(y) + "_" + strconv.Itoa(int(m))
			if _, exist := data[monthYear]; !exist {
				data[monthYear] = 0
			}
			data[monthYear] += len(block.BlockTransactions)
		}
	
		// Convert map to slice of key-value pairs.
		transactions := []txdata{}
		for key, value := range data {
			str := strconv.Itoa(value)
			var transaction txdata
			transaction.Time = key
			transaction.TransactionNumber = str
			transactions = append(transactions, transaction)
		}
		sendJSON, _ := json.Marshal(transactions)
		globalPkg.SendResponse(w, sendJSON)
		globalPkg.WriteLog(logobj, "get number of transaction per month ", "success")
	} else {
		responseObj := responses.FindResponseByID("2")
		globalPkg.SendError(w, responseObj.EngResponse)
		globalPkg.WriteLog(logobj, responseObj.EngResponse, "failed")
	}
}

// GetAlltransactionLastTenMinuteAPI endpoint to get All transaction on last ten minutes API
func GetAlltransactionLastTenMinuteAPI(w http.ResponseWriter, req *http.Request) {
	//log
	now, userIP := globalPkg.SetLogObj(req)
	logobj := logpkg.LogStruct{"_", now, userIP, "macAdress", "GetAlltransactionLastTenMinuteAPI", "adminModule", "_", "_", "_", 0}

	Adminobj := admin.Admin{}
	decoder := json.NewDecoder(req.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&Adminobj)

	if err != nil {
		responseObj := responses.FindResponseByID("1")
		globalPkg.SendError(w, responseObj.EngResponse)
		logobj.OutputData = "Faild to Decode Admin Object"
		logobj.Process = "faild"
		globalPkg.WriteLog(logobj, responseObj.EngResponse, "failed")
		return
	}
	if admin.ValidationAdmin(Adminobj) {
		data := make(map[string]int)
		var minuteHour string
		lastBlock := block.GetLastBlock()
		timeNow := time.Now().UTC()
		for {
			diff := timeNow.Sub(lastBlock.BlockTimeStamp).Minutes()
			if diff <= 10 {
				hour, min, _ := lastBlock.BlockTimeStamp.Clock()
				minuteHour = strconv.Itoa(hour) + ":" + strconv.Itoa(min) // 16 : 3
				if _, exist := data[minuteHour]; !exist {
					data[minuteHour] = 0
				}
				data[minuteHour] += len(lastBlock.BlockTransactions)
			} else {
				break
			}

			if lastBlock.BlockIndex == "000000000000000000000000000000" {
				break
			}
			beforeLastIndex, _ := globalPkg.ConvertIntToFixedLengthString(
				globalPkg.ConvertFixedLengthStringtoInt(lastBlock.BlockIndex)-1, globalPkg.GlobalObj.StringFixedLength,
			)
			lastBlock = block.GetBlockInfoByID(beforeLastIndex)
		}

		if len(data) == 0 {
			responseObj := responses.FindResponseByID("47")
			globalPkg.SendResponseMessage(w, responseObj.EngResponse)
			globalPkg.WriteLog(logobj, responseObj.EngResponse, "success")
			return
		}

	

		// Convert map to slice of key-value pairs.
		transactions := []txdata{}
		for key, value := range data {
			str := strconv.Itoa(value)
			var transaction txdata
			transaction.Time = key
			transaction.TransactionNumber = str
			transactions = append(transactions, transaction)
		}

		sendJSON, _ := json.Marshal(transactions)
		globalPkg.SendResponse(w, sendJSON)
		globalPkg.WriteLog(logobj, "get number of transaction on last ten minutes ", "success")

	} else {
		responseObj := responses.FindResponseByID("2")
		globalPkg.SendError(w, responseObj.EngResponse)
		globalPkg.WriteLog(logobj, responseObj.EngResponse, "failed")
	}
}

//GetValidatorListForAdmin get validator list for admin
func GetValidatorListForAdmin(w http.ResponseWriter, req *http.Request) {
	now, userIP := globalPkg.SetLogObj(req)
	logobj := logpkg.LogStruct{"_", now, userIP, "macAdress", "AddNewAdmin", "adminModule", "_", "_", "_", 0}

	AdminObj := admin.AdminStruct{}

	decoder := json.NewDecoder(req.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&AdminObj); err != nil {
		responseObj := responses.FindResponseByID("1")
		globalPkg.SendError(w, responseObj.EngResponse)
		globalPkg.WriteLog(logobj, responseObj.EngResponse, "failed")
		return
	}

	if AdminObj.AdminUsername == "" || AdminObj.AdminPassword == "" {
		responseObj := responses.FindResponseByID("11")
		globalPkg.SendError(w, responseObj.EngResponse)
		globalPkg.WriteLog(logobj, responseObj.EngResponse, "failed")
		return
	}

	existsAdminObj := admin.GetAdminsByUsername(AdminObj.AdminUsername)

	if existsAdminObj.AdminUsername != "" {
		if existsAdminObj.AdminPassword != AdminObj.AdminPassword {
			responseObj := responses.FindResponseByID("11")
			globalPkg.SendError(w, responseObj.EngResponse)
			globalPkg.WriteLog(logobj, responseObj.EngResponse, "failed")
			return
		}
	} else {
		responseObj := responses.FindResponseByID("11")
		globalPkg.SendError(w, responseObj.EngResponse)
		globalPkg.WriteLog(logobj, responseObj.EngResponse, "failed")
		return
	}

	validatorList := existsAdminObj.Validatorlst
	sendJSON, _ := json.Marshal(validatorList)
	globalPkg.SendResponse(w, sendJSON)
	globalPkg.WriteLog(logobj, "return validator list success", "success")
	
}
