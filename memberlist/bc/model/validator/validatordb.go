package validator

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	errorpk "../errorpk" //  write an error on the json file
	globalPkg "../globalPkg"

	"github.com/syndtr/goleveldb/leveldb"
)

//------------------------------------------------------------------------------------------------------------
//struct for object to be saved in db
//------------------------------------------------------------------------------------------------------------
type ValidatorStruct struct {
	ValidatorIP            string
	ValidatorSoketIP       string
	ValidatorPublicKey     string
	ValidatorPrivateKey    string
	ValidatorStakeCoins    float64
	ValidatorRegisterTime  time.Time
	ValidatorActive        bool
	ValidatorLastHeartBeat time.Time
	ValidatorRemove        bool
	Index                  string
}

var DB *leveldb.DB
var Open = false

//------------------------------------------------------------------------------------------------------------
// create or open DB if exist
//------------------------------------------------------------------------------------------------------------
func opendatabase() bool {
	if !Open {
		Open = true
		DBpath := "Database/ValidatorStruct"
		var err error
		DB, err = leveldb.OpenFile(DBpath, nil)
		if err != nil {
			errorpk.AddError("opendatabase ValidatorStruct package", "can't open the database", "Logic")
			return false
		}
		return true
	}
	return true

}

//------------------------------------------------------------------------------------------------------------
// close DB if exist
//------------------------------------------------------------------------------------------------------------
func closedatabase() bool {

	return true
}

//-------------------------------------------------------------------------------------------------------------
// insert Validator Struct
//-------------------------------------------------------------------------------------------------------------
func ValidatorCreate(data ValidatorStruct) bool {

	opendatabase()
	d, convert := globalPkg.ConvetToByte(data, "Validator create Validator package")
	if !convert {
		closedatabase()
		return false
	}
	err := DB.Put([]byte(data.ValidatorIP), d, nil)
	closedatabase()

	if err != nil {
		errorpk.AddError("validatorCreate ValidatorStruct package", "can't create ValidatorStructObj", "logic")
		return false
	}
	return true
}

//-------------------------------------------------------------------------------------------------------------
// select By key (VALIDATOR iP) BlockStruct
//-------------------------------------------------------------------------------------------------------------
func findValidatorByIP(key string) (ValidatorStructObj ValidatorStruct, err error) {

	opendatabase()

	data, err := DB.Get([]byte(key), nil)

	if err != nil {
		errorpk.AddError("ValidatorStructFindByKey  ValidatorStructObj package", "can't get ValidatorStruct", "logic")
		fmt.Println("Not foundddddddddddddddddddddddddddddddddddddddd")
		return ValidatorStructObj, err
	}

	json.Unmarshal(data, &ValidatorStructObj)
	closedatabase()
	return ValidatorStructObj, err
}

//-------------------------------------------------------------------------------------------------------------
// get all Validatorstructs
//-------------------------------------------------------------------------------------------------------------
func GetAllValidators() (values []ValidatorStruct) {
	opendatabase()
	iter := DB.NewIterator(nil, nil)
	for iter.Next() {

		value := iter.Value()
		var newdata ValidatorStruct
		json.Unmarshal(value, &newdata)
		values = append(values, newdata)
	}
	closedatabase()

	return values
}

//-------------------------------------------------------------------------------------------------------------
// delete ValidatorStruct by key
//-------------------------------------------------------------------------------------------------------------
func DeleteValidatorStruct(key string) (delete bool) {
	opendatabase()

	err := DB.Delete([]byte(key), nil)
	closedatabase()
	if err != nil {
		errorpk.AddError("ValidatorDeleted ErrorValidatorStruct package", "can't delete validatorstruct", "logic")
		return false
	}

	return true
}

///////////////////////////////////
//-----------------------------
//---update validator DataBase--------------------------------------------------

func updateValidatorStruct(ValidatorObj ValidatorStruct) bool {
	ValidatorStructObj, err := findValidatorByIP(ValidatorObj.ValidatorIP)
	if err != nil || !DeleteValidatorStruct(ValidatorStructObj.ValidatorIP) {
		return false
	}

	if ValidatorCreate(ValidatorObj) {
		return true
	}
	return false

}

//---------------------------------------------------------------------------------------
//                            Get Active validator
//------------------------------------------------------------------------------------
func GetActiveValidators() (values []ValidatorStruct) {
	opendatabase()
	iter := DB.NewIterator(nil, nil)
	for iter.Next() {

		value := iter.Value()
		var newdata ValidatorStruct
		json.Unmarshal(value, &newdata)
		if newdata.ValidatorActive {
			values = append(values, newdata)
		}
	}
	closedatabase()

	return values
}

//----------------------------------------------------------------------------------------------------
//     getvalidator By publick Key
//----------------------------------------------------------------------------------------------------

func getValidatorByPK(validatorPublickey string) (validatorstructobj ValidatorStruct) {
	validators := GetAllValidators()
	for _, validatorObj := range validators {
		if validatorObj.ValidatorPublicKey == validatorPublickey {
			return validatorObj
		}
	}
	return validatorstructobj
}

//NewIndex of a validaor
func NewIndex() (newIndex string) {
	lst := GetAllValidators()
	if lst == nil {
		newIndex = "1"
	} else {
		lastValidator := lst[len(lst)-1]
		i, _ := strconv.Atoi(lastValidator.Index)
		i = i + 1
		newIndex = strconv.Itoa(i)
	}
	return newIndex
}
