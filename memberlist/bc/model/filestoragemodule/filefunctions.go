package filestoragemodule

import (
	// "fmt"
	"math/rand"
	"strings"
	"time"

	"../transaction"

	"../accountdb"
	"../globalPkg"
	"../validator"
)

// randomValidator chooser 3 validatore at random
func randomValidator(validatorlist int) []int {
	rand.Seed(time.Now().Unix()) // initialize global pseudo random generator
	var indices []int
	var index int
	if validatorlist <= 3 {
		for j := 0; j < validatorlist; j++ {
			indices = append(indices, j)
		}
		return indices
	}
	// else

	//chunkthree:
	for i := 0; i < 3; i++ {
		index = rand.Intn(validatorlist)
		if !Containsindex(indices, index) {
			indices = append(indices, index)
		}
	}
	if len(indices) < 3 && validatorlist >= 3 {
		//goto chunkthree
		for k := 0; k < validatorlist; k++ {
			if !Containsindex(indices, k) {
				indices = append(indices, k)
				if len(indices) == 3 {
					break
				}
			}

		}
	}
	return indices
}

//Containsindex Contains tells whether a contains x.
func Containsindex(indices []int, index int) bool {
	for _, n := range indices {
		if index == n {
			return true
		}
	}
	return false
}
func checkIndxInTxPool(accountObj *accountdb.AccountStruct) string {
	txs := transaction.Pending_transaction
	// fmt.Println(len(txs))
	var fileIds []string
	for _, tx := range txs {
		if tx.SenderPK == accountObj.AccountPublicKey {
			fileObj := tx.Transaction.Filestruct
			if fileObj.FileSize != 0 {
				fileIds = append(fileIds, fileObj.Fileid)
			}
		}
	}
	if len(fileIds) != 0 {
		return fileIds[len(fileIds)-1]
	}
	return ""

}

//getLastIndexFile for account
func getLastIndexFile(accountObj accountdb.AccountStruct) (string, accountdb.AccountStruct) {

	fid := checkIndxInTxPool(&accountObj)
	if fid != "" {
		return fid, accountObj
	}
	if len(accountObj.Filelist) == 0 {
		return "-1", accountObj
	}

	lastindex := accountObj.Filelist[len(accountObj.Filelist)-1].Fileid
	return lastindex, accountObj
}

//FileIndex create index for file
func FileIndex(accountObj accountdb.AccountStruct) string {
	LastIndex, account := getLastIndexFile(accountObj)

	index := 0
	if LastIndex != "-1" {
		res := strings.Split(LastIndex, "_")
		// fmt.Println("================index ===============     ", res[len(res)-1])
		index = globalPkg.ConvertFixedLengthStringtoInt(res[len(res)-1]) + 1
	}
	timpIndex, _ := globalPkg.ConvertIntToFixedLengthString(index, globalPkg.GlobalObj.StringFixedLength)

	currentIndex := account.AccountIndex + "_" + globalPkg.GetHash([]byte(validator.CurrentValidator.ValidatorIP)) + "_" + timpIndex
	return currentIndex
}

// check if validator ip exist or not
func contains(actives []validator.ValidatorStruct, ip string) int {
	for i, v := range actives {
		if v.ValidatorIP == ip {
			return i
		}
	}
	return -1
}

// containspk tells whether a contains x.
func containspk(a []accountdb.AccountStruct, pk string) bool {
	for _, n := range a {
		if pk == n.AccountPublicKey {
			return true
		}
	}
	return false
}

//containsfileid tells whether a contains x.
func containsfileid(a []string, fileid string) bool {
	for _, n := range a {
		if fileid == n {
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

//containsfileidinfilelist tells whether a contains x.
func containsfileidinfilelist(a []accountdb.FileList, fileid string) int {
	for index, n := range a {
		if fileid == n.Fileid {
			return index
		}
	}
	return -1
}
