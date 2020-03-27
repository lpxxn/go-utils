package encryptutil

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
)

func EncryptOAEP(secretMessage string, publicKey *rsa.PublicKey) (string, error) {
	rng := rand.Reader
	cipherText, err := rsa.EncryptOAEP(sha256.New(), rng, publicKey, []byte(secretMessage), nil)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(cipherText), nil
}

func DecryptOAEP(cipherText string, privateKey *rsa.PrivateKey) (string, error) {
	ct, _ := base64.StdEncoding.DecodeString(cipherText)
	rng := rand.Reader
	plainText, err := rsa.DecryptOAEP(sha256.New(), rng, privateKey, ct, nil)
	if err != nil {
		return "", err
	}
	fmt.Printf("plainText: %s\n", string(plainText))

	return string(plainText), nil
}

// 私钥签名
func SignPKCS1v15(plaintext string, privateKey *rsa.PrivateKey) (string, error) {
	rng := rand.Reader
	hashed := sha256.Sum256([]byte(plaintext))
	signature, err := rsa.SignPKCS1v15(rng, privateKey, crypto.SHA256, hashed[:])
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(signature), nil
}

// 公钥验签
func VerifyPKCS1v15(signature string, plaintext string, publicKey *rsa.PublicKey) (bool, error) {
	sig, err := base64.StdEncoding.DecodeString(signature)
	if err != nil {
		return false, err
	}
	hashed := sha256.Sum256([]byte(plaintext))
	err = rsa.VerifyPKCS1v15(publicKey, crypto.SHA256, hashed[:], sig)
	if err != nil {
		if err == rsa.ErrVerification {
			return false, nil
		}
		return false, err
	}
	return true, err
}

// 公钥加密
func RsaEncrypt(origData string, publicKey *rsa.PublicKey) (string, error) {
	cipherText, err := rsa.EncryptPKCS1v15(rand.Reader, publicKey, []byte(origData))
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(cipherText), nil
}

// 私钥解密
func RsaDecrypt(cipherText string, privateKey *rsa.PrivateKey) (string, error) {
	data, err := base64.StdEncoding.DecodeString(cipherText)
	if err != nil {
		return "", err
	}
	revData, err := rsa.DecryptPKCS1v15(rand.Reader, privateKey, data)
	if err != nil {
		return "", err
	}
	return string(revData), nil
}

func ParsePrivateKey(privateKey []byte) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode(privateKey)
	if block == nil {
		return nil, errors.New("empty private key")
	}
	return x509.ParsePKCS1PrivateKey(block.Bytes)
}

func ParsePublicKey(publicKey []byte) (*rsa.PublicKey, error) {
	block, _ := pem.Decode(publicKey)
	if block == nil {
		return nil, errors.New("empty public key")
	}
	return x509.ParsePKCS1PublicKey(block.Bytes)
}
