package account

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"../broadcastTcp"
	"../filestorage"

	"../cryptogrpghy"

	error "../errorpk"
	globalpkg "../globalPkg"

	//globalPkg "github.com/tensor-programming/mining/../globalPkg"
	"../accountdb"
)

type errorMess struct {
	ErrorMessage string
}

type User struct {
	Account           accountdb.AccountStruct
	Oldpassword       string
	Reson             string
	CurrentTime       time.Time
	Confirmation_code string
	TextSearch        string
	Method            string
	PathApi           string
}

type searchResponse struct {
	UserName  string
	PublicKey string
}

/*----------------- function to get Account by public key -----------------*/
func GetAccountByAccountPubicKey(AccountPublicKey string) accountdb.AccountStruct {
	return accountdb.FindAccountByAccountPublicKey(AccountPublicKey)
}

/*----------------- function to check if account exists or not -----------------*/
func ifAccountExistsBefore(AccountPublicKey string) bool {
	// fmt.Println("  pk   1111   ", AccountPublicKey, "   pk")
	if (accountdb.FindAccountByAccountPublicKey(AccountPublicKey)).AccountPublicKey == "" {
		// fmt.Println("  LLLLL   false  ")
		return false //not exist
	}
	return true
}

/*----------------- function to save an account on json file -----------------*/
func AddAccount(accountObj accountdb.AccountStruct) string {

	// if !(ifAccountExistsBefore(accountObj.AccountPublicKey)) && validateAccount(accountObj) {
	if validateAccount(accountObj) {
		if accountdb.AccountCreate(accountObj) {
			return ""
		} else {
			return error.AddError("AddAccount account package", "Check your path or object to Add AccountStruct", "logical error")
		}
	}
	return error.AddError("AddAccount account package", "The account is already exists "+accountObj.AccountPublicKey, "hack error")

}

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

/*----------------- function to update an account on json file -----------------*/
func UpdateAccount(accountObj accountdb.AccountStruct) string {
	// if (ifAccountExistsBefore(accountObj.AccountPublicKey)) && validateAccount(accountObj) {
	if validateAccount(accountObj) {

		if accountdb.AccountUpdateUsingTmp(accountObj) {
			return ""
		} else {
			return error.AddError("UpdateAccount account package", "Check your path or object to Update AccountStruct", "logical error")

		}
		fmt.Println("iam update")
	}
	return error.AddError("FindjsonFile account package", "Can't find the account obj "+accountObj.AccountPublicKey, "hack error")

}

/*----------------- function to validate the account before register  -----------------*/
func validateAccount(accountObj accountdb.AccountStruct) bool {
	if len(accountObj.AccountName) < 8 || len(accountObj.AccountName) > 30 || (len(accountObj.AccountPassword) != 64) || len(accountObj.AccountAddress) < 5 || len(accountObj.AccountAddress) > 100 {
		fmt.Println("1")
		return false
	}

	re := regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")
	if !re.MatchString(accountObj.AccountEmail) && accountObj.AccountEmail != "" {
		fmt.Println("2")
		return false
	}
	return true
}

/////////////////*******func Add by Aya to get publickey using Any string ----

func getPublicKeyUsingString(Key string) string {

	existingAccountUsingName := accountdb.FindAccountByAccountName(Key)

	existingAccountusingEmail := accountdb.FindAccountByAccountEmail(Key)
	existingAccountusingPhoneNumber := accountdb.FindAccountByAccountPhoneNumber(Key)
	if existingAccountUsingName.AccountPublicKey != "" {
		return existingAccountUsingName.AccountPublicKey
	}
	if existingAccountusingEmail.AccountPublicKey != "" {
		return existingAccountusingEmail.AccountPublicKey
	}
	if existingAccountusingPhoneNumber.AccountPublicKey != "" {
		return existingAccountusingPhoneNumber.AccountPublicKey
	}

	return ""

}

/*-------------FUNCTION TO CHECK Acoount----*/

func checkAccount(userAccountObj accountdb.AccountStruct) string {

	existingAccountUsingName := accountdb.FindAccountByAccountName(userAccountObj.AccountName)
	existingAccountusingEmail := accountdb.FindAccountByAccountEmail(userAccountObj.AccountEmail)
	existingAccountusingPhoneNumber := accountdb.FindAccountByAccountPhoneNumber(userAccountObj.AccountPhoneNumber)

	if existingAccountUsingName.AccountPublicKey != "" && existingAccountUsingName.AccountPublicKey != userAccountObj.AccountPublicKey {

		return "UserName Found"
	}
	if existingAccountusingEmail.AccountPublicKey != "" && existingAccountusingEmail.AccountPublicKey != userAccountObj.AccountPublicKey && userAccountObj.AccountEmail != "" {
		fmt.Println("Email found", existingAccountusingEmail.AccountEmail, "  ", userAccountObj.AccountEmail)
		return "Email found"
	}
	if existingAccountusingPhoneNumber.AccountPublicKey != "" && existingAccountusingPhoneNumber.AccountPublicKey != userAccountObj.AccountPublicKey && userAccountObj.AccountPhoneNumber != "" {
		return "Phone Found "
	}
	return ""
}

func checkAccountbeforeRegister(userAccountObj accountdb.AccountStruct) string {

	existingAccountUsingName := accountdb.FindAccountByAccountName(userAccountObj.AccountName)
	existingAccountusingEmail := accountdb.FindAccountByAccountEmail(userAccountObj.AccountEmail)
	existingAccountusingPhoneNumber := accountdb.FindAccountByAccountPhoneNumber(userAccountObj.AccountPhoneNumber)

	if existingAccountUsingName.AccountName != "" && existingAccountUsingName.AccountName == userAccountObj.AccountName {

		return "UserName Found"
	}
	if existingAccountusingEmail.AccountEmail != "" && existingAccountusingEmail.AccountEmail == userAccountObj.AccountEmail && userAccountObj.AccountEmail != "" {
		fmt.Println("Email found", existingAccountusingEmail.AccountEmail, "  ", userAccountObj.AccountEmail)
		return "Email found"
	}
	if existingAccountusingPhoneNumber.AccountPhoneNumber != "" && existingAccountusingPhoneNumber.AccountPhoneNumber == userAccountObj.AccountPhoneNumber && userAccountObj.AccountPhoneNumber != "" {
		return "Phone Found "
	}
	return ""
}

/*-------------integrat account module with miner------*/
/*-------------check Befor Add------*/

func checkingIfAccountExixtsBeforeRegister(accountObj accountdb.AccountStruct) string {

	/*if IfAccountExistsBefore(accountObj.AccountPublicKey) {
		return "public key exists before"
	}*/
	Error := checkAccountbeforeRegister(accountObj)
	if Error != "" {
		return Error

	}

	if !validateAccount(accountObj) {
		return "please, check your data"
	}
	return ""

}
func checkingIfAccountExixtsBeforeAdd(accountObj accountdb.AccountStruct) string {

	//IfAccountExistsBefore(accountObj.AccountPublicKey)
	if ifAccountExistsBefore(accountObj.AccountPublicKey) {
		return "the publick key exist before "
	}

	return ""

}

/*-------------check Befor update------*/
func checkingIfAccountExixtsBeforeUpdating(accountObj accountdb.AccountStruct) string {
	if !(ifAccountExistsBefore(accountObj.AccountPublicKey)) {
		return "Please Check Your data to help me to find your account"
	}
	s := checkAccount(accountObj)
	if s != "" {
		return s
	}
	if !validateAccount(accountObj) {
		return "please, check your data"
	}
	return ""

}

/*-------------getAccountPassword-------*/
func getAccountPassword(AccountPublicKey string) string {
	return accountdb.FindAccountByAccountKey(AccountPublicKey).AccountPassword
}

/*-------------get AccountStruct using publicKey-------*/
// func getAccountByPublicKey(AccountPublicKey string) AccountStruct {
// 	return findAccountByAccountPublicKey(AccountPublicKey)
// }

// func GetAccountByPublicKey(AccountPublicKey string) AccountStruct {
// 	return findAccountByAccountPublicKey(AccountPublicKey)
// }

/*-------------get AccountStruct using email-------*/
func getAccountByEmail(AccountEmail string) accountdb.AccountStruct {
	return accountdb.FindAccountByAccountEmail(AccountEmail)
}
func getAccountByPhone(AccountPhoneNumber string) accountdb.AccountStruct {
	return accountdb.FindAccountByAccountPhoneNumber(AccountPhoneNumber)
}

/*-------------get AccountStruct using user name-------*/
func getAccountByName(AccountName string) accountdb.AccountStruct {
	return accountdb.FindAccountByAccountName(AccountName)
}

/*-------------get AccountStruct using user name-------*/
func GetAccountByName(AccountName string) accountdb.AccountStruct {
	return accountdb.FindAccountByAccountName(AccountName)
}

/*----------------AddBlockToAnaccount-----*/
func AddBlockToAccount(AccountPublicKey string, blockIndex string, tokenID string) {
	accountObj := accountdb.FindAccountByAccountPublicKey(AccountPublicKey)
	hashedIndex := cryptogrpghy.KeyEncrypt(globalpkg.EncryptAccount, blockIndex)
	accountObj.BlocksLst = append(accountObj.BlocksLst, hashedIndex)

	containid := ContainstokenID(accountObj.AccountTokenID, tokenID)
	if !containid {
		accountObj.AccountTokenID = append(accountObj.AccountTokenID, tokenID)
	}
	UpdateAccount2(accountObj)
}

func AddBlockFileToAccount(fileobj filestorage.FileStruct, blockindex string) {
	var accountObj accountdb.AccountStruct
	accountObj = accountdb.FindAccountByAccountPublicKey(fileobj.Ownerpk)
	hashedIndex := cryptogrpghy.KeyEncrypt(globalpkg.EncryptAccount, blockindex)
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

//ContainstokenID Contains tells whether a contains x.
func ContainstokenID(AccountTokenID []string, tokenid string) bool {
	for _, n := range AccountTokenID {
		if tokenid == n {
			return true
		}
	}
	return false
}

//containsfileid tells whether a contains x.
func containsfileidindex(a []string, fileid string) int {
	for index, n := range a {
		if fileid == n {
			return index
		}
	}
	return -1
}

func GetAccountByIndex(index string) accountdb.AccountStruct {
	return accountdb.FindAccountByAccountKey(index)
}

func UpdateAccount2(accountObj accountdb.AccountStruct) string {

	if (ifAccountExistsBefore(accountObj.AccountPublicKey)) && validateAccount(accountObj) {

		if accountdb.AccountUpdate2(accountObj) {

			return ""
		} else {
			return error.AddError("UpdateAccount account package", "Check your path or object to Update AccountStruct", "logical error")

		}
		fmt.Println("iam update22")
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

//----------to convert User Object to Account Object----
//convertUserTOAccount Deleted
