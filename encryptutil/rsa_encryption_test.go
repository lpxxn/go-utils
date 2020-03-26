package encryptutil

import (
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"testing"
)

func TestOAEP(t *testing.T) {
	// Generate Alice RSA keys Of 2048 Buts
	alicePrivateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		fmt.Println(err)
		return
	}
	// Extract Public Key from RSA Private Key
	alicePublicKey := &alicePrivateKey.PublicKey
	secretMessage := "Hello wrod"
	encryptedMessage, err := EncryptOAEP(secretMessage, alicePublicKey)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println("Cipher Text  ", encryptedMessage)
	decStr, err := DecryptOAEP(encryptedMessage, alicePrivateKey)
	if err != nil {
		t.Fatal(err)
	}
	if decStr != secretMessage {
		t.Error("not equal")
	}
}

func TestDecryptOAEP(t *testing.T) {
	alicePrivateKey, alicePublicKey, err := ParseKey()
	if err != nil {
		t.Fatal(err)
	}
	secretMessage := "Hello wrod"
	encryptedMessage, err := EncryptOAEP(secretMessage, alicePublicKey)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println("Cipher Text  ", encryptedMessage)
	decStr, err := DecryptOAEP(encryptedMessage, alicePrivateKey)
	if err != nil {
		t.Fatal(err)
	}
	if decStr != secretMessage {
		t.Error("not equal")
	}

}

func TestSignPKCS1v15(t *testing.T) {
	// Generate Alice RSA keys Of 2048 Buts
	alicePrivateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		fmt.Println(err)
		return
	}
	// Extract Public Key from RSA Private Key
	alicePublicKey := &alicePrivateKey.PublicKey
	secretMessage := "Hello wrod"
	fmt.Println("Original Text  ", secretMessage)
	signature, err := SignPKCS1v15(secretMessage, alicePrivateKey)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("Singature :  ", signature)
	verif, err := VerifyPKCS1v15(signature, secretMessage, alicePublicKey)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(verif)
}

func TestRsaDecrypt(t *testing.T) {
	// Generate Alice RSA keys Of 2048 Buts
	alicePrivateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		fmt.Println(err)
		return
	}
	// Extract Public Key from RSA Private Key
	alicePublicKey := &alicePrivateKey.PublicKey
	secretMessage := "Hello wrod"

	ciphertext, err := RsaEncrypt(secretMessage, alicePublicKey)
	if err != nil {
		t.Error(err)
		return
	}
	revBytes, err := RsaDecrypt(ciphertext, alicePrivateKey)
	if err != nil {
		t.Error(err)
		return
	}
	if revBytes != secretMessage {
		t.Error("not equal")
	}
}

func TestRsaEncrypt2(t *testing.T) {
	alicePrivateKey, alicePublicKey, err := ParseKey()
	if err != nil {
		t.Fatal(err)
	}
	secretMessage := "Hello wrod"

	cipherText, err := RsaEncrypt(secretMessage, alicePublicKey)
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println(cipherText)
	revBytes, err := RsaDecrypt(cipherText, alicePrivateKey)
	if err != nil {
		t.Error(err)
		return
	}
	if revBytes != secretMessage {
		t.Error("not equal")
	}
}

func TestRsaDecrypt2(t *testing.T) {
	encryptedMessage := "f3FwK+fyrYxBJIamOW3V1/eX+RYiU5Ex0SlqqnfAWh+A2TlI2fssCiPsDXl/p4zxdhTb7krPmD9D6J9M9BYsv5UonW8Tho+QxY5ySwQY1/jhoOpkoFjZTBBO2KuI3Ut+xLur4tfkaojhA7ikfFMT3uWneuDKpyIOMvF95+x+fFHxbx6Rn3qbP/VXTWhTc3laQkCncwAD9/J0wR4SFZZOEUjoRv4IBtdblxKziRSJ0mnjONcoSVwaALAsxT1UiyE36/q8vRvqQs+Jd3q/HDs2HpfRxRNFwtQWDBv39P4/nn3d7oUCwsyb7FhQ8as8yAFTmuwUTmMktJO5jIOlh2YVSg=="
	alicePrivateKey, _, err := ParseKey()
	if err != nil {
		t.Fatal(err)
	}

	decStr, err := RsaDecrypt(string(encryptedMessage), alicePrivateKey)
	if err != nil {
		t.Fatal(err)
	}
	secretMessage := "Hello wrod"

	if string(decStr) != secretMessage {
		t.Error("not equal")
	}
}

func ParseKey() (*rsa.PrivateKey, *rsa.PublicKey, error) {
	privateKey, err := ParsePrivateKey([]byte(privateKeyStr))
	if err != nil {
		fmt.Println(err)
		return nil, nil, err
	}

	publicKey, err := ParsePublicKey([]byte(publicKeyStr))
	if err != nil {
		fmt.Println(err)
		return nil, nil, err
	}
	return privateKey, publicKey, nil
}

// 注意不能有 \n
const privateKeyStr = `-----BEGIN RSA PRIVATE KEY-----
MIIEpQIBAAKCAQEAvAwmaK8akRfu+JBukLQyZBiWtF+g5Aekta83ADWe3M8j0N8M
8hcUL9d7UBp5xyZoS85JV7mcAY4Lb3nOw5keiwAP6/UBfZRAZefMVCNsb365+O3c
NrxT8zKPRdvSgScHmK4rcfZPbEBFUte1GSDH/XcTpm33c0B99PNqwlF0VU1zKAwf
G1nRIOuzmVC+B1WS562eGCuVagC4AeuNE1zWdEbQa+VDxHQWxMxF12BvoydXf9gR
gSsPuUxJl7U9k1vtejKXmKqavSXLJJThLtIl+3C6Sy+1htDm6shxcu3x6D6cOTRl
jR5jBHyNd3Q9W6ba3Q5OhIZ1afgOp/wvr0MBNQIDAQABAoIBAEqhn9TIOgj/sK4h
1F/FxIIJaDZqBZa6mdopkfCZV1VXOGW7QI4MLszV/nDKMS6ixZ3gXydb2NidIVi6
xR7H9GFCQw9oi5Dld7F6D5QNAwo1B2YOMOngUIkitc4J8j+j19X2ufNeyCK0V08L
oSo54mVsDvZsilrJa7P9r48zeLIpffTxJbesLmngdX0xyYGoQJUxo9rwFHEF9km2
V0iygTfNM0mvUG2/tIjf8L9dwcIvCIwGDGhhcXYKk3ZY0ZPBpbnrXtPp/ZGTsV7O
1EhLUQext+uG+iAUQrSy9YLymQpRcWvRnSEIqu8+Q+UaFNtxKlSqe5YkqbpcdTsA
gsgwkUECgYEA8K+9KUF3jarQKJ9JmuYy2Ncfpp9d/2ZG4HfwF5EtE4tmQ2REZv9r
PaZTRjE0DUd6GQnN06APPVDjdlKXLsfej/5TpF8WIYgdpwdwCqKMcN/EnHCq/qJD
x7lpRr9xiS4+50psa6peKj/YjjtmbqntazlxQ6dt3Z/wZv77x7fr/vkCgYEAyAMJ
HMeJ5dwf/3RlNCR8A7Dne4f2AT8DLKOqJrhSWSkGjSisUVoWCLsPX8G/KYdjhHgj
8hvKVCKlepTHXfEJQAAaJ4jIhmZyoEzHbFt/WdwBKJHLTqvgBrcc9weT/dRVD/ZU
dO28qLxceJ3tOsSkI7Jyno1UwDEd9jUjyB6E1x0CgYEA18DxlJX3EatZRdDkLlLE
qdTNrpOVs2h/iKB7POUKv0ZquWacWqgD/hOm+nkI7A5yyRccxuPoRVLJVDvdAjZw
sCuP1vzV3eEik6P7L81ej6BHouTso63ZjKQMVzsuD4bBJJx2bF0gZEcvXPCqdfEl
vsSTX84qkkzZN7rDANlCWCECgYEAjd4BV6V8/Upuc86Gfj6mrCONfYSJjIa6ZK5N
4Rr6Zf2AhR1lZGqmmFi+ZehSBE3g27QvountUFIm19SxuMNgEUJBSutteE8wXN04
0nXv1bgEJleLQmkNBRZa+Ckq4m76StEpRKrgFztLx84U14tk9WD8hdOvWoc8Pkeg
8rAa/00CgYEAhL69SjgRDUT6Son8nciJ31IbaNQWtYD0fSY7uIz855xCQWBrp8zl
slI7MftqbtISmdTE4PC6GWyPSBk79s5VHxoY7HlWQaPl6mGSz3uzqmRXedbx+szX
0/C52zG+3kvwquWqAIQWXRcKyR0nFqPaHteUzhQyWPG+qkOeyAZ7YEk=
-----END RSA PRIVATE KEY-----`
const publicKeyStr = `-----BEGIN RSA PUBLIC KEY-----
MIIBCgKCAQEAvAwmaK8akRfu+JBukLQyZBiWtF+g5Aekta83ADWe3M8j0N8M8hcU
L9d7UBp5xyZoS85JV7mcAY4Lb3nOw5keiwAP6/UBfZRAZefMVCNsb365+O3cNrxT
8zKPRdvSgScHmK4rcfZPbEBFUte1GSDH/XcTpm33c0B99PNqwlF0VU1zKAwfG1nR
IOuzmVC+B1WS562eGCuVagC4AeuNE1zWdEbQa+VDxHQWxMxF12BvoydXf9gRgSsP
uUxJl7U9k1vtejKXmKqavSXLJJThLtIl+3C6Sy+1htDm6shxcu3x6D6cOTRljR5j
BHyNd3Q9W6ba3Q5OhIZ1afgOp/wvr0MBNQIDAQAB
-----END RSA PUBLIC KEY-----`
