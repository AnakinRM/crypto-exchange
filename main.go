package main

import (
	"time"

	"github.com/anakinrm/crypto-exchange/client"
	"github.com/anakinrm/crypto-exchange/server"
)

func main() {
	go server.StartServer()
	time.Sleep(1 * time.Second)

	c := client.NewClient()

	for {
		limitOrderParams := &client.PlaceOrderParams{
			UserID: 1,
			Bid:    false,
			Price:  10_000,
			Size:   700_000,
		}

		_, err := c.PlaceLimitOrder(limitOrderParams)
		if err != nil {
			panic(err)
		}

		otherLimitOrderParams := &client.PlaceOrderParams{
			UserID: 666,
			Bid:    false,
			Price:  9_000,
			Size:   300_000,
		}

		_, err = c.PlaceLimitOrder(otherLimitOrderParams)
		if err != nil {
			panic(err)
		}

		//fmt.Println("placed limit order from the client =>", resp.OrderID)

		limitMarketParams := &client.PlaceOrderParams{
			UserID: 2,
			Bid:    true,
			Size:   1_000_000,
		}

		_, err = c.PlaceMarketOrder(limitMarketParams)
		if err != nil {
			panic(err)
		}

		//fmt.Println("placed market order from the client =>", resp.OrderID)

		time.Sleep(1 * time.Second)
	}

	select {}
}
