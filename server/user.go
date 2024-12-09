package server

import (
	"github.com/anakinrm/crypto-exchange/server/token"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID           int64
	UserName     string
	hashedPassWd string
	Email        string
	Phone        int64
	Wallet       map[token.Market]token.Token
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func NewUser(id int64, userName, passWd, email string, phone int64) *User {

	wallet, err := token.GenerateWallet()
	if err != nil {
		panic(err)
	}

	hashedPassWd, err := HashPassword(passWd)
	if err != nil {
		panic(err)
	}

	return &User{
		ID:           id,
		UserName:     userName,
		hashedPassWd: hashedPassWd,
		Email:        email,
		Phone:        phone,
		Wallet:       wallet,
	}
}
