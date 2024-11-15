package orderbook

import (
	"fmt"
	"reflect"
	"testing"
)

func assert(t *testing.T, a, b any) {
	if !reflect.DeepEqual(a, b) {
		t.Errorf("%+v != %+v", a, b)
	}
}

func TestLimit(t *testing.T) {
	l := NewLimit(10_000)
	buyOrderA := NewOrder(true, 5)
	buyOrderB := NewOrder(true, 8)
	buyOrderC := NewOrder(true, 10)

	l.AddOrder(buyOrderA)
	l.AddOrder(buyOrderB)
	l.AddOrder(buyOrderC)

	l.DeleteOrder(buyOrderB)

	fmt.Println(l)
}

func TestPlaceLimitOrder(t *testing.T) {
	ob := NewOrderbook()

	sellOrderA := NewOrder(false, 10)
	sellOrderB := NewOrder(false, 10)
	ob.PlaceLimitOrder(10_000, sellOrderA)
	ob.PlaceLimitOrder(10_000, sellOrderB)

	assert(t, len(ob.Orders), 2)
	assert(t, ob.Orders[sellOrderA.ID], sellOrderA)
	assert(t, ob.Orders[sellOrderB.ID], sellOrderB)
	assert(t, len(ob.asks), 1)
}

func TestPlaceMarketOrder(t *testing.T) {
	ob := NewOrderbook()

	sellOrder := NewOrder(false, 20)
	ob.PlaceLimitOrder(10_000, sellOrder)

	buyOrder := NewOrder(true, 10)
	matches := ob.PlaceMarketOrder(buyOrder)

	assert(t, len(matches), 1)
	assert(t, len(ob.asks), 1)
	assert(t, ob.AskTotalVolume(), 10.0)
	assert(t, matches[0].Ask, sellOrder)
	assert(t, matches[0].Bid, buyOrder)
	assert(t, matches[0].SizeFilled, 10.0)
	assert(t, matches[0].Price, 10_000.0)
	assert(t, buyOrder.IsFilled(), true)

	fmt.Printf("%+v", matches)
}

func TestPlaceMarketOrderMultiFill(t *testing.T) {
	ob := NewOrderbook()

	buyOrdersA := NewOrder(true, 5)
	buyOrdersB := NewOrder(true, 8)
	buyOrdersC := NewOrder(true, 10)
	buyOrdersD := NewOrder(true, 1)

	ob.PlaceLimitOrder(5_000, buyOrdersC)
	ob.PlaceLimitOrder(5_000, buyOrdersD)
	ob.PlaceLimitOrder(9_000, buyOrdersB)
	ob.PlaceLimitOrder(10_000, buyOrdersA)
	//ob.PlaceLimitOrder(5_000, buyOrdersD)

	assert(t, ob.BidTotalVolume(), 24.00)

	sellOrder := NewOrder(false, 20)
	matches := ob.PlaceMarketOrder(sellOrder)

	assert(t, ob.BidTotalVolume(), 4.0)
	assert(t, len(matches), 3)
	assert(t, len(ob.bids), 1)

	fmt.Printf("%+v", matches)
}

func TestCancelOrder(t *testing.T) {
	ob := NewOrderbook()
	buyOrder := NewOrder(true, 4)
	ob.PlaceLimitOrder(10000.0, buyOrder)

	assert(t, ob.BidTotalVolume(), 4.0)
	ob.CancelOrder(buyOrder)

	assert(t, ob.BidTotalVolume(), 0.0)
	_, ok := ob.Orders[buyOrder.ID]
	assert(t, ok, false)
}
