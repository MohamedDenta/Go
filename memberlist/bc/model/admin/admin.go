package admin

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"time"
)

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
	fmt.Println("**********admin", adminObj)
	if adminObj.AdminPassword == admin.PasswordAdmin && adminObj.AdminEndDate.After(time.Now().UTC()) {
		return true
	}
	return false
}

//AdminAccountExistsBefore   to check if admin account exists or not
func AdminAccountExistsBefore(AdminUsername string) bool {
	if (GetAdminsByUsername(AdminUsername)).AdminUsername == "" {
		return false //not exist
	}
	return true
}

//DataFound check email , phone exist before
func DataFound(AdminObj AdminStruct) string {

	adminList := GetAllAdmins()
	for _, admin := range adminList {
		if admin.AdminEmail == AdminObj.AdminEmail {
			return "this email exist before"
		}
		if admin.AdminPhone == AdminObj.AdminPhone {
			return "this phone exist before "
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
