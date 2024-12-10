package server

import (
	"github.com/anakinrm/crypto-exchange/server/db"
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

func GetUserbyID(id int64) (*User, error) {
	user := User{
		ID: id,
	}
	err := user.GetUserFromDataBase()
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (user *User) StoreUserInDataBase() error {
	userDB := db.User{
		ID:       user.ID,
		UserName: user.UserName,
		PassWD:   user.hashedPassWd,
		Email:    user.Email,
		Phone:    user.Phone,
	}
	err := userDB.InsertUser()
	if err != nil {
		return err
	}

	for _, v := range user.Wallet {
		err = v.StoreTokenToDataBase(user.ID)
		if err != nil {
			return err
		}
	}

	return nil
}

func (user *User) GetUserFromDataBase() error {
	getUser := db.User{
		ID: user.ID,
	}
	err := getUser.GetUserByID()
	if err != nil {
		return err
	}
	user.hashedPassWd = getUser.PassWD
	user.Email = getUser.Email
	user.Phone = getUser.Phone
	user.UserName = getUser.UserName
	user.Wallet, err = token.GetWalletFromDataBase(user.ID)
	if err != nil {
		return err
	}

	return nil
}
