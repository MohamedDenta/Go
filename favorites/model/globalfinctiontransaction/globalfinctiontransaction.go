package globalfinctiontransaction

import (
	"crypto/sha256"
	// "encoding/base64"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"

	"../block"
	"../validator"
)

var InputIndex string

/*----------function to convert Convert Fixed Length String to Int----------*/
func ConvertFixedLengthStringtoInt(key string) (stringform int) {
	for index := 0; index < len(key); index++ {
		if key[index:index+1] != "0" {
			number, _ := strconv.Atoi(key[index:len(key)])
			return number
		}

	}
	return 0
}

/*----------function to convert integar to fixed digits of string----------*/
func ConvertIntToFixedLengthString(key int, length int) (stringform string, err bool) {
	stringform = strconv.Itoa(key)
	stringlen := len(stringform)
	if stringlen > length {
		return "", false
	}
	for i := 0; i < length-stringlen; i++ {
		stringform = "0" + stringform
	}
	return stringform, true
}
func SetInputIndex(lastindex string, transactionhash string) string {
	validatorip := validator.CurrentValidator.ValidatorIP
	validatorhash := CreateValidatorHash(validatorip)
	InputIndex1 := validatorhash + "_" + lastindex + "_" + transactionhash
	InputIndex = InputIndex1
	AddTransactionIndexToTemMap(lastindex, validator.CurrentValidator.ValidatorIP)
	return InputIndex
}
func GetINputIndex() string {
	return InputIndex
}

func CreateValidatorHash(validator string) string {
	h := sha256.New()
	h.Write([]byte(validator))
	sum := h.Sum(nil)

	var hash [32]byte
	for i := 0; i < len(sum); i++ {
		hash[i] = sum[i]
	}
	return hex.EncodeToString(hash[:])
}

func GetlastTransactionIndexLinearSearch() string {
	var lastindex string
	blockobj := block.GetLastBlock()
	// fmt.Println("alaablockobj", blockobj)
research:
	blocktransactionslst := blockobj.BlockTransactions
	if len(blocktransactionslst) != 0 {
		for i := len(blocktransactionslst) - 1; i >= 0; i-- {
			transactionid := blocktransactionslst[i].TransactionID
			if transactionid != "" && blocktransactionslst[i].Validator == validator.CurrentValidator.ValidatorIP {
				transacionIDdata := strings.Split(transactionid, "_")
				lastindex, _ = ConvertIntToFixedLengthString(ConvertFixedLengthStringtoInt(transacionIDdata[1]), 5)
				AddTransactionIndexToTemMap(lastindex, validator.CurrentValidator.ValidatorIP)
				break
			} else {
				continue
			}
		}
		if lastindex == "" { //no Inputs in lastblock
			if blockobj.BlockIndex != "000000000000000000000000000000" {
				previousblockint := ConvertFixedLengthStringtoInt(blockobj.BlockIndex) - 1
				previousblockIndex, _ := ConvertIntToFixedLengthString(previousblockint, 30)
				blockobj = block.GetBlockInfoByID(previousblockIndex)
				goto research
			} else {
				lastindex = "00000 "
				AddTransactionIndexToTemMap("00000", validator.CurrentValidator.ValidatorIP)
			}
		}
	} else {
		fmt.Println("--validator.CurrentValidator.ValidatorIP", validator.CurrentValidator.ValidatorIP)
		lastindex = "00000"
		AddTransactionIndexToTemMap("00000", validator.CurrentValidator.ValidatorIP)
	}
	return lastindex
}

func GetlastTransactionIndexFromMap(validator string) string {
	lastindx, _ := ConvertIntToFixedLengthString(ConvertFixedLengthStringtoInt(ValidatorMap[validator])+1, 5)
	return lastindx
}

var ValidatorMap = make(map[string]string)

func AddTransactionIndexToTemMap(lastindex, validatorobj string) {
	oldmap := GetTransactionIndexTemMap()
	// fmt.Println("+____________________________________+oldmap+___________________________", oldmap)
	if oldmap != nil {
		for key, val := range oldmap {
			ValidatorMap[key] = val
		}
	}
	ValidatorMap[validatorobj] = lastindex
	// fmt.Println("jjjjjjjjjjjjjjjjjjjjjjjjjjjj", ValidatorMap, "current")
}

func GetTransactionIndexTemMap() map[string]string {
	return ValidatorMap
}

func SetTransactionIndexTemMap(M map[string]string) map[string]string {
	oldmap := GetTransactionIndexTemMap()
	for key, val := range oldmap {
		ValidatorMap[key] = val
	}
	for key, val := range M {
		ValidatorMap[key] = val
	}
	// fmt.Println("jjjjjjjjjjjjjjjjjjjjjjjjjjjjjjjj", ValidatorMap)
	return ValidatorMap
}
func FirstRun(transactionhash string) string {
	lastindex := GetlastTransactionIndexLinearSearch()
	transacionID := SetInputIndex(lastindex, transactionhash)
	return transacionID
}

func TransacionIDFromTemp(transactionhash string) string {
	lastindex := GetlastTransactionIndexFromMap(validator.CurrentValidator.ValidatorIP)
	transacionID := SetInputIndex(lastindex, transactionhash)
	return transacionID
}

func MapTempRollBack(validator string, index string) string {
	lastindex, _ := ConvertIntToFixedLengthString(ConvertFixedLengthStringtoInt(index)-1, 5)
	AddTransactionIndexToTemMap(lastindex, validator)

	return lastindex
}
func CheckTransactionID(lastindex, transactionhash, validator string) string {
	validatorhash := CreateValidatorHash(validator)
	InputIndex1 := validatorhash + "_" + lastindex + "_" + transactionhash
	tempmap := GetTransactionIndexTemMap()
	for key, val := range tempmap {
		if key == validator && val == lastindex {
			return InputIndex1
		}
	}
	return ""
}

func GetlastTransactionIndexLinearSearch2(validator2 string) string {
	var lastindex string
	blockobj := block.GetLastBlock()
	// fmt.Println("alaablockobj", blockobj)
research:
	blocktransactionslst := blockobj.BlockTransactions
	if len(blocktransactionslst) != 0 {
		for i := len(blocktransactionslst) - 1; i >= 0; i-- {
			transactionid := blocktransactionslst[i].TransactionID
			if transactionid != "" && blocktransactionslst[i].Validator == validator2 {
				transacionIDdata := strings.Split(transactionid, "_")
				lastindex, _ = ConvertIntToFixedLengthString(ConvertFixedLengthStringtoInt(transacionIDdata[1]), 5)
				AddTransactionIndexToTemMap(lastindex, validator2)
				break
			} else {
				continue
			}
		}
		if lastindex == "" { //no Inputs in lastblock
			if blockobj.BlockIndex != "000000000000000000000000000000" {
				previousblockint := ConvertFixedLengthStringtoInt(blockobj.BlockIndex) - 1
				previousblockIndex, _ := ConvertIntToFixedLengthString(previousblockint, 30)
				blockobj = block.GetBlockInfoByID(previousblockIndex)
				goto research
			} else {
				lastindex = "00000 "
				AddTransactionIndexToTemMap("00000", validator2)
			}
		}
	} else {
		fmt.Println("--validator.CurrentValidator.ValidatorIP", validator2)
		lastindex = "00000"
		AddTransactionIndexToTemMap("00000", validator2)
	}
	return lastindex
	//----------------------------------------------------------------------------------------
	//-----------------------do not remove this comments please-------------------------------
	//----------------------------------------------------------------------------------------
	// 	var lastindex string
	// 	blockobj := block.GetLastBlock()
	// 	fmt.Println("GetlastTransactionIndexLinearSearch2", blockobj)
	// research:
	// 	blocktransactionslst := blockobj.BlockTransactions
	// 	if len(blocktransactionslst) != 0 {
	// 		for i := len(blocktransactionslst) - 1; i >= 0; i-- {
	// 			transactionid := blocktransactionslst[i].TransactionID
	// 			if transactionid != "" && blocktransactionslst[i].Validator == validator2 {
	// 				transacionIDdata := strings.Split(transactionid, "_")
	// 				lastindex, _ = ConvertIntToFixedLengthString(ConvertFixedLengthStringtoInt(transacionIDdata[1]), 5)
	// 				AddTransactionIndexToTemMap(lastindex, validator2)
	// 				break
	// 			} else {
	// 				continue
	// 			}
	// 		}
	// 		if lastindex == "" { //no Inputs in lastblock
	// 			if blockobj.BlockIndex != "000000000000000000000000000000" {
	// 				previousblockint := ConvertFixedLengthStringtoInt(blockobj.BlockIndex) - 1
	// 				previousblockIndex, _ := ConvertIntToFixedLengthString(previousblockint, 30)
	// 				blockobj = block.GetBlockInfoByID(previousblockIndex)
	// 				goto research
	// 			} else {
	// 				lastindex = "00000 "
	// 				AddTransactionIndexToTemMap("00000", validator.CurrentValidator.ValidatorIP)
	// 			}
	// 		}
	// 	} else {
	// 		lastindex = "00000"
	// 		AddTransactionIndexToTemMap("00000", validator.CurrentValidator.ValidatorIP)
	// 	}
	// 	return lastindex
}
