package ledger

import (
	"encoding/json"
	"fmt"

	"net/http"
	"time"

	"../admin"
	"../globalPkg"
	globalfinctiontransaction "../globalfinctiontransaction"
	"../token"

	"../account"
	"../block"
	"../service"
	validator "../validator"

	"../accountdb"
	"../logpkg"
	transaction "../transaction"
)

//DataStruct data struct
type DataStruct struct {
	Date time.Time
}

//GetPartialLedger get partial ledger
func GetPartialLedger(dateTime time.Time) Ledger {
	ledgerObj := Ledger{}
	// NOTE: the validator public key is ava
	//	ledgerObj.ValidatorPubKey = validator.CurrentValidator.ValidatorPublicKey
	ledgerObj.ValidatorsLstObj = validator.ValidatorsLstObj

	// ledgerObj.ValidatorsLstObj = validator.GetValidatorsByTimeRange(dateTime) //decrepted
	// fmt.Println("ledgerObj.ValidatorsLstObj", ledgerObj.ValidatorsLstObj)
	ledgerObj.AccountsLstObj = accountdb.GetAccountsByLastUpdatedTime(dateTime)
	//GetTransactionByTimeRange
	ledgerObj.BlockchainObj = block.GetBlocksByTimeRange(dateTime)
	//ledgerObj.Tr = block.GetBlocks.ByTimeRange(dateTime)
	ledgerObj.ServiceTmp = service.GetAllservice()

	ledgerObj.TransactionLstObj = transaction.GetPendingTransactions()
	//ledgerObj.Temprequest = account.GetTempRequestedRegister()
	ledgerObj.UserObjects = account.GetUserObjLst()
	ledgerObj.UnconfirmedValidators = validator.TempValidatorlst
	//ledgerObj.ValidatorsLstObj = validator.ValidatorsLstObj
	ledgerObj.ResetPassArray = account.GetResetPasswordData()
	ledgerObj.LogDB = logpkg.GetlogDBaseByTimeRange(dateTime)
	ledgerObj.TokenObj = token.GetTokenByTimeRange(dateTime)
	ledgerObj.PurchasedService = service.GetserviceByTimeRange(dateTime)
	ledgerObj.AdminObj = admin.GetAdminByTimeRange(dateTime)
	ledgerObj.UserPK = account.GetSPKSByTimeRange(dateTime)
	ledgerObj.ValidatorMap = globalfinctiontransaction.GetTransactionIndexTemMap()

	return ledgerObj

}

//UpdateLedger update ledger
func UpdateLedger(y []byte) {
	// fmt.Println("opopopopopopopopopopopoopopopopopopopopoopopoppooopopopopopopopopopopopopopopopopopopop")
	//fmt.Println(validator.ValidatorsLstObj)
	//validatorObjlst := validator.GetAllValidators()
	validatorObjlst := validator.ValidatorsLstObj
	for _, node := range validatorObjlst {
		//fmt.Println("------", node.ValidatorIP)
		ledgerobj := Ledger{}
		if node.ValidatorIP != validator.CurrentValidator.ValidatorIP && node.ValidatorActive == true {
			_, ledgerbytes := globalPkg.SendLedger(y, node.ValidatorIP+"/GetPartialLegderAPI", "POST") ////get ledger
			json.Unmarshal(ledgerbytes, &ledgerobj)
			// fmt.Println(" len(ledgerobj.ValidatorsLstObj)", len(ledgerobj.ValidatorsLstObj))
			if len(ledgerobj.ValidatorsLstObj) != 0 {

				PostPartialLedger(ledgerobj) /////post
				fmt.Println("getPartialledger from :", node.ValidatorIP)
				break
			}
		}
		/*if len(ledgerobj.ValidatorsLstObj) != 0 {
			ledger.PostLedger(ledgerobj) /////post
			break
		}*/
	}
}

//GetPartialLegderAPI get partial ledger
func GetPartialLegderAPI(w http.ResponseWriter, req *http.Request) {
	// fmt.Println("hjhjhjhjhjhjhjhjjhjjhjhhjhhjj")
	datestructobj := DataStruct{}

	decoder := json.NewDecoder(req.Body)

	err := decoder.Decode(&datestructobj)

	if err != nil {
		globalPkg.SendError(w, "please enter your correct request")

		return
	}

	sendJSON, _ := json.Marshal(GetPartialLedger(datestructobj.Date))
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(sendJSON)
}

//PostPartialLedger post partial ledger
func PostPartialLedger(partialLedger Ledger) {

	for _, accountObj := range partialLedger.AccountsLstObj {
		// fmt.Println("account::", accountObj)
		if account.IfAccountExistsBefore(accountObj.AccountPublicKey) {
			// fmt.Println("iam updateddddddddddddsss")
			account.UpdateAccount(accountObj)
		} else {
			fmt.Println("Accounnt created?", account.AddAccount(accountObj))
		}
	}
	for _, transactionObj := range partialLedger.TransactionLstObj {
		transactionObj.AddTransaction()
	}
	for _, tokenobj := range partialLedger.TokenObj {
		token.TokenCreate(tokenobj)
	}
	for _, purchaseobj := range partialLedger.PurchasedService {
		(&purchaseobj).ServiceCreateOUpdate()
	}
	for _, blockObj := range partialLedger.BlockchainObj {
		// fmt.Println("block::", blockObj)
		// fmt.Println("block created?", (&blockObj).AddLedgerBlock())
		fmt.Println("block created?", block.AddLedgerBlock(&blockObj))
	}
	for _, AdminObjec := range partialLedger.AdminObj {

		fmt.Println("block created?", admin.CreateAdmin(AdminObjec))

	}
	for _, validatorobj := range partialLedger.ValidatorsLstObj {
		// value, _ := json.Marshal(validatorobj)
		// validatorobj.ECCPublicKey = ecc.UnmarshalECCPublicKey(value)
		validator.CreateValidator(&validatorobj)

	}
	for _, SPKS := range partialLedger.UserPK {
		account.SavePKAddress(SPKS)

	}

	//alaa
	globalfinctiontransaction.SetTransactionIndexTemMap(partialLedger.ValidatorMap) //globalfinctiontransaction.GetTransactionIndexTemMap()
	for _, validatorObject := range partialLedger.UnconfirmedValidators {
		validator.TempValidatorlst = append(validator.TempValidatorlst, validatorObject)
	}

	validator.ValidatorsLstObj = validator.GetAllValidatorsDecrypted()
	//validator.ValidatorsLstObj.ValidatorsLst = partialLedger.ValidatorsLstObj.ValidatorsLst
	//validator.CurrentValidator.ValidatorRegisterTime = time.Now()
	account.SetResetPasswordData(partialLedger.ResetPassArray)
	account.SetUserObjLst(partialLedger.UserObjects)
	service.SetserviceTemp(partialLedger.ServiceTmp)

}
