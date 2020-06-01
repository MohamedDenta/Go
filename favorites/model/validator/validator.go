package validator

import (
	"crypto/ecdsa"
	"crypto/rand"
	"fmt"
	"io"
	"log"
	"net/smtp"

	"../admin"
	errorpk "../errorpk"
	"../globalPkg"
)

//AddValidator function to add validator in the validators list
func (validatorObj *ValidatorStruct) AddValidator() string {
	validatorObj.Index = NewIndex()
	if validatorObj.validationAdd() {
		if validatorObj.ValidatorIP == CurrentValidator.ValidatorIP {
			ValidatorsLstObj = append(ValidatorsLstObj, *validatorObj) //add the current validator pivate key only in validator list
			validatorObj.ECCPublicKey = *new(ecdsa.PublicKey)          //delete both keys from validator to store in db
			validatorObj.ECCPrivateKey = new(ecdsa.PrivateKey)
		} else {
			validatorObj.ECCPrivateKey = new(ecdsa.PrivateKey)         //delete private key first
			ValidatorsLstObj = append(ValidatorsLstObj, *validatorObj) //add the other node validator with public key only to the list
			validatorObj.ECCPublicKey = *new(ecdsa.PublicKey)          // then delete the public key from validatore before saving it to the db
		}
		CreateValidator(validatorObj)
		return ""
	}
	return errorpk.AddError("Add Validator validator package", "enter correct validator object not exist before", "hack error")
}

//RemoveFromTemp removes confirmed validator from validator Timproray list
func (validatorObj *ValidatorStruct) RemoveFromTemp() {
	for i, timpValidator := range TempValidatorlst {
		if validatorObj.ValidatorIP == timpValidator.ValidatorObjec.ValidatorIP {
			TempValidatorlst = append(TempValidatorlst[:i], TempValidatorlst[i+1:]...) // delete validator from temp list
		}
	}
}

//UpdateValidator updates validator on the validators list
func (validatorObj *ValidatorStruct) UpdateValidator() string {
	for index, validatorExistsObj := range ValidatorsLstObj {
		if validatorExistsObj.ValidatorIP == validatorObj.ValidatorIP {
			decryptedValidatorList := GetAllValidatorsDecrypted()
			for _, myValidator := range decryptedValidatorList {
				if validatorObj.ValidatorIP == myValidator.ValidatorIP {
					if validatorObj.ValidatorIP == CurrentValidator.ValidatorIP {
						validatorObj.ECCPublicKey = myValidator.ECCPublicKey
						validatorObj.ECCPrivateKey = myValidator.ECCPrivateKey
						validatorObj.EncECCPriv = myValidator.EncECCPriv
					} else {
						validatorObj.ECCPublicKey = myValidator.ECCPublicKey
					}
				}
			}
			ValidatorsLstObj[index] = *validatorObj
			validatorObj.ECCPublicKey = *new(ecdsa.PublicKey)
			validatorObj.ECCPrivateKey = new(ecdsa.PrivateKey)
			CreateValidator(validatorObj)
			return ""
		}
	}
	return errorpk.AddError("Update Validator validator package", "Can't find the validator object "+validatorObj.ValidatorIP, "hack error")
}

//DeleteValidator function to delete validator from the validators list
func (validatorObj *ValidatorStruct) DeleteValidator() string {
	for index, validatorExistsObj := range ValidatorsLstObj {
		if validatorExistsObj.ValidatorIP == validatorObj.ValidatorIP {
			ValidatorsLstObj = append(ValidatorsLstObj[:index], ValidatorsLstObj[index+1:]...)
			return ""
		}
	}
	errorpk.AddError("Delete Validator validator package", "Can't find the validator object "+validatorObj.ValidatorIP, "hack error")
	return "Can't find the validator object" + validatorObj.ValidatorIP

}

//validationAdd validation  to add validator
func (validatorObj *ValidatorStruct) validationAdd() bool {
	existAdd := true
	for _, validatorExistsObj := range ValidatorsLstObj {
		if validatorExistsObj.ValidatorIP == validatorObj.ValidatorIP || validatorExistsObj.ValidatorSoketIP == validatorObj.ValidatorSoketIP || validatorExistsObj.EncECCPublic == validatorObj.EncECCPublic {
			errorpk.AddError("validation Add validator package", "The validator object already exists"+validatorExistsObj.ValidatorIP, "hack error")
			existAdd = false
			break
		}
	}
	return existAdd
}

//FindValidatorByValidatorIP find validator by validator ip
func FindValidatorByValidatorIP(validatorip string) ValidatorStruct {
	validatorObj, _ := findValidatorByIP(validatorip)
	return validatorObj
}

//AddValidatorTemporary add th validator to temp list and send confirmation email to admin
func (validator *TempValidator) AddValidatorTemporary() {
	TempValidatorlst = append(TempValidatorlst, *validator)
}

//SendConfMail send confirmation email to admin
func SendConfMail(admn admin.AdminStruct, validator TempValidator) {

	body := "Dear " + `,
Thank you for joining Inovatian’s InoChain, your request has been processed and your wallet has been created successfully.
Your confirmation code is: ` + `
Please follow the following link to activate your wallet:
(If this link is not clickable, please copy and paste into a new browser)  
` +
		globalPkg.GlobalObj.Downloadfileip + "/ConfirmedValidatorAPI?confirmationcode=" + validator.ConfirmationCode +
		`
 This is a no-reply email; for any enquiries please contact info@inovatian.com
If you did not create this wallet, please disregard this email.
Regards,
Inovatian Team`
	fmt.Println("---------*    Confirmation Code     *--------   ", validator.ConfirmationCode)
	sendEmail(body, admn.AdminEmail)
}

//sendEmail send email
func sendEmail(Body string, Email string) {
	from := "noreply@inovatian.com" ///// "inovatian.tech@gmail.com"
	pass := "ino13579$"             /////your passward   ////

	to := Email //Email of User

	msg := "From: " + from + "\n" +
		"To: " + to + "\n" +
		"Subject: Inovatian Validator Verification\n\n" + Body

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

var randomTable = [...]byte{'1', '2', '3', '4', '5', '6', '7', '8', '9', '0'}

//EncodeToString return confirmation code
func EncodeToString(max int) string {
	buffer := make([]byte, max)
	_, err := io.ReadAtLeast(rand.Reader, buffer, max)
	if err != nil {
		errorpk.AddError("account encodeToString", "the string is more than the max", "runtime error")
	}

code:
	for index := 0; index < len(buffer); index++ {
		buffer[index] = randomTable[int(buffer[index])%len(randomTable)]
	}
	for _, validObjec := range TempValidatorlst {
		if validObjec.ConfirmationCode == string(buffer) {
			goto code
		}
	}
	return string(buffer)
}
