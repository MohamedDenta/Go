package cryptogrpghy

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"fmt"

	"github.com/Luzifer/go-openssl"
)

// "crypto/aes"
// "crypto/cipher"
// "encoding/base64"
// "fmt"
//angular 7
// AESDecrypt decrypts cipher text string into plain text string
func AESDecrypt(encrypted string, CIPHER_KEY string) (string, error) {
	o := openssl.New()

	dec, err := o.DecryptBytes(CIPHER_KEY, []byte(encrypted), openssl.DigestMD5Sum)
	if err != nil {
		fmt.Printf("An error occurred: %s\n", err)
		return "", err
	}

	fmt.Printf("Decrypted text: %s\n", string(dec))
	return fmt.Sprintf("%s", dec), nil
}

//AES 256 stable for flutter

//256
func FAESDecrypt(encrypted string, CIPHER_KEY string) (string, error) {
	key := []byte(CIPHER_KEY)
	cipherText, _ := base64.StdEncoding.DecodeString(encrypted) ////hex.DecodeString(encrypted) //

	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}

	if len(cipherText) < aes.BlockSize {
		panic("cipherText too short")
	}
	// iv := cipherText[:aes.BlockSize]
	// iv := []byte("9c7b7c3826a92fb3")
	iv := []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	cipherText = cipherText[:]
	if len(cipherText)%aes.BlockSize != 0 {
		panic("cipherText is not a multiple of the block size")
	}

	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(cipherText, cipherText)
	dec := PKCS5Trimming(cipherText)
	// dec, _ := pkcs7.Pad(cipherText, len(cipherText))
	return fmt.Sprintf("%s", dec), nil
}

//pad for flutter encrytion pkcs5
func PKCS5Trimming(encrypt []byte) []byte {
	padding := encrypt[len(encrypt)-1]
	return encrypt[:len(encrypt)-int(padding)]
}
