package server

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/anakinrm/crypto-exchange/server/db"
	"github.com/anakinrm/crypto-exchange/server/token"
)

func assert(t *testing.T, a, b any) {
	if !reflect.DeepEqual(a, b) {
		t.Errorf("%+v != %+v", a, b)
	}
}

func TestDataBase(t *testing.T) {

	user := NewUser(1, "Anakin", "123456@ABC", "anakinrm@gmail.com", 123456789)

	db.InitializeMongo("mongodb://localhost:27017")
	user.StoreUserInDataBase()
	fmt.Println("User", user.Wallet[token.MarketETH])

	getUser, err := GetUserbyID(1)
	if err != nil {
		panic(err)
	}

	fmt.Println(getUser)
	fmt.Println("getUser", getUser.Wallet[token.MarketETH])

}
