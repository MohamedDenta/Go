package accountdb

import (
	// "crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"../globalPkg"

	"../validator"
	"github.com/syndtr/goleveldb/leveldb/util"

	"../errorpk" //  write an error on the json file
	"github.com/syndtr/goleveldb/leveldb"
)

//------------------------------------------------------------------------------------------------------------
//struct for object to be saved in db
//------------------------------------------------------------------------------------------------------------
type AccountStruct struct {
	AccountName                string
	AccountInitialUserName     string
	AccountInitialPassword     string
	AccountAuthenticationType  string
	AccountAuthenticationValue string
	AccountIndex               string
	AccountPassword            string
	AccountEmail               string
	AccountPhoneNumber         string
	AccountAddress             string
	AccountPublicKey           string
	AccountPrivateKey          string
	AccountStatus              bool
	AccountRole                string
	AccountLastUpdatedTime     time.Time
	AccountDeactivatedReason   string
	AccountBalance             string
	BlocksLst                  []string
	SessionID                  string
	AccountTokenID             []string
	Filelist                   []FileList
}

//FileList store file list in account
type FileList struct {
	Fileid     string
	FileName   string
	FileType   string
	FileSize   int64
	Blockindex string
	Filehash   string
	// []permissionlist pk
	PermissionList []string
}

var DB *leveldb.DB

var Open = false

//------------------------------------------------------------------------------------------------------------
// create or open db if exist
//------------------------------------------------------------------------------------------------------------
func Opendatabase() bool {
	if !Open {
		Open = true
		dbpath := "Database/AccountStruct"
		var err error
		DB, err = leveldb.OpenFile(dbpath, nil)
		if err != nil {

			errorpk.AddError("opendatabase AccountStruct package", "can't open the database", "critical error")
			return false
		}
		return true

	}
	return true

}

//------------------------------------------------------------------------------------------------------------
// create or open db if exist for ancidate use
//------------------------------------------------------------------------------------------------------------
// this is function to create a connection to private database that is created by defauilt in your code
// dont use global vars in the path instead you can use Opendatabase()
func opendatabaseCandidate(path string) (bool, *leveldb.DB) {
	var err error
	dbobj, err := leveldb.OpenFile(path, nil)
	if err != nil {
		errorpk.AddError("opencandidatedatabase package", "can't open the database on this path : "+path, "critical error")
		return false, nil
	}
	return true, dbobj

}

//-------------------------------------------------------------------------------------------------------------
// insert AccountStruct
//-------------------------------------------------------------------------------------------------------------
func AccountCreate(data AccountStruct) bool {

	Opendatabase()
	var err error
	d, convert := globalPkg.ConvetToByte(data, "accountCreate account package")
	if !convert {
		return false
	}
	err = DB.Put([]byte(data.AccountIndex), d, nil)
	if err != nil {
		errorpk.AddError("AccountStructCreate  AccountStruct package", "can't create AccountStruct", "runtime error")
		return false
	}
	//return true
	closedatabase()
	if err == nil {
		AccountStructObj := FindAccountByAccountKey(data.AccountIndex)
		AccountNameStructObj := AccountNameStruct{AccountName: AccountStructObj.AccountName, AccountIndex: AccountStructObj.AccountIndex}
		err := accountNameCreate(AccountNameStructObj)
		if !err {
			errorpk.AddError("AccountNameStructCreate  AccountNameStruct package", "can't create AccountNameStruct", "runtime error")
			return false
		}
		if err {
			AccountEmailStructObj := AccountEmailStruct{AccountEmail: AccountStructObj.AccountEmail, AccountIndex: AccountStructObj.AccountIndex}
			err = accountEmailCreate(AccountEmailStructObj)
			if !err {
				errorpk.AddError("AccountEmailStructCreate  AccountEmailStruct package", "can't create AccountEmailStruct", "runtime error")
				return false
			}
			if err {
				AccountPhoneNumberStructObj := AccountPhoneNumberStruct{AccountPhoneNumber: AccountStructObj.AccountPhoneNumber, AccountIndex: AccountStructObj.AccountIndex}
				err = accountPhoneNumberCreate(AccountPhoneNumberStructObj)
				if !err {
					errorpk.AddError("AccountPhoneNumberStructCreate  AccountPhoneNumberStruct package", "can't create AccountPhoneNumberStruct", "runtime error")
					return false
				}
				if err {
					AccountLastUpdatedTimestructObj := AccountLastUpdatedTimestruct{AccountLastUpdatedTime: AccountStructObj.AccountLastUpdatedTime, AccountIndex: AccountStructObj.AccountIndex}
					err = accountLastUpdatedTimeCreate(AccountLastUpdatedTimestructObj)
					if !err {
						errorpk.AddError("AccountLastUpdatedTimestructCreate  AccountLastUpdatedTimestruct package", "can't create AccountLastUpdatedTimestruct", "runtime error")
						return false
					}
					if err {
						AccountPublicKeystructObj := AccountPublicKeyStruct{AccountPublicKey: AccountStructObj.AccountPublicKey, AccountIndex: AccountStructObj.AccountIndex}
						err = accountPublicKeyCreate(AccountPublicKeystructObj)
						if !err {
							errorpk.AddError("AccountPublicKeyStructCreate  AccountPublicKeyStruct package", "can't create AccountPublicKeyStruct", "runtime error")
							return false
						}
					}
				}
			}
		}
	}
	return true
}
func accountCreate2(data AccountStruct) bool {

	Opendatabase()
	var err error
	d, convert := globalPkg.ConvetToByte(data, "accountCreate account package")
	if !convert {
		return false
	}
	err = DB.Put([]byte(data.AccountIndex), d, nil)
	if err != nil {
		errorpk.AddError("AccountStructCreate  AccountStruct package", "can't create AccountStruct", "runtime error")
		return false
	}
	//return true
	closedatabase()

	return true
}

//-------------------------------------------------------------------------------------------------------------
// select By key AccountStruct
//-------------------------------------------------------------------------------------------------------------
func FindAccountByAccountKey(key string) (AccountStructObj AccountStruct) {

	Opendatabase()

	data, err := DB.Get([]byte(key), nil)

	if err != nil {
		errorpk.AddError("AccountStructFindByKey  AccountStruct package", "can't get AccountStruct", "runtime error")
	}
	json.Unmarshal(data, &AccountStructObj)
	closedatabase()
	return AccountStructObj
}

//-------------------------------------------------------------------------------------------------------------
// get all AccountStruct
//-------------------------------------------------------------------------------------------------------------
func GetAllAccounts() (values []AccountStruct) {
	Opendatabase()
	iter := DB.NewIterator(nil, nil)
	for iter.Next() {

		value := iter.Value()
		var newdata AccountStruct
		json.Unmarshal(value, &newdata)
		values = append(values, newdata)
	}
	closedatabase()

	return values
}

func GetLastAccount() AccountStruct {
	Opendatabase()
	var result AccountStruct
	iter := DB.NewIterator(util.BytesPrefix([]byte(GetHash([]byte(validator.CurrentValidator.ValidatorIP))+"_")), nil)
	for iter.Last() {
		value := iter.Value()
		//fmt.Println("******     value   ", value )
		json.Unmarshal(value, &result)
		break
	}

	//fmt.Println("   ------------ result  --       "   , result)
	closedatabase()
	return result
}

func GetFirstAccount() (values AccountStruct) {
	Opendatabase()
	iter := DB.NewIterator(nil, nil)
	iter.First()
	value := iter.Value()
	var newdata AccountStruct
	json.Unmarshal(value, &newdata)
	values = newdata

	closedatabase()

	return values
}

func AccountUpdateUsingTmp(data AccountStruct) bool {
	fmt.Println("+++++++++++++++++++++++++++")
	AccountStructObj := FindAccountByAccountKey(data.AccountIndex)

	if data.AccountName != AccountStructObj.AccountName {
		accountNameDelete(AccountStructObj.AccountName)
		AccountNameStructObj := AccountNameStruct{AccountName: data.AccountName, AccountIndex: AccountStructObj.AccountIndex}
		err := accountNameCreate(AccountNameStructObj)
		if !err {
			errorpk.AddError("AccountNameStructCreate  AccountNameStruct package", "can't create AccountNameStruct", "runtime error")
			return false
		}

		AccountStructObj.AccountName = data.AccountName
	}

	if data.AccountEmail != AccountStructObj.AccountEmail {
		accountEmailDelete(AccountStructObj.AccountEmail)
		AccountEmailStructObj := AccountEmailStruct{AccountEmail: data.AccountEmail, AccountIndex: AccountStructObj.AccountIndex}
		err := accountEmailCreate(AccountEmailStructObj)
		if !err {
			errorpk.AddError("AccountEmailStructCreate  AccountEmailStruct package", "can't create AccountEmailStruct", "logical error")
			return false
		}

	}
	if data.AccountPhoneNumber != AccountStructObj.AccountPhoneNumber {

		AccountPhoneStructObj := AccountPhoneNumberStruct{AccountPhoneNumber: data.AccountPhoneNumber, AccountIndex: AccountStructObj.AccountIndex}
		accountPhoneNumberDelete(AccountStructObj.AccountPhoneNumber)
		err := accountPhoneNumberCreate(AccountPhoneStructObj)
		if !err {
			errorpk.AddError("AccountPhoneNumberStructCreate  AccountPhoneNumberStruct package", "can't create AccountPhoneNumberStruct", "runtime error")
			return false
		}

	}
	AccountLastUpdatedTimestructObj := AccountLastUpdatedTimestruct{AccountLastUpdatedTime: AccountStructObj.AccountLastUpdatedTime, AccountIndex: AccountStructObj.AccountIndex}
	accountLastUpdatedTimeDelete(AccountStructObj.AccountLastUpdatedTime.String())
	err := accountLastUpdatedTimeCreate(AccountLastUpdatedTimestructObj)
	if !err {
		errorpk.AddError("AccountLastUpdatedTimestructCreate  AccountLastUpdatedTimestruct package", "can't create AccountLastUpdatedTimestruct", "runtime error")
		return false
	}
	accountCreate2(data)

	return true

}

//------------------------------------------------------------------------------------------------------------
// close db if exist
//------------------------------------------------------------------------------------------------------------
func closedatabase() bool {

	return true
}

func AccountUpdate2(data AccountStruct) bool {

	Opendatabase()
	var err error
	d, convert := globalPkg.ConvetToByte(data, "accountCreate account package")
	if !convert {
		return false
	}
	err = DB.Put([]byte(data.AccountIndex), d, nil)
	if err != nil {
		errorpk.AddError("AccountStructCreate  AccountStruct package", "can't create AccountStruct", "runtime error")
		return false
	}
	closedatabase()

	return true

}

//addBKey adding public key to the account when it's confirm that the user dowenloaded private key
func AddBKey(accountObj AccountStruct) bool {
	accountObj.AccountLastUpdatedTime = globalPkg.UTCtime()
	if AccountCreate(accountObj) {
		return true
	}
	return false
}

//newIndex return the next empty index to be used

func GetHash(str []byte) string {
	hasher := sha256.New()
	hasher.Write(str)
	sha := base64.URLEncoding.EncodeToString(hasher.Sum(nil))
	return sha
}
