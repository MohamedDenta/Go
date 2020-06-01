package transaction

import (
	"fmt"
	"log"
	"time"

	"../accountdb"
	"../filestorage"
	"../globalPkg"
)

// Type :- 0 is normal transaction.
//  	   1 is token creation transaction.
// 		   2 is refund to flat currency.
// 		   3 is refund to ino token.
/*----------------Transaction structure----------------- */
type Transaction struct {
	TransactionID     string
	Type              int8
	TransactionInput  []TXInput
	TransactionOutPut []TXOutput
	TransactionTime   time.Time
	ServiceId         string
	Filestruct        filestorage.FileStruct //file struct data to store in block
}

type TransactionDB struct {
	TransactionID     string
	TransactionInput  []TXInput
	TransactionOutPut []TXOutput
	TransactionTime   time.Time
	TransactionKey    string
	Filestruct        filestorage.FileStruct //file struct data to store in block
}

type TXOutput struct {
	OutPutValue       float64 // amount of tokens
	RecieverPublicKey string
	IsFee             bool
	TokenID           string //
	// if token id == 1 is ino coin
}

// token initiator, inovation account will have the token and send transaction, the sender is inovation and the receiver is the initiator
// inovation take token total supply token value

type TXInput struct {
	InputID         string
	InputValue      float64 // amount of tokens
	SenderPublicKey string
	TokenID         string
}

type DigitalwalletTransaction struct {
	Sender    string
	Receiver  string
	TokenID   string
	Amount    float64
	Time      time.Time
	Signature string
	ServiceId string
}

type RefundDigitalWalletTx struct {
	Sender       string
	Receiver     string
	TokenID      string
	FlatCurrency bool
	Amount       float64
	Time         time.Time
	Signature    string
}
type PendingTransaction struct {
	Transaction
	SenderPK string
	Deleted  bool
}

type MixedTxStruct struct {
	TxObj        Transaction
	DigitalTxObj DigitalwalletTransaction
}

var Pending_transaction []PendingTransaction
var PendingValidationTxs = make(map[string][]string)

/*----------------function to add the transaction----------------- */
func AddTransaction(TransactionObj Transaction) string {
	// swap the empty senderPk with receiverPk in the case of adding coins to inovatian account
	var senderPK string
	inoAccPK := accountdb.GetFirstAccount().AccountPublicKey
	// if tx.filestruct  nil {
	// empty struct
	fileObj := TransactionObj.Filestruct
	if fileObj.FileSize == 0 {
		if len(TransactionObj.TransactionInput) == 0 {
			for _, txOut := range TransactionObj.TransactionOutPut {
				if txOut.RecieverPublicKey == inoAccPK {
					senderPK = inoAccPK
				}
			}
		} else {
			senderPK = TransactionObj.TransactionInput[0].SenderPublicKey
		}
		// add empty struct in tx obj2
		TransactionObj2 := TransactionDB{TransactionObj.TransactionID, TransactionObj.TransactionInput, TransactionObj.TransactionOutPut, TransactionObj.TransactionTime, "", fileObj}
		AddTransactiondb(TransactionObj2)
		pendingTx := PendingTransaction{TransactionObj, senderPK, false}
		Pending_transaction = append(Pending_transaction, pendingTx)
	} else {

		// add file struct into tx pool
		TransactionObj2 := TransactionDB{TransactionObj.TransactionID, TransactionObj.TransactionInput, TransactionObj.TransactionOutPut, TransactionObj.TransactionTime, "", fileObj}
		AddTransactiondb(TransactionObj2)
		pendingTx := PendingTransaction{TransactionObj, fileObj.Ownerpk, false}
		Pending_transaction = append(Pending_transaction, pendingTx)
	}
	return ""
}

// func AddTransaction(TransactionObj Transaction) string {
// 	// fmt.Println("add transaction    ", TransactionObj)
// 	// Timew := "2019-08-28 17:50:40 +0200 EET"
// 	// t, _:= time.Parse(time.RFC3339, Timew)
// 	// hashtransaction := globalPkg.CreateHash(t,  fmt.Sprintf("%s", TransactionObj), 3)
// 	// fmt.Println("   00time ---    "  ,TransactionObj.TransactionTime)
// 	// fmt.Println("   ###   isra   _____    " ,hashtransaction )
// 	// fmt.Println("    ***       " ,TransactionObj.TransactionID )
// 	// if TransactionObj.TransactionID == hashtransaction{
// 	// 	fmt.Println("=====================add transaction====================")
// 	TransactionObj2 := TransactionDB{TransactionObj.TransactionID, TransactionObj.TransactionInput, TransactionObj.TransactionOutPut, TransactionObj.TransactionTime, ""}
// 	AddTransactiondb(TransactionObj2)
// 	Pending_transaction = append(Pending_transaction, TransactionObj)

// 	return ""
// 	}
// 	fmt.Println("===================== not add transaction====================")
// 	return "hash not equal"
// }

/*----------------function to delete the transaction----------------- */
func DeleteTransaction(TransactionObj Transaction) string {
	for index, transactionExistsObj := range Pending_transaction {
		if ConvertTransactionToStr(transactionExistsObj.Transaction) == ConvertTransactionToStr(TransactionObj) {
			Pending_transaction = append(Pending_transaction[:index], Pending_transaction[index+1:]...)

			for _, txInput := range TransactionObj.TransactionInput {
				for index, pendingTxInputID := range PendingValidationTxs[transactionExistsObj.SenderPK] {
					if txInput.InputID == pendingTxInputID {
						PendingValidationTxs[transactionExistsObj.SenderPK] = append(
							PendingValidationTxs[transactionExistsObj.SenderPK][:index], PendingValidationTxs[transactionExistsObj.SenderPK][index+1:]...,
						)
					}
				}
			}
			if len(PendingValidationTxs[transactionExistsObj.SenderPK]) == 0 {
				delete(PendingValidationTxs, transactionExistsObj.SenderPK)
			}
			return ""
		}
	}
	return ""
}

func GetPendingTransactions() []Transaction {
	var Txs []Transaction
	for _, pendingTx := range Pending_transaction {
		Txs = append(Txs, pendingTx.Transaction)
	}
	return Txs
}

func ConvertTransactionToStr(trans Transaction) string {
	strID := trans.TransactionID
	strInp := ""
	strOut := ""
	strTime := trans.TransactionTime.String()
	for _, input := range trans.TransactionInput {
		strInp = strInp + input.InputID + fmt.Sprint(input.InputValue) + input.SenderPublicKey + input.TokenID
	}
	for _, output := range trans.TransactionOutPut {
		strOut = strOut + fmt.Sprint(output.OutPutValue) + fmt.Sprint(output.IsFee) + output.TokenID
	}
	log.Println("ConvertTransactionToStr ", strID+strInp+strOut+strTime)
	return "" + strID + strInp + strOut + strTime
}

func CheckReadyTransaction() bool {
	for _, transactionObj := range Pending_transaction {
		t := globalPkg.UTCtime()
		Subtime := (t.Sub(transactionObj.TransactionTime)).Seconds()
		if Subtime > 10 {
			return true
		}
	}
	return false
}
