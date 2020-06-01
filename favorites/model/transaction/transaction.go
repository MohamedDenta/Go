package transaction

import (
	"fmt"
	"time"

	"../accountdb"
	"../filestorage"
	"../globalPkg"
)

// Type :- 0 is normal transaction.
//  	   1 is token creation transaction.
// 		   2 is refund to flat currency.
// 		   3 is refund to ino token.

//Transaction structure
type Transaction struct {
	Validator           string
	TransactionID       string
	Type                int8
	TransactionInput    []TXInput
	TransactionOutPut   []TXOutput
	TransactionTime     time.Time
	TransactionTimeUser time.Time
	ServiceId           string
	NotBilling          bool
	OwnershipID         string                 // store token id that be owned
	Filestruct          filestorage.FileStruct //file struct data to store in block
}

//TXOutput transaction output
type TXOutput struct {
	OutPutValue       float64 // amount of tokens
	RecieverPublicKey string
	IsFee             bool
	TokenID           string //
	// if token id == 1 is ino coin
}

// token initiator, inovation account will have the token and send transaction, the sender is inovation and the receiver is the initiator
// inovation take token total supply token value

//TXInput transaction input
type TXInput struct {
	InputID         string
	InputValue      float64 // amount of tokens
	SenderPublicKey string
	TokenID         string
}

// PendingTransaction pool of transaction before added in block
type PendingTransaction struct {
	Transaction
	SenderPK string
	Deleted  bool
}

//Pending_transaction arr pending tx
var Pending_transaction []PendingTransaction

//PendingValidationTxs validate tx pool
var PendingValidationTxs = make(map[string][]string)

// AddTransaction to add the transaction
func (TransactionObj Transaction) AddTransaction() string {
	// swap the empty senderPk with receiverPk in the case of adding coins to inovatian account
	var senderPK string
	inoAccPK := accountdb.GetFirstAccount().AccountPublicKey
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

		TransactionObj2 := TransactionDB{TransactionObj.TransactionID, TransactionObj.TransactionInput, TransactionObj.TransactionOutPut, TransactionObj.TransactionTimeUser, "", fileObj}
		TransactionObj2.AddTransactiondb()
		pendingTx := PendingTransaction{TransactionObj, senderPK, false}
		Pending_transaction = append(Pending_transaction, pendingTx)
	} else {

		// add file struct into tx pool
		TransactionObj2 := TransactionDB{TransactionObj.TransactionID, TransactionObj.TransactionInput, TransactionObj.TransactionOutPut, TransactionObj.TransactionTime, "", fileObj}
		TransactionObj2.AddTransactiondb()
		pendingTx := PendingTransaction{TransactionObj, fileObj.Ownerpk, false}
		Pending_transaction = append(Pending_transaction, pendingTx)
	}
	return ""
}

// DeleteTransaction function to delete the transaction
func (TransactionObj Transaction) DeleteTransaction() string {
	for index, transactionExistsObj := range Pending_transaction {
		if transactionExistsObj.Transaction.ConvertTransactionToStr() == TransactionObj.ConvertTransactionToStr() {
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

//GetPendingTransactions get pending tx
func GetPendingTransactions() []Transaction {
	var Txs []Transaction
	for _, pendingTx := range Pending_transaction {
		Txs = append(Txs, pendingTx.Transaction)
	}
	return Txs
}

//ConvertTransactionToStr convert transaction to string
func (TransactionObj Transaction) ConvertTransactionToStr() string {
	strID := TransactionObj.TransactionID
	strInp := ""
	strOut := ""
	strTime := TransactionObj.TransactionTime.String()
	for _, input := range TransactionObj.TransactionInput {
		strInp = strInp + input.InputID + fmt.Sprint(input.InputValue) + input.SenderPublicKey + input.TokenID
	}
	for _, output := range TransactionObj.TransactionOutPut {
		strOut = strOut + fmt.Sprint(output.OutPutValue) + fmt.Sprint(output.IsFee) + output.TokenID
	}
	// log.Println("ConvertTransactionToStr ", strID+strInp+strOut+strTime)
	return "" + strID + strInp + strOut + strTime
}

//CheckReadyTransaction check Ready transaction
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
