package BackUp

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"../accountdb"
	"os"
	"reflect"
	"time"

	"../admin"
	"../block"
	"../errorpk"
	"../heartbeat"
	"../token"
	"../transaction"
	"../validator"
)

type Structure struct {
	Name        string
	Type        string
	Initalvalue string
}

//copy src file to dst filoder if not exist it will creat it
//then return the num of bytes copid and the first error that happen
// if there is no error then it will be nil
func copy(src, dst string) (int64, error) {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		// fmt.Println("error at state of folder")
		// fmt.Println(err)
		return 0, err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return 0, fmt.Errorf("%s is not a regular file", src)
	}

	source, err := os.Open(src)
	if err != nil {
		// fmt.Println("error at open  folder")
		// fmt.Println(err)
		return 0, err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		// fmt.Println("error at creat folder")
		// fmt.Println(err)
		return 0, err
	}
	defer destination.Close()
	nBytes, err := io.Copy(destination, source)
	return nBytes, err
}

var fileName string

func CreatBackup() {
	for {
		// fmt.Println("will sleep now")
		time.Sleep(24 * time.Hour)
		//fmt.Println("omar")

		t := time.Now()
		fileName = t.Format("20060102")
		//fmt.Println("done sleep")
		if createFolder("DatabaseBackUp") {
			if createFolder("DatabaseBackUp/" + fileName) {

				// fmt.Println("will create file now")

				copy("Database/AccountStruct/CURRENT.bak", "DatabaseBackUp/"+fileName+"/AccountStruct.bak")

				copy("Database/BlockStruct/CURRENT.bak", "DatabaseBackUp/"+fileName+"/BlockStruct.bak")

				copy("Database/ValidatorStruct/CURRENT.bak", "DatabaseBackUp/"+fileName+"/ValidatorStruct.bak") //********

				copy("Database/errorPk/CURRENT.bak", "DatabaseBackUp/"+fileName+"/errorPk.bak")

				copy("Database/HeartBeatStruct/CURRENT.bak", "DatabaseBackUp/"+fileName+"/HeartBeatStruct.bak")

				copy("Database/TempAccount/AccountEmailStruct/CURRENT.bak", "DatabaseBackUp/"+fileName+"/AccountEmailStruct.bak")

				copy("Database/TempAccount/AccountLastUpdatedTimestruct/CURRENT.bak", "DatabaseBackUp/"+fileName+"/AccountLastUpdatedTimestruct.bak")

				copy("Database/TempAccount/AccountNameStruct/CURRENT.bak", "DatabaseBackUp/"+fileName+"/AccountNameStruct.bak")

				copy("Database/TempAccount/AccountPhoneNumberStruct/CURRENT.bak", "DatabaseBackUp/"+fileName+"/AccountPhoneNumberStruct.bak")

				copy("Database/TempAccount/AccountPublicKeyStruct/CURRENT.bak", "DatabaseBackUp/"+fileName+"/AccountPublicKeyStruct.bak")

				copy("Database/TokenStruct/CURRENT.bak", "DatabaseBackUp/"+fileName+"/TokenStruct.bak") //***

				copy("Database/AdminStruct/CURRENT.bak", "DatabaseBackUp/"+fileName+"/adminStruct.bak")

				copy("Database/TransactionStruct/CURRENT.bak", "DatabaseBackUp/"+fileName+"/TransactionStruct.bak") //**

			}
		}
		CreateDatabaseStruct()
	}
}

func createFolder(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		os.Mkdir(path, 0700)
		// fmt.Println("folder is not exist and created successfully")
		return true // folder is not exist and created successfully
	}
	// fmt.Println("cannot creat file")

	return false // the folder exist
}

/*----------function to get the structure of a struct----------*/
func CreateStructure(data interface{}) []byte {
	structlst := []Structure{}
	e := reflect.ValueOf(data).Elem()
	for i := 0; i < e.NumField(); i++ {
		varName := e.Type().Field(i).Name
		varType := e.Type().Field(i).Type
		varValue := e.Field(i).Interface()
		obj := Structure{}
		obj.Name = string(varName)
		// blockstruct.Type = string(varType.Name())
		str := fmt.Sprintf("%v", varType)
		obj.Type = str
		str2 := fmt.Sprintf("%v", varValue)
		obj.Initalvalue = str2
		structlst = append(structlst, obj)

	}
	//fmt.Println(structlst)
	file, _ := json.MarshalIndent(structlst, "", " ")
	//fmt.Println(file)
	return file
}

func CreateDatabaseStruct() {
	timpPath := "DatabaseBackUp/" + fileName //filename
	ioutil.WriteFile(timpPath+"/block.json", CreateStructure(&block.BlockStruct{}), 0644)
	ioutil.WriteFile(timpPath+"/account.json", CreateStructure(&accountdb.AccountStruct{}), 0644)
	ioutil.WriteFile(timpPath+"/heartbeat.json", CreateStructure(&heartbeat.HeartBeatStruct{}), 0644)
	ioutil.WriteFile(timpPath+"/errorstruct.json", CreateStructure(&errorpk.ErrorStruct{}), 0644)
	ioutil.WriteFile(timpPath+"/AccountEmailStruct.json", CreateStructure(&accountdb.AccountEmailStruct{}), 0644)
	ioutil.WriteFile(timpPath+"/AccountNameStruct.json", CreateStructure(&accountdb.AccountNameStruct{}), 0644)
	ioutil.WriteFile(timpPath+"/AccountPublicKeyStruct.json", CreateStructure(&accountdb.AccountPublicKeyStruct{}), 0644)
	ioutil.WriteFile(timpPath+"/AccountPhoneNumberStruct.json", CreateStructure(&accountdb.AccountPhoneNumberStruct{}), 0644)
	ioutil.WriteFile(timpPath+"/AccountLastUpdatedTimestruct.json", CreateStructure(&accountdb.AccountLastUpdatedTimestruct{}), 0644)
	ioutil.WriteFile(timpPath+"/admin.json", CreateStructure(&admin.AdminStruct{}), 0644)
	ioutil.WriteFile(timpPath+"/validator.json", CreateStructure(&validator.ValidatorStruct{}), 0644)
	ioutil.WriteFile(timpPath+"/token.json", CreateStructure(&token.StructToken{}), 0644)
	ioutil.WriteFile(timpPath+"/transaction.json", CreateStructure(&transaction.Transaction{}), 0644)

}
