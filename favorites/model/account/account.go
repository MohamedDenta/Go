package account

import (
	"fmt"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"../accountdb"
	"../broadcastTcp"
	"../cryptogrpghy"
	"../errorpk"
	error "../errorpk"
	"../filestorage"
	"../responses"
)

//User contain user data
type User struct {
	Account          accountdb.AccountStruct
	Oldpassword      string
	CurrentTime      time.Time
	Confirmationcode string
	TextSearch       string
	Method           string
	PathAPI          string //dynamic api
}

// searchResponse search reponse
type searchResponse struct {
	UserName  string
	PublicKey string
}

// GetAccountByAccountPubicKey function to get Account by public key
func GetAccountByAccountPubicKey(AccountPublicKey string) accountdb.AccountStruct {
	return accountdb.FindAccountByAccountPublicKey(AccountPublicKey)
}

// IfAccountExistsBefore if account exist before pk
func IfAccountExistsBefore(AccountPublicKey string) bool {
	if (accountdb.FindAccountByAccountPublicKey(AccountPublicKey)).AccountPublicKey == "" {
		return false //not exist
	}
	return true
}

// AddAccount function to save an account on json file
func AddAccount(accountObj accountdb.AccountStruct) string {
	_, b := validateAccount(accountObj)
	//if !(ifAccountExistsBefore(accountObj.AccountPublicKey)) && b {
	if b {
		if accountdb.AccountCreate(accountObj) {
			return ""
		}
		return error.AddError("AddAccount account package", "Check your path or object to Add AccountStruct", "logical error")

	}
	return error.AddError("AddAccount account package", "error in user data validation", "logical error")

}

//getLastIndex get last index from db with account name
func getLastIndex() string {

	var Account accountdb.AccountStruct
	Account = accountdb.GetLastAccount()
	//if Account.AccountPublicKey == "" {
	//	return "-1"
	//}
	if Account.AccountName == "" {
		return "-1"
	}
	return Account.AccountIndex
}

// UpdateAccount function to update an account on json file
func UpdateAccount(accountObj accountdb.AccountStruct) string {
	_, b := validateAccount(accountObj)
	if b {
		if accountdb.AccountUpdateUsingTmp(accountObj) {
			return ""
		}
		return error.AddError("UpdateAccount account package", "Check your path or object to Update AccountStruct", "logical error")

	}
	return error.AddError("FindjsonFile account package", "error in validate account data", "logical error")
}

// validateAccount function to validate the account before register
func validateAccount(accountObj accountdb.AccountStruct) (string, bool) {
	if utf8.RuneCountInString(accountObj.AccountName) < 8 || utf8.RuneCountInString(accountObj.AccountName) > 30 {
		responseObj := responses.FindResponseByID("147")
		return responseObj.EngResponse, false
	}
	if len(accountObj.AccountPassword) != 64 {
		responseObj := responses.FindResponseByID("148")
		return responseObj.EngResponse, false
	}
	if utf8.RuneCountInString(accountObj.AccountAddress) < 5 || utf8.RuneCountInString(accountObj.AccountAddress) > 100 {
		responseObj := responses.FindResponseByID("149")
		return responseObj.EngResponse, false
	}
	re := regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")
	if !re.MatchString(accountObj.AccountEmail) && accountObj.AccountEmail != "" {
		responseObj := responses.FindResponseByID("150")
		return responseObj.EngResponse, false
	}
	re = regexp.MustCompile("^[a-zA-Z0-9\u0621-\u064A\u0660-\u0669]+(([ -_][a-zA-Z0-9\u0621-\u064A\u0660-\u0669 _])?[a-zA-Z0-9\u0621-\u064A\u0660-\u0669]*)$")
	if !re.MatchString(accountObj.AccountName) {
		responseObj := responses.FindResponseByID("151")
		return responseObj.EngResponse, false
	}
	re = regexp.MustCompile("^[a-zA-Z0-9]+(([. ,-][a-zA-Z _])?[a-zA-Z]*)*$")
	if !re.MatchString(accountObj.AccountAddress) {
		responseObj := responses.FindResponseByID("152")
		return responseObj.EngResponse, false
	}
	return "", true
}

//getPublicKeyUsingString func Add by Aya to get publickey using Any string
func getPublicKeyUsingString(Key string) string {

	existingAccountUsingName := accountdb.FindAccountByAccountName(Key)
	if existingAccountUsingName.AccountPublicKey != "" {
		return existingAccountUsingName.AccountPublicKey
	}
	existingAccountusingEmail := accountdb.FindAccountByAccountEmail(Key)
	if existingAccountusingEmail.AccountPublicKey != "" {
		return existingAccountusingEmail.AccountPublicKey
	}
	existingAccountusingPhoneNumber := accountdb.FindAccountByAccountPhoneNumber(Key)
	if existingAccountusingPhoneNumber.AccountPublicKey != "" {
		return existingAccountusingPhoneNumber.AccountPublicKey
	}
	return ""
}

// checkAccount to check account
func checkAccount(userAccountObj accountdb.AccountStruct) string {

	existingAccountUsingName := accountdb.FindAccountByAccountName(userAccountObj.AccountName)
	if existingAccountUsingName.AccountPublicKey != "" && existingAccountUsingName.AccountPublicKey != userAccountObj.AccountPublicKey {
		responseObj := responses.FindResponseByID("33")
		return responseObj.EngResponse
	}
	existingAccountusingEmail := accountdb.FindAccountByAccountEmail(userAccountObj.AccountEmail)
	if existingAccountusingEmail.AccountPublicKey != "" && existingAccountusingEmail.AccountPublicKey != userAccountObj.AccountPublicKey && userAccountObj.AccountEmail != "" {
		responseObj := responses.FindResponseByID("31")
		return responseObj.EngResponse
	}
	existingAccountusingPhoneNumber := accountdb.FindAccountByAccountPhoneNumber(userAccountObj.AccountPhoneNumber)
	if existingAccountusingPhoneNumber.AccountPublicKey != "" && existingAccountusingPhoneNumber.AccountPublicKey != userAccountObj.AccountPublicKey && userAccountObj.AccountPhoneNumber != "" {
		responseObj := responses.FindResponseByID("32")
		return responseObj.EngResponse
	}
	return ""
}

//checkAccountbeforeRegister to check account before register
func checkAccountbeforeRegister(userAccountObj accountdb.AccountStruct) string {

	existingAccountUsingName := accountdb.FindAccountByAccountName(userAccountObj.AccountName)
	existingAccountusingEmail := accountdb.FindAccountByAccountEmail(userAccountObj.AccountEmail)
	existingAccountusingPhoneNumber := accountdb.FindAccountByAccountPhoneNumber(userAccountObj.AccountPhoneNumber)

	if existingAccountUsingName.AccountName != "" && existingAccountUsingName.AccountName == userAccountObj.AccountName {
		responseObj := responses.FindResponseByID("33")
		return responseObj.EngResponse
	}
	if existingAccountusingEmail.AccountEmail != "" && existingAccountusingEmail.AccountEmail == userAccountObj.AccountEmail && userAccountObj.AccountEmail != "" {
		responseObj := responses.FindResponseByID("31")
		return responseObj.EngResponse
	}
	if existingAccountusingPhoneNumber.AccountPhoneNumber != "" && existingAccountusingPhoneNumber.AccountPhoneNumber == userAccountObj.AccountPhoneNumber && userAccountObj.AccountPhoneNumber != "" {
		responseObj := responses.FindResponseByID("32")
		return responseObj.EngResponse
	}
	return ""
}

//checkingIfAccountExixtsBeforeRegister check account exist before register
func checkingIfAccountExixtsBeforeRegister(accountObj accountdb.AccountStruct) string {
	/*if IfAccountExistsBefore(accountObj.AccountPublicKey) {
		return "public key exists before"
	}*/
	Error := checkAccountbeforeRegister(accountObj)
	if Error != "" {
		return Error

	}

	if r, b := validateAccount(accountObj); !b {
		return r
	}
	return ""
}

// checkingIfAccountExixtsBeforeUpdating check account if exist before update
func checkingIfAccountExixtsBeforeUpdating(accountObj accountdb.AccountStruct) string {
	if !(IfAccountExistsBefore(accountObj.AccountPublicKey)) {
		responseObj := responses.FindResponseByID("142")
		return responseObj.EngResponse
	}
	s := checkAccount(accountObj)
	if s != "" {
		return s
	}
	if r, b := validateAccount(accountObj); !b {
		return r
	}
	return ""
}

// getAccountPassword get password
func getAccountPassword(AccountPublicKey string) string {
	return accountdb.FindAccountByAccountKey(AccountPublicKey).AccountPassword
}

// getAccountByEmail get AccountStruct using email
func getAccountByEmail(AccountEmail string) accountdb.AccountStruct {
	return accountdb.FindAccountByAccountEmail(AccountEmail)
}

// getAccountByPhone get account by phone
func getAccountByPhone(AccountPhoneNumber string) accountdb.AccountStruct {
	return accountdb.FindAccountByAccountPhoneNumber(AccountPhoneNumber)
}

// GetAccountByName get AccountStruct using user name
func GetAccountByName(AccountName string) accountdb.AccountStruct {
	return accountdb.FindAccountByAccountName(AccountName)
}

//AddBlockToAccount add the hashed index to the account block list
func AddBlockToAccount(AccountPublicKey string, blockIndex string, tokenID string) {
	accountObj := accountdb.FindAccountByAccountPublicKey(AccountPublicKey)
	// hashedIndex := cryptogrpghy.KeyEncrypt(globalPkg.RSAPrivate, blockIndex)
	hashedIndex := cryptogrpghy.AESEncrypt("123456789", blockIndex)

	accountObj.BlocksLst = append(accountObj.BlocksLst, hashedIndex)

	containid := ContainstokenID(accountObj.AccountTokenID, tokenID)

	if !containid {
		accountObj.AccountTokenID = append(accountObj.AccountTokenID, tokenID)
	}
	var tokenObj accountdb.TokenList
	var exist bool
	indextoken := findTokenid(tokenID, accountObj)
	if indextoken == -1 {
		tokenObj.TokenID = tokenID
		tokenObj.BlockList = append(tokenObj.BlockList, hashedIndex)
		accountObj.TokenListArr = append(accountObj.TokenListArr, tokenObj)
	} else {
		for _, blockindex := range accountObj.TokenListArr[indextoken].BlockList {
			decryptIndex := cryptogrpghy.KeyDecrypt("123456789", blockindex)
			if blockindex == decryptIndex {
				exist = true
			}
		}
		if exist == false {
			accountObj.TokenListArr[indextoken].BlockList = append(accountObj.TokenListArr[indextoken].BlockList, hashedIndex)
		}
	}

	UpdateAccount2(accountObj)
}

// AddBlockFileToAccount add file list to account
func AddBlockFileToAccount(fileobj filestorage.FileStruct, blockindex string) {
	var accountObj accountdb.AccountStruct
	accountObj = accountdb.FindAccountByAccountPublicKey(fileobj.Ownerpk)
	hashedIndex := cryptogrpghy.AESEncrypt("123456789", blockindex)
	accountObj.BlocksLst = append(accountObj.BlocksLst, hashedIndex)
	if !fileobj.Deleted {
		var filelist accountdb.FileList
		filelist.Fileid = fileobj.Fileid
		filelist.FileName = fileobj.FileName
		filelist.FileType = fileobj.FileType
		filelist.FileSize = fileobj.FileSize
		filelist.Blockindex = hashedIndex
		filelist.Filehash = fileobj.FileHash

		accountObj.Filelist = append(accountObj.Filelist, filelist)
	} else { // delete from list
		for i, item := range accountObj.Filelist {
			if item.Fileid == fileobj.Fileid {
				if len(item.PermissionList) != 0 {
					for _, pk := range item.PermissionList {
						accIndex := GetAccountByAccountPubicKey(pk).AccountIndex
						sharefile := filestorage.FindSharedfileByAccountIndex(accIndex)
						for sharefileindex, sharefileObj := range sharefile.OwnerSharefile {
							fileindex := containsfileidindex(sharefileObj.Fileid, fileobj.Fileid)
							if fileindex != -1 {
								sharefileObj.Fileid = append(sharefileObj.Fileid[:fileindex], sharefileObj.Fileid[fileindex+1:]...)
								sharefile.OwnerSharefile = append(sharefile.OwnerSharefile[:sharefileindex], sharefile.OwnerSharefile[sharefileindex+1:]...)
								if len(sharefileObj.Fileid) != 0 && len(sharefile.OwnerSharefile) >= 1 {
									sharefile.OwnerSharefile = append(sharefile.OwnerSharefile, sharefileObj)
								} else if len(sharefileObj.Fileid) != 0 && len(sharefile.OwnerSharefile) == 0 {
									sharefile.OwnerSharefile = append(sharefile.OwnerSharefile, sharefileObj)
								}
								broadcastTcp.BoardcastingTCP(sharefile, "updatesharefile", "file")

								if len(sharefile.OwnerSharefile) == 0 {
									broadcastTcp.BoardcastingTCP(sharefile, "deleteaccountindex", "file")
								}
							}
						}
					}
				}
				accountObj.Filelist = append(accountObj.Filelist[:i], accountObj.Filelist[i+1:]...)
				break
			}
		}
		// delete chunks if exists
		values := filestorage.FindChanksByFileId(fileobj.Fileid)
		for _, value := range values {
			filestorage.DeleteChunk(value.Chunkid)
		}

	}
	UpdateAccount2(accountObj)
}

//containsfileidindex tells whether a contains x.
func containsfileidindex(a []string, fileid string) int {
	for index, n := range a {
		if fileid == n {
			return index
		}
	}
	return -1
}

//findTokenid  find token id in token list arr
func findTokenid(tokenID string, accountObj accountdb.AccountStruct) int {
	for index, tokenlistobj := range accountObj.TokenListArr {
		if tokenlistobj.TokenID == tokenID {
			return index
		}
	}
	return -1
}

//ContainstokenID Contains tells whether a contains x.
func ContainstokenID(AccountTokenID []string, tokenid string) bool {
	for _, n := range AccountTokenID {
		if tokenid == n {
			return true
		}
	}
	return false
}

// GetAccountByIndex get account by index
func GetAccountByIndex(index string) accountdb.AccountStruct {
	return accountdb.FindAccountByAccountKey(index)
}

//UpdateAccount2 update account
func UpdateAccount2(accountObj accountdb.AccountStruct) string {
	oldAccount := GetAccountByAccountPubicKey(accountObj.AccountPublicKey)

	_, b := validateAccount(accountObj)
	if b && (IfAccountExistsBefore(accountObj.AccountPublicKey)) {
		// if (ifAccountExistsBefore(accountObj.AccountPublicKey)) && validateAccount(accountObj) {
		if accountdb.AccountUpdate2(accountObj) {
			AccountLastUpdatedTimestructObj := accountdb.AccountLastUpdatedTimestruct{AccountLastUpdatedTime: accountObj.AccountLastUpdatedTime, AccountIndex: accountObj.AccountIndex}
			accountdb.AccountLastUpdatedTimeDelete(oldAccount.AccountLastUpdatedTime.String())
			err := accountdb.AccountLastUpdatedTimeCreate(AccountLastUpdatedTimestructObj)

			if !err {
				errorpk.AddError("AccountLastUpdatedTimestructCreate  AccountLastUpdatedTimestruct package", "can't create AccountLastUpdatedTimestruct", "runtime error")
				return error.AddError("FindjsonFile account package", "Can't create last updated time  "+accountObj.AccountPublicKey, "hack error")

			}
			return ""
		}
		return error.AddError("UpdateAccount account package", "Check your path or object to Update AccountStruct", "logical error")
	}

	return error.AddError("FindjsonFile account package", "Can't find the account obj "+accountObj.AccountPublicKey, "hack error")
}

//SetPublicKey update the public key into the database
func SetPublicKey(accountObjc accountdb.AccountStruct) {
	if accountdb.AddBKey(accountObjc) {
		fmt.Println("public key added successfully")
	} else {
		fmt.Println("failed to add public key")
	}
}

//convertStringTolowerCaseAndtrimspace approve username , email is lowercase and trim spaces
func convertStringTolowerCaseAndtrimspace(stringphrase string) string {
	stringphrase = strings.ToLower(stringphrase)
	stringphrase = strings.TrimSpace(stringphrase)
	return stringphrase
}
