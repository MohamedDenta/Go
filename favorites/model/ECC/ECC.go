package ecc

import (
	"crypto/aes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/hmac"
	"crypto/md5"
	"crypto/rand"
	"crypto/sha1"
	"crypto/sha256"
	"errors"
	"fmt"
	"hash"
	"io"
	"log"
	"math/big"
	"os"

	"github.com/cloudflare/redoctober/padding"
	"github.com/cloudflare/redoctober/symcrypt"
)

//Curve is the type of our curve
var Curve = elliptic.P256

//PublicPath is the public key path
var PublicPath = "ECC keys/PUBLIC KEY.txt"

//PrivatePath is the private key path
var PrivatePath = "ECC keys/PRIVATE KEY.pem"

var Message = []byte("This is phase to encrypt Public ckey with...")

// Encrypt secures and authenticates its input using the public key
// using ECDHE with AES-128-CBC-HMAC-SHA1.
func Encrypt(pub *ecdsa.PublicKey, in []byte) (out []byte, err error) {
	ephemeral, err := ecdsa.GenerateKey(Curve(), rand.Reader)
	if err != nil {
		return
	}
	x, _ := pub.Curve.ScalarMult(pub.X, pub.Y, ephemeral.D.Bytes())
	if x == nil {
		return nil, errors.New("Failed to generate encryption key")
	}
	shared := sha256.Sum256(x.Bytes())
	iv, err := symcrypt.MakeRandom(16)
	if err != nil {
		return
	}

	paddedIn := padding.AddPadding(in)
	ct, err := symcrypt.EncryptCBC(paddedIn, iv, shared[:16])
	if err != nil {
		return
	}

	ephPub := elliptic.Marshal(pub.Curve, ephemeral.PublicKey.X, ephemeral.PublicKey.Y)
	out = make([]byte, 1+len(ephPub)+16)
	out[0] = byte(len(ephPub))
	copy(out[1:], ephPub)
	copy(out[1+len(ephPub):], iv)
	out = append(out, ct...)

	h := hmac.New(sha1.New, shared[16:])
	h.Write(iv)
	h.Write(ct)
	out = h.Sum(out)
	return
}

// Decrypt authenticates and recovers the original message from
// its input using the private key and the ephemeral key included in
// the message.
func Decrypt(priv *ecdsa.PrivateKey, in []byte) (out []byte, err error) {
	ephLen := int(in[0])
	ephPub := in[1 : 1+ephLen]
	ct := in[1+ephLen:]
	if len(ct) < (sha1.Size + aes.BlockSize) {
		return nil, errors.New("Invalid ciphertext")
	}

	x, y := elliptic.Unmarshal(Curve(), ephPub)
	ok := Curve().IsOnCurve(x, y) // Rejects the identity point too.
	if x == nil || !ok {
		return nil, errors.New("Invalid public key")
	}

	x, _ = priv.Curve.ScalarMult(x, y, priv.D.Bytes())
	if x == nil {
		return nil, errors.New("Failed to generate encryption key")
	}
	shared := sha256.Sum256(x.Bytes())

	tagStart := len(ct) - sha1.Size
	h := hmac.New(sha1.New, shared[16:])
	h.Write(ct[:tagStart])
	mac := h.Sum(nil)
	if !hmac.Equal(mac, ct[tagStart:]) {
		return nil, errors.New("Invalid MAC")
	}

	paddedOut, err := symcrypt.DecryptCBC(ct[aes.BlockSize:tagStart], ct[:aes.BlockSize], shared[:16])
	if err != nil {
		return
	}
	out, err = padding.RemovePadding(paddedOut)
	return
}

//Verification  verifies the signature in r, s of hash using the public key, pub. Its return value records whether the signature is valid.
func Verification(pub ecdsa.PublicKey, hash []byte, r, s *big.Int) bool {
	verifystatus := ecdsa.Verify(&pub, hash, r, s)
	return verifystatus
}

//Sign signs a hash (which should be the result of hashing a larger message) using the private key.
func Sign(text string, priv *ecdsa.PrivateKey) ([]byte, *big.Int, *big.Int, []byte) {
	var h hash.Hash
	h = md5.New()
	r := big.NewInt(0)
	s := big.NewInt(0)

	io.WriteString(h, text)
	signhash := h.Sum(nil)

	r, s, err := ecdsa.Sign(rand.Reader, priv, signhash)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	signature := r.Bytes()
	signature = append(signature, s.Bytes()...)
	return signature, r, s, signhash
}

//GenerateECCKey generate and return ECC ket paiers
func GenerateECCKey() (privateKey *ecdsa.PrivateKey, publicKey ecdsa.PublicKey) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		log.Fatalf("Failed to generate ECDSA key: %s\n", err)
	}
	publicKey = privateKey.PublicKey

	return
}
