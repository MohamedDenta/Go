package filestoragemodule

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"../account"
	"../accountdb"
	"../block"
	"../broadcastTcp"
	"../cryptogrpghy"
	"../filestorage"
	"../globalPkg"
	"../logpkg"
	"../transaction"
)

// DeleteFile delete a specific file
func DeleteFile(w http.ResponseWriter, r *http.Request) {
	//log
	now, userIP := globalPkg.SetLogObj(r)
	logobj := logpkg.LogStruct{"_", now, userIP, "macAdress", "DeleteFile", "file", "_", "_", "_", 0}
	var obj RetrieveBody
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&obj)
	if err != nil {
		globalPkg.SendError(w, "please enter your correct request")
		globalPkg.WriteLog(logobj, "please enter your correct request", "failed")
		return
	}
	time.Sleep(time.Millisecond * 10)
	acc := account.GetAccountByAccountPubicKey(obj.Publickey)
	if acc.AccountPublicKey != obj.Publickey {
		globalPkg.SendError(w, "error in public key")
		globalPkg.WriteLog(logobj, "error in public key", "failed")
		return
	}
	if acc.AccountPassword != obj.Password {
		globalPkg.SendError(w, "error in password")
		globalPkg.WriteLog(logobj, "error in password", "failed")
		return
	}
	// 	tnow := globalPkg.UTCtime()
	// 	t, _ := time.Parse("2006-01-02T15:04:05Z07:00", obj.Time)
	// 	tfile := globalPkg.UTCtimefield(t)
	// 	timeDifference := tnow.Sub(tfile).Seconds()
	// 	if timeDifference > float64(globalPkg.GlobalObj.TxValidationTimeInSeconds) {
	// 		globalPkg.SendError(w, "please check your time")
	// 		globalPkg.WriteLog(logobj, "please check your time", "failed")
	// 		return
	// 	}

	// Signture string
	validSig := false
	pk := account.FindpkByAddress(acc.AccountPublicKey).Publickey
	if pk != "" {
		publickey := cryptogrpghy.ParsePEMtoRSApublicKey(pk)
		// signatureData := obj.FileID + obj.Publickey + obj.Password + obj.Time
		signatureData := obj.Publickey + obj.Password + obj.FileID
		validSig = cryptogrpghy.VerifyPKCS1v15(obj.Signture, signatureData, *publickey)
		validSig = true
	} else {
		validSig = false
	}
	if !validSig {
		globalPkg.SendError(w, "you are not allowed to delete")
		globalPkg.WriteLog(logobj, "you are not allowed to delete", "failed")
		return
	}
	// check user own this file id
	files := acc.Filelist
	found := false
	foundtx := false
	var selectedFile accountdb.FileList
	for _, file := range files {
		if file.Fileid == obj.FileID {
			found = true
			selectedFile = file
			break
		}
	}
	if !found {
		globalPkg.SendError(w, "You don't have this file")
		globalPkg.WriteLog(logobj, "You don't have this file", "failed")
		return
	}

	//check files in transaction pool with status deleted
	txs := transaction.Pending_transaction
	for _, tx := range txs {
		// owned files => owner delete files and get explore files
		if tx.SenderPK == acc.AccountPublicKey {
			fileObj := tx.Transaction.Filestruct
			if fileObj.FileSize != 0 && fileObj.Deleted == true && fileObj.Fileid == obj.FileID {
				foundtx = true
				break
			}
		}
	}
	if foundtx {
		globalPkg.SendError(w, "You don't have this file!")
		globalPkg.WriteLog(logobj, "You don't have this file", "failed")
		return
	}

	decryptIndexBlock1 := cryptogrpghy.KeyDecrypt("123456789", selectedFile.Blockindex)
	fmt.Println("Block Index to be deletd data ", decryptIndexBlock1)
	blkObj := block.GetBlockInfoByID(decryptIndexBlock1)
	var fStrct filestorage.FileStruct
	for _, tx := range blkObj.BlockTransactions {
		fStrct = tx.Filestruct
		if fStrct.Fileid == selectedFile.Fileid {
			fStrct = tx.Filestruct
			break
		}
	}
	fStrct.Deleted = true
	fStrct.Transactionid = globalPkg.CreateHash(fStrct.Timefile, fmt.Sprintf("%s", fStrct), 3)
	// add on trsansaction pool
	broadcastTcp.BoardcastingTCP(fStrct, "deletefile", "file")
	// delete file if share from share file table

	globalPkg.SendResponseMessage(w, "File Deleted Successfully")
	globalPkg.WriteLog(logobj, "File Deleted Successfully", "success")
}
