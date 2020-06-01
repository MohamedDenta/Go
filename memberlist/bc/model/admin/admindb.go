package admin

import (
	"encoding/json"
	// "reflect"
	"time"
	// "fmt"
	errorpk "../errorpk"
	"../globalPkg"

	"github.com/syndtr/goleveldb/leveldb"
)

//AdminStruct struct for object to be saved in db
type AdminStruct struct {
	AdminID              string
	AdminUsername        string
	AdminPassword        string
	AdminEmail           string
	AdminPhone           string
	AdminStartDate       time.Time
	AdminEndDate         time.Time
	AdminActive          bool
	AdminRole            string
	Validatorlst         []string
	ValiatorIPtoDeactive string
	SuperAdminUsername   string
	SuperAdminPassword   string
	OldUsername          string
	OldPassword          string
	AdminLastUpdateTime  time.Time
}

//DB name leveldb
var DB *leveldb.DB

//Open flag open db or not
var Open = false

// opendatabase create or open DB if exist
func opendatabase() bool {
	if !Open {
		Open = true
		DBpath := "Database/AdminStruct"
		var err error
		DB, err = leveldb.OpenFile(DBpath, nil)
		if err != nil {
			errorpk.AddError("opendatabase AdminStruct package", "can't open the database", "critical error")
			return false
		}
		return true
	}
	return true

}

// close DB if exist
func closedatabase() bool {
	return true
}

//CreateAdmin insert AdminStruct
func CreateAdmin(data AdminStruct) bool {

	opendatabase()
	d, convert := globalPkg.ConvetToByte(data, "Admin create Admin package")
	if !convert {
		closedatabase()
		return false
	}
	err := DB.Put([]byte(data.AdminID), d, nil)
	if err != nil {
		errorpk.AddError("AdminCreate  AdminStruct package", "can't create AdminStruct", "runtime error")
		return false
	}
	closedatabase()
	return true
}

// FindAdminByid select By key (AdminUsername) AdminStruct
func FindAdminByid(id string) (AdminStructObj AdminStruct) {
	opendatabase()
	data, err := DB.Get([]byte(id), nil)
	if err != nil {
		errorpk.AddError("FindAdminByid  AdminStruct package", "can't get AdminStruct", "runtime error")
	}

	json.Unmarshal(data, &AdminStructObj)
	closedatabase()
	return AdminStructObj
}

// // FindAdminByid select By key (AdminUsername) AdminStruct
// func findAdminByUsername(id string) (AdminStructObj AdminStruct) {
// 	opendatabase()
// 	x := id
// 	data, err := DB.Get([]byte(x), nil)
// 	if err != nil {
// 		errorpk.AddError("FindAdminByid  AdminStruct package", "can't get AdminStruct", "runtime error")
// 	}

// 	json.Unmarshal(data, &AdminStructObj)
// 	closedatabase()
// 	return AdminStructObj
// }
// GetAllAdmins get all AdminStruct
func GetAllAdmins() (values []AdminStruct) {

	opendatabase()
	iter := DB.NewIterator(nil, nil)
	for iter.Next() {

		value := iter.Value()
		var newdata AdminStruct
		json.Unmarshal(value, &newdata)
		values = append(values, newdata)
	}
	closedatabase()
	return values
}

// GetAdmins get all AdminStruct
func GetAdmins(data AdminStruct) (values AdminStruct) {

	opendatabase()
	iter := DB.NewIterator(nil, nil)
	for iter.Next() {

		value := iter.Value()

		var newdata AdminStruct
		json.Unmarshal(value, &newdata)
		if data.AdminUsername == newdata.AdminUsername {
			values = newdata
		}

	}
	closedatabase()
	return values
}

// GetAdminsByUsername get by username
func GetAdminsByUsername(username string) (values AdminStruct) {
	opendatabase()
	iter := DB.NewIterator(nil, nil)
	for iter.Next() {

		value := iter.Value()

		var newdata AdminStruct
		json.Unmarshal(value, &newdata)
		if newdata.AdminUsername == username {
			values = newdata
		}

	}
	closedatabase()
	return values
}

// GetAdminsBySuperUsername get by username
func GetAdminsBySuperUsername(username string) (values AdminStruct) {
	opendatabase()
	iter := DB.NewIterator(nil, nil)
	for iter.Next() {

		value := iter.Value()

		var newdata AdminStruct
		json.Unmarshal(value, &newdata)
		if newdata.SuperAdminUsername == username {
			values = newdata
		}

	}
	closedatabase()
	return values
}

//UpdateAdmindb update AdminStruct by Admin OldUsername and OldPassword
func UpdateAdmindb(data AdminStruct) bool {

	err := CreateAdmin(data)
	if err {
		return true
	}
	return false

}

//-------------------------------------------------------------------------------------------------------------
// get all AccountStruct
//-------------------------------------------------------------------------------------------------------------
func getLastAdmin() (values AdminStruct) {
	opendatabase()
	iter := DB.NewIterator(nil, nil)
	iter.Last()
	value := iter.Value()
	var newdata AdminStruct
	json.Unmarshal(value, &newdata)
	values = newdata

	closedatabase()

	return values
}

//UpdateAdmindb update AdminStruct by Admin OldUsername
// func UpdateAdmindb(data AdminStruct) (bool,string) {
// var errorfound string
// exist :=false
// adminobj := AdminStruct{}
// adminobj.AdminUsername = data.OldUsername
// 	AdminStructObj := FindAdminByid(adminobj.AdminUsername)

// 	if AdminStructObj.AdminUsername == "" {
// 		errorfound = "username  don't exist before "
// 		return false,errorfound
// 	}
// 	// //update except the  AdminStartDate,
// 	// data.AdminStartDate = AdminStructObj.AdminStartDate

// 	// data.AdminEndDate = AdminStructObj.AdminEndDate
// 	// data.AdminActive = AdminStructObj.AdminActive
// 	// data.AdminRole = AdminStructObj.AdminRole
// 	// data.Validatorlst = AdminStructObj.Validatorlst
// 	// data.ValiatorIPtoDeactive = AdminStructObj.ValiatorIPtoDeactive
// 	// data.SuperAdminUsername = AdminStructObj.SuperAdminUsername
// 	// data.SuperAdminPassword = AdminStructObj.SuperAdminPassword

// 	err := adminstructDelete(AdminStructObj.AdminUsername)
// 	if err {
// 			 lst := GetAllAdmins()
// 			for _, obj := range lst{
// 				if data.AdminUsername == obj.AdminUsername{
// 					exist =true

// 					}
// 			}

// 			if exist == true {
// 				errorfound ="Username already exist please change to another username"
// 				err = AdminCreate(AdminStructObj)
// 				// if err{
// 				// 	return true,""
// 				// }
// 				return false,errorfound
// 			}else{
// 				return true,""
// 			}
// 	//    data.OldUsername = data.AdminUsername
// 	//    data.AdminLastUpdateTime, _ = time.Parse("2006-01-02 03:04:05 PM -0000", time.Now().UTC().Format("2006-01-02 03:04:05 PM -0000"))
// 	// 	err = AdminCreate(data)
// 	// 	if err{
// 	// 		return true,""
// 	// 	}

// 	}

// 	return false,errorfound

// }

// func Updateadmin(data AdminStruct)bool{
// 	data.OldUsername = data.AdminUsername

// 	// data.AdminEndDate = data.AdminEndDate
// 	// data.AdminActive = data.AdminActive
// 	// data.AdminRole = data.AdminRole
// 	// data.Validatorlst = data.Validatorlst
// 	// data.ValiatorIPtoDeactive = data.ValiatorIPtoDeactive
// 	// data.SuperAdminUsername = data.SuperAdminUsername
// 	// data.SuperAdminPassword = data.SuperAdminPassword
// 	data.AdminLastUpdateTime, _ = time.Parse("2006-01-02 03:04:05 PM -0000", time.Now().UTC().Format("2006-01-02 03:04:05 PM -0000"))
// 	 err := AdminCreate(data)
// 	 if err{
// 		 return true
// 	 }

// 	 return false
// }
// //adminstructDelete delete adminstruct that do update
// func adminstructDelete(AdminUsername string) bool {

// 	opendatabase()
// 	_, convert := globalPkg.ConvetToByte(AdminUsername, "Admin delete Admin package")
// 	if !convert {
// 		closedatabase()
// 		return false
// 	}
// 	err := DB.Delete([]byte(AdminUsername), nil)
// 	if err != nil {
// 		errorpk.AddError("AdminCreate  AdminStruct package", "can't create AdminStruct", "runtime error")
// 		return false
// 	}
// 	closedatabase()
// 	return true

// }
