package filestoragemodule

import (
	"encoding/json"
	"net/http"
	"time"

	"../filestorage"
	"../globalPkg"

	"../account"
	"../logpkg"
	"../transaction"
)

// ExploreFiles explore all file for some user
func ExploreFiles(w http.ResponseWriter, r *http.Request) {
	//log
	now, userIP := globalPkg.SetLogObj(r)
	logobj := logpkg.LogStruct{"_", now, userIP, "macAdress", "ExploreFiles", "file", "_", "_", "_", 0}
	var obj ExploreBody
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&obj)
	if err != nil {
		globalPkg.SendError(w, "please enter your correct request")
		globalPkg.WriteLog(logobj, "please enter your correct request", "failed")
		return
	}
	time.Sleep(time.Millisecond * 10) // for handle unknown issue
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
	FilesData := ExploreResponse{}
	files := acc.Filelist
	for _, file := range files {
		FilesData.TotalSizeOwned += file.FileSize
		FilesData.OwnedFiles = append(FilesData.OwnedFiles, file)
	}

	sharefiles := filestorage.FindSharedfileByAccountIndex(acc.AccountIndex)
	// if sharefiles.AccountIndex == "" {
	// 	fmt.Println("-----------------not take share file ----------")
	// }
	if len(sharefiles.OwnerSharefile) != 0 {
		for _, sharefileObj := range sharefiles.OwnerSharefile {
			accountObj := account.GetAccountByAccountPubicKey(sharefileObj.OwnerPublicKey)
			for _, filelistObj := range accountObj.Filelist {
				if containsfileid(sharefileObj.Fileid, filelistObj.Fileid) {
					FilesData.TotalSizeShared += filelistObj.FileSize
					FilesData.SharedFile = append(FilesData.SharedFile, filelistObj)
				}
			}
		}
	}

	//check files in transaction pool with status deleted
	txs := transaction.Pending_transaction
	for _, tx := range txs {
		// owned files => owner delete files and get explore files
		if tx.SenderPK == acc.AccountPublicKey {
			fileObj := tx.Transaction.Filestruct
			if fileObj.FileSize != 0 && fileObj.Deleted == true {
				indexfileid := containsfileidinfilelist(FilesData.OwnedFiles, fileObj.Fileid)
				if indexfileid != -1 {
					FilesData.TotalSizeOwned -= fileObj.FileSize
					FilesData.OwnedFiles = append(FilesData.OwnedFiles[:indexfileid], FilesData.OwnedFiles[indexfileid+1:]...)
				}
			}
		}
		//check for shared files
		fileObj1 := tx.Transaction.Filestruct
		if fileObj1.FileSize != 0 && fileObj1.Deleted == true {
			indexfileidshare := containsfileidinfilelist(FilesData.SharedFile, fileObj1.Fileid)
			if indexfileidshare != -1 {
				FilesData.TotalSizeShared -= fileObj1.FileSize
				FilesData.SharedFile = append(FilesData.SharedFile[:indexfileidshare], FilesData.SharedFile[indexfileidshare+1:]...)
			}
		}
	}
	sendJSON, _ := json.Marshal(FilesData)
	globalPkg.SendResponse(w, sendJSON)
	globalPkg.WriteLog(logobj, "get files success", "success")
}
