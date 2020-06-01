package startPkg

import (
	"crypto/rand"
	"encoding/json"

	"../responses"

	ecc "../ECC"
	"github.com/ego008/xrsa"
	"github.com/tealeg/xlsx"

	"fmt"
	"rsapk"

	"../globalfinctiontransaction"

	//"time"

	"../account"
	"../accountdb"
	"../admin"
	"../cryptogrpghy"
	"../globalPkg"
	"../heartbeat"
	"../ledger"
	"../systemupdate"
	"../token"
	transaction "../transaction"
	"../validator"
)

type Config struct {
	GlobalData   globalPkg.GlobalVariables
	Server       server
	Updatestruct systemupdate.Updatestruct
}

type server struct {
	Ip                string
	Ips               []string
	PrivIP            string
	Port              string
	SoketPort         string
	DigitalwalletIp   string
	Digitalwalletport string
	UserName          string
	Password          string //0000_1111 hashed
	UserName2         string
	Password2         string
	PublicKey         string
	PrivateKey        string
	FirstMiner        bool
	InitialStakeCoin  float64
	InitialMinerCoins float64
}

var Conf Config
var Gkeys_list []string // for check duplicated keys

func is_duplicated_key(key string) bool {
	if contains(key) {
		return true
	}
	return false
}
func contains(e string) bool {
	for _, a := range Gkeys_list {
		if a == e {
			return true
		}
	}
	return false
}

func Timp() int {
	return len(accountdb.GetAllAccounts())
}

func InitTheaccount() accountdb.AccountStruct {

	if account.GetAccountByName(Conf.Server.UserName).AccountName == "" {

		readFileResponse() // load responses from file to database

		//var currentIndex1 = ""
		//currentIndex1 = account.NewIndex()
		bitSize := 1024
		reader := rand.Reader
		key, err := rsapk.GenerateKey(reader, bitSize)
		// save pk and address in db
		var savePKObj account.SavePKStruct

		cryptogrpghy.CheckError(err)
		Privatekey := cryptogrpghy.GetPrivatePEMKey(key)
		PublicKey := cryptogrpghy.GetPublicPEMKey(key.PublicKey)
		pk2 := []byte(PublicKey)
		address := cryptogrpghy.Address(pk2)
		add := string(address)

		// save pk, address
		savePKObj.Publickey = PublicKey
		savePKObj.Address = add
		savePKObj.CurrentTime = globalPkg.UTCtime()
		savePKObj.Index = "000000000000000000000000000000"
		account.SavePKAddress(savePKObj)
		Gkeys_list = append(Gkeys_list, Privatekey)
		Gkeys_list = append(Gkeys_list, PublicKey)
		var accountObj accountdb.AccountStruct
		accountObj.AccountInitialUserName = "inovatian"
		accountObj.AccountName = "inovatian"
		accountObj.AccountPassword = Conf.Server.Password
		accountObj.AccountInitialPassword = Conf.Server.Password
		accountObj.AccountIndex = "000000000000000000000000000000"
		accountObj.AccountEmail = "hatim@inovatian.com"
		accountObj.AccountAddress = "Cairo,Egypt"
		accountObj.AccountPrivateKey = Privatekey
		accountObj.AccountPublicKey = string(address)
		accountObj.AccountStatus = true
		accountObj.AccountRole = "admin"
		accountObj.AccountLastUpdatedTime = globalPkg.UTCtime()
		account.AddAccount(accountObj)

		firstTokenID, _ := globalPkg.ConvertIntToFixedLengthString(0, globalPkg.GlobalObj.TokenIDStringFixedLength)
		token.TokenCreate(
			token.StructToken{
				TokenID: firstTokenID, TokensTotalSupply: (Conf.Server.InitialMinerCoins), TokenName: "InoToken",
				TokenValue: 1.0,
			},
		)

		var transactionObj transaction.Transaction
		transactionObj.TransactionTime = globalPkg.UTCtime()
		transactionObj.TransactionOutPut = append(transactionObj.TransactionOutPut, transaction.TXOutput{
			OutPutValue: (Conf.Server.InitialMinerCoins), RecieverPublicKey: accountObj.AccountPublicKey,
			TokenID: firstTokenID,
		})
		transactionObj.Validator = validator.CurrentValidator.ValidatorIP
		transactionObj.TransactionID = ""
		transactonhash := globalPkg.CreateHash(transactionObj.TransactionTime, fmt.Sprintf("%s", transactionObj), 3)
		transactionObj.TransactionID = globalfinctiontransaction.FirstRun(transactonhash)
		transactionObj.AddTransaction()

		//currentIndex1 = "000000000000000000000000000001"
		key, err = rsapk.GenerateKey(reader, bitSize)
		cryptogrpghy.CheckError(err)
		Privatekey = cryptogrpghy.GetPrivatePEMKey(key)
		PublicKey = cryptogrpghy.GetPublicPEMKey(key.PublicKey)
		pk2 = []byte(PublicKey)
		address = cryptogrpghy.Address(pk2)
		add = string(address)
		// save pk, address

		savePKObj.Publickey = PublicKey
		savePKObj.Address = string(address)

		for is_duplicated_key(Privatekey) || is_duplicated_key(add) {
			// fmt.Println("Duplicated Key (#)")
			key, err = rsapk.GenerateKey(reader, bitSize)
			Privatekey = cryptogrpghy.GetPrivatePEMKey(key)
			PublicKey = cryptogrpghy.GetPublicPEMKey(key.PublicKey)
			pk2 = []byte(PublicKey)
			address = cryptogrpghy.Address(pk2)
			// save pk, address
			savePKObj.Publickey = PublicKey
			savePKObj.Address = string(address)
		}
		savePKObj.CurrentTime = globalPkg.UTCtime()
		savePKObj.Index = "000000000000000000000000000001"
		account.SavePKAddress(savePKObj)
		Gkeys_list = append(Gkeys_list, Privatekey)
		Gkeys_list = append(Gkeys_list, PublicKey)
		// fmt.Println("Add Key (#)")
		accountObj.AccountInitialUserName = "inovatian-fees"
		accountObj.AccountName = "inovatian-fees"
		accountObj.AccountPassword = "EDBF8506E6E2BE91E76AD20406C36807011A6DBBB190046427D00E2D30E1D773"
		// inovatian-fees
		accountObj.AccountInitialPassword = "EDBF8506E6E2BE91E76AD20406C36807011A6DBBB190046427D00E2D30E1D773"
		accountObj.AccountIndex = "000000000000000000000000000001"
		accountObj.AccountEmail = "hatim.fees@inovatian.com"
		accountObj.AccountAddress = "Cairo,Egypt"
		accountObj.AccountPublicKey = string(address)
		accountObj.AccountPrivateKey = Privatekey
		accountObj.AccountStatus = true
		accountObj.AccountRole = "admin"
		accountObj.AccountLastUpdatedTime = globalPkg.UTCtime()
		account.AddAccount(accountObj)

		//currentIndex1 = "000000000000000000000000000002"
		key, err = rsapk.GenerateKey(reader, bitSize)
		cryptogrpghy.CheckError(err)
		Privatekey = cryptogrpghy.GetPrivatePEMKey(key)
		PublicKey = cryptogrpghy.GetPublicPEMKey(key.PublicKey)
		pk2 = []byte(PublicKey)
		address = cryptogrpghy.Address(pk2)
		add = string(address)
		// save pk, address
		savePKObj.Publickey = PublicKey
		savePKObj.Address = string(address)

		// check depulicated keys
		for is_duplicated_key(Privatekey) || is_duplicated_key(add) {
			fmt.Println("Duplicated Key (#)")
			key, err = rsapk.GenerateKey(reader, bitSize)
			Privatekey = cryptogrpghy.GetPrivatePEMKey(key)
			PublicKey = cryptogrpghy.GetPublicPEMKey(key.PublicKey)
			pk2 = []byte(PublicKey)
			address = cryptogrpghy.Address(pk2)
			// save pk, address
			savePKObj.Publickey = PublicKey
			savePKObj.Address = string(address)
			//	account.SavePKAddress(savePKObj)
		}
		savePKObj.CurrentTime = globalPkg.UTCtime()
		savePKObj.Index = "000000000000000000000000000002"
		account.SavePKAddress(savePKObj)
		Gkeys_list = append(Gkeys_list, Privatekey)
		Gkeys_list = append(Gkeys_list, PublicKey)
		// fmt.Println("Add Key (#)")
		accountObj.AccountInitialUserName = "inovatian-refund-fees"
		accountObj.AccountName = "inovatian-refund-fees"
		accountObj.AccountPassword = "1D3B54DD108C13388A7D82F39FD41970AF48341F4C891973F5FACE49B8A1A4F7" // inovatian-refund-fees
		accountObj.AccountInitialPassword = "1D3B54DD108C13388A7D82F39FD41970AF48341F4C891973F5FACE49B8A1A4F7"
		//currentIndex1 = "000000000000000000000000000003"
		accountObj.AccountIndex = "000000000000000000000000000002"
		accountObj.AccountEmail = "hatim.refund@inovatian.com"
		accountObj.AccountAddress = "Cairo,Egypt"
		accountObj.AccountPublicKey = string(address)
		accountObj.AccountPrivateKey = Privatekey
		accountObj.AccountStatus = true
		accountObj.AccountRole = "admin"
		accountObj.AccountLastUpdatedTime = globalPkg.UTCtime()
		account.AddAccount(accountObj)
		//-----------------------
		//currentIndex1 = account.NewIndex()
		key, err = rsapk.GenerateKey(rand.Reader, 1024)
		cryptogrpghy.CheckError(err)
		Privatekey = cryptogrpghy.GetPrivatePEMKey(key)
		PublicKey = cryptogrpghy.GetPublicPEMKey(key.PublicKey)
		pk2 = []byte(PublicKey)
		address = cryptogrpghy.Address(pk2)
		add = string(address)
		// save pk, address
		savePKObj.Publickey = PublicKey
		savePKObj.Address = string(address)
		//account.SavePKAddress(savePKObj)
		// check depulicated keys
		for is_duplicated_key(Privatekey) || is_duplicated_key(add) {
			// fmt.Println("Duplicated Key (#)")
			key, err = rsapk.GenerateKey(reader, bitSize)
			Privatekey = cryptogrpghy.GetPrivatePEMKey(key)
			PublicKey = cryptogrpghy.GetPublicPEMKey(key.PublicKey)
			pk2 = []byte(PublicKey)
			address = cryptogrpghy.Address(pk2)
			// save pk, address
			savePKObj.Publickey = PublicKey
			savePKObj.Address = string(address)

		}
		savePKObj.CurrentTime = globalPkg.UTCtime()
		savePKObj.Index = "000000000000000000000000000003"
		account.SavePKAddress(savePKObj)
		Gkeys_list = append(Gkeys_list, Privatekey)
		Gkeys_list = append(Gkeys_list, PublicKey)
		// fmt.Println("Add Key (#)")

		accountObj.AccountInitialUserName = Conf.Server.UserName2
		accountObj.AccountName = Conf.Server.UserName2
		accountObj.AccountPassword = Conf.Server.Password2
		accountObj.AccountInitialPassword = Conf.Server.Password2
		accountObj.AccountIndex = "000000000000000000000000000003"
		accountObj.AccountEmail = "web.service@account.com"
		accountObj.AccountAddress = "account address"
		accountObj.AccountPublicKey = string(address)
		accountObj.AccountPrivateKey = Privatekey
		accountObj.AccountStatus = true
		accountObj.AccountRole = "service"
		accountObj.AccountLastUpdatedTime = globalPkg.UTCtime()
		account.AddAccount(accountObj)

	}
	tempAcc := account.GetAccountByName(Conf.Server.UserName)
	tempAcc.AccountInitialUserName = ""
	tempAcc.AccountInitialPassword = ""
	tempAcc.AccountRole = ""
	tempAcc.AccountLastUpdatedTime = globalPkg.UTCtime()
	tempAcc.AccountBalance = ""
	tempAcc.BlocksLst = nil
	tempAcc.SessionID = ""
	return tempAcc
}
func readFileResponse() {
	excelFileName := "responsessheet.xlsx"
	xlFile, err := xlsx.OpenFile(excelFileName)
	if err != nil {
		fmt.Println("error in open file")
	}
	var res responses.Response
	for _, sheet := range xlFile.Sheets {
		for _, row := range sheet.Rows {
			for i, cell := range row.Cells {
				text := cell.String()
				switch i {
				case 0:
					// res.ID = strconv.Itoa(j + 1)
					res.ID = text
				case 1:
					res.EngResponse = text
				}

				//fmt.Printf("%s\n", text)
			}
			responses.AddResponse(res)
			// fmt.Println("=== res ", res)
		}
	}
}
func Init() {
	public, private, _ := xrsa.GenRsaKeyPair(1024, xrsa.PKCS1)
	globalPkg.RSAPublic = string(public)
	globalPkg.RSAPrivate = string(private)
	// fmt.Println("---------------------------globalPkg.RSAPublic--------------------", globalPkg.RSAPublic)
	ledger.AdminObjec = admin.Admin1{"jkjdsfjgjdsfgjdsf", "fkhdfhdfkf"}
	validator.DigitalWalletIpObj = validator.DigitalWalletIp{Conf.Server.DigitalwalletIp, Conf.Server.Digitalwalletport}
	globalPkg.GlobalObj = Conf.GlobalData
	globalPkg.GlobalServerObj = globalPkg.GlobalServerIp{Ip: Conf.Server.PrivIP, Port: Conf.Server.Port}
	// fmt.Println("validator.CurrentValidator befoe : ", validator.CurrentValidator)
	adminIndex := admin.GetHash([]byte(validator.CurrentValidator.ValidatorIP)) + "_" + "0000000000"
	//plz put admin name & pswd and Superadmin name and password
	AdminObj := admin.AdminStruct{adminIndex, "inoadmin", "a5601de47276914b0b2bc40e9555d826b382001897f9cf065cc147ab1a3b483b", "ayaelhawary14@gmail.com", "01001873464", globalPkg.UTCtime(), globalPkg.UTCtime().AddDate(1, 0, 0), true, "SuperAdmin", nil, "", "inoadmin", "a5601de47276914b0b2bc40e9555d826b382001897f9cf065cc147ab1a3b483b", "", "", globalPkg.UTCtime()}
	fmt.Println("v PublicKey", globalPkg.RSAPublic)
	fmt.Println("v privateKey", globalPkg.RSAPrivate)

	if len(accountdb.GetAllAccounts()) == 0 {
		now := globalPkg.UTCtime()

		//Generate ECC key
		ECCPrivate, ECCPublic := ecc.GenerateECCKey()
		EncPubECCKey, EncPrivECCKey := globalPkg.EncryptECCKeys(ECCPublic, ECCPrivate)

		if Conf.Server.FirstMiner {
			validator.CurrentValidator = validator.ValidatorStruct{"http://" + Conf.Server.Ip + ":" + Conf.Server.Port, "http://" + Conf.Server.Ip + ":" + Conf.Server.Port, ECCPublic, ECCPrivate, "", "", Conf.Server.InitialStakeCoin, now, true, now, false, ""}
			v := validator.ValidatorStruct{"http://" + Conf.Server.Ip + ":" + Conf.Server.Port, "http://" + Conf.Server.Ip + ":" + Conf.Server.Port, ECCPublic, ECCPrivate, EncPubECCKey, EncPrivECCKey, Conf.Server.InitialStakeCoin, now, true, now, false, ""}
			(&v).AddValidator()
			InitTheaccount()
		} else {
			validator.CurrentValidator = validator.ValidatorStruct{"http://" + Conf.Server.Ip + ":" + Conf.Server.Port, "http://" + Conf.Server.Ip + ":" + Conf.Server.Port, ECCPublic, ECCPrivate, EncPubECCKey, EncPrivECCKey, Conf.Server.InitialStakeCoin, now, true, now, false, ""}
			validator.CreateValidator(&validator.CurrentValidator)
		}
		admin.CreateAdmin(AdminObj)
	} else {
		globalPkg.IsDown = true

		ip := "http://" + Conf.Server.Ip + ":" + Conf.Server.Port
		validator2Objlst := validator.GetAllValidators()
		for index := range validator2Objlst {
			if ip == validator2Objlst[index].ValidatorIP {
				validator.CurrentValidator = validator2Objlst[index]
			}
		}

		//--------------------------------
		var obj admin.Admin
		obj.UsernameAdmin = AdminObj.AdminUsername
		obj.PasswordAdmin = AdminObj.AdminPassword
		//y, _ := json.Marshal(obj)
		//alaa
		mapobj := globalfinctiontransaction.GetTransactionIndexTemMap()
		// fmt.Println("mapobj", mapobj)
		if mapobj[validator.CurrentValidator.ValidatorIP] == "" {
			globalfinctiontransaction.GetlastTransactionIndexLinearSearch()
		}

		//-------------------updateledger-----------------------------------------
		//
		var minersInfoObj heartbeat.MinersInfo
		minersInfoObj.Message.TimeStamp = validator.CurrentValidator.ValidatorLastHeartBeat
		minersInfoObj.Message.UpdateExist = false
		minersInfoObj.Message.MinerIP = validator.CurrentValidator.ValidatorIP

		//	heartbeatObj.HeartBeatIp_Time = validator.CurrentValidator.ValidatorLastHeartBeat.String()
		//	heartbeatObj.HeartBeatStatus = false

		// hb := heartbeat.ConvertminerInfoTOhbdatabase(minersInfoObj)
		hb := minersInfoObj.ConvertminerInfoTOhbdatabase()

		heartbeat.CreateHeartbeatAfterSystemdown(hb)
		var DObj ledger.DataStruct
		DObj.Date = validator.CurrentValidator.ValidatorLastHeartBeat

		//GEt ECC key pairs decrypted to append them to current validator
		DecryptedValidaorList := validator.GetAllValidatorsDecrypted()
		for _, myValidator := range DecryptedValidaorList {
			ip := "http://" + Conf.Server.Ip + ":" + Conf.Server.Port
			if myValidator.ValidatorIP == ip {
				validator.CurrentValidator.ECCPublicKey = myValidator.ECCPublicKey
				validator.CurrentValidator.ECCPrivateKey = myValidator.ECCPrivateKey
				validator.CurrentValidator.EncECCPublic = myValidator.EncECCPublic
				validator.CurrentValidator.EncECCPriv = myValidator.EncECCPriv
			}
		}
		b, _ := json.Marshal(DObj)
		if len(validator.ValidatorsLstObj) == 0 {
			validator.ValidatorsLstObj = DecryptedValidaorList
		}
		ledger.UpdateLedger(b)
		//heartbeatObj heartbeat.HeartBeatStruct
		//heartbeatObj. HeartBeatIp_Time=validator.CurrentValidator.ValidatorLastHeartBeat
		//heartbeatObj.HeartBeatStatus=false

		for i := 0; i < len(validator2Objlst); i++ {
			v := validator2Objlst[i].ValidatorIP
			globalfinctiontransaction.GetlastTransactionIndexLinearSearch2(v)
		}

	}
}
