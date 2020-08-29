package utils

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

//RSAEncrypt 加密
func RSAEncrypt(plain string, publicKey []byte) (cipher string, err error) {

	block, _ := pem.Decode(publicKey)
	if block == nil {
		return "", errors.New("public key error ")
	}
	pubInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return "", err

	}
	pub := pubInterface.(*rsa.PublicKey)
	data, err := rsa.EncryptPKCS1v15(rand.Reader, pub, []byte(plain))
	return string(data), err
}

//RSADecrypt 解密
func RSADecrypt(cipher string, privateKey []byte) (plain string, err error) {
	cipherByte, _ := base64.StdEncoding.DecodeString(cipher)
	block, _ := pem.Decode(privateKey)
	if block == nil {
		return "", errors.New("private key error! ")
	}
	priv, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return "", err
	}
	data, err := rsa.DecryptPKCS1v15(rand.Reader, priv, cipherByte)
	return string(data), err
}

//RSASignature ...
func RSASignature(plain []byte, privateKey []byte) []byte {
	block, _ := pem.Decode(privateKey)
	if block == nil {
		return nil
	}
	priv, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil
	}
	myHash := sha256.New()
	myHash.Write(plain)
	hashText := myHash.Sum(nil)
	signature, err := rsa.SignPKCS1v15(rand.Reader, priv, crypto.SHA256, hashText)
	return signature
}

//RSAVerifySignature ...
func RSAVerifySignature(plain []byte, signature []byte, publicKey []byte) (verified bool) {

	pubInterface, err := x509.ParsePKIXPublicKey(publicKey)
	if err != nil {
		fmt.Println(err)
		return false

	}
	pub := pubInterface.(*rsa.PublicKey)
	myHash := sha256.New()
	myHash.Write(plain)
	hashText := myHash.Sum(nil)
	err = rsa.VerifyPKCS1v15(pub, crypto.SHA256, hashText, signature)
	if err != nil {
		fmt.Println(err)
		return false
	}
	return true
}

var (
// 	privateKey = []byte(`-----BEGIN RSA PRIVATE KEY-----
// MIICXAIBAAKBgQDped0GfIWVTsncpvkpN2reAa6uGOckrO3s4BFdOSGPwfVynytF
// MS0J+ZNm+G4UrKLu0Q3u/FUjqVorp8v+JMLvI9g9m2qd49Mi41+O/MaTf0jBvQsw
// Z4UyYWIT95M0UikTPAFWeQm4pNiE+AWqB7EoexPkPE5IZeap0loefv0+mQIDAQAB
// AoGAaRngs5DOmZ30JQ5NT46Q3wum1NyFAO+P03gymOKlBw/rLAQW+Hjgq4LcJhhj
// hY8JcbROL20L7pfH7asFm4x061HahrF8+qx807U+ADUEf7uVMUlntlH6b4eUZZdY
// PSLXWkRv+oj51Cct2e4PQJZF4p3sdycaxyAFIldPG5a9dZkCQQD9LV7HXMyvDEgG
// /+nNWqu3vnaUcmtoltmPFiaI6qzGS5xdafM13mvTT58z9tJ/2xTg3IFIrmt4tacf
// Z7MvGcwXAkEA7BRC1/qfFxq/AX96iSglWnmSyzgxU69jX37agB5hpsuhoijzdGs1
// NRByJjTKiThz9SnJ1ZnxamLrcJZPSROIzwJAFxAU/DA17RQ/U3Pohm5mChztjGRH
// 6IUlWGV6KSrHhmDI47GNGDEkvWEZbZBkaIU6h6lOlaJd4+cYTEIUDoxZDQJBALbp
// 8+iX3I/wPzIP7Yc7vcVeEOi3/zAR4nLpPK6r24l6mR+ljwwSzMTymx8TJCIxxVad
// LC79+dkuD7HKJGBAbG0CQDfuWWI46RySJY5EyGDNKRDqCl7o335BwskNM+OT/waK
// XhYGh6Yzfjg30kqAlh+uUIkCJvmvWdNnCx2LIXpfzw8=
// -----END RSA PRIVATE KEY-----`)

// 	publicKey = []byte(`-----BEGIN PUBLIC KEY-----
// MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQDped0GfIWVTsncpvkpN2reAa6u
// GOckrO3s4BFdOSGPwfVynytFMS0J+ZNm+G4UrKLu0Q3u/FUjqVorp8v+JMLvI9g9
// m2qd49Mi41+O/MaTf0jBvQswZ4UyYWIT95M0UikTPAFWeQm4pNiE+AWqB7EoexPk
// PE5IZeap0loefv0+mQIDAQAB
// -----END PUBLIC KEY-----`)
)
