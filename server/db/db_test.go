package db

import (
	"reflect"
	"testing"
)

func assert(t *testing.T, a, b any) {
	if !reflect.DeepEqual(a, b) {
		t.Errorf("%+v != %+v", a, b)
	}
}

func TestDataBase(t *testing.T) {

	// InitializeMongo("mongodb://localhost:27017")
	// user := User{
	// 	ID:       1,
	// 	UserName: "Anakin",
	// 	Email:    "anakinrm@gmail.com",
	// 	Phone:    123456789,
	// }
	// err := InsertUser(user)
	// if err != nil {
	// 	panic(err)
	// }
	// getUser, err := GetUserByID(1)

	// fmt.Println(getUser)

}
