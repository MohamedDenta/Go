package admin

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"../globalPkg"
	"../logpkg"

	"../responses"
)

//function that used if API need admin permition
//TODO if !AdminAPIDecoderAndValidation(w,req.Body,logobj){
//		return
//	}
//after filling logstruct
func AdminAPIDecoderAndValidation(w http.ResponseWriter, reqbody io.ReadCloser, logobj logpkg.LogStruct) bool {
	decoder := json.NewDecoder(reqbody)
	decoder.DisallowUnknownFields()
	adminObj := Admin{}
	err := decoder.Decode(&adminObj)
	if err != nil {
		responseObj := responses.FindResponseByID("1")
		globalPkg.SendNotFound(w, responseObj.EngResponse)
		logobj.OutputData = responseObj.EngResponse
		logobj.Process = "failed"
		logobj.WriteLog()
		return false
	} else if !ValidationAdmin(adminObj) {
		globalPkg.SendNotFound(w, responses.FindResponseByID("2").EngResponse)
		logobj.OutputData = responses.FindResponseByID("2").EngResponse
		logobj.Process = "failed"
		logobj.WriteLog()
		return false
	}
	return true
}

//Admin permission
type Admin struct {
	UsernameAdmin   string
	PasswordAdmin   string
	ObjectInterface interface{}
}

//Admin1 permission
type Admin1 struct {
	UsernameAdmin string
	PasswordAdmin string
}

//ValidationAdmin validate admin
func ValidationAdmin(admin Admin) bool {

	// adminObj := FindAdminByid(admin.UsernameAdmin)
	adminObj := GetAdminsByUsername(admin.UsernameAdmin)
	if adminObj.AdminPassword == admin.PasswordAdmin && adminObj.AdminEndDate.After(time.Now().UTC()) {
		return true
	}
	return false
}

//CheckAdminExistsBefore   to check if admin account exists or not
func CheckAdminExistsBefore(AdminUsername string) bool {
	if (GetAdminsByUsername(AdminUsername)).AdminUsername == "" {
		return false //not exist
	}
	return true
}

//DataFound check email , phone exist before
func (AdminObj *AdminStruct) DataFound() string {

	adminList := GetAllAdmins()
	for _, admin := range adminList {
		if admin.AdminEmail == AdminObj.AdminEmail {
			responseObj := responses.FindResponseByID("31")
			return responseObj.EngResponse
		}
		if admin.AdminPhone == AdminObj.AdminPhone {
			responseObj := responses.FindResponseByID("32")
			return responseObj.EngResponse
		}
	}

	return ""
}

//AdminUPdate update admin
// func AdminUPdate (AdminObj AdminStruct){
// 	updateAdmindb(AdminObj)
// }

//GetHash get hash to index
func GetHash(str []byte) string {
	hasher := sha256.New()
	hasher.Write(str)
	sha := base64.URLEncoding.EncodeToString(hasher.Sum(nil))
	return sha
}

//GetLastIndex get last index for admin
func GetLastIndex() string {
	var Admin AdminStruct
	Admin = getLastAdmin()
	if Admin.AdminID == "" {
		return "-1"
	}
	return Admin.AdminID
}

//NewAdminIndex for admin
func NewAdminIndex() int {
	LastIndex := GetLastIndex()

	index := 0
	if LastIndex != "-1" {
		res := strings.Split(LastIndex, "_")

		if len(res) != 0 {
			index = globalPkg.ConvertFixedLengthStringtoInt(res[len(res)-1]) + 1
		} else {
			index = globalPkg.ConvertFixedLengthStringtoInt(LastIndex) + 1
		}
	}
	return index
}
