package cryptoClient

import (
	"reflect"
	"testing"
)

func assert(t *testing.T, a, b any) {
	if !reflect.DeepEqual(a, b) {
		t.Errorf("%+v != %+v", a, b)
	}
}

// func TestFloatToBigInt(t *testing.T) {
// 	floatValue := 1.2345
// 	bigIntValue := FloatToBigInt(floatValue)

// 	assert(t, bigIntValue, big.NewInt(1234500000000000000))
// }

// func TestBigIntToFloat(t *testing.T) {
// 	weiValue := big.NewInt(1234500000000000000)
// 	floatValue := BigIntToFloat(weiValue)

// 	assert(t, floatValue, 1.2345)
// }
