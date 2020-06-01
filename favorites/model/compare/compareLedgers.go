package compare

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/smtp"
	"strings"
	"time"

	"../cryptogrpghy"

	"../accountdb"

	"../account"
	"../admin"
	tcp "../broadcastTcp"
	"../globalPkg"
	g "../globalPkg"
	"../ledger"
	"../logpkg"
	"../validator"
	"github.com/mitchellh/mapstructure"
)

var flag = false
//delUser delete user 
func delUser(lst *[]account.User, accountObjc accountdb.AccountStruct) {
	for index, data := range *lst {
		if data.Account.AccountName == accountObjc.AccountName {
			(*lst) = append((*lst)[:index], (*lst)[index+1:]...)
			fmt.Println("delete user ", accountObjc)
			break
		}
	}
}
//getsoketIpPublickey get socket ip pk 
func getsoketIpPublickey(ledg ledger.Ledger, ip string) (string, ecdsa.PublicKey) {
	for i := 0; i < len(ledg.ValidatorsLstObj); i++ {
		if ledg.ValidatorsLstObj[i].ValidatorIP == ip {
			encPub := ledg.ValidatorsLstObj[i].EncECCPublic
			timpPublic := cryptogrpghy.KeyDecrypt(globalPkg.AESEncryptionKey, encPub)

			rt := new(validator.RetrieveECCPublicKey)
			errmarsh := json.Unmarshal([]byte(timpPublic), &rt)
			if errmarsh != nil {
				fmt.Println("can't umarshal this obj check if this obj != nil")
				panic(errmarsh)
			}

			var public ecdsa.PublicKey
			var pub ecdsa.PublicKey

			public.Curve = rt.CurveParams
			public.X = rt.MyX
			public.Y = rt.MyY
			mapstructure.Decode(public, &pub)
			// fmt.Println("finshed unmarshalling key : ", pub)

			return ledg.ValidatorsLstObj[i].ValidatorSoketIP, pub
		}
	}
	fmt.Println("at getsoketIpPublickey () , can't get public key")
	return "", *new(ecdsa.PublicKey)
}
//sendEmail send Email 
func sendEmail(Body string, Email string) {
	//mime := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n";
	from := "noreply@inovatian.com" ///// "inovatian.tech@gmail.com"
	pass := "ino13579$"             /////your passward   ////

	to := Email //Email of User

	msg := "From: " + from + "\n" +
		"To: " + to + "\n" +
		"Subject: Inovatian Digital Wallet Verification\n\n" + Body

	///confirmation link

	err := smtp.SendMail("mail.inovatian.com:26",
		smtp.PlainAuth("", from, pass, "mail.inovatian.com"),
		from, []string{to}, []byte(msg))

	if err != nil {
		log.Printf("smtp error: %s", err)
		return
	}

	log.Println("sent, visit", Email)
}
//compareLedgerAccount compare ledger account 
func compareLedgerAccount(ledgerLst *[]ledger.Ledger, ledgerObj *ledger.Ledger, lstSoket []string, objSoket string, lstPublicKey []ecdsa.PublicKey, objPublicKey ecdsa.PublicKey) {
	//fmt.Println("compareLedgerAccount ")
	lstLedger, _ := json.Marshal((*ledgerLst)[0].AccountsLstObj)
	objLedjer, _ := json.Marshal(ledgerObj.AccountsLstObj)
	////fmt.Println("1")
	if bytes.Compare(lstLedger, objLedjer) != 0 {
		//fmt.Println("compareLedgerAccount 1")
		//		//fmt.Println("2")
		if len((*ledgerLst)[0].AccountsLstObj) == len(ledgerObj.AccountsLstObj) {
			//			//fmt.Println("3")
			for i := 0; i < len(ledgerObj.AccountsLstObj); i++ {
				//fmt.Println("compareLedgerAccount 1")
				accountlstByte, _ := json.Marshal((*ledgerLst)[0].AccountsLstObj[i])
				accountObjByte, _ := json.Marshal(ledgerObj.AccountsLstObj[i])
				////fmt.Println(bytes.Compare(accountlstByte,accountObjByte))
				if bytes.Compare(accountlstByte, accountObjByte) != 0 {
					check := (*ledgerLst)[0].AccountsLstObj[i].AccountLastUpdatedTime
					if check.After(ledgerObj.AccountsLstObj[i].AccountLastUpdatedTime) {
						currentTime, _ := time.Parse("2006-01-02 03:04:05 PM -0000", time.Now().UTC().Format("2006-01-02 03:04:05 PM -0000"))
						var currentTimeString []string
						currentTimeString = append(currentTimeString, currentTime.String())
						tcp.SendObject((*ledgerLst)[0].AccountsLstObj[i], objPublicKey, "PUT", "account", objSoket)
						ledgerObj.AccountsLstObj[i] = (*ledgerLst)[0].AccountsLstObj[i]
						//fmt.Println("check ..... ")
					} else {
						currentTime, _ := time.Parse("2006-01-02 03:04:05 PM -0000", time.Now().UTC().Format("2006-01-02 03:04:05 PM -0000"))
						var currentTimeString []string
						currentTimeString = append(currentTimeString, currentTime.String())
						for j := 0; j < len(*ledgerLst); j++ {
							tcp.SendObject(ledgerObj.AccountsLstObj[i], lstPublicKey[j], "PUT", "account", lstSoket[j])
							(*ledgerLst)[j].AccountsLstObj[i] = ledgerObj.AccountsLstObj[i]
						}
					}

				}

			}

		} else if len((*ledgerLst)[0].AccountsLstObj) > len(ledgerObj.AccountsLstObj) {
			diff := len((*ledgerLst)[0].AccountsLstObj) - len(ledgerObj.AccountsLstObj)

			for i := diff; i > 0; i-- {

				currentTime, _ := time.Parse("2006-01-02 03:04:05 PM -0000", time.Now().UTC().Format("2006-01-02 03:04:05 PM -0000"))
				var currentTimeString []string
				currentTimeString = append(currentTimeString, currentTime.String())

				//fmt.Println(tcp.SendObject((*ledgerLst)[0].AccountsLstObj[len((*ledgerLst)[0].AccountsLstObj)-i], objPublicKey, "POST", "account", objSoket))
				ledgerObj.AccountsLstObj = append(ledgerObj.AccountsLstObj, (*ledgerLst)[0].AccountsLstObj[len((*ledgerLst)[0].AccountsLstObj)-i])
				delUser(&ledgerObj.UserObjects, (*ledgerLst)[0].AccountsLstObj[len((*ledgerLst)[0].AccountsLstObj)-i])

			}

		} else {

			diff := len(ledgerObj.AccountsLstObj) - len((*ledgerLst)[0].AccountsLstObj)

			for i := diff; i > 0; i-- {

				currentTime, _ := time.Parse("2006-01-02 03:04:05 PM -0000", time.Now().UTC().Format("2006-01-02 03:04:05 PM -0000"))
				var currentTimeString []string
				currentTimeString = append(currentTimeString, currentTime.String())

				for j := 0; j < len(*ledgerLst); j++ {
					//fmt.Println("object", ledgerObj.AccountsLstObj[len(ledgerObj.AccountsLstObj)-i])
					//fmt.Println("pk", lstPublicKey[j])
					//fmt.Println("sk", lstSoket[j])
					// //fmt.Println("tcp", tcp.SendObject(ledgerObj.AccountsLstObj[len(ledgerObj.AccountsLstObj)-i], lstPublicKey[j], "POST", "account", lstSoket[j]))
					(*ledgerLst)[j].AccountsLstObj = append((*ledgerLst)[j].AccountsLstObj, ledgerObj.AccountsLstObj[len(ledgerObj.AccountsLstObj)-i])
					delUser(&(*ledgerLst)[j].UserObjects, ledgerObj.AccountsLstObj[len(ledgerObj.AccountsLstObj)-i])
					//if err != nil {
					//	//fmt.Println(err,"append account fail")
					//	//fmt.Println("(*ledgerLst)[j].AccountsLstObj",(*ledgerLst)[j].AccountsLstObj)
					//}
				}

			}

		}
	} else {
		//fmt.Println("compareLedgerAccount IS equal")
	}

}
//compareLedgerUser compare ledger user 
func compareLedgerUser(ledgerLst *[]ledger.Ledger, ledgerObj *ledger.Ledger, lstSoket []string, objSoket string, lstPublicKey []ecdsa.PublicKey, objPublicKey ecdsa.PublicKey) {
	//fmt.Println("in : compareLedgerUser")
	lstLedger, _ := json.Marshal((*ledgerLst)[0].UserObjects)
	objLedjer, _ := json.Marshal(ledgerObj.UserObjects)
	if bytes.Compare(lstLedger, objLedjer) != 0 {
		//fmt.Println("in : compareLedgerUser not equall")
		if len((*ledgerLst)[0].UserObjects) == len(ledgerObj.UserObjects) {
			for i := 0; i < len(ledgerObj.UserObjects); i++ {
				userLsttByte, _ := json.Marshal((*ledgerLst)[0].UserObjects[i])
				userObjByte, _ := json.Marshal(ledgerObj.UserObjects[i])
				if bytes.Compare(userLsttByte, userObjByte) != 0 {
					check := (*ledgerLst)[0].UserObjects[i].CurrentTime
					if check.After(ledgerObj.UserObjects[i].CurrentTime) {
						currentTime, _ := time.Parse("2006-01-02 03:04:05 PM -0000", time.Now().UTC().Format("2006-01-02 03:04:05 PM -0000"))
						var currentTimeString []string
						currentTimeString = append(currentTimeString, currentTime.String())
						tcp.SendObject((*ledgerLst)[0].UserObjects[i], objPublicKey, "adduser", "account module", objSoket)
						ledgerObj.UserObjects[i] = (*ledgerLst)[0].UserObjects[i]
						//fmt.Println("in : compareLedgerUser sent")
						break
					} else {
						currentTime, _ := time.Parse("2006-01-02 03:04:05 PM -0000", time.Now().UTC().Format("2006-01-02 03:04:05 PM -0000"))
						var currentTimeString []string
						currentTimeString = append(currentTimeString, currentTime.String())
						for j := 0; j < len(*ledgerLst); j++ {
							tcp.SendObject(ledgerObj.UserObjects[i], lstPublicKey[j], "adduser", "account module", lstSoket[j])
							(*ledgerLst)[j].UserObjects[i] = ledgerObj.UserObjects[i]
						}
						break
					}

				}

			}

		} else if len((*ledgerLst)[0].UserObjects) > len(ledgerObj.UserObjects) {
			diff := len((*ledgerLst)[0].UserObjects) - len(ledgerObj.UserObjects)

			for i := diff; i > 0; i-- {

				currentTime, _ := time.Parse("2006-01-02 03:04:05 PM -0000", time.Now().UTC().Format("2006-01-02 03:04:05 PM -0000"))
				var currentTimeString []string
				currentTimeString = append(currentTimeString, currentTime.String())
				tcp.SendObject((*ledgerLst)[0].UserObjects[len((*ledgerLst)[0].UserObjects)-i], objPublicKey, "adduser", "account module", objSoket)
				ledgerObj.UserObjects = append(ledgerObj.UserObjects, (*ledgerLst)[0].UserObjects[len((*ledgerLst)[0].UserObjects)-i])
				//if err != nil {
				//	//fmt.Println("append user fail 1",err)
				//	//fmt.Println("ledgerObj.UserObjects",ledgerObj.UserObjects)
				//}

			}

		} else {

			diff := len(ledgerObj.UserObjects) - len((*ledgerLst)[0].UserObjects)
			for i := diff; i > 0; i-- {
				currentTime, _ := time.Parse("2006-01-02 03:04:05 PM -0000", time.Now().UTC().Format("2006-01-02 03:04:05 PM -0000"))
				var currentTimeString []string
				currentTimeString = append(currentTimeString, currentTime.String())
				for j := 0; j < len(*ledgerLst); j++ {
					tcp.SendObject(ledgerObj.UserObjects[len(ledgerObj.UserObjects)-i], lstPublicKey[j], "adduser", "account module", lstSoket[j])
					(*ledgerLst)[j].UserObjects = append((*ledgerLst)[j].UserObjects, ledgerObj.UserObjects[len(ledgerObj.UserObjects)-i])
					//if err != nil {
					//	//fmt.Println("append user fail 2",err)
					//	//fmt.Println("(*ledgerLst)[j].UserObjects",(*ledgerLst)[j].UserObjects)
					//}
				}

			}

		}
	}

}
//compareLedgerToken compare ledger token 
func compareLedgerToken(ledgerLst *[]ledger.Ledger, ledgerObj *ledger.Ledger, lstSoket []string, objSoket string, lstPublicKey []ecdsa.PublicKey, objPublicKey ecdsa.PublicKey) {
	lstLedger, _ := json.Marshal((*ledgerLst)[0].TokenObj)
	objLedjer, _ := json.Marshal(ledgerObj.TokenObj)

	if bytes.Compare(lstLedger, objLedjer) != 0 {
		if len((*ledgerLst)[0].TokenObj) == len(ledgerObj.TokenObj) {
			for i := 0; i < len(ledgerObj.AccountsLstObj); i++ {
				tokenlstByte, _ := json.Marshal((*ledgerLst)[0].TokenObj[i])
				tokenObjByte, _ := json.Marshal(ledgerObj.TokenObj[i])
				if bytes.Compare(tokenlstByte, tokenObjByte) != 0 {
					check := (*ledgerLst)[0].TokenObj[i].TokenTime
					if check.After(ledgerObj.TokenObj[i].TokenTime) {
						currentTime, _ := time.Parse("2006-01-02 03:04:05 PM -0000", time.Now().UTC().Format("2006-01-02 03:04:05 PM -0000"))
						var currentTimeString []string
						currentTimeString = append(currentTimeString, currentTime.String())
						tcp.SendObject((*ledgerLst)[0].TokenObj[i], objPublicKey, "updatetoken", "token", objSoket)
						ledgerObj.TokenObj[i] = (*ledgerLst)[0].TokenObj[i]

					} else {
						currentTime, _ := time.Parse("2006-01-02 03:04:05 PM -0000", time.Now().UTC().Format("2006-01-02 03:04:05 PM -0000"))
						var currentTimeString []string
						currentTimeString = append(currentTimeString, currentTime.String())
						for j := 0; j < len(*ledgerLst); j++ {
							tcp.SendObject(ledgerObj.TokenObj[i], lstPublicKey[j], "updatetoken", "token", lstSoket[j])
							(*ledgerLst)[j].TokenObj[i] = ledgerObj.TokenObj[i]
						}
					}
				}
			}
		} else if len((*ledgerLst)[0].TokenObj) > len(ledgerObj.TokenObj) {
			diff := len((*ledgerLst)[0].TokenObj) - len(ledgerObj.TokenObj)
			for i := diff; i > 0; i-- {
				currentTime, _ := time.Parse("2006-01-02 03:04:05 PM -0000", time.Now().UTC().Format("2006-01-02 03:04:05 PM -0000"))
				var currentTimeString []string
				currentTimeString = append(currentTimeString, currentTime.String())
				tcp.SendObject((*ledgerLst)[0].TokenObj[len((*ledgerLst)[0].TokenObj)-i], objPublicKey, "addtoken", "token", objSoket)
				ledgerObj.TokenObj = append(ledgerObj.TokenObj, (*ledgerLst)[0].TokenObj[len((*ledgerLst)[0].TokenObj)-i])
			}
		} else {
			diff := len(ledgerObj.TokenObj) - len((*ledgerLst)[0].TokenObj)
			for i := diff; i > 0; i-- {
				currentTime, _ := time.Parse("2006-01-02 03:04:05 PM -0000", time.Now().UTC().Format("2006-01-02 03:04:05 PM -0000"))
				var currentTimeString []string
				currentTimeString = append(currentTimeString, currentTime.String())
				for j := 0; j < len(*ledgerLst); j++ {
					tcp.SendObject(ledgerObj.TokenObj[len(ledgerObj.TokenObj)-i], lstPublicKey[j], "addtoken", "token", lstSoket[j])
					(*ledgerLst)[j].TokenObj = append((*ledgerLst)[j].TokenObj, ledgerObj.TokenObj[len(ledgerObj.TokenObj)-i])
				}

			}

		}
	}
}
//compareLedgerAdmin compare ledger admin 
func compareLedgerAdmin(ledgerLst *[]ledger.Ledger, ledgerObj *ledger.Ledger, lstSoket []string, objSoket string, lstPublicKey []ecdsa.PublicKey, objPublicKey ecdsa.PublicKey) {
	fmt.Println("in : compareLedgerAdmin")

	lstLedger, _ := json.Marshal((*ledgerLst)[0].AdminObj)
	objLedjer, _ := json.Marshal(ledgerObj.AdminObj)

	if bytes.Compare(lstLedger, objLedjer) != 0 {
		if len((*ledgerLst)[0].AdminObj) == len(ledgerObj.AdminObj) {
			body := "Dear admin ,the lists of admins are not identical in servers with soket ip " + strings.Join(lstSoket, " , ") + " deffirence in " + objSoket
			for i := 0; i < len(ledgerObj.AdminObj); i++ {
				fmt.Println(body)
				sendEmail(body, ledgerObj.AdminObj[i].AdminEmail)
			}

		} else if len((*ledgerLst)[0].AdminObj) > len(ledgerObj.AdminObj) {
			diff := len((*ledgerLst)[0].AdminObj) - len(ledgerObj.AdminObj)
			halfLedgerLst := len(*ledgerLst) / 2
			if diff < halfLedgerLst || diff == halfLedgerLst {
				for i := diff; i > 0; i-- {
					tcp.SendObject((*ledgerLst)[0].AdminObj[len((*ledgerLst)[0].AdminObj)-i], objPublicKey, "addadmin", "admin", objSoket)
					ledgerObj.AdminObj = append(ledgerObj.AdminObj, (*ledgerLst)[0].AdminObj[len((*ledgerLst)[0].AdminObj)-i])
				}
			} else {
				fmt.Println("diffrence happend in less than half of the ledgers")
			}
		} else {
			diff := len(ledgerObj.AdminObj) - len((*ledgerLst)[0].AdminObj)
			halfLedgerLst := len(*ledgerLst) / 2

			if diff < halfLedgerLst || diff == halfLedgerLst {
				for i := diff; i > 0; i-- {
					for j := 0; j < len(*ledgerLst); j++ {
						tcp.SendObject(ledgerObj.AdminObj[len(ledgerObj.AdminObj)-i], lstPublicKey[j], "addadmin", "admin", lstSoket[j])
						(*ledgerLst)[j].AdminObj = append((*ledgerLst)[j].AdminObj, ledgerObj.AdminObj[len(ledgerObj.AdminObj)-i])
					}
				}
			} else {
				fmt.Println("diffrence happend in less than half of the ledgers")
			}
		}
	}
}
//compareLedgerBlock compare ledger block 
func compareLedgerBlock(ledgerLst *[]ledger.Ledger, ledgerObj *ledger.Ledger, lstSoket []string, objSoket string, lstPublicKey []ecdsa.PublicKey, objPublicKey ecdsa.PublicKey) {
	fmt.Println("in : compareLedgerBlock")

	lstLedger, _ := json.Marshal((*ledgerLst)[0].BlockchainObj)
	objLedjer, _ := json.Marshal(ledgerObj.BlockchainObj)

	if bytes.Compare(lstLedger, objLedjer) != 0 {
		if len((*ledgerLst)[0].BlockchainObj) == len(ledgerObj.BlockchainObj) {
			body := "Dear admin ,the lists of blocks are not identical in servers with soket ip " + strings.Join(lstSoket, " , ") + " deffirence in " + objSoket
			for i := 0; i < len(ledgerObj.AdminObj); i++ {
				fmt.Println(body)
				sendEmail(body, ledgerObj.AdminObj[i].AdminEmail)
			}
		} else if len((*ledgerLst)[0].BlockchainObj) > len(ledgerObj.BlockchainObj) {
			diff := len((*ledgerLst)[0].BlockchainObj) - len(ledgerObj.BlockchainObj)

			for i := diff; i > 0; i-- {
				tcp.SendObject((*ledgerLst)[0].BlockchainObj[len((*ledgerLst)[0].BlockchainObj)-i], objPublicKey, "AddBlock", "block", objSoket)
				ledgerObj.BlockchainObj = append(ledgerObj.BlockchainObj, (*ledgerLst)[0].BlockchainObj[len((*ledgerLst)[0].BlockchainObj)-i])

				if len((*ledgerLst)[0].BlockchainObj[len((*ledgerLst)[0].BlockchainObj)-i].BlockTransactions) > 0 {
					fmt.Println("first senario")
					for m := 0; m < len((*ledgerLst)[0].BlockchainObj[len((*ledgerLst)[0].BlockchainObj)-i].BlockTransactions); m++ {
						tcp.SendObject((*ledgerLst)[0].BlockchainObj[len((*ledgerLst)[0].BlockchainObj)-i].BlockTransactions[m], objPublicKey, "AddTransaction", "transaction", objSoket)
						ledgerObj.TransactionLstObj = append(ledgerObj.TransactionLstObj, (*ledgerLst)[0].BlockchainObj[len((*ledgerLst)[0].BlockchainObj)-i].BlockTransactions[m])
						fmt.Println("first senario0000001112121212121121121")
					}
				} else {
					fmt.Println(" block is wrong")
				}
			}

		} else {
			fmt.Println("second senario")
			diff := len(ledgerObj.BlockchainObj) - len((*ledgerLst)[0].BlockchainObj)

			for i := diff; i > 0; i-- {
				for j := 0; j < len(*ledgerLst); j++ {
					tcp.SendObject(ledgerObj.BlockchainObj[len(ledgerObj.BlockchainObj)-i], lstPublicKey[j], "AddBlock", "block", lstSoket[j])
					(*ledgerLst)[j].BlockchainObj = append((*ledgerLst)[j].BlockchainObj, ledgerObj.BlockchainObj[len(ledgerObj.BlockchainObj)-i])
					if len(ledgerObj.BlockchainObj[len(ledgerObj.BlockchainObj)-i].BlockTransactions) > 0 {
						for m := 0; m < len(ledgerObj.BlockchainObj[len(ledgerObj.BlockchainObj)-i].BlockTransactions); m++ {
							tcp.SendObject(ledgerObj.BlockchainObj[len(ledgerObj.BlockchainObj)-i].BlockTransactions[m], lstPublicKey[j], "AddTransaction", "transaction", lstSoket[j])
							(*ledgerLst)[j].TransactionLstObj = append((*ledgerLst)[j].TransactionLstObj, ledgerObj.BlockchainObj[len(ledgerObj.BlockchainObj)-i].BlockTransactions[m])
							//if err != nil {
							//	fmt.Println("append block transaction fail")
							//}
						}
					} else {
						fmt.Println(" block is wrong")
					}
				}
			}
		}
	}
}
//compareLedgerTransaction compare ledger transaction 
func compareLedgerTransaction(ledgerLst *[]ledger.Ledger, ledgerObj *ledger.Ledger, lstSoket []string, objSoket string, lstPublicKey []ecdsa.PublicKey, objPublicKey ecdsa.PublicKey) {
	//fmt.Println("in : compareLedgerTransaction")
	lstLedger, _ := json.Marshal((*ledgerLst)[0].TransactionLstObj)
	objLedjer, _ := json.Marshal(ledgerObj.TransactionLstObj)
	if bytes.Compare(lstLedger, objLedjer) != 0 {
		//fmt.Println("in : compareLedgerTransaction 2")
		if len((*ledgerLst)[0].TransactionLstObj) == len(ledgerObj.TransactionLstObj) {
			//fmt.Println("in : compareLedgerTransaction 3")
			body := "Dear admin ,the lists of Transactions are not identical in servers with soket ip " + strings.Join(lstSoket, " , ") + " difference in " + objSoket
			for i := 0; i < len(ledgerObj.AdminObj); i++ {
				sendEmail(body, ledgerObj.AdminObj[i].AdminEmail)
			}
		} else if len((*ledgerLst)[0].TransactionLstObj) > len(ledgerObj.TransactionLstObj) {
			//fmt.Println("in : compareLedgerTransaction 4")
			diff := len((*ledgerLst)[0].TransactionLstObj) - len(ledgerObj.TransactionLstObj)
			for i := diff; i > 0; i-- {
				currentTime, _ := time.Parse("2006-01-02 03:04:05 PM -0000", time.Now().UTC().Format("2006-01-02 03:04:05 PM -0000"))
				var currentTimeString []string
				currentTimeString = append(currentTimeString, currentTime.String())
				tcp.SendObject((*ledgerLst)[0].TransactionLstObj[len((*ledgerLst)[0].TransactionLstObj)-i], objPublicKey, "missed-transaction", "transaction", objSoket)
				ledgerObj.TransactionLstObj = append(ledgerObj.TransactionLstObj, (*ledgerLst)[0].TransactionLstObj[len((*ledgerLst)[0].TransactionLstObj)-i])
				//if err != nil {
				//	//fmt.Println("append Transaction fail")
				//}
			}

		} else {
			//fmt.Println("in : compareLedgerTransaction 5")
			diff := len(ledgerObj.TransactionLstObj) - len((*ledgerLst)[0].TransactionLstObj)
			for i := diff; i > 0; i-- {

				currentTime, _ := time.Parse("2006-01-02 03:04:05 PM -0000", time.Now().UTC().Format("2006-01-02 03:04:05 PM -0000"))
				var currentTimeString []string
				currentTimeString = append(currentTimeString, currentTime.String())

				for j := 0; j < len(*ledgerLst); j++ {
					tcp.SendObject(ledgerObj.TransactionLstObj[len(ledgerObj.TransactionLstObj)-i], lstPublicKey[j], "missed-transaction", "transaction", lstSoket[j])
					(*ledgerLst)[j].TransactionLstObj = append((*ledgerLst)[j].TransactionLstObj, ledgerObj.TransactionLstObj[len(ledgerObj.TransactionLstObj)-i])
					//if err != nil {
					//	//fmt.Println("append Transaction fail")
					//}
				}

			}

		}
	}

}

//compareLedgerValidator compare ledger validator
func compareLedgerValidator(ledgerLst *[]ledger.Ledger, ledgerObj *ledger.Ledger, lstSoket []string, objSoket string, lstPublicKey []ecdsa.PublicKey, objPublicKey ecdsa.PublicKey) {
	// fmt.Println("in : compareLedgerValidator")

	lstLedger, _ := json.Marshal((*ledgerLst)[0].ValidatorsLstObj)
	objLedjer, _ := json.Marshal(ledgerObj.ValidatorsLstObj)

	if bytes.Compare(lstLedger, objLedjer) != 0 {
		fmt.Println("First condition is wrong")

		if len((*ledgerLst)[0].ValidatorsLstObj) == len(ledgerObj.ValidatorsLstObj) { // not the same but the same length
			for i := 0; i < len(ledgerObj.ValidatorsLstObj); i++ {
				if (*ledgerLst)[0].ValidatorsLstObj[i].EncECCPublic != ledgerObj.ValidatorsLstObj[i].EncECCPublic ||
					(*ledgerLst)[0].ValidatorsLstObj[i].ValidatorSoketIP != ledgerObj.ValidatorsLstObj[i].ValidatorSoketIP ||
					(*ledgerLst)[0].ValidatorsLstObj[i].ValidatorIP != ledgerObj.ValidatorsLstObj[i].ValidatorSoketIP ||
					(*ledgerLst)[0].ValidatorsLstObj[i].ValidatorRegisterTime != ledgerObj.ValidatorsLstObj[i].ValidatorRegisterTime ||
					(*ledgerLst)[0].ValidatorsLstObj[i].ValidatorStakeCoins != ledgerObj.ValidatorsLstObj[i].ValidatorStakeCoins ||
					(*ledgerLst)[0].ValidatorsLstObj[i].ValidatorActive != ledgerObj.ValidatorsLstObj[i].ValidatorActive {
					check := (*ledgerLst)[0].ValidatorsLstObj[i].ValidatorLastHeartBeat
					if check.After(ledgerObj.ValidatorsLstObj[i].ValidatorLastHeartBeat) {
						tcp.SendObject((*ledgerLst)[0].ValidatorsLstObj[i], objPublicKey, "PUT", "validator", objSoket)
						ledgerObj.ValidatorsLstObj[i] = (*ledgerLst)[0].ValidatorsLstObj[i]
						fmt.Println("third condition is wrong")

					} else {
						fmt.Println("4th condition is wrong")

						for j := 0; j < len(*ledgerLst); j++ {
							tcp.SendObject(ledgerObj.ValidatorsLstObj[i], lstPublicKey[j], "PUT", "validator", lstSoket[j])
							(*ledgerLst)[j].ValidatorsLstObj[i] = ledgerObj.ValidatorsLstObj[i]
						}
					}

				}

			}

		} else if len((*ledgerLst)[0].ValidatorsLstObj) > len(ledgerObj.ValidatorsLstObj) {
			//ledger list contain more validators

			body := "Dear admin ,the lists of validators are not identical in servers with soket ip " + strings.Join(lstSoket, " , ") + " deffirence in " + objSoket
			for i := 0; i < len(ledgerObj.AdminObj); i++ {
				// fmt.Println(body)
				sendEmail(body, ledgerObj.AdminObj[i].AdminEmail)
			}

			// diff := len((*ledgerLst)[0].ValidatorsLstObj) - len(ledgerObj.ValidatorsLstObj)
			// for i := diff; i > 0; i-- {
			// 	fmt.Println("6th pblic key : ", objPublicKey)
			// 	fmt.Println("6th soket : ", objSoket)
			// 	tcp.SendObject((*ledgerLst)[0].ValidatorsLstObj[len((*ledgerLst)[0].ValidatorsLstObj)-i], objPublicKey, "POST", "confirmedvalidator", objSoket)
			// 	ledgerObj.ValidatorsLstObj = append(ledgerObj.ValidatorsLstObj, (*ledgerLst)[0].ValidatorsLstObj[len((*ledgerLst)[0].ValidatorsLstObj)-i])
			// }
		} else {
			//current ledger compaired have less validators
			body := "Dear admin ,the lists of validators are not identical in servers with soket ip " + strings.Join(lstSoket, " , ") + " deffirence in " + objSoket
			for i := 0; i < len(ledgerObj.AdminObj); i++ {
				// fmt.Println(body)
				sendEmail(body, ledgerObj.AdminObj[i].AdminEmail)
			}

			// diff := len(ledgerObj.ValidatorsLstObj) - len((*ledgerLst)[0].ValidatorsLstObj)
			// for i := diff; i > 0; i-- {
			// 	for j := 0; j < len(*ledgerLst); j++ {
			// 		fmt.Println("7th pblic key : ", lstPublicKey[j])
			// 		fmt.Println("7th soket : ", lstSoket[j])

			// 		tcp.SendObject(ledgerObj.ValidatorsLstObj[len(ledgerObj.ValidatorsLstObj)-i], lstPublicKey[j], "POST", "confirmedvalidator", lstSoket[j])
			// 		(*ledgerLst)[j].ValidatorsLstObj = append((*ledgerLst)[j].ValidatorsLstObj, ledgerObj.ValidatorsLstObj[len(ledgerObj.ValidatorsLstObj)-i])
			// 	}

			// }
		}
	}

}
//compareLedgerTransactionDb compare ledger transaction db 
func compareLedgerTransactionDb(ledgerLst *[]ledger.Ledger, ledgerObj *ledger.Ledger, lstSoket []string, objSoket string, lstPublicKey []ecdsa.PublicKey, objPublicKey ecdsa.PublicKey) {
	lstLedger, _ := json.Marshal((*ledgerLst)[0].TransactionLstDb)
	objLedjer, _ := json.Marshal(ledgerObj.TransactionLstDb)
	//fmt.Println("compareLedgerTransactionDb start")
	if bytes.Compare(lstLedger, objLedjer) != 0 {
		//fmt.Println("compareLedgerTransactionDb 1")
		if len((*ledgerLst)[0].TransactionLstDb) == len(ledgerObj.TransactionLstDb) {
			//fmt.Println("compareLedgerTransactionDb 2")
			body := "Dear admin ,the lists of Transactions saved in database are not identical in servers with soket ip " + strings.Join(lstSoket, " , ") + " and " + objSoket
			for i := 0; i < len(ledgerObj.AdminObj); i++ {
				sendEmail(body, ledgerObj.AdminObj[i].AdminEmail)
			}
		} else if len((*ledgerLst)[0].TransactionLstDb) > len(ledgerObj.TransactionLstDb) {
			//fmt.Println("compareLedgerTransactionDb 3")
			diff := len((*ledgerLst)[0].TransactionLstDb) - len(ledgerObj.TransactionLstDb)
			for i := diff; i > 0; i-- {
				currentTime, _ := time.Parse("2006-01-02 03:04:05 PM -0000", time.Now().UTC().Format("2006-01-02 03:04:05 PM -0000"))
				var currentTimeString []string
				currentTimeString = append(currentTimeString, currentTime.String())
				tcp.SendObject((*ledgerLst)[0].TransactionLstDb[len((*ledgerLst)[0].TransactionLstDb)-i], objPublicKey, "missed-transaction-db", "transaction", objSoket)

				ledgerObj.TransactionLstDb = append(ledgerObj.TransactionLstDb, (*ledgerLst)[0].TransactionLstDb[len((*ledgerLst)[0].TransactionLstDb)-i])
				// if err != nil {
				// 	//fmt.Println("append Transactiondb fail")
				// }

			}

		} else {
			//fmt.Println("compareLedgerTransactionDb 4")
			diff := len(ledgerObj.TransactionLstDb) - len((*ledgerLst)[0].TransactionLstDb)

			for i := diff; i > 0; i-- {

				currentTime, _ := time.Parse("2006-01-02 03:04:05 PM -0000", time.Now().UTC().Format("2006-01-02 03:04:05 PM -0000"))
				var currentTimeString []string
				currentTimeString = append(currentTimeString, currentTime.String())

				for j := 0; j < len(*ledgerLst); j++ {
					tcp.SendObject(ledgerObj.TransactionLstDb[len(ledgerObj.TransactionLstDb)-i], lstPublicKey[j], "missed-transaction-db", "transaction", lstSoket[j])
					(*ledgerLst)[j].TransactionLstDb = append((*ledgerLst)[j].TransactionLstDb, ledgerObj.TransactionLstDb[len(ledgerObj.TransactionLstDb)-i])
					// if err != nil {
					// 	//fmt.Println("append Transactiondb fail")
					// }
				}

			}

		}
	}

}
//compareLedgerResetPass compare ledger reset password 
func compareLedgerResetPass(ledgerLst *[]ledger.Ledger, ledgerObj *ledger.Ledger, lstSoket []string, objSoket string, lstPublicKey []ecdsa.PublicKey, objPublicKey ecdsa.PublicKey) {
	//fmt.Println("compareLedgerResetPass starts ")

	lstLedger, _ := json.Marshal((*ledgerLst)[0].ResetPassArray)
	objLedjer, _ := json.Marshal(ledgerObj.ResetPassArray)

	if bytes.Compare(lstLedger, objLedjer) != 0 {
		if len((*ledgerLst)[0].ResetPassArray) == len(ledgerObj.ResetPassArray) {
			for i := 0; i < len(ledgerObj.ResetPassArray); i++ {
				resetPassLsttByte, _ := json.Marshal((*ledgerLst)[0].ResetPassArray[i])
				resetPassObjByte, _ := json.Marshal(ledgerObj.ResetPassArray[i])
				if bytes.Compare(resetPassLsttByte, resetPassObjByte) != 0 {
					//fmt.Println("compareLedgerResetPass works ")
					check := (*ledgerLst)[0].ResetPassArray[i].CurrentTime
					if check.After(ledgerObj.ResetPassArray[i].CurrentTime) {
						currentTime, _ := time.Parse("2006-01-02 03:04:05 PM -0000", time.Now().UTC().Format("2006-01-02 03:04:05 PM -0000"))
						var currentTimeString []string
						currentTimeString = append(currentTimeString, currentTime.String())
						tcp.SendObject((*ledgerLst)[0].ResetPassArray, objPublicKey, "addRestPassword", "account module", objSoket)
						ledgerObj.ResetPassArray = (*ledgerLst)[0].ResetPassArray
						//fmt.Println("My compareLedgerResetPass updated ")
						break

					} else {
						currentTime, _ := time.Parse("2006-01-02 03:04:05 PM -0000", time.Now().UTC().Format("2006-01-02 03:04:05 PM -0000"))
						var currentTimeString []string
						currentTimeString = append(currentTimeString, currentTime.String())
						for j := 0; j < len(*ledgerLst); j++ {
							tcp.SendObject(ledgerObj.ResetPassArray, lstPublicKey[j], "addRestPassword", "account module", lstSoket[j])
							(*ledgerLst)[j].ResetPassArray = ledgerObj.ResetPassArray
							//fmt.Println("compareLedgerResetPass sent ")
						}
						break
					}

				}

			}

		} else if len((*ledgerLst)[0].ResetPassArray) > len(ledgerObj.ResetPassArray) {
			//fmt.Println("****************88********************")
			diff := len((*ledgerLst)[0].ResetPassArray) - len(ledgerObj.ResetPassArray)
			for i := diff; i > 0; i-- {
				currentTime, _ := time.Parse("2006-01-02 03:04:05 PM -0000", time.Now().UTC().Format("2006-01-02 03:04:05 PM -0000"))
				var currentTimeString []string
				currentTimeString = append(currentTimeString, currentTime.String())
				tcp.SendObject((*ledgerLst)[0].ResetPassArray[len((*ledgerLst)[0].ResetPassArray)-i], objPublicKey, "addRestPassword", "account module", objSoket)
				ledgerObj.ResetPassArray = append(ledgerObj.ResetPassArray, (*ledgerLst)[0].ResetPassArray[len((*ledgerLst)[0].ResetPassArray)-i])

				//if err != nil {
				//	//fmt.Println("append ResetPass fail")
				//}
			}

		} else {
			//fmt.Println("****************88********************")
			diff := len(ledgerObj.ResetPassArray) - len((*ledgerLst)[0].ResetPassArray)

			for i := diff; i > 0; i-- {

				currentTime, _ := time.Parse("2006-01-02 03:04:05 PM -0000", time.Now().UTC().Format("2006-01-02 03:04:05 PM -0000"))
				var currentTimeString []string
				currentTimeString = append(currentTimeString, currentTime.String())

				for j := 0; j < len(*ledgerLst); j++ {
					tcp.SendObject(ledgerObj.ResetPassArray[len(ledgerObj.ResetPassArray)-i], lstPublicKey[j], "addRestPassword", "account module", lstSoket[j])
					(*ledgerLst)[j].ResetPassArray = append((*ledgerLst)[j].ResetPassArray, ledgerObj.ResetPassArray[len(ledgerObj.ResetPassArray)-i])
					//if err != nil {
					//	//fmt.Println("append ResetPass fail")
					//}
				}

			}

		}
	}

}
//compareUpdateLedger compare update ledger 
func compareUpdateLedger(ledgerLst *[]ledger.Ledger, ledgerObj *ledger.Ledger, lstSoket []string, objSoket string, lstPublicKey []ecdsa.PublicKey, objPublicKey ecdsa.PublicKey) {

	compareLedgerUser(&(*ledgerLst), &(*ledgerObj), lstSoket, objSoket, lstPublicKey, objPublicKey)
	compareLedgerAccount(&(*ledgerLst), &(*ledgerObj), lstSoket, objSoket, lstPublicKey, objPublicKey)
	compareLedgerResetPass(&(*ledgerLst), &(*ledgerObj), lstSoket, objSoket, lstPublicKey, objPublicKey)

	compareLedgerToken(&(*ledgerLst), &(*ledgerObj), lstSoket, objSoket, lstPublicKey, objPublicKey)
	compareLedgerAdmin(&(*ledgerLst), &(*ledgerObj), lstSoket, objSoket, lstPublicKey, objPublicKey)
	compareLedgerTransaction(&(*ledgerLst), &(*ledgerObj), lstSoket, objSoket, lstPublicKey, objPublicKey)
	compareLedgerTransactionDb(&(*ledgerLst), &(*ledgerObj), lstSoket, objSoket, lstPublicKey, objPublicKey)

	compareLedgerBlock(&(*ledgerLst), &(*ledgerObj), lstSoket, objSoket, lstPublicKey, objPublicKey)

	compareLedgerValidator(&(*ledgerLst), &(*ledgerObj), lstSoket, objSoket, lstPublicKey, objPublicKey)
	compareLedgerServiceTmp(&(*ledgerLst), &(*ledgerObj), lstSoket, objSoket, lstPublicKey, objPublicKey)
	compareLedgerPurchaseService(&(*ledgerLst), &(*ledgerObj), lstSoket, objSoket, lstPublicKey, objPublicKey)

}

//Routine make it run without api like any onther go routine
// make server ips dynamic
// fix get ledger issue
func Routine() {
	for {
		time.Sleep(30 * time.Second)
		// fmt.Println("flag>>>", flag)
		// if flag ==true {
		fmt.Println("Will Compare Ledger now : ")
		allValidators := validator.GetAllValidators()
		ips := []string{}
		for _, myValidator := range allValidators {
			// fmt.Println("validator in validator list : ", myValidator.ValidatorIP)
			if myValidator.ValidatorActive == true {
				ips = append(ips, myValidator.ValidatorIP)
			}
		}
		// fmt.Println("my active validators in IPS list : ", ips)
		ledgerlist := []ledger.Ledger{}
		serverSoketIps := []string{}
		serverPublicKey := []ecdsa.PublicKey{}
		// fmt.Println("1")

		for _, ip := range ips {
			iptemp := ip
			ip = ip + "/2c3920b33633a95417ea"
			var adminstructObj = admin.Admin{
				UsernameAdmin: "inoadmin",
				PasswordAdmin: "a5601de47276914b0b2bc40e9555d826b382001897f9cf065cc147ab1a3b483b",
			}
			var dummy ledger.Ledger
			adminObj, _ := json.Marshal(adminstructObj)
			ledgerString := g.SendRequestAndGetResponse(adminObj, ip, "POST", dummy)
			ledgerbyte := []byte(ledgerString)
			ledgerTemp := ledger.Ledger{}
			ledger := ledger.Ledger{}
			json.Unmarshal(ledgerbyte, &ledgerTemp)
			mapstructure.Decode(ledgerTemp, &ledger)
			if ledger.ValidatorsLstObj == nil {
				fmt.Println("cant get leger of server" + iptemp)
				continue
			}
			ledgerlist = append(ledgerlist, ledger)
			soketip, publicKey := getsoketIpPublickey(ledger, iptemp)
			serverSoketIps = append(serverSoketIps, soketip)
			serverPublicKey = append(serverPublicKey, publicKey)

		}

		ledgerLstUpdated := []ledger.Ledger{}
		lenn := len(ledgerlist)
		fmt.Println("length of ledger list : ", lenn)

		if lenn > 1 {
			for i := 0; i < lenn-1; i++ {
				for j := 0; j < i+1; j++ {
					ledgerLstUpdated = append(ledgerLstUpdated, ledgerlist[j])
				}
				compareUpdateLedger(&ledgerLstUpdated, &ledgerlist[i+1], serverSoketIps, serverSoketIps[i+1], serverPublicKey, serverPublicKey[i+1])

				for j := 0; j < i+1; j++ {
					ledgerlist[j] = ledgerLstUpdated[j]
				}
				ledgerLstUpdated = nil
			}
			fmt.Println("Finished comparing ledger")

		} else {
			fmt.Println("cant copmpare less than 2 servers ledger")
		}
		// }
	}

}
//CompareLedgers compare all ledger 
func CompareLedgers(w http.ResponseWriter, req *http.Request) {
	//log
	now, userIP := globalPkg.GetIP(req)
	logobj := logpkg.LogStruct{"_", now, userIP, "macAdress", "GetLegderAPI", "ledger", "_", "_", "_", 0}

	if !admin.AdminAPIDecoderAndValidation(w,req.Body,logobj){
		return
	}
		//go Routine()
	flag = true
	sendJSON, _ := json.Marshal(flag)
	globalPkg.SendResponse(w, sendJSON)
	globalPkg.WriteLog(logobj, "get updade success", "success")

}
//compareLedgerServiceTmp compare ledger service temp 
func compareLedgerServiceTmp(ledgerLst *[]ledger.Ledger, ledgerObj *ledger.Ledger, lstSoket []string, objSoket string, lstPublicKey []ecdsa.PublicKey, objPublicKey ecdsa.PublicKey) {
	fmt.Println("in : compareLedgerServiceTmp")

	lstLedger, _ := json.Marshal((*ledgerLst)[0].ServiceTmp)
	objLedjer, _ := json.Marshal(ledgerObj.ServiceTmp)
	if len((*ledgerLst)[0].ServiceTmp) == 0 && len(ledgerObj.ServiceTmp) == 0 {
		fmt.Println("eMPTY lIST")
		return
	}
	fmt.Println("list11111", (*ledgerLst)[0].ServiceTmp)
	fmt.Println("list222222", ledgerObj.ServiceTmp)
	if bytes.Compare(lstLedger, objLedjer) != 0 {

		if len((*ledgerLst)[0].ServiceTmp) == len(ledgerObj.ServiceTmp) {

			fmt.Println("ffffff1")
			fmt.Println("in : forloooooooooooooooolooooop")
			body := "Dear admin ,the lists of ServiceTmp are not identical in servers with soket ip " + strings.Join(lstSoket, " , ") + " deffirence in " + objSoket
			for i := 0; i < len(ledgerObj.AdminObj); i++ {
				fmt.Println("body", body)

				sendEmail(body, ledgerObj.AdminObj[i].AdminEmail)
			}
		} else if len((*ledgerLst)[0].ServiceTmp) > len(ledgerObj.ServiceTmp) {

			diff := len((*ledgerLst)[0].ServiceTmp) - len(ledgerObj.ServiceTmp)

			for i := diff; i > 0; i-- {

				currentTime, _ := time.Parse("2006-01-02 03:04:05 PM -0000", time.Now().UTC().Format("2006-01-02 03:04:05 PM -0000"))
				var currentTimeString []string
				currentTimeString = append(currentTimeString, currentTime.String())
				tcp.SendObject((*ledgerLst)[0].ServiceTmp[len((*ledgerLst)[0].ServiceTmp)-i], objPublicKey, "Tmp", "Add Service", objSoket)
				ledgerObj.ServiceTmp = append(ledgerObj.ServiceTmp, (*ledgerLst)[0].ServiceTmp[len((*ledgerLst)[0].ServiceTmp)-i])
				//if err != nil {
				//	fmt.Println("append ServiceTmp fail")
				//}

			}

		} else {

			diff := len(ledgerObj.ServiceTmp) - len((*ledgerLst)[0].ServiceTmp)

			for i := diff; i > 0; i-- {

				currentTime, _ := time.Parse("2006-01-02 03:04:05 PM -0000", time.Now().UTC().Format("2006-01-02 03:04:05 PM -0000"))
				var currentTimeString []string
				currentTimeString = append(currentTimeString, currentTime.String())

				for j := 0; j < len(*ledgerLst); j++ {
					tcp.SendObject(ledgerObj.ServiceTmp[len(ledgerObj.ServiceTmp)-i], lstPublicKey[j], "Tmp", "Add Service", lstSoket[j])
					(*ledgerLst)[j].ServiceTmp = append((*ledgerLst)[j].ServiceTmp, ledgerObj.ServiceTmp[len(ledgerObj.ServiceTmp)-i])

				}

			}

		}
	}

}
//compareLedgerPurchaseService compare ledger purchase service  
func compareLedgerPurchaseService(ledgerLst *[]ledger.Ledger, ledgerObj *ledger.Ledger, lstSoket []string, objSoket string, lstPublicKey []ecdsa.PublicKey, objPublicKey ecdsa.PublicKey) {
	fmt.Println("in : compareLedgerPurchaseService")

	lstLedger, _ := json.Marshal((*ledgerLst)[0].PurchasedService)
	objLedjer, _ := json.Marshal(ledgerObj.PurchasedService)

	if bytes.Compare(lstLedger, objLedjer) != 0 {
		fmt.Println("differnt ledger")
		if len((*ledgerLst)[0].PurchasedService) == len(ledgerObj.PurchasedService) {
			fmt.Println("same length")
			body := "Dear admin ,the lists of PurchasedService are not identical in servers with soket ip " + strings.Join(lstSoket, " , ") + " deffirence in " + objSoket
			for i := 0; i < len(ledgerObj.AdminObj); i++ {
				fmt.Println(body)
				sendEmail(body, ledgerObj.AdminObj[i].AdminEmail)
			}
		} else if len((*ledgerLst)[0].PurchasedService) > len(ledgerObj.PurchasedService) {
			diff := len((*ledgerLst)[0].PurchasedService) - len(ledgerObj.PurchasedService)

			for i := diff; i > 0; i-- {

				currentTime, _ := time.Parse("2006-01-02 03:04:05 PM -0000", time.Now().UTC().Format("2006-01-02 03:04:05 PM -0000"))
				var currentTimeString []string
				currentTimeString = append(currentTimeString, currentTime.String())
				tcp.SendObject((*ledgerLst)[0].PurchasedService[len((*ledgerLst)[0].PurchasedService)-i], objPublicKey, "DB", "Add Service", objSoket)
				ledgerObj.PurchasedService = append(ledgerObj.PurchasedService, (*ledgerLst)[0].PurchasedService[len((*ledgerLst)[0].PurchasedService)-i])
				//if err != nil {
				//	fmt.Println("append PurchasedService fail")
				//}

			}

		} else {

			diff := len(ledgerObj.PurchasedService) - len((*ledgerLst)[0].PurchasedService)

			for i := diff; i > 0; i-- {

				currentTime, _ := time.Parse("2006-01-02 03:04:05 PM -0000", time.Now().UTC().Format("2006-01-02 03:04:05 PM -0000"))
				var currentTimeString []string
				currentTimeString = append(currentTimeString, currentTime.String())

				for j := 0; j < len(*ledgerLst); j++ {
					fmt.Println("sendledger1111122222222")
					tcp.SendObject(ledgerObj.PurchasedService[len(ledgerObj.PurchasedService)-i], lstPublicKey[j], "DB", "Add Service", lstSoket[j])
					(*ledgerLst)[j].PurchasedService = append((*ledgerLst)[j].PurchasedService, ledgerObj.PurchasedService[len(ledgerObj.PurchasedService)-i])
					//if err != nil {
					//	fmt.Println("append PurchasedService fail")
					//}
				}

			}

		}
	}

}
