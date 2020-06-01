package validator

import (
	"crypto/ecdsa"
	"errors"
	"fmt"
	"strconv"
	"time"

	errorpk "../errorpk"
	globalPkg "../globalPkg"

	"github.com/mitchellh/mapstructure"
	"github.com/syndtr/goleveldb/leveldb"
)

//DB database instance
var DB *leveldb.DB

//Open state
var Open = false

//opendatabase open db
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

// closedatabase close db
func closedatabase() bool {
	return true
}

// CreateValidator insert Validator Struct
func CreateValidator(data *ValidatorStruct) bool {
	opendatabase()
	d, convert := globalPkg.ConvetToByte(*data, "Validator create Validator package")
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

//findValidatorByIP find some validator using ip
func findValidatorByIP(key string) (ValidatorStructObj ValidatorStruct, err error) {
	opendatabase()
	data, err := DB.Get([]byte(key), nil)
	if err != nil {
		errorpk.AddError("ValidatorStructFindByKey  ValidatorStructObj package", "can't get ValidatorStruct", "logic")
		return ValidatorStructObj, err
	}

	// json.Unmarshal(data, &ValidatorStructObj)
	(&ValidatorStructObj).Decode(data[:])
	closedatabase()
	return ValidatorStructObj, err
}

//GetAllValidators encryptted
func GetAllValidators() (values []ValidatorStruct) {
	opendatabase()
	iter := DB.NewIterator(nil, nil)
	for iter.Next() {
		value := iter.Value()
		var newdata ValidatorStruct
		(&newdata).Decode(value[:])
		values = append(values, newdata)
	}
	closedatabase()
	return values
}

//GetAllValidatorsDecrypted return all Validators decrypted from where it's stored in db
func GetAllValidatorsDecrypted() (values []ValidatorStruct) {
	opendatabase()
	iter := DB.NewIterator(nil, nil)
	for iter.Next() {
		value := iter.Value()

		var newdata ValidatorStruct
		// json.Unmarshal(value, &newdata)
		(&newdata).Decode(value[:])
		//extracting ECC public key :
		var public ecdsa.PublicKey
		var pub ecdsa.PublicKey

		rt := new(RetrieveECCPublicKey)
		pk := (&newdata).DecryptPK()
		(rt).Decode([]byte(pk)[:])
		if rt.CurveParams == nil {
			fmt.Println("can't umarshal this obj check if this obj != nil")
			panic(errors.New("error in decrypting validator public key"))
		}
		public.Curve = rt.CurveParams
		public.X = rt.MyX
		public.Y = rt.MyY
		mapstructure.Decode(public, &pub)
		newdata.ECCPublicKey = pub

		//Extracting ECC private key
		var privateECCKey ecdsa.PrivateKey
		var priv ecdsa.PrivateKey
		rt1 := new(RetrieveECCPrivateKey)
		if newdata.EncECCPriv != "" {
			timpPrivate := (&newdata).DecryptPrvK()
			if !(rt1.Decode([]byte(timpPrivate))) {
				panic(errors.New("error in decrypting private key "))
			}
			privateECCKey.PublicKey = pub
			privateECCKey.D = rt1.MyD
			mapstructure.Decode(privateECCKey, &priv)
			newdata.ECCPrivateKey = &priv
		} else {
			fmt.Println("no encoded private key")
		}
		values = append(values, newdata)
	}
	closedatabase()

	return values
}

//DeleteValidatorStruct delete ValidatorStruct by key
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

//---update validator DataBase--------------------------------------------------
func (ValidatorObj *ValidatorStruct) updateValidatorStruct() bool {
	ValidatorStructObj, err := findValidatorByIP(ValidatorObj.ValidatorIP)
	if err != nil || !DeleteValidatorStruct(ValidatorStructObj.ValidatorIP) {
		return false
	}
	ValidatorObj.ECCPublicKey = *new(ecdsa.PublicKey)
	ValidatorObj.ECCPrivateKey = new(ecdsa.PrivateKey)
	if CreateValidator(ValidatorObj) {
		return true
	}
	return false

}

//GetActiveValidators Get Active validator
func GetActiveValidators() (values []ValidatorStruct) {
	opendatabase()
	iter := DB.NewIterator(nil, nil)
	for iter.Next() {
		value := iter.Value()
		var newdata ValidatorStruct
		(&newdata).Decode(value)
		if newdata.ValidatorActive {
			values = append(values, newdata)
		}
	}
	closedatabase()

	return values
}

//getValidatorByPK get validator data by public key

func getValidatorByPK(validatorPublickey string) (validatorstructobj ValidatorStruct) {
	validators := GetAllValidators()
	for _, validatorObj := range validators {
		if validatorObj.ValidatorIP == validatorObj.ValidatorIP {
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

// GetValidatorsByTimeRange ...
func GetValidatorsByTimeRange(dateTime time.Time) (validatorobj []ValidatorStruct) {
	opendatabase()
	iter := DB.NewIterator(nil, nil)
	for iter.Next() {
		value := iter.Value()
		var newdata ValidatorStruct
		(&newdata).Decode(value)
		if newdata.ValidatorRegisterTime.After(dateTime) || dateTime.Equal(newdata.ValidatorRegisterTime) {
			validatorobj = append(validatorobj, newdata)
		}
	}
	iter.Release()
	closedatabase()
	return validatorobj
}
