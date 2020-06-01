package ledger

import (
	"encoding/json" //read and send json data through api
	"fmt"
	"net/http" // using API request
	"os"
	"time"

	"../accountdb"
	"../admin"
	"../cryptogrpghy"
	"../globalPkg"
	"../heartbeat"
	"../logfunc"
	"../logpkg"
	"../service"
	"../systemupdate"
	"../token"
	"../tokenModule"

	account "../account" //use accounts in the ledger structure
	block "../block"     //use blocks in the ledger structure
	file "../filestorage"
	globalfinctiontransaction "../globalfinctiontransaction"
	"../responses"
	transaction "../transaction" // use transaction in ledger structure
	validator "../validator"     //use validator in ledger structure
	"github.com/BurntSushi/toml"
)

//ledger structure of the ledger
type Ledger struct {
	AccountsLstObj        []accountdb.AccountStruct
	ValidatorsLstObj      []validator.ValidatorStruct
	UnconfirmedValidators []validator.TempValidator
	ResetPassArray        []account.ResetPasswordData
	UserObjects           []account.User
	TransactionLstObj     []transaction.Transaction
	BlockchainObj         []block.BlockStruct
	AdminObj              []admin.AdminStruct
	TokenObj              []token.StructToken
	ServiceTmp            []service.ServiceStruct
	PurchasedService      []service.ServiceStruct
	LogDB                 []logpkg.LogStruct
	ValidatorMap          map[string]string
	Temprequest           []account.Tempregister //temp register requested
	UserPK                []account.SavePKStruct
	TransactionLstDb      []transaction.TransactionDB
	//	SPKS                  []account.SavePKStruct
	Chunks      []file.Chunkdb
	Sharedfiles []file.SharedFile
}

//MixedObjStruct mixed of admin obj, admin obj
type MixedObjStruct struct {
	AdminObject  admin.Admin1
	LedgerObject Ledger
}

//TmpAccount temp account
type TmpAccount struct {
	SessionDB   []accountdb.AccountSessionStruct
	EmailDB     []accountdb.AccountEmailStruct
	NameDB      []accountdb.AccountNameStruct
	PhoneDB     []accountdb.AccountPhoneNumberStruct
	LastApdated []accountdb.AccountLastUpdatedTimestruct
}

//AdminObjec admin object
var AdminObjec admin.Admin1

// GetLedger function to get the ledger of the miner
func GetLedger() Ledger {
	ledgerObj := Ledger{}
	ledgerObj.AccountsLstObj = accountdb.GetAllAccounts()

	ledgerObj.ValidatorsLstObj = validator.GetAllValidators()
	ledgerObj.UnconfirmedValidators = validator.TempValidatorlst
	// for index := range ledgerObj.ValidatorsLstObj {
	// 	if validator.CurrentValidator.ValidatorPublicKey == ledgerObj.ValidatorsLstObj[index].ValidatorPublicKey {
	// 		ledgerObj.ValidatorsLstObj[index].ValidatorPrivateKey = ""
	// 	}
	// }
	ledgerObj.BlockchainObj = block.GetBlockchain()
	ledgerObj.TransactionLstObj = transaction.GetPendingTransactions()
	ledgerObj.UserObjects = account.GetUserObjLst()
	ledgerObj.ResetPassArray = account.GetResetPasswordData()
	ledgerObj.AdminObj = admin.GetAllAdmins()
	ledgerObj.TokenObj = token.GetAllTokens()
	ledgerObj.ServiceTmp = service.GetAllservice()
	ledgerObj.PurchasedService = service.GetAllPurchusedservice()
	//alaa
	ledgerObj.ValidatorMap = globalfinctiontransaction.GetTransactionIndexTemMap()
	ledgerObj.LogDB = logfunc.GetAllLogs()
	//Temp registered requested
	//ledgerObj.Temprequest = account.GetTempRequestedRegister()
	ledgerObj.UserPK = account.GetAllsavepksave()
	ledgerObj.TransactionLstDb = transaction.GetAllTransaction() // Denta
	ledgerObj.Chunks = file.GetAllChunks()
	ledgerObj.Sharedfiles = file.GetAllSharedFile()
	return ledgerObj

}

//RemoveDatabase remove db
func RemoveDatabase() {
	if accountdb.Open {
		accountdb.DB.Close()
		// account.DBEmail.Close()
		// account.DBName.Close()
		// account.DBPublicKey.Close()
		// account.DBPhoneNo.Close()
		// account.DBLastUpdateTime.Close()
		accountdb.Open = false
	}

	if block.Open {
		block.DB.Close()
		block.Open = false
	}

	if admin.Open {
		admin.DB.Close()
		admin.Open = false
	}

	if token.Open {
		token.DB.Close()
		token.Open = false
	}
	if service.Open {
		service.DB.Close()
		service.Open = false
	}
	if logpkg.Open {
		logpkg.DB.Close()
		logpkg.Open = false
	}

	if account.Opensave {
		account.DBsave.Close()
		account.Opensave = false
	}

	if file.Open {
		file.DB.Close()
		file.Open = false
	}
	if file.Openshare {
		file.DBshare.Close()
		file.Openshare = false
	}
	fmt.Println("tessssssssssssssssssst")
	//err := os.RemoveAll("AccountStruct")
	os.RemoveAll("Database/AccountStruct")
	os.RemoveAll("Database/AdminStruct")
	os.RemoveAll("Database/BlockStruct")
	//os.RemoveAll("Database/TempAccount")
	os.RemoveAll("Database/TokenStruct")
	os.RemoveAll("Database/SessionStruct")
	os.RemoveAll("Database/Service")
	os.RemoveAll("Database/SavePKStruct")
	os.RemoveAll("Database/LoggerDB")
	os.RemoveAll("Database/SharedFile")

}

// PostLegderAPI API to set the ledger of the miner
func PostLegderAPI(w http.ResponseWriter, req *http.Request) {
	//log
	now, userIP := globalPkg.SetLogObj(req)
	logobj := logpkg.LogStruct{"_", now, userIP, "macAdress", "PostLegderAPI", "Ledger", "_", "_", "_", 0}

	recievedObj := MixedObjStruct{}
	adminReqObj := admin.Admin1{}
	ledgerReqObj := Ledger{}

	decoder := json.NewDecoder(req.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&recievedObj)
	if err != nil {
		responseObj := responses.FindResponseByID("1")
		globalPkg.SendError(w, responseObj.EngResponse)
		globalPkg.WriteLog(logobj, responseObj.EngResponse, "failed")
		return
	}
	adminReqObj = recievedObj.AdminObject
	ledgerReqObj = recievedObj.LedgerObject

	if adminReqObj == AdminObjec {
		fmt.Println("nice")
	}

	RemoveDatabase()
	for _, accountObj := range ledgerReqObj.AccountsLstObj {
		accountObj.BlocksLst = nil
		account.AddAccount(accountObj)
	}

	// encryptedValidators := validator.GetAllValidators()
	for index := range ledgerReqObj.ValidatorsLstObj {
		if validator.CurrentValidator.ValidatorIP == ledgerReqObj.ValidatorsLstObj[index].ValidatorIP {
			ledgerReqObj.ValidatorsLstObj[index].ECCPrivateKey = validator.CurrentValidator.ECCPrivateKey
			ledgerReqObj.ValidatorsLstObj[index].ECCPublicKey = validator.CurrentValidator.ECCPublicKey
		} else {
			ledgerReqObj.ValidatorsLstObj[index].ECCPublicKey = validator.CurrentValidator.ECCPublicKey
		}
		validator.CreateValidator(&(ledgerReqObj.ValidatorsLstObj[index]))
	}
	validator.ValidatorsLstObj = validator.GetAllValidatorsDecrypted()
	// validator.ValidatorsLstObj = ledgerReqObj.ValidatorsLstObj
	// for _, transactionObj := range ledgerReqObj.TransactionLstObj {
	// 	transaction.AddTransaction(transactionObj)
	// }
	for _, transactionObj := range ledgerReqObj.TransactionLstObj {
		transactionObj.AddTransaction()
	}
	for _, transactionObj := range ledgerReqObj.TransactionLstDb {
		transactionObj.AddTransactiondb()
	}
	for _, validatorObject := range ledgerReqObj.UnconfirmedValidators {
		validator.TempValidatorlst = append(validator.TempValidatorlst, validatorObject)
	}
	for _, blockObj := range ledgerReqObj.BlockchainObj {
		(&blockObj).AddBlock(true)
	}

	account.SetResetPasswordData(ledgerReqObj.ResetPassArray)
	account.SetUserObjLst(ledgerReqObj.UserObjects)

	for _, adminObj := range ledgerReqObj.AdminObj {
		admin.CreateAdmin(adminObj)
	}

	for _, tokenobj := range ledgerReqObj.TokenObj {
		tokenModule.AddToken(tokenobj)
	}
	for _, serviceObj := range ledgerReqObj.PurchasedService {
		(&serviceObj).AddAndUpdateServiceObj()
	}
	service.SetserviceTemp(ledgerReqObj.ServiceTmp)

	//alaa
	globalfinctiontransaction.SetTransactionIndexTemMap(ledgerReqObj.ValidatorMap) //globalfinctiontransaction.GetTransactionIndexTemMap()
	service.SetserviceTemp(ledgerReqObj.ServiceTmp)
	for _, UserPKObj := range ledgerReqObj.UserPK {
		account.SavePKAddress(UserPKObj)
	}

	for _, logdb := range ledgerReqObj.LogDB {
		logpkg.RecordLog(logdb)
	}

	for _, shareFileObj := range ledgerReqObj.Sharedfiles {
		file.AddSharedFile(shareFileObj)
	}
	//get permission list to append in owner pk
	GetPermissionByOwnerpk()

	responseObj := responses.FindResponseByID("49")
	globalPkg.SendResponseMessage(w, responseObj.EngResponse)
	globalPkg.WriteLog(logobj, responseObj.EngResponse, "success")

}

//PostLedger post ledger
func PostLedger(ledgerReqObj Ledger) {

	RemoveDatabase()
	for _, accountObj := range ledgerReqObj.AccountsLstObj {
		accountObj.BlocksLst = nil
		account.AddAccount(accountObj)
	}

	decryptedValidators := ledgerReqObj.ValidatorsLstObj
	for index := range decryptedValidators {
		// for index := range ledgerReqObj.ValidatorsLstObj {
		if validator.CurrentValidator.ValidatorIP == ledgerReqObj.ValidatorsLstObj[index].ValidatorIP {
			ledgerReqObj.ValidatorsLstObj[index].ECCPrivateKey = validator.CurrentValidator.ECCPrivateKey
			ledgerReqObj.ValidatorsLstObj[index].ECCPublicKey = validator.CurrentValidator.ECCPublicKey
		} else {
			// ledgerReqObj.ValidatorsLstObj[index].ECCPrivateKey = new(ecdsa.PrivateKey)
		}
		fmt.Println("index : ", index)
		fmt.Println("validator.CurrentValidator. : ", validator.CurrentValidator.ValidatorIP)
		validator.CreateValidator(&(ledgerReqObj.ValidatorsLstObj[index]))

	}

	validator.ValidatorsLstObj = validator.GetAllValidatorsDecrypted()

	for _, transactionObj := range ledgerReqObj.TransactionLstObj {
		transactionObj.AddTransaction()
	}
	for _, blockObj := range ledgerReqObj.BlockchainObj {
		(&blockObj).AddBlock(true)
	}

	account.SetResetPasswordData(ledgerReqObj.ResetPassArray)
	account.SetUserObjLst(ledgerReqObj.UserObjects)
	for _, adminObj := range ledgerReqObj.AdminObj {
		admin.CreateAdmin(adminObj)
	}
	for _, tokenobj := range ledgerReqObj.TokenObj {
		tokenModule.AddToken(tokenobj)
	}
	for _, serviceObj := range ledgerReqObj.PurchasedService {
		(&serviceObj).AddAndUpdateServiceObj()
	}
	service.SetserviceTemp(ledgerReqObj.ServiceTmp)

	for _, validatorObject := range ledgerReqObj.UnconfirmedValidators {
		validator.TempValidatorlst = append(validator.TempValidatorlst, validatorObject)
	}

	for _, UserPKObj := range ledgerReqObj.UserPK {
		account.SavePKAddress(UserPKObj)
	}
	for _, logdb := range ledgerReqObj.LogDB {
		logpkg.RecordLog(logdb)
	}
	globalfinctiontransaction.SetTransactionIndexTemMap(ledgerReqObj.ValidatorMap)

	for _, shareFileObj := range ledgerReqObj.Sharedfiles {
		file.AddSharedFile(shareFileObj)
	}
	//get permission list to append in owner pk
	GetPermissionByOwnerpk()

}

// GetPermissionByOwnerpk get permission list to append in owner pk
func GetPermissionByOwnerpk() {

	accountList := accountdb.GetAllAccounts()
	for _, accountObj := range accountList {
		var permissionOwnerMap map[string][]string
		permissionOwnerMap = make(map[string][]string)
		sharedfileOwned := file.FindSharedFileByownerpk(accountObj.AccountPublicKey)
		if len(sharedfileOwned) != 0 {
			for _, sharefileobj := range sharedfileOwned {
				for _, ownersharefileObj := range sharefileobj.OwnerSharefile {
					if ownersharefileObj.OwnerPublicKey == accountObj.AccountPublicKey {
						for _, fileid := range ownersharefileObj.Fileid {
							pk := account.GetAccountByIndex(sharefileobj.AccountIndex).AccountPublicKey
							if pk != "" {
								permissionOwnerMap[fileid] = append(permissionOwnerMap[fileid], pk)
							}
						}
					}
				}
			}

			for i, fileObj := range accountObj.Filelist {
				if permLst, ok := permissionOwnerMap[fileObj.Fileid]; ok {
					accountObj.Filelist[i].PermissionList = permLst
				}
			}
			account.UpdateAccount2(accountObj)
		}
	}
}

//GetLegderAPI API to get the ledger from the miner
func GetLegderAPI(w http.ResponseWriter, req *http.Request) {
	//log
	now, userIP := globalPkg.SetLogObj(req)
	logobj := logpkg.LogStruct{"_", now, userIP, "macAdress", "GetLegderAPI", "ledger", "_", "_", "_", 0}

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
		sendJSON, _ := json.Marshal(GetLedger())
		globalPkg.SendResponse(w, sendJSON)
		globalPkg.WriteLog(logobj, "get ledger success", "success")

		var UpdateDataObj systemupdate.UpdateData
		toml.DecodeFile("./config.toml", &UpdateDataObj)
		heartbeat.SendUpdateHeartBeat(UpdateDataObj.Updatestruct.Updateversion, UpdateDataObj.Updatestruct.Updateurl)
	} else {
		responseObj := responses.FindResponseByID("2")
		globalPkg.SendError(w, responseObj.EngResponse)
		globalPkg.WriteLog(logobj, responseObj.EngResponse, "failed")

	}

}

//GetAllTmpAccountDB get all temp account
func GetAllTmpAccountDB() TmpAccount {
	ledgerObj := TmpAccount{}
	ledgerObj.EmailDB = accountdb.GetAllEmails()
	ledgerObj.NameDB = accountdb.GetAllNames()
	ledgerObj.PhoneDB = accountdb.GetAllPhones()
	ledgerObj.SessionDB = accountdb.GetAllSessions()
	ledgerObj.LastApdated = accountdb.GetAllLastTimes()
	return ledgerObj
}

//GetTmpAccountDB get temp account from db
func GetTmpAccountDB(w http.ResponseWriter, req *http.Request) {
	//log
	now, userIP := globalPkg.SetLogObj(req)
	logobj := logpkg.LogStruct{"", now, userIP, "macAdress", "GetLegderAPI", "ledger", "", "", "", 0}

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
		sendJSON, _ := json.Marshal(GetAllTmpAccountDB())
		globalPkg.SendResponse(w, sendJSON)
		globalPkg.WriteLog(logobj, "get GetAllTmpAccountDB success", "success")
	} else {
		responseObj := responses.FindResponseByID("2")
		globalPkg.SendError(w, responseObj.EngResponse)
		globalPkg.WriteLog(logobj, responseObj.EngResponse, "failed")
	}

}

//PostLedgerTmpAccount post ledger temp account
func PostLedgerTmpAccount(ledgerReqObj TmpAccount) {

	RemoveDatabase()

	//validator.CurrentValidator.ValidatorRegisterTime, _ = time.Parse("2006-01-02 03:04:05 PM -0000", time.Now().UTC().Format("2006-01-02 03:04:05 PM -0000"))

	for _, sessionObj := range ledgerReqObj.SessionDB {
		accountdb.AddSessionIdStruct(sessionObj)
	}
	// for _, sessionObj := range ledgerReqObj.NameDB {
	// 	accountdb.AddSessionIdStruct(sessionObj)
	// }

}

//Blockchain struct is save accounts and blocks
type Blockchain struct {
	AccountList []accountdb.AccountStruct
	BlockList   []block.BlockStruct
}

//SaveBlockchainInTemp save blockchain in temp
func SaveBlockchainInTemp() Blockchain {

	blockchainObj := Blockchain{}
	blockchainObj.AccountList = accountdb.GetAllAccounts()
	blockchainObj.BlockList = block.GetBlockchain()

	return blockchainObj
}

//CompareBlockchain compare blockchain check block list in account as block
func CompareBlockchain() {
	for {

		time.Sleep(time.Hour * 24)
		// time.Sleep(time.Second * 5)
		fmt.Println("-----------     Start  check blockchain data")
		blockchainObj := SaveBlockchainInTemp()
		var pksender string
		var pkreciever []string
		var blocklstdecrypt []string
		var decryptblockIndexsender, decryptblockIndexreciever []string
		var tokenidSender, tokenidReciever string

		//for loop on blocks and then loop on transactions to get sender address and reciever address
		for _, blockobj := range blockchainObj.BlockList {

			for _, transactionObj := range blockobj.BlockTransactions {
				if transactionObj.TransactionInput != nil {
					for _, transactionInputObj := range transactionObj.TransactionInput {
						pksender = transactionInputObj.SenderPublicKey
						tokenidSender = transactionInputObj.TokenID
					}
				}

				for _, transactionOutPutObj := range transactionObj.TransactionOutPut {
					pkreciever = append(pkreciever, transactionOutPutObj.RecieverPublicKey)
					tokenidReciever = transactionOutPutObj.TokenID
				}
			}

			//for loop on account and then loop blocklst on account to decrypt the index and compare it to block index
			for _, accountobj := range blockchainObj.AccountList {
				// decrypt blocklst of account
				for _, blocklstindex := range accountobj.BlocksLst {
					decryptIndex := cryptogrpghy.KeyDecrypt("123456789", blocklstindex)
					containblockind := ContainsBlockindex(blocklstdecrypt, decryptIndex)
					if !containblockind {
						blocklstdecrypt = append(blocklstdecrypt, decryptIndex)
					}
				}

				containindex := ContainsBlockindex(blocklstdecrypt, blockobj.BlockIndex)

				if accountobj.AccountPublicKey == pksender && accountobj.AccountPublicKey != "" {
					//check for blocklst account sender
					if !containindex {
						fmt.Println("===================  Sender blocklist  ===================", containindex, "     ", blockobj.BlockIndex)
						hashedIndex := cryptogrpghy.AESEncrypt("123456789", blockobj.BlockIndex)
						accountobj.BlocksLst = append(accountobj.BlocksLst, hashedIndex)
						account.UpdateAccount2(accountobj)
					}
					//check for token id exist for sender
					containTokenid := account.ContainstokenID(accountobj.AccountTokenID, tokenidSender)
					if !containTokenid {
						fmt.Println("===================  Sender tokenid  =================== ", tokenidSender)
						accountobj.AccountTokenID = append(accountobj.AccountTokenID, tokenidSender)
						account.UpdateAccount2(accountobj)
					}
					//check data in tokenlist array is tokenid and blocklst
					var existblockindex, existtokenid bool
					tokenlist := accountdb.TokenList{}
					for _, tokenlistobj := range accountobj.TokenListArr {
						if tokenlistobj.TokenID == tokenidSender {
							existtokenid = true
							// decrypt blocklst in tokenlst
							for _, blockindexObj := range tokenlistobj.BlockList {
								decryptblockIndex := cryptogrpghy.KeyDecrypt("123456789", blockindexObj)
								containind := ContainsBlockindex(decryptblockIndexsender, decryptblockIndex)
								if !containind {
									decryptblockIndexsender = append(decryptblockIndexsender, decryptblockIndex)
								}
							}
							// compare blockindex in tokenlst and in blocklst of accunt
							for _, blocklisttoken := range decryptblockIndexsender {
								for _, blocklstaccount := range blocklstdecrypt {
									if blocklisttoken == blocklstaccount {
										existblockindex = true
										break
									}
								}
							}
						}
						break
					}

					//if token id exist in token list
					if existtokenid == false {
						fmt.Println("===================  Sender tokenid in token list =================== ", tokenidSender)
						tokenlist.TokenID = tokenidSender
						accountobj.TokenListArr = append(accountobj.TokenListArr, tokenlist)
						account.UpdateAccount2(accountobj)
					}
					//check block index exist in block list in token list
					if existblockindex == false {
						//get all blocklist account and find blockobj by index and loop for transaction input
						// and token id equal token id with sender append block index in tokenlist blocklst
						tokenlist.TokenID = tokenidSender
						for _, blocklistindex := range accountobj.BlocksLst {
							decryptIndexBlock := cryptogrpghy.KeyDecrypt("123456789", blocklistindex)
							blockobject := block.GetBlockInfoByID(decryptIndexBlock)
							for _, transactionObj := range blockobject.BlockTransactions {
								for _, transactionInputObject := range transactionObj.TransactionInput {
									if tokenidSender == transactionInputObject.TokenID {
										//check if exist in blockblist in tokenlist
										containsblockindex := ContainsEncryptBlockindex(tokenlist.BlockList, blocklistindex)
										if !containsblockindex {
											fmt.Println("===================  Sender block list in token list  =================== ", blocklistindex)
											tokenlist.BlockList = append(tokenlist.BlockList, blocklistindex)
											accountobj.TokenListArr = append(accountobj.TokenListArr, tokenlist)
											account.UpdateAccount2(accountobj)
										}
									}
								}
							}
						}
					}

				}

				if accountobj.AccountPublicKey == pkreciever[0] && accountobj.AccountPublicKey != "" {
					//check for blocklst account reciever
					if !containindex {
						fmt.Println("===================  Reciever blocklist ===================", containindex, "     ", blockobj.BlockIndex)
						hashedIndex := cryptogrpghy.AESEncrypt("123456789", blockobj.BlockIndex)
						accountobj.BlocksLst = append(accountobj.BlocksLst, hashedIndex)
						account.UpdateAccount2(accountobj)
					}
					//check for token id exist for reciever
					containTokenid := account.ContainstokenID(accountobj.AccountTokenID, tokenidReciever)
					if !containTokenid {
						fmt.Println("===================  reciever tokenid  =================== ", tokenidReciever)
						accountobj.AccountTokenID = append(accountobj.AccountTokenID, tokenidReciever)
						account.UpdateAccount2(accountobj)
					}

					//check data in tokenlist array is tokenid and blocklst
					var existblockindex, existtokenid bool
					tokenlist := accountdb.TokenList{}
					for _, tokenlistobj := range accountobj.TokenListArr {
						if tokenlistobj.TokenID == tokenidReciever {
							existtokenid = true
							// decrypt blocklst in tokenlst
							for _, blockindexObj := range tokenlistobj.BlockList {
								decryptblockIndex := cryptogrpghy.KeyDecrypt("123456789", blockindexObj)
								containind := ContainsBlockindex(decryptblockIndexreciever, decryptblockIndex)
								if !containind {
									decryptblockIndexreciever = append(decryptblockIndexreciever, decryptblockIndex)
								}
							}
							// compare blockindex in tokenlst and in blocklst of accunt
							for _, blocklisttoken := range decryptblockIndexreciever {
								for _, blocklstaccount := range blocklstdecrypt {
									if blocklisttoken == blocklstaccount {
										existblockindex = true
										break
									}
								}
							}
						}
						break
					}

					//if token id exist in token list
					if existtokenid == false {
						fmt.Println("===================  reciever tokenid in token list =================== ", tokenidReciever)
						tokenlist.TokenID = tokenidReciever
						accountobj.TokenListArr = append(accountobj.TokenListArr, tokenlist)
						account.UpdateAccount2(accountobj)
					}
					//check block index exist in block list in token list
					if existblockindex == false {
						//get all blocklist account and find blockobj by index and loop for transaction input
						// and token id equal token id with sender append block index in tokenlist blocklst
						for _, blocklistindex := range accountobj.BlocksLst {
							decryptIndexBlock1 := cryptogrpghy.KeyDecrypt("123456789", blocklistindex)
							blockobject := block.GetBlockInfoByID(decryptIndexBlock1)
							for _, transactionObj := range blockobject.BlockTransactions {
								for _, TransactionOutPutObject := range transactionObj.TransactionOutPut {
									if tokenidReciever == TransactionOutPutObject.TokenID {
										//check if exist in blockblist in tokenlist
										containsblockindex := ContainsEncryptBlockindex(tokenlist.BlockList, blocklistindex)
										if !containsblockindex {
											fmt.Println("===================  Reciever block list in token list  =================== ", blocklistindex)
											tokenlist.BlockList = append(tokenlist.BlockList, blocklistindex)
											accountobj.TokenListArr = append(accountobj.TokenListArr, tokenlist)
											account.UpdateAccount2(accountobj)
										}
									}
								}
							}
						}
					}

				}
			}
		}
		fmt.Println("  -----------------     End Ckeck blockchain data")
	}
}

//ContainsBlockindex Contains tells whether a contains x.  decrypt index
func ContainsBlockindex(AccountBlocklstdecrypt []string, blockindex string) bool {
	for _, n := range AccountBlocklstdecrypt {
		if blockindex == n {
			return true
		}
	}
	return false
}

//ContainsEncryptBlockindex Contains tells whether a contains x.  decrypt index
func ContainsEncryptBlockindex(AccountBlocklstEncrypt []string, blockindex string) bool {
	for _, n := range AccountBlocklstEncrypt {
		if blockindex == n {
			return true
		}
	}
	return false
}
