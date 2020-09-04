package epik

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/EpiK-Protocol/epik-explorer-backend/storage"
	"github.com/EpiK-Protocol/epik-explorer-backend/utils"
	"github.com/dgraph-io/badger/v2"
	jwt "github.com/dgrijalva/jwt-go"
)

const (
	//TokenSecretKey token 密钥
	tokenSecretKey = "epik.explorer"
	kvAccessToken  = "%s:%d"
)

//CreateToken 生成token
func CreateToken(userID int64, platform string, expired time.Duration) string {
	token := jwt.New(jwt.SigningMethodHS256)
	claims := make(jwt.MapClaims)
	claims["exp"] = time.Now().Add(expired).Unix()
	claims["iat"] = time.Now().Unix()
	claims["uid"] = userID
	claims["plt"] = platform
	token.Claims = claims

	tokenString, err := token.SignedString([]byte(tokenSecretKey))
	if err != nil {
		panic(err)
	}
	storage.TokenKV.Update(func(txn *badger.Txn) error {
		e := badger.NewEntry([]byte(fmt.Sprintf(kvAccessToken, platform, userID)), []byte(tokenString)).WithTTL(expired)
		txn.SetEntry(e)
		return nil
	})
	return tokenString
}

// ParseToken 解析token
func ParseToken(tokenString string) (userID int64, expired int64, platform string, valid bool) {

	token, err := jwt.Parse(tokenString, func(*jwt.Token) (interface{}, error) {
		return []byte(tokenSecretKey), nil
	})
	if err != nil || token == nil {
		valid = false
		return
	}
	claims := token.Claims.(jwt.MapClaims)
	userID = utils.ParseInt64(claims["uid"])
	expired = utils.ParseInt64(claims["exp"])
	platform = utils.ParseString(claims["plt"])
	originToken := ""
	storage.TokenKV.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(fmt.Sprintf(kvAccessToken, platform, userID)))
		if err != nil {
			return err
		}
		item.Value(func(val []byte) error {
			originToken = string(val)
			return nil
		})
		return nil
	})
	if originToken != tokenString {
		valid = false
		return
	}
	valid = true
	return
}

const (
	CHARNUMBER    = "0123456789"
	CHARCHARACTER = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
)

//RandomString 随机字符
func RandomString(chars string, length int) string {
	rand.Seed(time.Now().Unix())
	str := []byte("")
	for i := 0; i < length; i++ {
		str = append(str, chars[rand.Intn(len(chars))])
	}
	return string(str)
}
