package adminModule

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	admin "../admin"
	"../block"
	"../broadcastTcp"
	globalPkg "../globalPkg"
	// "../logfunc"
	logpkg "../logpkg"
	validator "../validator"
)

// GetAllAdminsAPI get all Admins info
func GetAllAdminsAPI(w http.ResponseWriter, req *http.Request) {
	// write log struct
	now, userIP := globalPkg.SetLogObj(req)
	logobj := logpkg.LogStruct{"_", now, userIP, "macAdress", "GetAllAdminsAPI", "adminModule", "_", "_", "_", 0}

	AdminObj := admin.AdminStruct{}
	decoder := json.NewDecoder(req.Body)

	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&AdminObj); err != nil {
		globalPkg.SendError(w, "please enter your correct request")
		globalPkg.WriteLog(logobj, "please enter your correct request", "failed")
		return
	}
	if AdminObj.AdminUsername == "" || AdminObj.AdminPassword == "" {
		globalPkg.SendError(w, "Please enter username and password")
		globalPkg.WriteLog(logobj, "Please enter username and password", "failed")
		return
	}
	Adminexist := admin.GetAdminsByUsername(AdminObj.AdminUsername)
	if AdminObj.AdminUsername == Adminexist.AdminUsername && AdminObj.AdminPassword == Adminexist.AdminPassword {
		lst := admin.GetAllAdmins()
		for index, _ := range lst {
			lst[index].AdminPassword = ""
			lst[index].SuperAdminPassword = ""
		}
		sendJSON, _ := json.Marshal(lst)
		globalPkg.SendResponse(w, sendJSON)
		globalPkg.WriteLog(logobj, "list of all Admin ", "success")

	} else {
		globalPkg.SendError(w, "please check password and username")
		globalPkg.WriteLog(logobj, "please check password and username ", "failed")

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
		globalPkg.SendError(w, "please enter your correct request")
		globalPkg.WriteLog(logobj, "please enter your correct request", "failed")
		return
	}
	logobj.InputData = AdminObj.AdminUsername
	if AdminObj.SuperAdminUsername == "" || AdminObj.SuperAdminPassword == "" {
		globalPkg.SendError(w, "please enter superAdmin username and superAdmin password")
		globalPkg.WriteLog(logobj, "please enter superAdmin username and superAdmin password", "failed")
		return
	}
	if len(AdminObj.AdminPassword) != 64 {
		globalPkg.SendError(w, "please enter correct password")
		globalPkg.WriteLog(logobj, "please enter correct password", "failed")
		return
	}
	if !admin.ValidationAdmin(admin.Admin{AdminObj.SuperAdminUsername, AdminObj.SuperAdminPassword, ""}) {
		globalPkg.SendError(w, "You are not an admin")
		globalPkg.WriteLog(logobj, "You are not an admin", "failed")
		return
	}

	if admin.AdminAccountExistsBefore(AdminObj.AdminUsername) {
		globalPkg.SendError(w, "this username exists before")
		globalPkg.WriteLog(logobj, "this username exists before", "failed")
		return
	}
	//check for email , phone is exist before
	errorfound := admin.DataFound(AdminObj)
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

			globalPkg.SendError(w, "please enter correct validator IP")
			globalPkg.WriteLog(logobj, "please enter correct validator IP", "failed")
			return
		}
	}

	if AdminObj.AdminStartDate.Before(time.Now().UTC()) || AdminObj.AdminEndDate.Before(AdminObj.AdminStartDate) {
		globalPkg.SendError(w, "start date not before date now and End date not before start date")
		logobj.OutputData = "start date not before date now"
		logobj.Process = "faild"
		logpkg.WriteOnlogFile(logobj)
		return
	}
	AdminObj.AdminStartDate = globalPkg.UTCtimefield(AdminObj.AdminStartDate) //globalPkg.UTCtime(AdminObj.AdminStartDate)
	AdminObj.AdminEndDate = globalPkg.UTCtimefield(AdminObj.AdminEndDate)     //globalPkg.UTCtime()
	AdminObj.AdminLastUpdateTime = globalPkg.UTCtime()

	// index
	LastIndex := admin.GetLastIndex()

	index := 0
	if LastIndex != "-1" {
		res := strings.Split(LastIndex, "_")

		if len(res) != 0 {
			index = globalPkg.ConvertFixedLengthStringtoInt(res[len(res)-1]) + 1
		} else {
			index = globalPkg.ConvertFixedLengthStringtoInt(LastIndex) + 1
		}
	}
	AdminObj.AdminID, _ = globalPkg.ConvertIntToFixedLengthString(index, globalPkg.GlobalObj.StringFixedLength)
	i, _ := strconv.Atoi(AdminObj.AdminID)

	var currentIndex = ""
	if i > 0 {
		currentIndex = admin.GetHash([]byte(validator.CurrentValidator.ValidatorIP)) + "_" + AdminObj.AdminID
	} else {
		currentIndex = AdminObj.AdminID
	}

	// fmt.Println("current index  ** ", currentIndex)
	AdminObj.AdminID = currentIndex

	broadcastTcp.BoardcastingTCP(AdminObj, "addadmin", "admin")
	// admin.AdminCreate(AdminObj)
	sendJSON, _ := json.Marshal(AdminObj)
	globalPkg.SendResponse(w, sendJSON)
	globalPkg.WriteLog(logobj, AdminObj.AdminUsername, "success")
}

//LoginAdmin to login admin
func LoginAdmin(w http.ResponseWriter, req *http.Request) {

	now, userIP := globalPkg.SetLogObj(req)
	logobj := logpkg.LogStruct{"", now, userIP, "macAdress", "LoginAdmin", "adminModule", "_", "_", "_", 0}
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

	// 	logobj = logpkg.LogStruct{Logindex, now, userIP, "macAdress", "LoginAdmin", "adminModule", "_", "_", "_", 0}
	// }
	// logobj = logfunc.ReplaceLog(logobj, "LoginAdmin", "adminModule")

	AdminObj := admin.AdminStruct{}
	decoder := json.NewDecoder(req.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&AdminObj); err != nil {
		globalPkg.SendError(w, "please enter your correct request")
		globalPkg.WriteLog(logobj, "please enter your correct request", "failed")
		return
	}
	InputData := AdminObj.AdminUsername + "," + AdminObj.AdminEmail
	logobj.InputData = InputData
	logobj.InputData = AdminObj.AdminUsername
	if AdminObj.AdminUsername == "" || AdminObj.AdminPassword == "" {
		globalPkg.SendError(w, "Please enter username and password")
		globalPkg.WriteLog(logobj, "please enter your correct request", "failed")
		return
	}
	Adminexist := admin.GetAdmins(AdminObj)
	if AdminObj.AdminUsername == Adminexist.AdminUsername && AdminObj.AdminPassword == Adminexist.AdminPassword {
		Adminexist.SuperAdminPassword = ""
		sendJSON, _ := json.Marshal(Adminexist)
		w.Header().Set("jwt-token", globalPkg.GenerateJwtToken(Adminexist.AdminUsername, true)) // set jwt token
		globalPkg.SendResponse(w, sendJSON)
		// if logobj.Count > 0 {
		// 	logobj.Count = 0
		// 	broadcastTcp.BoardcastingTCP(logobj, "", "AddAndUpdateLog")
		// }
		globalPkg.WriteLog(logobj, Adminexist.AdminUsername, "success")
	} else {
		globalPkg.SendError(w, "please check password and username")
		// logobj.Count = logobj.Count + 1

		// broadcastTcp.BoardcastingTCP(logobj, "", "AddAndUpdateLog")
		globalPkg.WriteLog(logobj, "please check password and username", "failed")

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
		globalPkg.SendError(w, "please enter your correct request")
		globalPkg.WriteLog(logobj, "please enter your correct request", "failed")
		return
	}

	if len(AdminObj.AdminPassword) != 64 {
		globalPkg.SendError(w, "please enter admin password")
		globalPkg.WriteLog(logobj, "please enter admin password", "failed")
		return
	}

	exist := true
	username := AdminObj.OldUsername
	existsAdminObj := admin.GetAdminsByUsername(username)

	if AdminObj.OldUsername == "" || existsAdminObj.AdminUsername != AdminObj.OldUsername || AdminObj.OldPassword == "" || existsAdminObj.AdminPassword != AdminObj.OldPassword {
		exist = false
		globalPkg.SendError(w, "this admin is not found")
		globalPkg.WriteLog(logobj, "this admin is not found", "failed")
		return
	}

	adminlist := admin.GetAllAdmins()
	for _, admin := range adminlist {
		if admin.AdminUsername != existsAdminObj.AdminUsername {
			if admin.AdminUsername == AdminObj.AdminUsername || admin.AdminEmail == AdminObj.AdminEmail || admin.AdminPhone == AdminObj.AdminPhone {
				globalPkg.SendError(w, "username or email or phone exist before ")
				globalPkg.WriteLog(logobj, "email or phone exist before", "failed")
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
		globalPkg.SendResponseMessage(w, "Your admin updated successfully ")
		globalPkg.WriteLog(logobj, "Your admin updated successfully", "success")
		return
	}
	globalPkg.SendError(w, "Your admin not updated  ")
	globalPkg.WriteLog(logobj, "Your admin not updated", "failed")
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
		globalPkg.SendError(w, "please enter your correct request")
		globalPkg.WriteLog(logobj, "please enter your correct request", "failed")
		return
	}

	if len(AdminObj.AdminPassword) != 64 {
		globalPkg.SendError(w, "please enter admin password")
		globalPkg.WriteLog(logobj, "please enter admin password", "failed")
		return
	}

	if AdminObj.AdminUsername == "" || AdminObj.AdminEmail == "" || AdminObj.AdminPhone == "" || AdminObj.AdminRole == "" {
		globalPkg.SendError(w, "enter required data admin username or email or phone or admin role must be not empty")
		globalPkg.WriteLog(logobj, "enter required data admin username or email or phone or admin role must be not empty", "failed")
		return
	}

	exist := true
	username := AdminObj.OldUsername
	existsAdminObj := admin.GetAdminsByUsername(username)

	if AdminObj.SuperAdminUsername == "" || existsAdminObj.SuperAdminUsername != AdminObj.SuperAdminUsername || AdminObj.SuperAdminPassword == "" || existsAdminObj.SuperAdminPassword != AdminObj.SuperAdminPassword {
		exist = false
		globalPkg.SendError(w, "this Super admin is not found")
		globalPkg.WriteLog(logobj, "this Super admin is not found", "failed")
		return
	}
	superusername := AdminObj.SuperAdminUsername
	existsSuperAdminObj := admin.GetAdminsByUsername(superusername)
	// fmt.Println("----    ***    ---    ", existsSuperAdminObj)
	if AdminObj.AdminEndDate.After(existsSuperAdminObj.AdminEndDate) {
		globalPkg.SendError(w, "end date of admin  must  before the end date of the super admin ")
		globalPkg.WriteLog(logobj, "end date of admin  must  before the end date of the super admin", "failed")
		return
	}

	adminlist := admin.GetAllAdmins()
	for _, admin := range adminlist {
		if admin.AdminUsername != existsAdminObj.AdminUsername {
			if admin.AdminUsername == AdminObj.AdminUsername || admin.AdminEmail == AdminObj.AdminEmail || admin.AdminPhone == AdminObj.AdminPhone {
				globalPkg.SendError(w, "username or email or phone exist before ")
				globalPkg.WriteLog(logobj, "email or phone exist before", "failed")
				return
			}
		}
	}
	if exist == true {

		//update except the  AdminStartDate,
		AdminObj.AdminID = existsAdminObj.AdminID
		AdminObj.AdminStartDate = globalPkg.UTCtimefield(AdminObj.AdminStartDate)
		AdminObj.AdminEndDate = globalPkg.UTCtimefield(AdminObj.AdminEndDate)
		AdminObj.AdminLastUpdateTime = globalPkg.UTCtime()

		broadcastTcp.BoardcastingTCP(AdminObj, "updateadmin", "admin")
		globalPkg.SendResponseMessage(w, "Your admin updated successfully ")
		globalPkg.WriteLog(logobj, "Your admin updated successfully", "success")
		return
	}
	globalPkg.SendError(w, "Your admin not updated  ")
	globalPkg.WriteLog(logobj, "Your admin not updated", "failed")
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
		globalPkg.SendError(w, "please enter your correct request ")
		globalPkg.WriteLog(logobj, "Faild to Decode Admin Object", "failed")
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
		type txdata struct {
			Time              string
			TransactionNumber string
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
		globalPkg.SendError(w, "you are not the admin ")
		globalPkg.WriteLog(logobj, "you are not the admin to  get All transaction Per Month ", "failed")
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
		globalPkg.SendError(w, "please enter your correct request ")
		logobj.OutputData = "Faild to Decode Admin Object"
		logobj.Process = "faild"
		globalPkg.WriteLog(logobj, "Faild to Decode Admin Object", "failed")
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
			// fmt.Println("___   ", beforeLastIndex)
			lastBlock = block.GetBlockInfoByID(beforeLastIndex)
		}

		if len(data) == 0 {
			globalPkg.SendResponseMessage(w, "no block created in last ten minutes")
			globalPkg.WriteLog(logobj, "no block in last ten minutes", "failed")
			return
		}

		type txdata struct {
			Time              string
			TransactionNumber string
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
		globalPkg.SendError(w, "you are not the admin ")
		globalPkg.WriteLog(logobj, "you are not the admin to  get All transaction on last ten minutes  ", "failed")
	}
}

// GetadminsofSuperAdminAPI superadmin want to see who admins controlled by you
func GetadminsofSuperAdminAPI(w http.ResponseWriter, req *http.Request) {
	// write log struct
	now, userIP := globalPkg.SetLogObj(req)
	logobj := logpkg.LogStruct{"_", now, userIP, "macAdress", "GetAllAdminsAPI", "adminModule", "_", "_", "_", 0}

	AdminObj := admin.AdminStruct{}
	decoder := json.NewDecoder(req.Body)

	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&AdminObj); err != nil {
		globalPkg.SendError(w, "please enter your correct request")
		globalPkg.WriteLog(logobj, "please enter your correct request", "failed")
		return
	}
	if AdminObj.AdminUsername == "" || AdminObj.AdminPassword == "" {
		globalPkg.SendError(w, "Please enter username and password")
		globalPkg.WriteLog(logobj, "Please enter username and password", "failed")
		return
	}
	Adminexist := admin.GetAdminsByUsername(AdminObj.AdminUsername)
	admnlst := []admin.AdminStruct{}
	if AdminObj.AdminUsername == Adminexist.AdminUsername && AdminObj.AdminPassword == Adminexist.AdminPassword {
		lst := admin.GetAllAdmins()
		for _, adm := range lst {
		adnOnj	:= admin.GetAdminsBySuperUsername(adm.SuperAdminUsername)
		if adnOnj.SuperAdminUsername == Adminexist.AdminUsername{
			admnlst = append(admnlst , adm)
		   }
		}
		sendJSON, _ := json.Marshal(admnlst)
		globalPkg.SendResponse(w, sendJSON)
		globalPkg.WriteLog(logobj, "list of all Admin ", "success")

	} else {
		globalPkg.SendError(w, "please check password and username")
		globalPkg.WriteLog(logobj, "please check password and username ", "failed")

	}

}