package main

import (
	"fmt"
	"log"
	"math"
	"time"

	"github.com/anakinrm/crypto-exchange/client"
	"github.com/anakinrm/crypto-exchange/server"
)

// resycle the time tick

const (
	maxOrders = 3
)

var (
	tick   = 2 * time.Second
	myAsks = make(map[float64]int64)
	myBids = make(map[float64]int64)
)

func marketOrderPlacer(c *client.Client) {
	ticker := time.NewTicker(5 * time.Second)
	for {
		//Sell
		marketSellOrder := &client.PlaceOrderParams{
			UserID: 666,
			Bid:    false,
			Size:   1000,
		}

		sellOrderResp, err := c.PlaceMarketOrder(marketSellOrder)
		if err != nil {
			log.Println(sellOrderResp)
		}

		//Buy
		marketBuyOrder := &client.PlaceOrderParams{
			UserID: 666,
			Bid:    true,
			Size:   1000,
		}

		orderResp, err := c.PlaceMarketOrder(marketBuyOrder)
		if err != nil {
			log.Println(orderResp)
		}

		<-ticker.C
	}
}

func makeMarketSimple(c *client.Client) {
	ticker := time.NewTicker(tick)

	for {

		orders, err := c.GetOrders(1)
		if err != nil {
			log.Println(err)
		}
		fmt.Println("=-----------------------------------")
		fmt.Printf("%+v\n", orders)
		fmt.Println("=-----------------------------------")

		bestAsk, err := c.GetBestAskPrice()
		if err != nil {
			log.Println(err)
		}

		bestBid, err := c.GetBestBidPrice()
		if err != nil {
			log.Println(err)
		}

		spread := math.Abs(bestBid - bestAsk)
		fmt.Println("exchange spread", spread)

		if len(myBids) < 3 {
			//place the bid
			bidLimit := &client.PlaceOrderParams{
				UserID: 2,
				Bid:    true,
				Price:  bestBid + 100,
				Size:   1000,
			}

			bidOrderResp, err := c.PlaceLimitOrder(bidLimit)
			if err != nil {
				log.Println(bidOrderResp)
			}

			myBids[bidLimit.Price] = bidOrderResp.OrderID

		}

		if len(myAsks) < 3 {
			//place the ask
			askLimit := &client.PlaceOrderParams{
				UserID: 2,
				Bid:    false,
				Price:  bestAsk - 100,
				Size:   1000,
			}

			askOrderResp, err := c.PlaceLimitOrder(askLimit)
			if err != nil {
				log.Println(askOrderResp)
			}

			myAsks[askLimit.Price] = askOrderResp.OrderID
		}

		fmt.Println("best ask price", bestAsk)
		fmt.Println("best bid price", bestBid)

		//its a channel which will better than sleep
		<-ticker.C
	}
}

func seedMarket(c *client.Client) error {
	ask := &client.PlaceOrderParams{
		UserID: 1,
		Bid:    false,
		Price:  10_000,
		Size:   5_000_000,
	}

	bid := &client.PlaceOrderParams{
		UserID: 1,
		Bid:    true,
		Price:  9_000,
		Size:   5_000_000,
	}
	_, err := c.PlaceLimitOrder(ask)
	if err != nil {
		return err
	}

	_, err = c.PlaceLimitOrder(bid)
	if err != nil {
		return err
	}
	return nil
}

func main() {
	go server.StartServer()
	time.Sleep(1 * time.Second)

	c := client.NewClient()

	if err := seedMarket(c); err != nil {
		panic(err)
	}

	go makeMarketSimple(c)

	time.Sleep(5 * time.Second)
	marketOrderPlacer(c)

	// limitOrderParams := &client.PlaceOrderParams{
	// 	UserID: 1,
	// 	Bid:    false,
	// 	Price:  10_000,
	// 	Size:   5_000_000,
	// }

	// _, err := c.PlaceLimitOrder(limitOrderParams)
	// if err != nil {
	// 	panic(err)
	// }

	// otherLimitOrderParams := &client.PlaceOrderParams{
	// 	UserID: 666,
	// 	Bid:    false,
	// 	Price:  9_000,
	// 	Size:   500_000,
	// }

	// _, err = c.PlaceLimitOrder(otherLimitOrderParams)
	// if err != nil {
	// 	panic(err)
	// }

	// buyLimitOrder := &client.PlaceOrderParams{
	// 	UserID: 666,
	// 	Bid:    true,
	// 	Price:  11_000,
	// 	Size:   500_000,
	// }

	// _, err = c.PlaceLimitOrder(buyLimitOrder)
	// if err != nil {
	// 	panic(err)
	// }

	// //fmt.Println("placed limit order from the client =>", resp.OrderID)

	// limitMarketParams := &client.PlaceOrderParams{
	// 	UserID: 2,
	// 	Bid:    true,
	// 	Size:   1_000_000,
	// }

	// _, err = c.PlaceMarketOrder(limitMarketParams)
	// if err != nil {
	// 	panic(err)
	// }

	// bestBidPrice, err := c.GetBestBidPrice()
	// if err != nil {
	// 	panic(err)
	// }

	// bestAskPrice, err := c.GetBestAskPrice()
	// if err != nil {
	// 	panic(err)
	// }

	// fmt.Println(bestBidPrice, bestAskPrice)

	// //fmt.Println("placed market order from the client =>", resp.OrderID)

	// time.Sleep(1 * time.Second)

	select {}
}
