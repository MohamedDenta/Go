package cryptogrpghy

import (
	"bytes"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/asn1"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"math/big"
	"strings"
)

type pkcs1PrivateKey struct {
	Version int
	N       *big.Int
	E       int
	D       *big.Int
	P       *big.Int
	Q       *big.Int
	// We ignore these values, if present, because rsa will calculate them.
	Dp   *big.Int `asn1:"optional"`
	Dq   *big.Int `asn1:"optional"`
	Qinv *big.Int `asn1:"optional"`

	AdditionalPrimes []pkcs1AdditionalRSAPrime `asn1:"optional,omitempty"`
}

type pkcs1AdditionalRSAPrime struct {
	Prime *big.Int

	// We ignore these values because rsa will calculate them.
	Exp   *big.Int
	Coeff *big.Int
}
type pkcsType int64

const (
	rsaAlgorithmSign = crypto.SHA256

	PKCS1 pkcsType = iota
	PKCS8
)

type XRsa struct {
	keyLen         int
	privateKeyType pkcsType
	publicKey      *rsa.PublicKey
	privateKey     *rsa.PrivateKey
}
type DigitalwalletTransaction struct {
	Sender    string
	Receiver  string
	TokenID   string
	Amount    float32
	Time      string
	Signature string
}

// func main() {
// 	dat, _ := ioutil.ReadFile("vnode5/public.pem")
// 	publicKeyData := string(dat)
// 	encryptedkeydata := "d5fce9709e9cae77db9ecf540758e7c3d6787acd"
// 	hashedkeyenc, _ := PublicEncrypt(publicKeyData, encryptedkeydata)
// 	fmt.Println("hashedkeyenc :", hashedkeyenc, " the length is ", len(hashedkeyenc))
// 	dat2, _ := ioutil.ReadFile("vnode5/private.pem")
// 	privKeyData := string(dat2)
// 	d := "PKADjI3qSMYN69l1/xsRm5NJIKhNv3WZXyF59eyzDPcVQUc9pK/wPSJ7BHjejMiD9+2lAHPHIxWEL2JGpcrFCr24T1WjbpdedaqLRrjWKW7cqF10KMYy35Wc2aMjcUrN8XMXFGMbviGWAtz6qpBuoJTsDURBEdKLFsxGoSTLTmg=" //"u2fQKh9t808lWGEC14313BOfexOFopMNOyHDGS6KL92BVvsytF0gK4JD1Pf08YXI1BQ/EhJBksWYzdS+FUEorRld9wBcsOrXo48FwmPHpYXzCieECiJuxQntjaIazZUHMDZv33Tmrik+xynLVJYofWNBtIDMqg/CzyEH0nlXLhM=" //"kmhokCiegUjDB7hoHotdfZrrwnU1Y/aia6QxDyJBcfUUgwgLN07oSvZtVvGk9CwTciRfYZ3WKFe23Fvr9XtAw83AA8byVkGc6EGHhgRBB9D9SEFtWinjQeDUkfsr0A5lQy1dlZJhjz5IWAI2+X0z4btKPwXudAueSH5JgllK+k0=" //"mWwsPzqPhPTr/l/JCVa7SsGL2Ws0e/em0iWQUhsifT45jxnnkA40f2wAjQlvs/7z2NXR/+wKQJZCE0KdGSHIMsI09Bp5uN/rsanUYGrE85WSD7YpG+kuubOyCafzz7HcdFyR+on9Mz+5ABi9BvcVmB1nFPPjczFO6VlgS5RzM5E="
// 	originalData, _ := PrivateDecrypt(publicKeyData, privKeyData, d, PKCS1)
// 	fmt.Println("originalData :", originalData)
// }
func PublicEncrypt(publicKey, data string) (string, error) {
	block, _ := pem.Decode([]byte(publicKey))
	if block == nil {
		return "", errors.New("public key error")
	}
	pubInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return "", err
	}
	pub := pubInterface.(*rsa.PublicKey)

	partLen := pub.N.BitLen()/8 - 11
	p := string(partLen)
	chunks := strings.Split(data, p)

	buffer := bytes.NewBufferString("")
	for _, chunk := range chunks {
		bts, err := rsa.EncryptPKCS1v15(rand.Reader, pub, []byte(chunk))
		if err != nil {
			return "", err
		}
		buffer.Write(bts)
	}

	return base64.StdEncoding.EncodeToString(buffer.Bytes()), nil
}

//decryption function
func Decrypt(publicKey, privateKey, encrypted string) (originalData string, err error) {
	return PrivateDecrypt(publicKey, privateKey, encrypted, PKCS1)
}
func PrivateDecrypt(publicKey, privateKey, encrypted string, privateKeyType pkcsType) (originalData string, err error) {
	block, _ := pem.Decode([]byte(publicKey))
	if block == nil {
		err = errors.New("public key error")
		return
	}
	pubInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return "", err
	}
	pub := pubInterface.(*rsa.PublicKey)

	block, _ = pem.Decode([]byte(privateKey))
	if block == nil {
		err = errors.New("private key error")
		return
	}
	var pri *rsa.PrivateKey
	pri, err = ParsePKCS1PrivateKey(block.Bytes)
	partLen := pub.N.BitLen() / 8
	raw, err := base64.StdEncoding.DecodeString(encrypted)
	// chunks := split([]byte(raw), partLen)
	p := string(partLen)
	chunks := strings.Split(string([]byte(raw)), p)
	buffer := bytes.NewBufferString("")
	var decrypted []byte
	for _, chunk := range chunks {
		decrypted, err = rsa.DecryptPKCS1v15(rand.Reader, pri, []byte(chunk))
		if err != nil {
			return
		}
		buffer.Write(decrypted)
	}

	originalData = buffer.String()
	return
}

func parsePrivateKey(blockBytes []byte, privateKeyType pkcsType) (private *rsa.PrivateKey, err error) {
	switch privateKeyType {
	case PKCS1:
		private, err = ParsePKCS1PrivateKey(blockBytes)
	case PKCS8:
		var privateKey interface{}
		privateKey, err = x509.ParsePKCS8PrivateKey(blockBytes)
		if err != nil {
			return nil, err
		}
		var ok bool
		private, ok = privateKey.(*rsa.PrivateKey)
		if !ok {
			err = errors.New("private key not supported")
		}
	default:
		err = errors.New("unsupported private key type")
	}
	return
}
func ParsePKCS1PrivateKey(der []byte) (*rsa.PrivateKey, error) {
	var priv pkcs1PrivateKey
	rest, err := asn1.Unmarshal(der, &priv)
	if len(rest) > 0 {
		return nil, asn1.SyntaxError{Msg: "trailing data"}
	}
	if err != nil {
		return nil, err
	}

	if priv.Version > 1 {
		return nil, errors.New("x509: unsupported private key version")
	}

	if priv.N.Sign() <= 0 || priv.D.Sign() <= 0 || priv.P.Sign() <= 0 || priv.Q.Sign() <= 0 {
		return nil, errors.New("x509: private key contains zero or negative value")
	}

	key := new(rsa.PrivateKey)
	key.PublicKey = rsa.PublicKey{
		E: priv.E,
		N: priv.N,
	}

	key.D = priv.D
	key.Primes = make([]*big.Int, 2+len(priv.AdditionalPrimes))
	key.Primes[0] = priv.P
	key.Primes[1] = priv.Q
	for i, a := range priv.AdditionalPrimes {
		if a.Prime.Sign() <= 0 {
			return nil, errors.New("x509: private key contains zero or negative prime")
		}
		key.Primes[i+2] = a.Prime
		// We ignore the other two values because rsa will calculate
		// them as needed.
	}

	err = key.Validate()
	if err != nil {
		return nil, err
	}
	key.Precompute()

	return key, nil
}
