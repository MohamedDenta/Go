package startPkg

import (
	"crypto/rand"
	// "encoding/json"

	"fmt"
	"io/ioutil"
	"rsapk"

	//"time"

	"../accountdb"
	"../admin"
	"../cryptogrpghy"
	"../globalPkg"
	"../ledger"
	"../systemupdate"
	"../token"
	"../transaction"
	"../validator"

	// "../chunk"
	"../account"
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
func InitTheaccount() accountdb.AccountStruct {

	if account.GetAccountByName(Conf.Server.UserName).AccountName == "" {
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

		for is_duplicated_key(Privatekey) || is_duplicated_key(add) {
			key, err = rsapk.GenerateKey(reader, bitSize)
			Privatekey = cryptogrpghy.GetPrivatePEMKey(key)
			PublicKey = cryptogrpghy.GetPublicPEMKey(key.PublicKey)
			pk2 = []byte(PublicKey)
			address = cryptogrpghy.Address(pk2)

			// save pk, address
			savePKObj.Publickey = PublicKey
			savePKObj.Address = string(address)
		}
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
		transactionObj.TransactionID = ""
		transactionObj.TransactionID = globalPkg.CreateHash(transactionObj.TransactionTime, fmt.Sprintf("%s", transactionObj), 3)
		transaction.AddTransaction(transactionObj)

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
			fmt.Println("Duplicated Key (#)")
			key, err = rsapk.GenerateKey(reader, bitSize)
			Privatekey = cryptogrpghy.GetPrivatePEMKey(key)
			PublicKey = cryptogrpghy.GetPublicPEMKey(key.PublicKey)
			pk2 = []byte(PublicKey)
			address = cryptogrpghy.Address(pk2)
			// save pk, address
			savePKObj.Publickey = PublicKey
			savePKObj.Address = string(address)

		}
		account.SavePKAddress(savePKObj)
		Gkeys_list = append(Gkeys_list, Privatekey)
		Gkeys_list = append(Gkeys_list, PublicKey)
		fmt.Println("Add Key (#)")
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

		}
		account.SavePKAddress(savePKObj)
		Gkeys_list = append(Gkeys_list, Privatekey)
		Gkeys_list = append(Gkeys_list, PublicKey)
		fmt.Println("Add Key (#)")
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

		}
		account.SavePKAddress(savePKObj)
		Gkeys_list = append(Gkeys_list, Privatekey)
		Gkeys_list = append(Gkeys_list, PublicKey)
		fmt.Println("Add Key (#)")

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

func Init() {

	ledger.AdminObjec = admin.Admin1{"jkjdsfjgjdsfgjdsf", "fkhdfhdfkf"}
	validator.DigitalWalletIpObj = validator.DigitalWalletIp{Conf.Server.DigitalwalletIp, Conf.Server.Digitalwalletport}
	globalPkg.GlobalObj = Conf.GlobalData
	globalPkg.GlobalServerObj = globalPkg.GlobalServerIp{Ip: Conf.Server.PrivIP, Port: Conf.Server.Port}
	adminIndex := admin.GetHash([]byte(validator.CurrentValidator.ValidatorIP)) + "_" + "0000000000"
	AdminObj := admin.AdminStruct{adminIndex, "inoadmin", "a5601de47276914b0b2bc40e9555d826b382001897f9cf065cc147ab1a3b483b", "isra.elghazawy@gmail.com", "01001873464", globalPkg.UTCtime(), globalPkg.UTCtime().AddDate(1, 0, 0), true, "SuperAdmin", nil, "", "", "", "", "", globalPkg.UTCtime()}
	if len(accountdb.GetAllAccounts()) == 0 {
		now := globalPkg.UTCtime()
		dat1, _ := ioutil.ReadFile("validator/public.pem")
		dat2, _ := ioutil.ReadFile("validator/private.pem")

		if Conf.Server.FirstMiner {
			InitTheaccount()
			validator.CurrentValidator = validator.ValidatorStruct{"http://" + Conf.Server.Ip + ":" + Conf.Server.Port, "http://" + Conf.Server.Ip + ":" + Conf.Server.Port, string(dat1), string(dat2), Conf.Server.InitialStakeCoin, now, true, now, false, ""}

			validator.AddValidator(validator.ValidatorStruct{"http://" + Conf.Server.Ip + ":" + Conf.Server.Port, "http://" + Conf.Server.Ip + ":" + Conf.Server.Port, validator.CurrentValidator.ValidatorPublicKey, validator.CurrentValidator.ValidatorPrivateKey, Conf.Server.InitialStakeCoin, now, true, now, false, ""})
		} else {
			validator.CurrentValidator = validator.ValidatorStruct{"http://" + Conf.Server.Ip + ":" + Conf.Server.Port, "http://" + Conf.Server.Ip + ":" + Conf.Server.Port, string(dat1), string(dat2), Conf.Server.InitialStakeCoin, now, true, now, false, validator.NewIndex()}
			validator.ValidatorCreate(validator.CurrentValidator)
		}
		admin.CreateAdmin(AdminObj)
	} else {

		globalPkg.IsDown = true
		ip := "http://" + Conf.Server.Ip + ":" + Conf.Server.Port
		validator2Objlst := validator.GetAllValidators()
		for index := range validator2Objlst {
			if ip == validator2Objlst[index].ValidatorIP {
				validator.CurrentValidator = validator2Objlst[index]
				private, _ := ioutil.ReadFile("validator/private.pem")
				validator.CurrentValidator.ValidatorPrivateKey = string(private)
			}
		}

		// var obj admin.Admin
		// obj.UsernameAdmin = AdminObj.AdminUsername
		// obj.PasswordAdmin = AdminObj.AdminPassword
		// y, _ := json.Marshal(obj)

		// for _, node := range validator2Objlst {
		// 	ledgerobj := ledger.Ledger{}
		// 	if node.ValidatorIP != validator.CurrentValidator.ValidatorIP {
		// 		_, ledgerbytes := globalPkg.SendLedger(y, node.ValidatorIP+"/2c3920b33633a95417ea", "POST")
		// 		json.Unmarshal(ledgerbytes, &ledgerobj)
		// 	}
		// 	if len(ledgerobj.ValidatorsLstObj) != 0 {
		// 		ledger.PostLedger(ledgerobj)
		// 		break
		// 	}
		// }

		if len(validator.ValidatorsLstObj) == 0 {
			validator.ValidatorsLstObj = validator2Objlst
		}

	}
}
