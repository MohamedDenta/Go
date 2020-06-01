package ecc

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"math/big"
	"os"

	"github.com/mitchellh/mapstructure"
)

//ToPublicKey converting string to public key in ecdsa format
func ToPublicKey(keyInString string) (publicKey ecdsa.PublicKey) {
	pub, err := x509.ParsePKIXPublicKey([]byte(keyInString))
	if err != nil {
		panic("failed to parse DER encoded public key: " + err.Error())
	}
	mapstructure.Decode(pub, &publicKey)
	return publicKey
}

//ParseIntoPrivateKey convert Private key from string to *ecdsa.PrivateKey format
func ParseIntoPrivateKey(keyInString string) (privateKey *ecdsa.PrivateKey) {
	block1, _ := pem.Decode([]byte(keyInString))
	if block1 == nil {
		panic("failed to parse PEM block containing the PRIVATE KEY")
	}

	privateKey, err := x509.ParseECPrivateKey([]byte(block1.Bytes))
	check(err)
	return privateKey
}

//GetKeyPairs read then return the public and private keys from it's files
func GetKeyPairs() *ecdsa.PrivateKey {
	//new part :
	// dat, err := ioutil.ReadFile(PublicPath)
	// if err != nil {
	// 	panic(err)
	// }

	// decryptedPublic := cryptogrpghy.KeyDecrypt(string(Message), string(dat))
	// fmt.Println("decryptedPublic", decryptedPublic)
	// var PublickKey ecdsa.PublicKey

	/////////////////////////////////////////////////////////////////////////////////////////////
	//the old part
	// PublickKey := ToPublicKey(decryptedPublic)

	//Reading public key
	// dat, err := ioutil.ReadFile(PublicPath)
	// if err != nil {
	// 	panic(err)
	// }

	// block, _ := pem.Decode([]byte(dat))
	// if block == nil {
	// 	panic("failed to parse PEM block containing the public key")
	// }

	// pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	// if err != nil {
	// 	panic("failed to parse DER encoded public key: " + err.Error())
	// }

	// var PublickKey ecdsa.PublicKey
	// mapstructure.Decode(pub, &PublickKey)

	//Reading  private key
	dat1, err1 := ioutil.ReadFile(PrivatePath)
	if err1 != nil {
		panic(err1)
	}

	block1, _ := pem.Decode([]byte(dat1))
	if block1 == nil {
		panic("failed to parse PEM block containing the PRIVATE KEY")
	}

	priv, err := x509.ParseECPrivateKey(block1.Bytes)
	if err != nil {
		panic("failed to parse DER encoded PRIVATE KEY: " + err.Error())
	}

	return priv
}

//HandleGenerateKeyPaires will call GenerateECCKey() to generate both public and private keys if they donot exist
func HandleGenerateKeyPaires() {
	//check if the folder containse Public and Private exist
	path := "ECC keys"
	if _, err := os.Stat(path); os.IsNotExist(err) {
		os.Mkdir(path, 0700)
	}

	// check for key pairs
	if _, err := os.Stat(PublicPath); err == nil {
		// path exists
	} else if os.IsNotExist(err) {
		// path does *not* exist
		GenerateECCKey()
	}

}

type retrievePublicKey struct {
	Key retrieve `json:"ECCPublicKey"`
}

type retrieve struct {
	CurveParams *elliptic.CurveParams `json:"Curve"`
	MyX         *big.Int              `json:"X"`
	MyY         *big.Int              `json:"Y"`
}

//UnmarshalECCPublicKey extract ECC public key from marshaled objects
func UnmarshalECCPublicKey(object []byte) (pub ecdsa.PublicKey) {
	// fmt.Println("at UnmarshalECCPublicKey() function")
	var public ecdsa.PublicKey
	rt := new(retrievePublicKey)

	errmarsh := json.Unmarshal(object, &rt)
	if errmarsh != nil {
		fmt.Println("err at UnmarshalECCPublicKey()")
		panic(errmarsh)
	}

	public.Curve = rt.Key.CurveParams
	public.X = rt.Key.MyX
	public.Y = rt.Key.MyY
	mapstructure.Decode(public, &pub)

	// fmt.Println("Unmarshalled ECC public key : ", pub)
	return
}

//check if err == nil
func check(err error) {
	if err != nil {
		panic(err)
	}
}
