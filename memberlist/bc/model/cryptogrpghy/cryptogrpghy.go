package cryptogrpghy

import (
	"bytes"
	"crypto"
	"crypto/rand"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/x509"
	"encoding/asn1"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"log"
	"os"
	rsapk "rsapk"

	"github.com/mr-tron/base58"
	"golang.org/x/crypto/ripemd160"
)

// func RSAENC(publickey string, data []byte) string {
// 	hash := sha1.New()
// 	random := rand.Reader
// 	var pub *rsapk.PublicKey
// 	pub = ParsePEMtoRSApublicKey(publickey)
// 	encryptedData, encryptErr := rsapk.EncryptOAEP(hash, random, pub, data, nil)
// 	if encryptErr != nil {
// 		fmt.Println("Encrypt data error")
// 		//panic(encryptErr)
// 		return ""
// 	}
// 	ncodedData := base64.StdEncoding.EncodeToString(encryptedData)
// 	return ncodedData
// }
// func RSADEC(privatekey string, ciphertext string) []byte {
// 	hash := sha1.New()
// 	random := rand.Reader
// 	var pri *rsapk.PrivateKey
// 	pri = ParsePEMtoRSAprivateKey(privatekey)
// 	sss, _ := base64.StdEncoding.DecodeString(ciphertext)
// 	decryptedData, decryptErr := rsapk.DecryptOAEP(hash, random, pri, sss, nil)
// 	if decryptErr != nil {
// 		fmt.Println("Decrypt data error")
// 		//panic(decryptErr)
// 		return nil
// 	}
// 	// fmt.Println("Key decrypted from RSA :", string(decryptedData))
// 	return decryptedData
// }

func GetPrivatePEMKey(privkey *rsapk.PrivateKey) string {
	privkey_bytes := x509.MarshalPKCS1PrivateKey(privkey)
	privkey_pem := pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: privkey_bytes,
		},
	)
	return string(privkey_pem)

}
func GetPublicPEMKey(pubkey rsapk.PublicKey) string {
	asn1Bytes, err := asn1.Marshal(pubkey)
	CheckError(err)
	pemkey := pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PUBLIC KEY",
			Bytes: asn1Bytes,
		},
	)

	return string(pemkey)

}

func CheckError(err error) {
	if err != nil {
		fmt.Println("Fatal error ", err.Error())
		os.Exit(1)
	}
}

func SignPKCS1v15(plaintext string, privKey rsapk.PrivateKey) string {
	// crypto/rand.Reader is a good source of entropy for blinding the RSA
	// operation.
	// rng := rand.Reader
	// hashed := sha256.Sum256([]byte(plaintext))
	// signature, err := rsapk.SignPKCS1v15(rng, &privKey, crypto.SHA256, hashed[:])
	// if err != nil {
	// 	fmt.Fprintf(os.Stderr, "Error from signing: %s\n", err)
	// 	return "Error from signing"
	// }
	// return base64.StdEncoding.EncodeToString(signature)
	rng := rand.Reader
	hashed := sha1.Sum([]byte(plaintext))
	signature, err := rsapk.SignPKCS1v15(rng, &privKey, crypto.SHA1, hashed[:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error from signing: %s\n", err)
		return "Error from signing"
	}
	return base64.StdEncoding.EncodeToString(signature)
}

//Signature Verification is performed using RSA Public key, and verification is done along with the original message.
// func VerifyPKCS1v15(signature string, plaintext string, pubkey rsapk.PublicKey) bool {
// 	sig, _ := base64.StdEncoding.DecodeString(signature)
// 	hashed := sha256.Sum256([]byte(plaintext))
// 	err := rsapk.VerifyPKCS1v15(&pubkey, crypto.SHA256, hashed[:], sig)
// 	if err != nil {
// 		return false
// 	}
// 	return true
// }
func VerifyPKCS1v15(signature string, plaintext string, pubkey rsapk.PublicKey) bool {
	sig, _ := base64.StdEncoding.DecodeString(signature)
	hashed := sha1.Sum([]byte(plaintext))
	err := rsapk.VerifyPKCS1v15(&pubkey, crypto.SHA1, hashed[:], sig)
	if err != nil {
		return false
	}
	return true
}
func ParsePEMtoRSAprivateKey(pemPrivateKey string) *rsapk.PrivateKey {
	privblock, _ := pem.Decode([]byte(pemPrivateKey))
	privateKey, _ := x509.ParsePKCS1PrivateKey(privblock.Bytes)
	return privateKey
}

func ParsePEMtoRSApublicKey(pemPublicKey string) *rsapk.PublicKey {
	publicblock, _ := pem.Decode([]byte(pemPublicKey))
	publicKey, _ := x509.ParsePKCS1PublicKey(publicblock.Bytes)
	return publicKey
}

//Address for public key
//GetPublicPEMKeybyte get public
func GetPublicPEMKeybyte(pubkey rsapk.PublicKey) []byte {
	asn1Bytes, err := asn1.Marshal(pubkey)
	CheckError(err)
	pemkey := pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PUBLIC KEY",
			Bytes: asn1Bytes,
		},
	)
	return pemkey
}

//PublicKeyHash sha256 for publickey and ripemd160 hash
func PublicKeyHash(pubKey []byte) []byte {
	pubHash := sha256.Sum256(pubKey)

	hasher := ripemd160.New()
	_, err := hasher.Write(pubHash[:])
	if err != nil {
		log.Panic(err)
	}

	publicRipMD := hasher.Sum(nil)

	return publicRipMD
}

const (
	checksumLength = 4
	version        = byte(0x00)
)

//Checksum make double hashing
func Checksum(payload []byte) []byte {
	firstHash := sha256.Sum256(payload)
	secondHash := sha256.Sum256(firstHash[:])

	return secondHash[:checksumLength]
}

//Address return []byte.
func Address(PublicKey []byte) []byte {
	pubHash := PublicKeyHash(PublicKey)

	versionedHash := append([]byte{version}, pubHash...)
	checksum := Checksum(versionedHash)
	fullHash := append(versionedHash, checksum...)
	address := Base58Encode(fullHash)
	return address
}

//Base58Encode to encode version +checksum +pk hash
func Base58Encode(input []byte) []byte {
	encode := base58.Encode(input)

	return []byte(encode)
}

//Base58Decode to decode
func Base58Decode(input []byte) []byte {
	decode, err := base58.Decode(string(input[:]))
	if err != nil {
		log.Panic(err)
	}

	return decode
}

//ValidateAddress validate address
func ValidateAddress(address string) bool {
	pubKeyHash := Base58Decode([]byte(address))
	actualChecksum := pubKeyHash[len(pubKeyHash)-checksumLength:]
	version := pubKeyHash[0]
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-checksumLength]
	targetChecksum := Checksum(append([]byte{version}, pubKeyHash...))

	return bytes.Compare(actualChecksum, targetChecksum) == 0
}
